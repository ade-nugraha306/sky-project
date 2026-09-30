package tui

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ade-nugraha306/sky-project/internal/config"
	"github.com/ade-nugraha306/sky-project/internal/db"
	"github.com/ade-nugraha306/sky-project/internal/library"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

const statusFlashDuration = 3 * time.Second

func (m Model) flash(s string) Model {
	m.toast = s
	m.toastUntil = time.Now().Add(statusFlashDuration)
	return m
}

type scanDoneMsg struct {
	count int
	err   error
}

type tracksLoadedMsg struct {
	tracks []library.Track
	err    error
}

type playResultMsg struct {
	track    library.Track
	resumeAt time.Duration
	err      error
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

		listWidth := msg.Width * 40 / 100
		if listWidth < 20 {
			listWidth = 20
		}
		if listWidth > 60 {
			listWidth = 60
		}
		m.browser.SetSize(listWidth, msg.Height-4)
		m.folderList.SetSize(msg.Width, msg.Height-4)

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

		// Set status dasar lebih dulu, supaya kalau resume trigger,
		// kita sudah punya teks fallback yang benar saat toast hilang.
		m.status = formatStatus(len(items))

		// Resume: hanya sekali per sesi.
		if !m.resumeChecked && m.cfg.LastTrackPath != "" && m.currentTrack == nil {
			m.resumeChecked = true

			tracks := make([]library.Track, len(msg.tracks))
			copy(tracks, msg.tracks)

			for i, t := range tracks {
				if filepath.ToSlash(t.Path) != m.cfg.LastTrackPath {
					continue
				}
				m.queue = tracks
				m.queueIndex = i
				pos := m.cfg.ResumePosition()
				paused := m.cfg.LastPaused
				track := t
				return m, func() tea.Msg {
					err := m.player.LoadAt(track.Path, pos, paused)
					return playResultMsg{track: track, resumeAt: pos, err: err}
				}
			}
			// Track tidak ditemukan — bersihkan resume lama.
			m.cfg.ClearResume()
			m.cfg.Save()
		}
		return m, nil

	case playResultMsg:
		if msg.err != nil {
			m.status = "gagal putar: " + msg.err.Error()
			return m, nil
		}
		t := msg.track
		m.currentTrack = &t

		// Reset status dasar — supaya kalau sebelumnya "queue selesai"
		// atau pesan permanen lain, kembali ke status library yang benar
		// setelah toast hilang.
		m.status = formatStatus(len(m.list.Items()))

		if msg.resumeAt > 0 {
			dur := m.player.Duration()
			icon := "▶"
			if m.player.IsPaused() {
				icon = "⏸"
			}
			m = m.flash(fmt.Sprintf("%s %s · %s / %s",
				icon,
				t.Title,
				formatDuration(msg.resumeAt),
				formatDuration(dur),
			))
		} else {
			m = m.flash("▶ " + t.Title)
		}
		return m, nil

	case dirLoadedMsg:
		if msg.err != nil {
			m.status = "gagal buka: " + msg.err.Error()
			return m, nil
		}
		m.browserPath = msg.path
		m.browser.SetItems(msg.items)
		m.status = msg.path

		if item, ok := m.browser.SelectedItem().(dirItem); ok {
			return m, loadPreviewCmd(item.path)
		}
		return m, nil

	case previewLoadedMsg:
		if msg.err != nil {
			m.preview = nil
			m.previewFor = msg.path
			m.previewTotal = 0
			m.previewErr = msg.err.Error()
			return m, nil
		}
		m.preview = msg.entries
		m.previewFor = msg.path
		m.previewTotal = msg.total
		m.previewErr = ""
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case modeBrowser:
			return m.updateBrowser(msg)
		case modeFolders:
			return m.updateFolders(msg)
		default:
			return m.updateLibrary(msg)
		}

	case tickMsg:
		if m.toast != "" && time.Now().After(m.toastUntil) {
			m.toast = ""
		}

		if m.mode != modeLibrary {
			return m, tickCmd()
		}
		if m.currentTrack != nil && m.player.IsFinished() {
			return m.handleTrackEnd()
		}
		return m, tickCmd()
	}

	var cmd tea.Cmd
	switch m.mode {
	case modeBrowser:
		m.browser, cmd = m.browser.Update(msg)
	case modeFolders:
		m.folderList, cmd = m.folderList.Update(msg)
	default:
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

// saveResume menyimpan posisi playback sekarang ke config.
func (m Model) saveResume() {
	if m.currentTrack != nil {
		m.cfg.SetResume(
			m.currentTrack.Path,
			m.player.Position(),
			m.player.IsPaused(),
		)
	} else {
		m.cfg.ClearResume()
	}
	m.cfg.Save()
}

// handleTrackEnd dipanggil saat lagu selesai secara alami (bukan
// karena user menekan n/p). Perilakunya tergantung m.repeat.
func (m Model) handleTrackEnd() (tea.Model, tea.Cmd) {
	switch m.repeat {
	case RepeatOne:
		next, cmd := m.playAt(m.queueIndex)
		return next, tea.Batch(cmd, tickCmd())

	case RepeatAll:
		nextIdx := m.queueIndex + 1
		if nextIdx >= len(m.queue) {
			nextIdx = 0
		}
		next, cmd := m.playAt(nextIdx)
		return next, tea.Batch(cmd, tickCmd())

	default: // RepeatOff
		nextIdx := m.queueIndex + 1
		if nextIdx >= len(m.queue) {
			m.currentTrack = nil
			m = m.flash("queue selesai")
			return m, tickCmd()
		}
		next, cmd := m.playAt(nextIdx)
		return next, tea.Batch(cmd, tickCmd())
	}
}

func (m Model) updateLibrary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "q", "ctrl+c", "esc":
		m.saveResume()
		return m, tea.Quit

	case "r":
		m.status = "scanning..."
		return m, scanCmd(m.cfg, m.db)

	case "a":
		m.mode = modeBrowser
		start := m.browserPath
		if start == "" {
			start = homeDir()
		}
		m.status = "memuat " + start + "..."
		return m, loadDirCmd(start)

	case "d":
		m.mode = modeFolders
		items := make([]list.Item, 0, len(m.cfg.MusicFolders))
		for _, p := range m.cfg.MusicFolders {
			items = append(items, folderItem{path: p})
		}
		m.folderList.SetItems(items)
		return m, nil

	case "enter":
		if _, ok := m.list.SelectedItem().(trackItem); ok {
			tracks := make([]library.Track, 0, len(m.list.Items()))
			for _, it := range m.list.Items() {
				if ti, ok := it.(trackItem); ok {
					tracks = append(tracks, ti.track)
				}
			}
			idx := m.list.Index()
			m.queue = tracks
			return m.playAt(idx)
		}

	case "n", ">":
		next, cmd := m.playNext()
		return next, cmd

	case "p", "<":
		prev, cmd := m.playPrev()
		return prev, cmd

	case ",", "[":
		if m.currentTrack == nil {
			break
		}
		m.player.Seek(-5 * time.Second)
		m = m.flash("⏪ " + formatDuration(m.player.Position()))

	case ".", "]":
		if m.currentTrack == nil {
			break
		}
		m.player.Seek(5 * time.Second)
		m = m.flash("⏩ " + formatDuration(m.player.Position()))

	case "m":
		m.repeat = (m.repeat + 1) % 3
		m = m.flash(m.repeat.String())

	case " ":
		m.player.TogglePause()
		if m.player.IsPaused() {
			m = m.flash("⏸ paused")
		} else {
			m = m.flash("▶ playing")
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.browser.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.browser, cmd = m.browser.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c":
		m.saveResume()
		return m, tea.Quit

	case "esc":
		m.mode = modeLibrary
		return m, loadTracksCmd(m.db)

	case "s":
		if err := m.cfg.AddFolder(m.browserPath); err != nil {
			m.status = "gagal simpan: " + err.Error()
			return m, nil
		}
		m.mode = modeLibrary
		m.status = "menambahkan " + m.browserPath + "..."
		return m, scanCmd(m.cfg, m.db)

	case "backspace", "left":
		parent := parentDir(m.browserPath)
		if parent == "" {
			return m, nil
		}
		m.status = "memuat " + parent + "..."
		return m, loadDirCmd(parent)

	case "enter":
		if item, ok := m.browser.SelectedItem().(dirItem); ok {
			m.status = "memuat " + item.path + "..."
			return m, loadDirCmd(item.path)
		}
	}

	beforeIdx := m.browser.Index()
	var cmd tea.Cmd
	m.browser, cmd = m.browser.Update(msg)

	if m.browser.Index() != beforeIdx {
		if item, ok := m.browser.SelectedItem().(dirItem); ok {
			previewCmd := loadPreviewCmd(item.path)
			return m, tea.Batch(cmd, previewCmd)
		}
	}
	return m, cmd
}

