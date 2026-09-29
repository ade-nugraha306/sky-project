package db

import (
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