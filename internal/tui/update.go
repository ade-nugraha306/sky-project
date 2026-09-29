package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ade-nugraha306/sky-project/internal/config"
	"github.com/ade-nugraha306/sky-project/internal/db"
	"github.com/ade-nugraha306/sky-project/internal/library"
)

type scanDoneMsg struct {
	count int
	err   error
}

type tracksLoadedMsg struct {
	tracks []library.Track
	err    error
}

type playResultMsg struct {
	track library.Track
	err   error
}

func scanCmd(cfg *config.Config, database *db.DB) tea.Cmd {
	return func() tea.Msg {
		for _, folder := range cfg.MusicFolders {
			paths, err := library.ScanFolder(folder)
			if err != nil {
				continue
			}
			for _, p := range paths {
				track := library.ParseMetadata(p)
				if err := database.UpsertTrack(track); err != nil {
					return scanDoneMsg{err: err}
				}
			}
		}
		n, err := database.CountTracks()
		return scanDoneMsg{count: n, err: err}
	}
}

func loadTracksCmd(database *db.DB) tea.Cmd {
	return func() tea.Msg {
		tracks, err := database.AllTracks()
		return tracksLoadedMsg{tracks: tracks, err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)

	case scanDoneMsg:
		if msg.err != nil {
			m.status = "scan gagal: " + msg.err.Error()
			return m, nil
		}
		m.status = "scan selesai, memuat lagu..."
		return m, loadTracksCmd(m.db)

	case tracksLoadedMsg:
		if msg.err != nil {
			m.status = "load gagal: " + msg.err.Error()
			return m, nil
		}
		items := make([]list.Item, len(msg.tracks))
		for i, t := range msg.tracks {
			items[i] = trackItem{track: t}
		}
		m.list.SetItems(items)
		m.status = formatStatus(len(items))
		return m, nil

	case playResultMsg:
		if msg.err != nil {
			m.status = "gagal putar: " + msg.err.Error()
			return m, nil
		}
		m.nowPlaying = &msg.track
		m.status = "▶ " + msg.track.Title
		return m, nil

	case tea.KeyMsg:
		// Jangan intercept key saat user sedang filter.
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "r":
			m.status = "scanning..."
			return m, scanCmd(m.cfg, m.db)

		case "enter":
			if item, ok := m.list.SelectedItem().(trackItem); ok {
				t := item.track
				return m, func() tea.Msg {
					err := m.player.Load(t.Path)
					return playResultMsg{track: t, err: err}
				}
			}

		case " ":
			m.player.TogglePause()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func formatStatus(n int) string {
	if n == 0 {
		return "belum ada lagu • tekan r untuk scan"
	}
	return "siap • " + itoa(n) + " lagu"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}