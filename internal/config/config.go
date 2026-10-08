package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	MusicFolders   []string `json:"music_folders"`
	Volume         int      `json:"volume"`
	LastTrackPath  string   `json:"last_track_path,omitempty"`
	LastPositionMs int64    `json:"last_position_ms,omitempty"`
	RepeatMode     string   `json:"repeat_mode,omitempty"`
	Shuffle        bool     `json:"shuffle,omitempty"`
	LastPlaylistID int64    `json:"last_playlist_id,omitempty"`
	LastActiveFolder string   `json:"last_active_folder,omitempty"`
	SortMode         string   `json:"sort_mode,omitempty"`
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sky")
}

func DefaultConfigPath() string {
	return filepath.Join(configDir(), "config.json")
}

func DefaultDBPath() string {
	return filepath.Join(configDir(), "sky.db")
}

func normalizePath(p string) string {
	if p == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(p))
}

func Load() (*Config, error) {
	path := DefaultConfigPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := &Config{
			MusicFolders: []string{},
			Volume:       100,
			RepeatMode:   "off",
			SortMode:     "title",
		}
		if err := cfg.Save(); err != nil {
			return nil, err
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf(
			"config rusak di %s: %w\n"+
				"petunjuk: di JSON, backslash harus di-escape (\\\\) atau "+
				"pakai forward slash (/)",
			path, err,
		)
	}
	for i, p := range cfg.MusicFolders {
		cfg.MusicFolders[i] = normalizePath(p)
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	cleaned := make([]string, len(c.MusicFolders))
	for i, p := range c.MusicFolders {
		cleaned[i] = normalizePath(p)
	}
	out := Config{
		MusicFolders:   cleaned,
		Volume:         c.Volume,
		LastTrackPath:  c.LastTrackPath,
		LastPositionMs: c.LastPositionMs,
		RepeatMode:     c.RepeatMode,
		Shuffle:        c.Shuffle,
		LastPlaylistID: c.LastPlaylistID,
		LastActiveFolder: normalizePath(c.LastActiveFolder),
		SortMode:         c.SortMode,
	}
	path := DefaultConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (c *Config) AddFolder(p string) error {
	p = normalizePath(p)
	for _, f := range c.MusicFolders {
		if f == p {
			return nil
		}
	}
	c.MusicFolders = append(c.MusicFolders, p)
	return c.Save()
}

func (c *Config) RemoveFolder(p string) error {
	p = normalizePath(p)
	out := c.MusicFolders[:0]
	for _, f := range c.MusicFolders {
		if f != p {
			out = append(out, f)
		}
	}
	c.MusicFolders = out
	return c.Save()
}

// SetResume menyimpan track dan posisi terakhir. Path dinormalisasi
// jadi forward slash supaya JSON-nya valid tanpa escaping.
func (c *Config) SetResume(trackPath string, pos time.Duration, playlistID int64) {
	c.LastTrackPath = filepath.ToSlash(trackPath)
	c.LastPositionMs = pos.Milliseconds()
	c.LastPlaylistID = playlistID
}

func (c *Config) ClearResume() {
	c.LastTrackPath = ""
	c.LastPositionMs = 0
	c.LastPlaylistID = 0
}

func (c *Config) ResumePosition() time.Duration {
	return time.Duration(c.LastPositionMs) * time.Millisecond
}
