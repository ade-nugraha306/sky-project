package tests

import (
	"errors"
	"testing"

	"github.com/ade-nugraha306/sky-project/internal/db"
	"github.com/ade-nugraha306/sky-project/internal/library"
)

func TestCreatePlaylist(t *testing.T) {
	d := newTestDB(t)

	id, err := d.CreatePlaylist("My Favorites")
	if err != nil {
		t.Fatalf("CreatePlaylist: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}
}

func TestCreatePlaylist_DuplicateName(t *testing.T) {
	d := newTestDB(t)
	d.CreatePlaylist("Favorites")

	_, err := d.CreatePlaylist("Favorites")
	if !errors.Is(err, db.ErrPlaylistNameExists) {
		t.Errorf("expected ErrPlaylistNameExists, got %v", err)
	}
}

func TestCreatePlaylist_TrimsWhitespace(t *testing.T) {
	d := newTestDB(t)
	id, err := d.CreatePlaylist("  Chill  ")
	if err != nil {
		t.Fatalf("CreatePlaylist: %v", err)
	}
	p, _ := d.PlaylistByID(id)
	if p.Name != "Chill" {
		t.Errorf("name not trimmed: got %q", p.Name)
	}
}

func TestCreatePlaylist_EmptyName(t *testing.T) {
	d := newTestDB(t)
	_, err := d.CreatePlaylist("   ")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestAllPlaylists_Empty(t *testing.T) {
	d := newTestDB(t)
	ps, err := d.AllPlaylists()
	if err != nil {
		t.Fatalf("AllPlaylists: %v", err)
	}
	if len(ps) != 0 {
		t.Errorf("expected 0 playlists, got %d", len(ps))
	}
}

func TestAllPlaylists_Ordered(t *testing.T) {
	d := newTestDB(t)
	d.CreatePlaylist("Zebra")
	d.CreatePlaylist("apple")
	d.CreatePlaylist("Mango")

	ps, err := d.AllPlaylists()
	if err != nil {
		t.Fatalf("AllPlaylists: %v", err)
	}
	if len(ps) != 3 {
		t.Fatalf("expected 3, got %d", len(ps))
	}
	// NOCASE: apple, Mango, Zebra
	want := []string{"apple", "Mango", "Zebra"}
	for i, w := range want {
		if ps[i].Name != w {
			t.Errorf("playlist[%d]: got %q, want %q", i, ps[i].Name, w)
		}
	}
}

func TestAllPlaylists_IncludesEmpty(t *testing.T) {
	d := newTestDB(t)
	d.CreatePlaylist("Empty Playlist")

	ps, _ := d.AllPlaylists()
	if len(ps) != 1 {
		t.Fatalf("expected 1, got %d", len(ps))
	}
	if ps[0].TrackCount != 0 {
		t.Errorf("expected 0 tracks, got %d", ps[0].TrackCount)
	}
}

func TestRenamePlaylist(t *testing.T) {
	d := newTestDB(t)
	id, _ := d.CreatePlaylist("Old Name")

	if err := d.RenamePlaylist(id, "New Name"); err != nil {
		t.Fatalf("RenamePlaylist: %v", err)
	}
	p, _ := d.PlaylistByID(id)
	if p.Name != "New Name" {
		t.Errorf("name: got %q, want %q", p.Name, "New Name")
	}
}

func TestRenamePlaylist_Conflict(t *testing.T) {
	d := newTestDB(t)
	d.CreatePlaylist("A")
	id2, _ := d.CreatePlaylist("B")

	err := d.RenamePlaylist(id2, "A")
	if !errors.Is(err, db.ErrPlaylistNameExists) {
		t.Errorf("expected ErrPlaylistNameExists, got %v", err)
	}
}

func TestRenamePlaylist_NotFound(t *testing.T) {
	d := newTestDB(t)
	err := d.RenamePlaylist(999, "Whatever")
	if !errors.Is(err, db.ErrPlaylistNotFound) {
		t.Errorf("expected ErrPlaylistNotFound, got %v", err)
	}
}

func TestDeletePlaylist(t *testing.T) {
	d := newTestDB(t)
	id, _ := d.CreatePlaylist("To Delete")

	if err := d.DeletePlaylist(id); err != nil {
		t.Fatalf("DeletePlaylist: %v", err)
	}

	_, err := d.PlaylistByID(id)
	if !errors.Is(err, db.ErrPlaylistNotFound) {
		t.Errorf("playlist should be gone, got err=%v", err)
	}
}

func TestDeletePlaylist_NotFound(t *testing.T) {
	d := newTestDB(t)
	err := d.DeletePlaylist(999)
	if !errors.Is(err, db.ErrPlaylistNotFound) {
		t.Errorf("expected ErrPlaylistNotFound, got %v", err)
	}
}

func TestPlaylistByID(t *testing.T) {
	d := newTestDB(t)
	id, _ := d.CreatePlaylist("Test")

	p, err := d.PlaylistByID(id)
	if err != nil {
		t.Fatalf("PlaylistByID: %v", err)
	}
	if p.Name != "Test" {
		t.Errorf("name: got %q", p.Name)
	}
	if p.TrackCount != 0 {
		t.Errorf("track count: got %d", p.TrackCount)
	}
}

func TestTracksInPlaylist_Empty(t *testing.T) {
	d := newTestDB(t)
	id, _ := d.CreatePlaylist("Empty")

	tracks, err := d.TracksInPlaylist(id)
	if err != nil {
		t.Fatalf("TracksInPlaylist: %v", err)
	}
	if len(tracks) != 0 {
		t.Errorf("expected 0 tracks, got %d", len(tracks))
	}
}

// Catatan: test insert track ke playlist (AddTrackToPlaylist)
// akan ditambahkan di Batch 3 bersama logika duplikat.

func insertTestTrack(t *testing.T, d *db.DB, path string) {
	t.Helper()
	if err := d.UpsertTrack(library.Track{
		Path:  path,
		Title: path,
	}); err != nil {
		t.Fatalf("UpsertTrack: %v", err)
	}
}

func TestAddTrackToPlaylist(t *testing.T) {
	d := newTestDB(t)
	pid, _ := d.CreatePlaylist("Test")
	insertTestTrack(t, d, "D:/Music/song.mp3")

	if err := d.AddTrackToPlaylist(pid, "D:/Music/song.mp3"); err != nil {
		t.Fatalf("AddTrackToPlaylist: %v", err)
	}

	tracks, _ := d.TracksInPlaylist(pid)
	if len(tracks) != 1 {
		t.Errorf("expected 1 track, got %d", len(tracks))
	}
	if tracks[0].Path != "D:/Music/song.mp3" {
		t.Errorf("wrong track: %s", tracks[0].Path)
	}
}

func TestAddTrackToPlaylist_NotFound(t *testing.T) {
	d := newTestDB(t)
	pid, _ := d.CreatePlaylist("Test")

	err := d.AddTrackToPlaylist(pid, "D:/Music/ghost.mp3")
	if err == nil {
		t.Error("expected error for non-existent track")
	}
}

func TestAddTrackToPlaylist_AllowsDuplicate(t *testing.T) {
	d := newTestDB(t)
	pid, _ := d.CreatePlaylist("Test")
	insertTestTrack(t, d, "D:/Music/song.mp3")

	// Tambah dua kali — harus sukses keduanya.
	if err := d.AddTrackToPlaylist(pid, "D:/Music/song.mp3"); err != nil {
		t.Fatalf("add #1: %v", err)
	}
	if err := d.AddTrackToPlaylist(pid, "D:/Music/song.mp3"); err != nil {
		t.Fatalf("add #2: %v", err)
	}

	tracks, _ := d.TracksInPlaylist(pid)
	if len(tracks) != 2 {
		t.Errorf("expected 2 tracks (duplicate), got %d", len(tracks))
	}
}

func TestAddTrackToPlaylist_PositionAppends(t *testing.T) {
	d := newTestDB(t)
	pid, _ := d.CreatePlaylist("Test")
	insertTestTrack(t, d, "a")
	insertTestTrack(t, d, "b")
	insertTestTrack(t, d, "c")

	d.AddTrackToPlaylist(pid, "a")
	d.AddTrackToPlaylist(pid, "b")
	d.AddTrackToPlaylist(pid, "c")

	tracks, _ := d.TracksInPlaylist(pid)
	// Urutan harus a, b, c (sesuai urutan add).
	want := []string{"a", "b", "c"}
	for i, w := range want {
		if tracks[i].Path != w {
			t.Errorf("track[%d]: got %s, want %s", i, tracks[i].Path, w)
		}
	}
}

func TestTrackInPlaylist(t *testing.T) {
	d := newTestDB(t)
	pid, _ := d.CreatePlaylist("Test")
	insertTestTrack(t, d, "D:/Music/song.mp3")
	insertTestTrack(t, d, "D:/Music/other.mp3")

	// Belum ada.
	exists, err := d.TrackInPlaylist(pid, "D:/Music/song.mp3")
	if err != nil {
		t.Fatalf("TrackInPlaylist: %v", err)
	}
	if exists {
		t.Error("should not exist yet")
	}

	// Tambah, cek lagi.
	d.AddTrackToPlaylist(pid, "D:/Music/song.mp3")
	exists, _ = d.TrackInPlaylist(pid, "D:/Music/song.mp3")
	if !exists {
		t.Error("should exist now")
	}

	// Track lain belum ditambahkan.
	exists, _ = d.TrackInPlaylist(pid, "D:/Music/other.mp3")
	if exists {
		t.Error("other should not exist")
	}
}

func TestTrackInPlaylist_WrongPlaylist(t *testing.T) {
	d := newTestDB(t)
	pid1, _ := d.CreatePlaylist("A")
	pid2, _ := d.CreatePlaylist("B")
	insertTestTrack(t, d, "D:/Music/song.mp3")

	d.AddTrackToPlaylist(pid1, "D:/Music/song.mp3")

	// Cek di playlist lain — harus false.
	exists, _ := d.TrackInPlaylist(pid2, "D:/Music/song.mp3")
	if exists {
		t.Error("should not exist in playlist B")
	}
}

func TestAddTrackToPlaylist_CascadeOnDelete(t *testing.T) {
	d := newTestDB(t)
	pid, _ := d.CreatePlaylist("Test")
	insertTestTrack(t, d, "D:/Music/song.mp3")
	d.AddTrackToPlaylist(pid, "D:/Music/song.mp3")

	// Hapus playlist — track di dalamnya ikut terhapus (CASCADE).
	if err := d.DeletePlaylist(pid); err != nil {
		t.Fatalf("DeletePlaylist: %v", err)
	}

	tracks, _ := d.TracksInPlaylist(pid)
	if len(tracks) != 0 {
		t.Errorf("expected 0 tracks after cascade delete, got %d", len(tracks))
	}
}

func TestDeleteTrack_CascadeToPlaylist(t *testing.T) {
	d := newTestDB(t)
	pid, _ := d.CreatePlaylist("Test")
	insertTestTrack(t, d, "D:/Music/song.mp3")
	d.AddTrackToPlaylist(pid, "D:/Music/song.mp3")

	// Hapus track dari library — harus hilang dari playlist (CASCADE).
	_, err := d.DeleteTracksUnderFolder("D:/Music")
	if err != nil {
		t.Fatalf("DeleteTracksUnderFolder: %v", err)
	}

	tracks, _ := d.TracksInPlaylist(pid)
	if len(tracks) != 0 {
		t.Errorf("expected 0 tracks after track cascade, got %d", len(tracks))
	}
}

func TestRemoveTrackFromPlaylistAt(t *testing.T) {
	d := newTestDB(t)
	pid, _ := d.CreatePlaylist("Test")
	insertTestTrack(t, d, "a")
	insertTestTrack(t, d, "b")
	insertTestTrack(t, d, "c")

	d.AddTrackToPlaylist(pid, "a")
	d.AddTrackToPlaylist(pid, "b")
	d.AddTrackToPlaylist(pid, "c")

	// Hapus track di index 1 (b).
	if err := d.RemoveTrackFromPlaylistAt(pid, 1); err != nil {
		t.Fatalf("RemoveTrackFromPlaylistAt: %v", err)
	}

	tracks, _ := d.TracksInPlaylist(pid)
	if len(tracks) != 2 {
		t.Fatalf("expected 2 tracks, got %d", len(tracks))
	}
	want := []string{"a", "c"}
	for i, w := range want {
		if tracks[i].Path != w {
			t.Errorf("track[%d]: got %s, want %s", i, tracks[i].Path, w)
		}
	}
}

func TestRemoveTrackFromPlaylistAt_FirstAndLast(t *testing.T) {
	d := newTestDB(t)
	pid, _ := d.CreatePlaylist("Test")
	insertTestTrack(t, d, "a")
	insertTestTrack(t, d, "b")
	insertTestTrack(t, d, "c")

	d.AddTrackToPlaylist(pid, "a")
	d.AddTrackToPlaylist(pid, "b")
	d.AddTrackToPlaylist(pid, "c")

	d.RemoveTrackFromPlaylistAt(pid, 0) // hapus a
	d.RemoveTrackFromPlaylistAt(pid, 1) // sekarang c di index 1

	tracks, _ := d.TracksInPlaylist(pid)
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}
	if tracks[0].Path != "b" {
		t.Errorf("expected b, got %s", tracks[0].Path)
	}
}

func TestRemoveTrackFromPlaylistAt_OnlyOneDuplicate(t *testing.T) {
	d := newTestDB(t)
	pid, _ := d.CreatePlaylist("Test")
	insertTestTrack(t, d, "song")
	d.AddTrackToPlaylist(pid, "song")
	d.AddTrackToPlaylist(pid, "song")
	d.AddTrackToPlaylist(pid, "song")

	// Hapus 1 instance.
	if err := d.RemoveTrackFromPlaylistAt(pid, 1); err != nil {
		t.Fatalf("RemoveTrackFromPlaylistAt: %v", err)
	}

	tracks, _ := d.TracksInPlaylist(pid)
	if len(tracks) != 2 {
		t.Errorf("expected 2 remaining, got %d", len(tracks))
	}
}

func TestRemoveTrackFromPlaylistAt_OutOfRange(t *testing.T) {
	d := newTestDB(t)
	pid, _ := d.CreatePlaylist("Test")
	insertTestTrack(t, d, "a")
	d.AddTrackToPlaylist(pid, "a")

	// Index 5 tidak ada — harus no-op tanpa error.
	if err := d.RemoveTrackFromPlaylistAt(pid, 5); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	tracks, _ := d.TracksInPlaylist(pid)
	if len(tracks) != 1 {
		t.Errorf("expected 1 track (unchanged), got %d", len(tracks))
	}
}