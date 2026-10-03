package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS tracks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT UNIQUE NOT NULL,
    title TEXT,
    artist TEXT,
    album TEXT,
    duration_ms INTEGER,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_tracks_artist ON tracks(artist);
CREATE INDEX IF NOT EXISTS idx_tracks_album ON tracks(album);

CREATE TABLE IF NOT EXISTS playlists (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS playlist_tracks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    playlist_id INTEGER NOT NULL,
    track_id    INTEGER NOT NULL,
    position    INTEGER NOT NULL,
    added_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE,
    FOREIGN KEY (track_id)    REFERENCES tracks(id)    ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_playlist_tracks_playlist
    ON playlist_tracks(playlist_id, position);
`

func New(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	// DSN dengan pragma foreign_keys=ON supaya CASCADE bekerja.
	// Path dinormalisasi ke forward slash agar DSN valid di Windows.
	dsn := "file:" + filepath.ToSlash(path) +
    "?_pragma=foreign_keys(1)" +
    "&_pragma=journal_mode(WAL)" +
    "&_pragma=busy_timeout(5000)" +
    "&_pragma=synchronous(NORMAL)"

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(0)
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, err
	}
	return &DB{conn: conn}, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) Conn() *sql.DB {
	return d.conn
}