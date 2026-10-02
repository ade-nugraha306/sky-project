package tests

import (
	"path/filepath"
	"testing"

	"github.com/ade-nugraha306/sky-project/internal/db"
	"github.com/ade-nugraha306/sky-project/internal/library"
)

func newTestDB(t *testing.T) *db.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	d, err := db.New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestNew_EmptyDB(t *testing.T) {
	d := newTestDB(t)
	n, err := d.CountTracks()
	if err != nil {
		t.Fatalf("CountTracks: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 tracks, got %d", n)
	}
}

func TestUpsertTrack_Insert(t *testing.T) {
	d := newTestDB(t)
	track := library.Track{
		Path:       "D:/Music/song.mp3",
		Title:      "Song",
		Artist:     "Artist",
		Album:      "Album",
		DurationMs: 180000,
	}
	if err := d.UpsertTrack(track); err != nil {
		t.Fatalf("UpsertTrack: %v", err)
	}
	n, _ := d.CountTracks()
	if n != 1 {
		t.Errorf("expected 1 track, got %d", n)
	}
}

func TestUpsertTrack_Idempotent(t *testing.T) {
	d := newTestDB(t)
	track := library.Track{Path: "D:/Music/song.mp3", Title: "Song"}

	for i := 0; i < 3; i++ {
		if err := d.UpsertTrack(track); err != nil {
			t.Fatalf("UpsertTrack #%d: %v", i, err)
		}
	}
	n, _ := d.CountTracks()
	if n != 1 {
		t.Errorf("expected 1 track after 3 upserts, got %d", n)
	}
}

func TestUpsertTrack_UpdatesOnConflict(t *testing.T) {
	d := newTestDB(t)
	path := "D:/Music/song.mp3"

	d.UpsertTrack(library.Track{Path: path, Title: "Original", Artist: "Old"})
	d.UpsertTrack(library.Track{Path: path, Title: "Updated", Artist: "New"})

	tracks, err := d.AllTracks()
	if err != nil {
		t.Fatalf("AllTracks: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}
	if tracks[0].Title != "Updated" {
		t.Errorf("Title: got %q, want %q", tracks[0].Title, "Updated")
	}
	if tracks[0].Artist != "New" {
		t.Errorf("Artist: got %q, want %q", tracks[0].Artist, "New")
	}
}

func TestAllTracks_Ordering(t *testing.T) {
	d := newTestDB(t)
	d.UpsertTrack(library.Track{Path: "a", Title: "Z", Artist: "B"})
	d.UpsertTrack(library.Track{Path: "b", Title: "A", Artist: "B"})
	d.UpsertTrack(library.Track{Path: "c", Title: "M", Artist: "A"})

	tracks, err := d.AllTracks()
	if err != nil {
		t.Fatalf("AllTracks: %v", err)
	}
	if len(tracks) != 3 {
		t.Fatalf("expected 3, got %d", len(tracks))
	}
	want := []struct{ artist, title string }{
		{"A", "M"}, {"B", "A"}, {"B", "Z"},
	}
	for i, w := range want {
		if tracks[i].Artist != w.artist || tracks[i].Title != w.title {
			t.Errorf("track[%d]: got (%s,%s), want (%s,%s)",
				i, tracks[i].Artist, tracks[i].Title, w.artist, w.title)
		}
	}
}

func TestDeleteTracksUnderFolder_Prefix(t *testing.T) {
	d := newTestDB(t)

	d.UpsertTrack(library.Track{Path: "D:/Music/a.mp3"})
	d.UpsertTrack(library.Track{Path: "D:/Music/sub/b.mp3"})
	d.UpsertTrack(library.Track{Path: "D:/Music/sub/deep/c.mp3"})
	d.UpsertTrack(library.Track{Path: "D:/Music2/d.mp3"})
	d.UpsertTrack(library.Track{Path: "D:/Other/e.mp3"})

	n, err := d.DeleteTracksUnderFolder("D:/Music")
	if err != nil {
		t.Fatalf("DeleteTracksUnderFolder: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 deleted, got %d", n)
	}

	remaining, _ := d.CountTracks()
	if remaining != 2 {
		t.Errorf("expected 2 remaining, got %d", remaining)
	}
}

func TestDeleteTracksUnderFolder_NoMatch(t *testing.T) {
	d := newTestDB(t)
	d.UpsertTrack(library.Track{Path: "D:/Music/a.mp3"})

	n, err := d.DeleteTracksUnderFolder("D:/Nonexistent")
	if err != nil {
		t.Fatalf("DeleteTracksUnderFolder: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 deleted, got %d", n)
	}
}

func TestDeleteTracksUnderFolder_ExactMatch(t *testing.T) {
	d := newTestDB(t)
	d.UpsertTrack(library.Track{Path: "D:/Music"})
	d.UpsertTrack(library.Track{Path: "D:/Music/a.mp3"})

	n, err := d.DeleteTracksUnderFolder("D:/Music")
	if err != nil {
		t.Fatalf("DeleteTracksUnderFolder: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 deleted (exact + child), got %d", n)
	}
}