package db

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"database/sql"

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

func sortClause(mode library.SortMode) string {
	switch mode {
	case library.SortByArtist:
		return "artist COLLATE NOCASE, album COLLATE NOCASE, title COLLATE NOCASE, path"
	case library.SortByAlbum:
		return "album COLLATE NOCASE, artist COLLATE NOCASE, title COLLATE NOCASE, path"
	case library.SortByDateDesc:
		return "added_at DESC, path"
	default: // SortByTitle
		return "title COLLATE NOCASE, artist COLLATE NOCASE, path"
	}
}

func (d *DB) AllTracks(sort library.SortMode) ([]library.Track, error) {
	query := "SELECT path, title, artist, album, duration_ms FROM tracks ORDER BY " + sortClause(sort)
	rows, err := d.conn.Query(query)
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

func (d *DB) AllTracksInFolder(folder string, sort library.SortMode) ([]library.Track, error) {
	if folder == "" {
		return d.AllTracks(sort)
	}
	folder = filepath.ToSlash(filepath.Clean(folder))
	prefix := folder + "/"

	query := "SELECT path, title, artist, album, duration_ms FROM tracks ORDER BY " + sortClause(sort)
	rows, err := d.conn.Query(query)
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

func (d *DB) TrackByPath(path string) (library.Track, error) {
	// Normalisasi input ke forward slash.
	forwardPath := filepath.ToSlash(filepath.Clean(path))
	// Versi native (backslash di Windows, sama dengan forward di Unix).
	// DB mungkin menyimpan dengan salah satu format tergantung versi
	// saat track di-insert.
	nativePath := filepath.FromSlash(forwardPath)

	var t library.Track
	err := d.conn.QueryRow(`
		SELECT path, title, artist, album, duration_ms
		FROM tracks
		WHERE path = ? OR path = ?
		LIMIT 1
	`, forwardPath, nativePath).Scan(&t.Path, &t.Title, &t.Artist, &t.Album, &t.DurationMs)

	if errors.Is(err, sql.ErrNoRows) {
		return library.Track{}, fmt.Errorf("track tidak ada di library: %s", path)
	}
	return t, err
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
