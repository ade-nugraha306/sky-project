package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	MusicFolders []string `json:"music_folders"`
	Volume       int      `json:"volume"`
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

// normalizePath membersihkan path dan mengubah semua separator
// jadi forward slash, supaya JSON tidak perlu escaping.
func normalizePath(p string) string {
	return filepath.ToSlash(filepath.Clean(p))
}

func Load() (*Config, error) {
	path := DefaultConfigPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := &Config{MusicFolders: []string{}, Volume: 100}
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

	// Normalisasi semua path setelah baca.
	for i, p := range cfg.MusicFolders {
		cfg.MusicFolders[i] = normalizePath(p)
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	// Normalisasi sebelum tulis, supaya file JSON selalu bersih.
	cleaned := make([]string, len(c.MusicFolders))
	for i, p := range c.MusicFolders {
		cleaned[i] = normalizePath(p)
	}

	// Salinan sementara supaya tidak memutasi struct asli user.
	out := Config{
		MusicFolders: cleaned,
		Volume:       c.Volume,
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

// AddFolder menambahkan folder (kalau belum ada) dan simpan.
func (c *Config) AddFolder(p string) error {
	p = normalizePath(p)
	for _, f := range c.MusicFolders {
		if f == p {
			return nil // sudah ada
		}
	}
	c.MusicFolders = append(c.MusicFolders, p)
	return c.Save()
}

// RemoveFolder menghapus folder dari daftar.
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