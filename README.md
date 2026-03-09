# save

Save articles for reading later — from the terminal.

A CLI read-later tool. Fetch web articles, extract clean text, store locally in SQLite, search with full-text search, and read rendered in your terminal. No account, no cloud, no subscription.

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

## Install

```bash
# Homebrew
brew install julian0xff/tap/save

# Go
go install github.com/julian0xff/save@latest

# From source
git clone https://github.com/julian0xff/save.git
cd save
make build
make install
```

## Usage

```bash
# Save an article
save https://go.dev/blog/go1.22
save https://example.com/article --tag rust,async

# List saved articles
save list
save list --unread
save list --tag rust

# Read in terminal (auto-marks as read)
save read 1

# Search across all saved articles
save search "database performance"

# Open original URL in browser
save open 1

# Mark as read without reading
save mark 1

# Delete
save delete 1
```

## Example

```
$ save https://sqlite.org/whentouse.html --tag database
Saved: "Appropriate Uses For SQLite" (ID #2, 1969 words, ~8 min read)
Tags: database

$ save list

 #   Title                              Tags     Saved    Words
 2 ○ Appropriate Uses For SQLite        database just now 1969
 1 ✓ Go 1.22 is released!              go       2h ago   422

2 articles (2391 words, ~10 min total)

$ save search "SQLite"
1 result(s):
  #2  Appropriate Uses For SQLite
      SQLite is not directly comparable to client/server SQL...
```

## Storage

Articles are stored locally in SQLite at `~/.local/share/save/articles.db`.

Override with `$SAVE_DB` or `--db <path>`.

## How it works

1. Fetches the URL
2. Extracts clean article text using [Mozilla Readability](https://github.com/go-shiori/go-readability)
3. Converts to Markdown and stores in SQLite with metadata
4. Full-text search powered by SQLite FTS5
5. Renders in terminal using [Glamour](https://github.com/charmbracelet/glamour)

## License

MIT