func (m Model) updateFolders(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.saveResume()
		return m, tea.Quit

	case "esc", "q":
		m.mode = modeLibrary
		return m, loadTracksCmd(m.db)

	case "d", "delete", "backspace":
		item, ok := m.folderList.SelectedItem().(folderItem)
		if !ok {
			return m, nil
		}

		if err := m.cfg.RemoveFolder(item.path); err != nil {
			m.status = "gagal hapus folder: " + err.Error()
			return m, nil
		}

		n, err := m.db.DeleteTracksUnderFolder(item.path)
		if err != nil {
			m.status = "folder dihapus, tapi gagal bersihkan lagu: " + err.Error()
		} else {
			m.status = fmt.Sprintf("dihapus: %s (%d lagu dibuang)", item.path, n)
		}

		// Rebuild folder list.
		items := make([]list.Item, 0, len(m.cfg.MusicFolders))
		for _, p := range m.cfg.MusicFolders {
			items = append(items, folderItem{path: p})
		}
		m.folderList.SetItems(items)
		return m, nil
	}

	var cmd tea.Cmd
	m.folderList, cmd = m.folderList.Update(msg)
	return m, cmd
}

func (m Model) playAt(index int) (Model, tea.Cmd) {
	if index < 0 || index >= len(m.queue) {
		return m, nil
	}
	m.queueIndex = index
	t := m.queue[index]
	return m, func() tea.Msg {
		err := m.player.Load(t.Path)
		return playResultMsg{track: t, err: err}
	}
}

func (m Model) playNext() (Model, tea.Cmd) {
	return m.playAt(m.queueIndex + 1)
}

func (m Model) playPrev() (Model, tea.Cmd) {
	return m.playAt(m.queueIndex - 1)
}

func parentDir(path string) string {
	if path == "" {
		return ""
	}
	parent := filepath.Dir(path)
	if parent == path {
		return ""
	}
	return parent
}

func formatStatus(n int) string {
	if n == 0 {
		return "belum ada lagu • a tambah folder • r rescan"
	}
	return fmt.Sprintf("siap • %d lagu • a tambah folder • d kelola folder • r rescan", n)
}