package library

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
)

func ParseMetadata(path string) Track {
	t := Track{
		Path:  path,
		Title: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
	}
	f, err := os.Open(path)
	if err != nil {
		return t
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return t
	}
	if v := m.Title(); v != "" {
		t.Title = v
	}
	t.Artist = m.Artist()
	t.Album = m.Album()
	return t
}