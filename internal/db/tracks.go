package db

import (
	"path/filepath"
	"strings"

	"github.com/ade-nugraha306/sky-project/internal/library"
)

func (d *DB) UpsertTrack(t library.Track) error {
	_, err := d.conn.Exec(`
		INSERT INTO tracks (path, title, artist, album, duration_ms)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			title       = excluded.title,
			artist      = excluded.artist,
			album       = excluded.album,
			duration_ms = excluded.duration_ms
	`, t.Path, t.Title, t.Artist, t.Album, t.DurationMs)
	return err
}

func (d *DB) AllTracks() ([]library.Track, error) {
	rows, err := d.conn.Query(`
		SELECT path, title, artist, album, duration_ms
		FROM tracks
		ORDER BY artist, album, title
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []library.Track
	for rows.Next() {
		var t library.Track
		if err := rows.Scan(&t.Path, &t.Title, &t.Artist, &t.Album, &t.DurationMs); err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}
	return tracks, rows.Err()
}

func (d *DB) CountTracks() (int, error) {
	var n int
	err := d.conn.QueryRow(`SELECT COUNT(*) FROM tracks`).Scan(&n)
	return n, err
}

func (d *DB) DeleteTracksUnderFolder(folder string) (int, error) {
	folder = filepath.ToSlash(filepath.Clean(folder))
	prefix := folder + "/"

	rows, err := d.conn.Query(`SELECT id, path FROM tracks`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		var p string
		if err := rows.Scan(&id, &p); err != nil {
			return 0, err
		}
		normalized := filepath.ToSlash(filepath.Clean(p))
		if normalized == folder || strings.HasPrefix(normalized, prefix) {
			ids = append(ids, id)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	if len(ids) == 0 {
		return 0, nil
	}

	tx, err := d.conn.Begin()
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare(`DELETE FROM tracks WHERE id = ?`)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer stmt.Close()

	for _, id := range ids {
		if _, err := stmt.Exec(id); err != nil {
			tx.Rollback()
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(ids), nil
}