package db

import (
	"path/filepath"
	"strings"

	"github.com/ade-nugraha306/sky-project/internal/library"
)

func (d *DB) UpsertTracks(tracks []library.Track) error {
	if len(tracks) == 0 {
		return nil
	}
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT INTO tracks (path, title, artist, album, duration_ms)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			title       = excluded.title,
			artist      = excluded.artist,
			album       = excluded.album,
			duration_ms = excluded.duration_ms
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, t := range tracks {
		if _, err := stmt.Exec(t.Path, t.Title, t.Artist, t.Album, t.DurationMs); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
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

func (d *DB) AllTracksInFolder(folder string) ([]library.Track, error) {
	if folder == "" {
		return d.AllTracks()
	}
	folder = filepath.ToSlash(filepath.Clean(folder))
	prefix := folder + "/"

	rows, err := d.conn.Query(`
		SELECT path, title, artist, album, duration_ms
		FROM tracks
		ORDER BY artist, album, title
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []library.Track
	for rows.Next() {
		var t library.Track
		if err := rows.Scan(&t.Path, &t.Title, &t.Artist, &t.Album, &t.DurationMs); err != nil {
			return nil, err
		}
		normalized := filepath.ToSlash(filepath.Clean(t.Path))
		if normalized == folder || strings.HasPrefix(normalized, prefix) {
			out = append(out, t)
		}
	}
	return out, rows.Err()
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

func (d *DB) DeleteTracksNotIn(folder string, foundPaths map[string]bool) (int, error) {
	folder = filepath.ToSlash(filepath.Clean(folder))
	prefix := folder + "/"

	// Fase 1: kumpulkan ID yang perlu dihapus (di luar transaction).
	rows, err := d.conn.Query(`SELECT id, path FROM tracks`)
	if err != nil {
		return 0, err
	}

	var ids []int64
	for rows.Next() {
		var id int64
		var p string
		if err := rows.Scan(&id, &p); err != nil {
			rows.Close()
			return 0, err
		}
		normalized := filepath.ToSlash(filepath.Clean(p))

		// Hanya proses track yang ada di bawah folder ini.
		inFolder := normalized == folder || strings.HasPrefix(normalized, prefix)
		if !inFolder {
			continue
		}

		// Kalau tidak ditemukan saat scan, hapus.
		if !foundPaths[normalized] {
			ids = append(ids, id)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()

	if len(ids) == 0 {
		return 0, nil
	}

	// Fase 2: hapus dalam satu transaction.
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
