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

// upsertOne membungkus UpsertTracks untuk single track. Dipakai
// oleh test yang hanya butuh insert satu track.
func upsertOne(t *testing.T, d *db.DB, track library.Track) {
	t.Helper()
	if err := d.UpsertTracks([]library.Track{track}); err != nil {
		t.Fatalf("UpsertTracks: %v", err)
	}
}

func setAddedAt(t *testing.T, d *db.DB, path, ts string) {
	t.Helper()
	_, err := d.Conn().Exec(`UPDATE tracks SET added_at = ? WHERE path = ?`, ts, path)
	if err != nil {
		t.Fatalf("setAddedAt: %v", err)
	}
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

// ===== UpsertTrack (single, via upsertOne helper) =====

func TestUpsertTrack_Insert(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{
		Path:       "D:/Music/song.mp3",
		Title:      "Song",
		Artist:     "Artist",
		Album:      "Album",
		DurationMs: 180000,
	})
	n, _ := d.CountTracks()
	if n != 1 {
		t.Errorf("expected 1 track, got %d", n)
	}
}

func TestUpsertTrack_Idempotent(t *testing.T) {
	d := newTestDB(t)
	track := library.Track{Path: "D:/Music/song.mp3", Title: "Song"}

	for i := 0; i < 3; i++ {
		upsertOne(t, d, track)
	}
	n, _ := d.CountTracks()
	if n != 1 {
		t.Errorf("expected 1 track after 3 upserts, got %d", n)
	}
}

func TestUpsertTrack_UpdatesOnConflict(t *testing.T) {
	d := newTestDB(t)
	path := "D:/Music/song.mp3"

	upsertOne(t, d, library.Track{Path: path, Title: "Original", Artist: "Old"})
	upsertOne(t, d, library.Track{Path: path, Title: "Updated", Artist: "New"})

	tracks, err := d.AllTracks(library.SortByTitle)
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

// ===== UpsertTracks (batch) =====

func TestUpsertTracks_Batch(t *testing.T) {
	d := newTestDB(t)
	tracks := []library.Track{
		{Path: "D:/Music/a.mp3", Title: "A", Artist: "X"},
		{Path: "D:/Music/b.mp3", Title: "B", Artist: "Y"},
		{Path: "D:/Music/c.mp3", Title: "C", Artist: "Z"},
	}
	if err := d.UpsertTracks(tracks); err != nil {
		t.Fatalf("UpsertTracks: %v", err)
	}
	n, _ := d.CountTracks()
	if n != 3 {
		t.Errorf("expected 3 tracks, got %d", n)
	}
}

func TestUpsertTracks_Idempotent(t *testing.T) {
	d := newTestDB(t)
	tracks := []library.Track{
		{Path: "a", Title: "A"},
		{Path: "b", Title: "B"},
	}
	for i := 0; i < 3; i++ {
		if err := d.UpsertTracks(tracks); err != nil {
			t.Fatalf("UpsertTracks #%d: %v", i, err)
		}
	}
	n, _ := d.CountTracks()
	if n != 2 {
		t.Errorf("expected 2 tracks after 3 batch upserts, got %d", n)
	}
}

func TestUpsertTracks_Empty(t *testing.T) {
	d := newTestDB(t)
	if err := d.UpsertTracks(nil); err != nil {
		t.Fatalf("UpsertTracks nil: %v", err)
	}
	if err := d.UpsertTracks([]library.Track{}); err != nil {
		t.Fatalf("UpsertTracks empty slice: %v", err)
	}
	n, _ := d.CountTracks()
	if n != 0 {
		t.Errorf("expected 0 tracks, got %d", n)
	}
}

func TestUpsertTracks_UpdatesOnConflict(t *testing.T) {
	d := newTestDB(t)
	first := []library.Track{{Path: "a", Title: "Old"}}
	second := []library.Track{{Path: "a", Title: "New"}}

	d.UpsertTracks(first)
	d.UpsertTracks(second)

	tracks, _ := d.AllTracks(library.SortByTitle)
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}
	if tracks[0].Title != "New" {
		t.Errorf("Title: got %q, want %q", tracks[0].Title, "New")
	}
}

// ===== AllTracks / AllTracksInFolder =====

func TestAllTracks_Ordering(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "a", Title: "Z", Artist: "B"})
	upsertOne(t, d, library.Track{Path: "b", Title: "A", Artist: "B"})
	upsertOne(t, d, library.Track{Path: "c", Title: "M", Artist: "A"})

	tracks, err := d.AllTracks(library.SortByArtist)
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

func TestAllTracksInFolder_EmptyMeansAll(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "D:/Music/a.mp3"})
	upsertOne(t, d, library.Track{Path: "E:/Other/b.mp3"})

	// Folder kosong = semua track.
	tracks, err := d.AllTracksInFolder("", library.SortByArtist)
	if err != nil {
		t.Fatalf("AllTracksInFolder: %v", err)
	}
	if len(tracks) != 2 {
		t.Errorf("expected 2 (all), got %d", len(tracks))
	}
}

