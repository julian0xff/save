package db

import "fmt"

type migration struct {
	version int
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		sql: `
CREATE TABLE IF NOT EXISTS articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    author TEXT DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    word_count INTEGER DEFAULT 0,
    read_time INTEGER DEFAULT 0,
    tags TEXT DEFAULT '',
    saved_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    read_at DATETIME,
    archived INTEGER DEFAULT 0
);

CREATE VIRTUAL TABLE IF NOT EXISTS articles_fts USING fts5(
    title, content, content=articles, content_rowid=id
);

CREATE TRIGGER IF NOT EXISTS articles_ai AFTER INSERT ON articles BEGIN
    INSERT INTO articles_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
END;

CREATE TRIGGER IF NOT EXISTS articles_ad AFTER DELETE ON articles BEGIN
    INSERT INTO articles_fts(articles_fts, rowid, title, content) VALUES('delete', old.id, old.title, old.content);
END;

CREATE TRIGGER IF NOT EXISTS articles_au AFTER UPDATE ON articles BEGIN
    INSERT INTO articles_fts(articles_fts, rowid, title, content) VALUES('delete', old.id, old.title, old.content);
    INSERT INTO articles_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
END;
`,
	},
}

func (d *DB) migrate() error {
	var currentVersion int
	err := d.conn.QueryRow("PRAGMA user_version").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("reading schema version: %w", err)
	}

	for _, m := range migrations {
		if m.version <= currentVersion {
			continue
		}
		if _, err := d.conn.Exec(m.sql); err != nil {
			return fmt.Errorf("migration v%d: %w", m.version, err)
		}
		if _, err := d.conn.Exec(fmt.Sprintf("PRAGMA user_version = %d", m.version)); err != nil {
			return fmt.Errorf("updating schema version: %w", err)
		}
	}

	return nil
}
