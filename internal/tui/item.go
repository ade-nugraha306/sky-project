package tui

import (
	"path/filepath"

	"github.com/ade-nugraha306/sky-project/internal/library"
)

type trackItem struct {
	track library.Track
}

func (i trackItem) Title() string {
	return i.track.Title
}

func (i trackItem) Description() string {
	desc := i.track.Artist
	if i.track.Album != "" {
		if desc != "" {
			desc += " — "
		}
		desc += i.track.Album
	}
	if desc == "" {
		desc = filepath.Base(i.track.Path)
	}
	return desc
}

func (i trackItem) FilterValue() string {
	return i.track.Title + " " + i.track.Artist + " " + i.track.Album
}