func TestAllTracksInFolder_Filter(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "D:/Music/a.mp3"})
	upsertOne(t, d, library.Track{Path: "D:/Music/sub/b.mp3"})
	upsertOne(t, d, library.Track{Path: "D:/Music2/c.mp3"})
	upsertOne(t, d, library.Track{Path: "E:/Other/d.mp3"})

	tracks, err := d.AllTracksInFolder("D:/Music", library.SortByArtist)
	if err != nil {
		t.Fatalf("AllTracksInFolder: %v", err)
	}
	if len(tracks) != 2 {
		t.Errorf("expected 2 (a + sub/b), got %d", len(tracks))
	}
	for _, tr := range tracks {
		if tr.Path != "D:/Music/a.mp3" && tr.Path != "D:/Music/sub/b.mp3" {
			t.Errorf("unexpected track: %s", tr.Path)
		}
	}
}

func TestAllTracksInFolder_PrefixNoMatch(t *testing.T) {
	d := newTestDB(t)
	// "D:/Music2" tidak boleh match dengan prefix "D:/Music"
	upsertOne(t, d, library.Track{Path: "D:/Music2/a.mp3"})
	upsertOne(t, d, library.Track{Path: "D:/Music/a.mp3"})

	tracks, _ := d.AllTracksInFolder("D:/Music", library.SortByTitle)
	if len(tracks) != 1 {
		t.Errorf("expected 1, got %d", len(tracks))
	}
	if tracks[0].Path != "D:/Music/a.mp3" {
		t.Errorf("wrong track: %s", tracks[0].Path)
	}
}

func TestAllTracksInFolder_NoMatch(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "D:/Music/a.mp3"})

	tracks, err := d.AllTracksInFolder("E:/Nonexistent", library.SortByTitle)
	if err != nil {
		t.Fatalf("AllTracksInFolder: %v", err)
	}
	if len(tracks) != 0 {
		t.Errorf("expected 0, got %d", len(tracks))
	}
}

// ===== DeleteTracksUnderFolder =====

func TestDeleteTracksUnderFolder_Prefix(t *testing.T) {
	d := newTestDB(t)

	upsertOne(t, d, library.Track{Path: "D:/Music/a.mp3"})
	upsertOne(t, d, library.Track{Path: "D:/Music/sub/b.mp3"})
	upsertOne(t, d, library.Track{Path: "D:/Music/sub/deep/c.mp3"})
	upsertOne(t, d, library.Track{Path: "D:/Music2/d.mp3"})
	upsertOne(t, d, library.Track{Path: "D:/Other/e.mp3"})

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
	upsertOne(t, d, library.Track{Path: "D:/Music/a.mp3"})

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
	upsertOne(t, d, library.Track{Path: "D:/Music"})
	upsertOne(t, d, library.Track{Path: "D:/Music/a.mp3"})

	n, err := d.DeleteTracksUnderFolder("D:/Music")
	if err != nil {
		t.Fatalf("DeleteTracksUnderFolder: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 deleted (exact + child), got %d", n)
	}
}

// ===== DeleteTracksNotIn (scan reconciliation) =====

func TestDeleteTracksNotIn(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "D:/Music/a.mp3"})
	upsertOne(t, d, library.Track{Path: "D:/Music/b.mp3"})
	upsertOne(t, d, library.Track{Path: "D:/Music/c.mp3"})
	upsertOne(t, d, library.Track{Path: "D:/Music/d.mp3"})
	upsertOne(t, d, library.Track{Path: "D:/Music/e.mp3"})

	foundPaths := map[string]bool{
		"D:/Music/a.mp3": true,
		"D:/Music/b.mp3": true,
	}

	n, err := d.DeleteTracksNotIn("D:/Music", foundPaths)
	if err != nil {
		t.Fatalf("DeleteTracksNotIn: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 deleted (c, d, e), got %d", n)
	}

	remaining, _ := d.CountTracks()
	if remaining != 2 {
		t.Errorf("expected 2 remaining (a, b), got %d", remaining)
	}
}

func TestDeleteTracksNotIn_OnlyAffectedFolder(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "D:/Music/a.mp3"})
	upsertOne(t, d, library.Track{Path: "D:/Music/b.mp3"})
	upsertOne(t, d, library.Track{Path: "E:/Other/x.mp3"})

	// Scan folder "D:/Music" — hanya "a.mp3" ada di disk.
	foundPaths := map[string]bool{"D:/Music/a.mp3": true}

	n, err := d.DeleteTracksNotIn("D:/Music", foundPaths)
	if err != nil {
		t.Fatalf("DeleteTracksNotIn: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 deleted (b), got %d", n)
	}

	// Track dari folder lain harus tetap ada.
	remaining, _ := d.CountTracks()
	if remaining != 2 {
		t.Errorf("expected 2 remaining (a + x), got %d", remaining)
	}
}

