package tests

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/ade-nugraha306/sky-project/internal/config"
)

func TestLoad_CreatesDefaultConfig(t *testing.T) {
	setTempHome(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Volume != 100 {
		t.Errorf("Volume: got %d, want 100", cfg.Volume)
	}
	if cfg.RepeatMode != "off" {
		t.Errorf("RepeatMode: got %q, want %q", cfg.RepeatMode, "off")
	}
	if _, err := os.Stat(config.DefaultConfigPath()); err != nil {
		t.Errorf("config file not created: %v", err)
	}
}

func TestSave_RoundTrip(t *testing.T) {
	setTempHome(t)

	cfg, _ := config.Load()
	cfg.Volume = 42
	cfg.RepeatMode = "one"
	cfg.LastTrackPath = "D:/Music/song.mp3"
	cfg.LastPositionMs = 12345

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	cfg2, err := config.Load()
	if err != nil {
		t.Fatalf("Load #2: %v", err)
	}
	if cfg2.Volume != 42 {
		t.Errorf("Volume: got %d, want 42", cfg2.Volume)
	}
	if cfg2.RepeatMode != "one" {
		t.Errorf("RepeatMode: got %q, want %q", cfg2.RepeatMode, "one")
	}
	if cfg2.LastTrackPath != "D:/Music/song.mp3" {
		t.Errorf("LastTrackPath: got %q", cfg2.LastTrackPath)
	}
	if cfg2.LastPositionMs != 12345 {
		t.Errorf("LastPositionMs: got %d", cfg2.LastPositionMs)
	}
}

func TestAddFolder_Deduplicates(t *testing.T) {
	setTempHome(t)
	cfg, _ := config.Load()

	if err := cfg.AddFolder("D:/Music"); err != nil {
		t.Fatalf("AddFolder #1: %v", err)
	}
	if err := cfg.AddFolder("D:/Music"); err != nil {
		t.Fatalf("AddFolder #2: %v", err)
	}
	if len(cfg.MusicFolders) != 1 {
		t.Errorf("expected 1 folder after duplicate add, got %d", len(cfg.MusicFolders))
	}

	cfg.AddFolder("D:/Music/Other")
	if len(cfg.MusicFolders) != 2 {
		t.Errorf("expected 2 folders, got %d", len(cfg.MusicFolders))
	}
}

func TestAddFolder_NormalizesTrailingSlash(t *testing.T) {
	setTempHome(t)
	cfg, _ := config.Load()

	if runtime.GOOS == "windows" {
		cfg.AddFolder(`D:\Music\Album\`)
	} else {
		cfg.AddFolder("/tmp/Music/Album/")
	}

	for _, f := range cfg.MusicFolders {
		if len(f) > 0 && f[len(f)-1] == '/' {
			t.Errorf("folder still has trailing slash: %q", f)
		}
	}
}

func TestRemoveFolder(t *testing.T) {
	setTempHome(t)
	cfg, _ := config.Load()
	cfg.AddFolder("D:/Music/A")
	cfg.AddFolder("D:/Music/B")
	cfg.AddFolder("D:/Music/C")

	if err := cfg.RemoveFolder("D:/Music/B"); err != nil {
		t.Fatalf("RemoveFolder: %v", err)
	}
	if len(cfg.MusicFolders) != 2 {
		t.Errorf("expected 2 folders, got %d", len(cfg.MusicFolders))
	}
	for _, f := range cfg.MusicFolders {
		if f == "D:/Music/B" {
			t.Errorf("folder B still present")
		}
	}
}

func TestRemoveFolder_NonExistent(t *testing.T) {
	setTempHome(t)
	cfg, _ := config.Load()
	cfg.AddFolder("D:/Music/A")

	if err := cfg.RemoveFolder("D:/Music/Nonexistent"); err != nil {
		t.Fatalf("RemoveFolder non-existent: %v", err)
	}
	if len(cfg.MusicFolders) != 1 {
		t.Errorf("expected 1 folder, got %d", len(cfg.MusicFolders))
	}
}

func TestSetResume(t *testing.T) {
	setTempHome(t)
	cfg, _ := config.Load()

	cfg.SetResume("D:/Music/Album/song.mp3", 76200*time.Millisecond, 0)

	if cfg.LastTrackPath != "D:/Music/Album/song.mp3" {
		t.Errorf("LastTrackPath: got %q", cfg.LastTrackPath)
	}
	if cfg.LastPositionMs != 76200 {
		t.Errorf("LastPositionMs: got %d, want 76200", cfg.LastPositionMs)
	}
	if cfg.LastPlaylistID != 0 {
		t.Errorf("LastPlaylistID: got %d, want 0", cfg.LastPlaylistID)
	}
}

func TestSetResume_NormalizesPath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("path backslash hanya valid di Windows")
	}
	setTempHome(t)
	cfg, _ := config.Load()

	cfg.SetResume(`D:\Music\song.mp3`, time.Second, 0)
	if cfg.LastTrackPath != "D:/Music/song.mp3" {
		t.Errorf("path not normalized: %q", cfg.LastTrackPath)
	}
}

func TestClearResume(t *testing.T) {
	setTempHome(t)
	cfg, _ := config.Load()
	cfg.SetResume("D:/Music/song.mp3", 5*time.Second, 7)
	cfg.ClearResume()

	if cfg.LastTrackPath != "" {
		t.Errorf("LastTrackPath not cleared: %q", cfg.LastTrackPath)
	}
	if cfg.LastPositionMs != 0 {
		t.Errorf("LastPositionMs not cleared: %d", cfg.LastPositionMs)
	}
	if cfg.LastPlaylistID != 0 {
		t.Errorf("LastPlaylistID not cleared: %d", cfg.LastPlaylistID)
	}
}

func TestResumePosition(t *testing.T) {
	cfg := &config.Config{LastPositionMs: 76200}
	pos := cfg.ResumePosition()
	if pos != 76200*time.Millisecond {
		t.Errorf("ResumePosition: got %v, want 76.2s", pos)
	}
}