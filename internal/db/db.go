package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

type Article struct {
	ID        int
	URL       string
	Title     string
	Author    string
	Content   string
	WordCount int
	ReadTime  int
	Tags      string
	SavedAt   time.Time
	ReadAt    *time.Time
	Archived  bool
}

type SearchResult struct {
	ID      int
	Title   string
	Snippet string
	Tags    string
	SavedAt time.Time
}

type ListOpts struct {
	Unread bool
	Tag    string
	Limit  int
}

func defaultPath() (string, error) {
	if p := os.Getenv("SAVE_DB"); p != "" {
		return p, nil
	}
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting home directory: %w", err)
		}
		dataDir = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataDir, "save", "articles.db"), nil
}

func OpenDefault() (*DB, error) {
	path, err := defaultPath()
	if err != nil {
		return nil, err
	}
	return Open(path)
}

func Open(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating data directory %s: %w", dir, err)
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := conn.Exec(p); err != nil {
			conn.Close()
			return nil, fmt.Errorf("setting pragma: %w", err)
		}
	}

	d := &DB{conn: conn}
	if err := d.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return d, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) SaveArticle(a *Article) (int64, error) {
	res, err := d.conn.Exec(
		`INSERT INTO articles (url, title, author, content, word_count, read_time, tags)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		a.URL, a.Title, a.Author, a.Content, a.WordCount, a.ReadTime, a.Tags,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) GetArticle(id int) (*Article, error) {
	a := &Article{}
	var readAt sql.NullTime
	err := d.conn.QueryRow(
		`SELECT id, url, title, author, content, word_count, read_time, tags, saved_at, read_at, archived
		 FROM articles WHERE id = ?`, id,
	).Scan(&a.ID, &a.URL, &a.Title, &a.Author, &a.Content, &a.WordCount, &a.ReadTime, &a.Tags, &a.SavedAt, &readAt, &a.Archived)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if readAt.Valid {
		a.ReadAt = &readAt.Time
	}
	return a, nil
}

func (d *DB) GetArticleByURL(url string) (*Article, error) {
	a := &Article{}
	err := d.conn.QueryRow(`SELECT id, url, title FROM articles WHERE url = ?`, url).Scan(&a.ID, &a.URL, &a.Title)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (d *DB) ListArticles(opts ListOpts) ([]Article, error) {
	query := `SELECT id, url, title, author, word_count, read_time, tags, saved_at, read_at, archived FROM articles WHERE 1=1`
	var args []any

	if opts.Unread {
		query += ` AND read_at IS NULL`
	}
	if opts.Tag != "" {
		query += ` AND (',' || tags || ',' LIKE '%,' || ? || ',%')`
		args = append(args, opts.Tag)
	}

	query += ` ORDER BY saved_at DESC`

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	query += ` LIMIT ?`
	args = append(args, limit)

	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		var readAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.URL, &a.Title, &a.Author, &a.WordCount, &a.ReadTime, &a.Tags, &a.SavedAt, &readAt, &a.Archived); err != nil {
			return nil, err
		}
		if readAt.Valid {
			a.ReadAt = &readAt.Time
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
}

func (d *DB) SearchArticles(query string) ([]SearchResult, error) {
	rows, err := d.conn.Query(
		`SELECT a.id, a.title, snippet(articles_fts, 1, '»', '«', '...', 32), a.tags, a.saved_at
		 FROM articles_fts
		 JOIN articles a ON a.id = articles_fts.rowid
		 WHERE articles_fts MATCH ?
		 ORDER BY bm25(articles_fts)
		 LIMIT 20`, query,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ID, &r.Title, &r.Snippet, &r.Tags, &r.SavedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (d *DB) DeleteArticle(id int) error {
	res, err := d.conn.Exec(`DELETE FROM articles WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("no article with ID %d", id)
	}
	return nil
}

func (d *DB) MarkRead(id int) error {
	res, err := d.conn.Exec(`UPDATE articles SET read_at = CURRENT_TIMESTAMP WHERE id = ? AND read_at IS NULL`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		// Check if article exists but is already read
		a, err := d.GetArticle(id)
		if err != nil {
			return err
		}
		if a == nil {
			return fmt.Errorf("no article with ID %d", id)
		}
		// Already read — not an error
	}
	return nil
}