func TestDeleteTracksNotIn_NoMatch(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "D:/Music/a.mp3"})

	foundPaths := map[string]bool{
		"D:/Music/a.mp3": true,
		"D:/Music/b.mp3": true,
	}

	n, err := d.DeleteTracksNotIn("D:/Music", foundPaths)
	if err != nil {
		t.Fatalf("DeleteTracksNotIn: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 deleted, got %d", n)
	}
}

func TestDeleteTracksNotIn_Empty(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "D:/Music/a.mp3"})
	upsertOne(t, d, library.Track{Path: "D:/Music/b.mp3"})

	// Scan menemukan kosong — semua track di bawah folder dihapus.
	foundPaths := map[string]bool{}

	n, err := d.DeleteTracksNotIn("D:/Music", foundPaths)
	if err != nil {
		t.Fatalf("DeleteTracksNotIn: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 deleted, got %d", n)
	}

	remaining, _ := d.CountTracks()
	if remaining != 0 {
		t.Errorf("expected 0 remaining, got %d", remaining)
	}
}

func TestDeleteTracksNotIn_Subdirectory(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "D:/Music/sub/a.mp3"})
	upsertOne(t, d, library.Track{Path: "D:/Music/sub/deep/b.mp3"})

	foundPaths := map[string]bool{
		"D:/Music/sub/a.mp3": true,
	}

	n, err := d.DeleteTracksNotIn("D:/Music", foundPaths)
	if err != nil {
		t.Fatalf("DeleteTracksNotIn: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 deleted (deep/b), got %d", n)
	}
}

func TestAllTracks_SortByTitle(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "a", Title: "Zebra"})
	upsertOne(t, d, library.Track{Path: "b", Title: "apple"})
	upsertOne(t, d, library.Track{Path: "c", Title: "Mango"})

	tracks, _ := d.AllTracks(library.SortByTitle)
	want := []string{"apple", "Mango", "Zebra"} // NOCASE
	for i, w := range want {
		if tracks[i].Title != w {
			t.Errorf("track[%d]: got %q, want %q", i, tracks[i].Title, w)
		}
	}
}

func TestAllTracks_SortByAlbum(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "a", Album: "Zebra"})
	upsertOne(t, d, library.Track{Path: "b", Album: "apple"})
	upsertOne(t, d, library.Track{Path: "c", Album: "Mango"})

	tracks, _ := d.AllTracks(library.SortByAlbum)
	want := []string{"apple", "Mango", "Zebra"}
	for i, w := range want {
		if tracks[i].Album != w {
			t.Errorf("track[%d]: got %q, want %q", i, tracks[i].Album, w)
		}
	}
}

func TestAllTracks_SortByDateDesc(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "a", Title: "First"})
	upsertOne(t, d, library.Track{Path: "b", Title: "Second"})
	upsertOne(t, d, library.Track{Path: "c", Title: "Third"})

	// Set timestamp eksplisit supaya berbeda (insert cepat bisa
	// punya added_at identik karena presisi detik).
	setAddedAt(t, d, "a", "2026-01-01 10:00:00")
	setAddedAt(t, d, "b", "2026-01-02 10:00:00")
	setAddedAt(t, d, "c", "2026-01-03 10:00:00")

	tracks, err := d.AllTracks(library.SortByDateDesc)
	if err != nil {
		t.Fatalf("AllTracks: %v", err)
	}
	// Terbaru di depan: c (2026-01-03), b (2026-01-02), a (2026-01-01)
	want := []string{"Third", "Second", "First"}
	for i, w := range want {
		if tracks[i].Title != w {
			t.Errorf("track[%d]: got %q, want %q", i, tracks[i].Title, w)
		}
	}
}

func TestTrackByPath_ForwardSlash(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "D:/Music/song.mp3", Title: "Song"})

	// Query dengan forward slash.
	track, err := d.TrackByPath("D:/Music/song.mp3")
	if err != nil {
		t.Fatalf("TrackByPath: %v", err)
	}
	if track.Title != "Song" {
		t.Errorf("Title: got %q, want Song", track.Title)
	}
}

func TestTrackByPath_NormalizesInput(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "D:/Music/song.mp3", Title: "Song"})

	// Query dengan trailing slash dan backslash — harus tetap match.
	_, err := d.TrackByPath("D:/Music/song.mp3/")
	if err != nil {
		t.Errorf("TrackByPath with trailing slash: %v", err)
	}
}

func TestTrackByPath_NotFound(t *testing.T) {
	d := newTestDB(t)
	upsertOne(t, d, library.Track{Path: "D:/Music/song.mp3"})

	_, err := d.TrackByPath("D:/Music/ghost.mp3")
	if err == nil {
		t.Error("expected error for non-existent track")
	}
}