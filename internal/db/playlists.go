package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ade-nugraha306/sky-project/internal/library"
)

// ErrPlaylistNotFound dikembalikan saat operasi membutuhkan playlist
// yang ada tapi ID/nama tidak ditemukan.
var ErrPlaylistNotFound = errors.New("playlist tidak ditemukan")

// ErrPlaylistNameExists dikembalikan saat CreatePlaylist atau
// RenamePlaylist memakai nama yang sudah dipakai playlist lain.
var ErrPlaylistNameExists = errors.New("nama playlist sudah dipakai")

type Playlist struct {
	ID        int64
	Name      string
	TrackCount int
}

// CreatePlaylist membuat playlist baru dengan nama tertentu.
// Nama di-trim whitespace. Kalau nama kosong atau duplikat,
// mengembalikan error.
func (d *DB) CreatePlaylist(name string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("nama playlist tidak boleh kosong")
	}

	res, err := d.conn.Exec(
		`INSERT INTO playlists (name) VALUES (?)`,
		name,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return 0, ErrPlaylistNameExists
		}
		return 0, err
	}
	return res.LastInsertId()
}

// RenamePlaylist mengubah nama playlist berdasarkan ID.
func (d *DB) RenamePlaylist(id int64, newName string) error {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return fmt.Errorf("nama playlist tidak boleh kosong")
	}

	res, err := d.conn.Exec(
		`UPDATE playlists SET name = ? WHERE id = ?`,
		newName, id,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return ErrPlaylistNameExists
		}
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrPlaylistNotFound
	}
	return nil
}

// DeletePlaylist menghapus playlist dan semua track di dalamnya
// (via FK CASCADE).
func (d *DB) DeletePlaylist(id int64) error {
	res, err := d.conn.Exec(`DELETE FROM playlists WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrPlaylistNotFound
	}
	return nil
}

// AllPlaylists mengembalikan semua playlist dengan jumlah track
// masing-masing, diurutkan berdasarkan nama.
func (d *DB) AllPlaylists() ([]Playlist, error) {
	rows, err := d.conn.Query(`
		SELECT p.id, p.name, COUNT(pt.id)
		FROM playlists p
		LEFT JOIN playlist_tracks pt ON pt.playlist_id = p.id
		GROUP BY p.id, p.name
		ORDER BY p.name COLLATE NOCASE
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Playlist
	for rows.Next() {
		var p Playlist
		if err := rows.Scan(&p.ID, &p.Name, &p.TrackCount); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// PlaylistByID mengambil satu playlist berdasarkan ID.
func (d *DB) PlaylistByID(id int64) (Playlist, error) {
	var p Playlist
	err := d.conn.QueryRow(`
		SELECT p.id, p.name, COUNT(pt.id)
		FROM playlists p
		LEFT JOIN playlist_tracks pt ON pt.playlist_id = p.id
		WHERE p.id = ?
		GROUP BY p.id, p.name
	`, id).Scan(&p.ID, &p.Name, &p.TrackCount)
	if errors.Is(err, sql.ErrNoRows) {
		return Playlist{}, ErrPlaylistNotFound
	}
	return p, err
}

// TracksInPlaylist mengembalikan track-track dalam playlist,
// urut sesuai kolom position.
func (d *DB) TracksInPlaylist(playlistID int64) ([]library.Track, error) {
	rows, err := d.conn.Query(`
		SELECT t.path, t.title, t.artist, t.album, t.duration_ms
		FROM playlist_tracks pt
		JOIN tracks t ON t.id = pt.track_id
		WHERE pt.playlist_id = ?
		ORDER BY pt.position
	`, playlistID)
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
		out = append(out, t)
	}
	return out, rows.Err()
}

func (d *DB) TrackInPlaylist(playlistID int64, trackPath string) (bool, error) {
	var n int
	err := d.conn.QueryRow(`
		SELECT COUNT(*) FROM playlist_tracks pt
		JOIN tracks t ON t.id = pt.track_id
		WHERE pt.playlist_id = ? AND t.path = ?
	`, playlistID, trackPath).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// TrackPaths mengembalikan semua path di playlist, urut posisi.
// Dipakai untuk lookup index saat ada duplikat — path saja tidak unik,
// kita perlu tahu instance ke-N dari path yang sama.
func (d *DB) TrackPaths(playlistID int64) ([]string, error) {
	rows, err := d.conn.Query(`
		SELECT t.path FROM playlist_tracks pt
		JOIN tracks t ON t.id = pt.track_id
		WHERE pt.playlist_id = ?
		ORDER BY pt.position
	`, playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, rows.Err()
}

func (d *DB) AddTrackToPlaylist(playlistID int64, trackPath string) error {
	var trackID int64
	err := d.conn.QueryRow(
		`SELECT id FROM tracks WHERE path = ?`, trackPath,
	).Scan(&trackID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("track tidak ada di library")
	}
	if err != nil {
		return err
	}

	// Position = MAX + 1 (atau 0 kalau playlist masih kosong).
	var maxPos sql.NullInt64
	err = d.conn.QueryRow(
		`SELECT MAX(position) FROM playlist_tracks WHERE playlist_id = ?`,
		playlistID,
	).Scan(&maxPos)
	if err != nil {
		return err
	}
	nextPos := 0
	if maxPos.Valid {
		nextPos = int(maxPos.Int64) + 1
	}

	_, err = d.conn.Exec(`
		INSERT INTO playlist_tracks (playlist_id, track_id, position)
		VALUES (?, ?, ?)
	`, playlistID, trackID, nextPos)
	return err
}

func (d *DB) RemoveTrackFromPlaylistAt(playlistID int64, index int) error {
	_, err := d.conn.Exec(`
		DELETE FROM playlist_tracks
		WHERE id = (
			SELECT id FROM playlist_tracks
			WHERE playlist_id = ?
			ORDER BY position
			LIMIT 1 OFFSET ?
		)
	`, playlistID, index)
	return err
}

func (d *DB) SwapPositions(playlistID int64, i, j int) error {
	if i == j {
		return nil
	}

	type row struct {
		id  int64
		pos float64
	}
	get := func(idx int) (row, error) {
		var r row
		err := d.conn.QueryRow(`
			SELECT id, position FROM playlist_tracks
			WHERE playlist_id = ?
			ORDER BY position
			LIMIT 1 OFFSET ?
		`, playlistID, idx).Scan(&r.id, &r.pos)
		return r, err
	}

	a, err := get(i)
	if err != nil {
		return fmt.Errorf("index %d tidak ditemukan: %w", i, err)
	}
	b, err := get(j)
	if err != nil {
		return fmt.Errorf("index %d tidak ditemukan: %w", j, err)
	}

	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`UPDATE playlist_tracks SET position = ? WHERE id = ?`, b.pos, a.id,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`UPDATE playlist_tracks SET position = ? WHERE id = ?`, a.pos, b.id,
	); err != nil {
		return err
	}
	return tx.Commit()
}

// isUniqueConstraintError mendeteksi error unique constraint dari
// modernc.org/sqlite. Pesan errornya mengandung "UNIQUE constraint failed".
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}