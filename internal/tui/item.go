package tui

import (
	"path/filepath"
	"fmt"

	"github.com/ade-nugraha306/sky-project/internal/db"
	"github.com/ade-nugraha306/sky-project/internal/library"
)

type trackItem struct {
	track   library.Track
	playing bool
}

func (i trackItem) Title() string {
	if i.playing {
		return "▶ " + i.track.Title
	}
	return "  " + i.track.Title
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

type folderItem struct {
	path string
}

type playlistItem struct {
	playlist db.Playlist
}

func (p playlistItem) Title() string { return p.playlist.Name }
func (p playlistItem) Description() string {
	if p.playlist.TrackCount == 0 {
		return "kosong"
	}
	if p.playlist.TrackCount == 1 {
		return "1 lagu"
	}
	return fmt.Sprintf("%d lagu", p.playlist.TrackCount)
}
func (p playlistItem) FilterValue() string { return p.playlist.Name }

func (f folderItem) Title() string       { return f.path }
func (f folderItem) Description() string { return "" }
func (f folderItem) FilterValue() string { return f.path }