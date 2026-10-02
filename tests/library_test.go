package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ade-nugraha306/sky-project/internal/library"
)

func TestScanFolder_Empty(t *testing.T) {
	dir := t.TempDir()
	paths, err := library.ScanFolder(dir)
	if err != nil {
		t.Fatalf("ScanFolder error: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(paths))
	}
}

func TestScanFolder_OnlySupported(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "a.mp3")
	createFile(t, dir, "b.flac")
	createFile(t, dir, "c.ogg")
	createFile(t, dir, "d.wav")
	createFile(t, dir, "e.txt")
	createFile(t, dir, "f.m4a") // tidak didukung
	createFile(t, dir, "g.MP3") // uppercase harus tetap match

	paths, err := library.ScanFolder(dir)
	if err != nil {
		t.Fatalf("ScanFolder error: %v", err)
	}
	if len(paths) != 5 {
		t.Errorf("expected 5 paths, got %d: %v", len(paths), paths)
	}
}

func TestScanFolder_Recursive(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "root.mp3")

	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	createFile(t, sub, "nested.mp3")

	deeper := filepath.Join(sub, "deeper")
	if err := os.MkdirAll(deeper, 0o755); err != nil {
		t.Fatal(err)
	}
	createFile(t, deeper, "deepest.flac")

	paths, err := library.ScanFolder(dir)
	if err != nil {
		t.Fatalf("ScanFolder error: %v", err)
	}
	if len(paths) != 3 {
		t.Errorf("expected 3 paths, got %d: %v", len(paths), paths)
	}
}

func TestScanFolder_SkipsHiddenDirs(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "visible.mp3")

	hidden := filepath.Join(dir, ".hidden")
	if err := os.MkdirAll(hidden, 0o755); err != nil {
		t.Fatal(err)
	}
	createFile(t, hidden, "should-not-appear.mp3")

	paths, err := library.ScanFolder(dir)
	if err != nil {
		t.Fatalf("ScanFolder error: %v", err)
	}
	if len(paths) != 1 {
		t.Errorf("expected 1 path, got %d: %v", len(paths), paths)
	}
	if filepath.Base(paths[0]) != "visible.mp3" {
		t.Errorf("expected visible.mp3, got %s", paths[0])
	}
}

func TestScanFolder_NonExistentRoot(t *testing.T) {
	dir := t.TempDir()
	nonExistent := filepath.Join(dir, "does-not-exist")

	paths, err := library.ScanFolder(nonExistent)
	if err != nil {
		t.Logf("ScanFolder returned error: %v (acceptable)", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(paths))
	}
}

func TestParseMetadata_FallbackToFilename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "My Song.mp3")
	os.WriteFile(path, []byte{}, 0o644)

	track := library.ParseMetadata(path)
	if track.Path != path {
		t.Errorf("Path: got %q, want %q", track.Path, path)
	}
	if track.Title != "My Song" {
		t.Errorf("Title: got %q, want %q", track.Title, "My Song")
	}
	if track.Artist != "" {
		t.Errorf("Artist should be empty, got %q", track.Artist)
	}
}

func TestParseMetadata_NonExistentFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ghost.mp3")

	track := library.ParseMetadata(path)
	if track.Path != path {
		t.Errorf("Path: got %q, want %q", track.Path, path)
	}
	if track.Title != "ghost" {
		t.Errorf("Title: got %q, want %q", track.Title, "ghost")
	}
}

func TestParseMetadata_ExtensionStripped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	os.WriteFile(path, []byte{}, 0o644)

	track := library.ParseMetadata(path)
	if track.Title != "track" {
		t.Errorf("Title: got %q, want %q", track.Title, "track")
	}
}