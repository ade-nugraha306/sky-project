package tui

import (
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

		// Browser pakai ~40% lebar; sisanya untuk preview.
		listWidth := msg.Width * 40 / 100
		if listWidth < 20 {
			listWidth = 20
		}
		if listWidth > 60 {
			listWidth = 60
		}
		m.browser.SetSize(listWidth, msg.Height-4)

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
		t := msg.track
		m.currentTrack = &t
		m.status = "▶ " + t.Title
		return m, nil

	case dirLoadedMsg:
		if msg.err != nil {
			m.status = "gagal buka: " + msg.err.Error()
			return m, nil
		}
		m.browserPath = msg.path
		m.browser.SetItems(msg.items)
		m.status = msg.path

		// Setelah SetItems, cursor reset ke index 0.
		// Trigger preview untuk folder pertama.
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
		if m.mode == modeBrowser {
			return m.updateBrowser(msg)
		}
		return m.updateLibrary(msg)

	case tickMsg:
		// Kalau mode browser, tidak perlu update progress — hemat CPU.
		if m.mode == modeBrowser {
			return m, tickCmd()
		}

		// Cek apakah lagu selesai → auto-next.
		if m.currentTrack != nil && m.player.IsFinished() {
			next, cmd := m.playNext()
			if cmd == nil {
				// Queue habis — stop.
				m.currentTrack = nil
				m.status = "queue selesai"
				return m, tickCmd()
			}
			return next, tea.Batch(cmd, tickCmd())
		}
		return m, tickCmd()
	}
	// Teruskan ke komponen yang sedang aktif.
	var cmd tea.Cmd
	if m.mode == modeBrowser {
		m.browser, cmd = m.browser.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m Model) updateLibrary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.list.FilterState() == list.Filtering {
		// Biarkan list yang proses.
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "q", "ctrl+c":
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

	case "enter":
		if _, ok := m.list.SelectedItem().(trackItem); ok {
			// Set seluruh library jadi queue, mulai dari index yang disorot.
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

	case " ":
		m.player.TogglePause()
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
		return m, tea.Quit

	case "esc":
		m.mode = modeLibrary
		return m, loadTracksCmd(m.db)

	case "s":
		// Tambahkan folder yang sedang dibuka ke config.
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

	// Kalau cursor berpindah, refresh preview.
	if m.browser.Index() != beforeIdx {
		if item, ok := m.browser.SelectedItem().(dirItem); ok {
			previewCmd := loadPreviewCmd(item.path)
			return m, tea.Batch(cmd, previewCmd)
		}
	}
	return m, cmd
}

// playAt mengubah queueIndex dan mulai memutar track di index tsb.
// Aman dipanggil kalau index di luar batas — tidak ada aksi.
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
		return "" // sudah di root
	}
	return parent
}

func formatStatus(n int) string {
	if n == 0 {
		return "belum ada lagu • tekan a untuk tambah folder • r untuk rescan"
	}
	return "siap • " + itoa(n) + " lagu • a tambah folder • r rescan"
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