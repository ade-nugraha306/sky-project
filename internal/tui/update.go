package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ade-nugraha306/sky-project/internal/config"
	"github.com/ade-nugraha306/sky-project/internal/db"
	"github.com/ade-nugraha306/sky-project/internal/library"
	"github.com/ade-nugraha306/sky-project/internal/queue"
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

type playlistsLoadedMsg struct {
	playlists []db.Playlist
	err       error
}

type playlistTracksMsg struct {
	playlistName string
	tracks       []library.Track
	err          error
}

type playlistDetailMsg struct {
	playlistID   int64
	playlistName string
	tracks       []library.Track
	err          error
}

func loadPlaylistsCmd(database *db.DB) tea.Cmd {
	return func() tea.Msg {
		ps, err := database.AllPlaylists()
		return playlistsLoadedMsg{playlists: ps, err: err}
	}
}

func scanCmd(cfg *config.Config, database *db.DB) tea.Cmd {
	return func() tea.Msg {
		var allTracks []library.Track
		for _, folder := range cfg.MusicFolders {
			paths, err := library.ScanFolder(folder)
			if err != nil {
				continue
			}
			for _, p := range paths {
				allTracks = append(allTracks, library.ParseMetadata(p))
			}
		}
		if err := database.UpsertTracks(allTracks); err != nil {
			return scanDoneMsg{err: err}
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
		m.playlistList.SetSize(msg.Width, msg.Height-4)
		m.playlistPicker.SetSize(msg.Width, msg.Height-4)
		m.playlistTrackList.SetSize(msg.Width, msg.Height-6)

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

		// Refresh highlight untuk track yang sedang diputar (misal
		// setelah balik dari playlist detail atau browser).
		m = m.rebuildLibraryList()

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

				// Tentukan sumber pemutaran dari config.
				// Kalau LastPlaylistID > 0, lookup nama dari DB.
				// Kalau playlist sudah dihapus, fallback ke Library.
				if m.cfg.LastPlaylistID > 0 {
					pl, err := m.db.PlaylistByID(m.cfg.LastPlaylistID)
					if err == nil {
						m.playbackSourceID = pl.ID
						m.playbackSourceName = pl.Name
					} else {
						// Playlist hilang — fallback Library.
						m.playbackSourceID = 0
						m.playbackSourceName = "Library"
					}
				} else {
					m.playbackSourceID = 0
					m.playbackSourceName = "Library"
				}

				pos := m.cfg.ResumePosition()
				track := t
				return m, func() tea.Msg {
					err := m.player.LoadAt(track.Path, pos, true)
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

		if m.mode != modePlaylistDetail {
			m.status = formatStatus(len(m.list.Items()))
		}

		// Refresh highlight di list yang sedang aktif.
		if m.mode == modePlaylistDetail {
			m = m.rebuildPlaylistTrackList()
		} else {
			m = m.rebuildLibraryList()
		}

		if msg.resumeAt > 0 {
			dur := m.player.Duration()
			m = m.flash(fmt.Sprintf("⏸ %s · %s / %s",
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

	case playlistsLoadedMsg:
		if msg.err != nil {
			m.status = "gagal load playlist: " + msg.err.Error()
			return m, nil
		}
		items := make([]list.Item, len(msg.playlists))
		for i, p := range msg.playlists {
			items[i] = playlistItem{playlist: p}
		}
		m.playlistList.SetItems(items)
		m.playlistPicker.SetItems(items)
		m.status = fmt.Sprintf("%d playlist • n baru • r rename • d hapus", len(msg.playlists))
		return m, nil

	case playlistDetailMsg:
		if msg.err != nil {
			m.status = "gagal load playlist: " + msg.err.Error()
			return m, nil
		}
		m.currentPlaylistID = msg.playlistID
		m.currentPlaylistName = msg.playlistName

		items := make([]list.Item, len(msg.tracks))
		for i, t := range msg.tracks {
			items[i] = trackItem{track: t}
		}
		m.playlistTrackList.SetItems(items)
		m.playlistTrackList.Title = msg.playlistName

		// Refresh highlight untuk track yang sedang diputar.
		m = m.rebuildPlaylistTrackList()

		m.status = fmt.Sprintf("%s • %d lagu", msg.playlistName, len(msg.tracks))
		return m, nil

	case tea.KeyMsg:
		// Help overlay kalau aktif: hanya beberapa tombol yang di-handle.
		// Tombol lain di-swallow supaya user tidak tidak sengaja
		// mengubah state saat membaca panduan.
		if m.confirmDuplicate {
			switch msg.String() {
			case "y", "Y":
				return m.handleConfirmDuplicateYes()
			case "n", "N", "esc":
				m.confirmDuplicate = false
				m.confirmPlaylistID = 0
				m.confirmPlaylistName = ""
				m.pendingTrack = library.Track{}
				m.mode = modeLibrary
				return m, nil
			}
			return m, nil
		}
		if m.inputMode != inputNone {
			return m.updateInput(msg)
		}
		if m.showHelp {
			switch msg.String() {
			case "ctrl+c":
				m.saveResume()
				return m, tea.Quit
			case "?", "esc", "q":
				m.showHelp = false
				return m, nil
			}
			return m, nil
		}

		if msg.String() == "?" {
			m.showHelp = true
			return m, nil
		}

		if !m.isFiltering() {
			switch msg.String() {
			case "L":
				return m.gotoLibrary()
			case "P":
				return m.gotoPlaylists()
			case "B":
				return m.gotoBrowser()
			case "F":
				return m.gotoFolders()
			case "x":
				return m.clearActiveFilter()
			}
		}

		switch m.mode {
		case modeBrowser:
			return m.updateBrowser(msg)
		case modeFolders:
			return m.updateFolders(msg)
		case modePlaylists:
			return m.updatePlaylists(msg)
		case modePlaylistPicker:
			return m.updatePlaylistPicker(msg)
		case modePlaylistDetail:
			return m.updatePlaylistDetail(msg)
		default:
			return m.updateLibrary(msg)
		}

	case tickMsg:
		if m.toast != "" && time.Now().After(m.toastUntil) {
			m.toast = ""
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
	case modePlaylists:
		m.playlistList, cmd = m.playlistList.Update(msg)
	case modePlaylistPicker:
		m.playlistPicker, cmd = m.playlistPicker.Update(msg)
	case modePlaylistDetail:
		m.playlistTrackList, cmd = m.playlistTrackList.Update(msg)
	default:
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m Model) saveResume() {
	if m.currentTrack != nil {
		if !m.player.IsPaused() {
			m.player.TogglePause()
		}
		m.cfg.SetResume(
			m.currentTrack.Path,
			m.player.Position(),
			m.playbackSourceID,
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
		// Set currentTrack = nil dulu supaya tick berikutnya
		// tidak trigger handleTrackEnd lagi sebelum playResultMsg
		// dari Load ini masuk (race condition ~250ms).
		m.currentTrack = nil
		next, cmd := m.playAt(m.queueIndex)
		return next, tea.Batch(cmd, tickCmd())

	case RepeatAll:
		nextIdx := m.queueIndex + 1
		if nextIdx >= len(m.queue) {
			nextIdx = 0
		}
		m.currentTrack = nil
		next, cmd := m.playAt(nextIdx)
		return next, tea.Batch(cmd, tickCmd())

	default: // RepeatOff
		nextIdx := m.queueIndex + 1
		if nextIdx >= len(m.queue) {
			if m.shuffle && len(m.queue) > 0 {
				m.currentTrack = nil
				m.queue = queue.Shuffle(m.queue)
				m.queueIndex = 0
				next, cmd := m.playAt(0)
				return next, tea.Batch(cmd, tickCmd())
			}
			// Queue habis — reset semua state playback.
			m.currentTrack = nil
			m.playbackSourceID = 0
			m.playbackSourceName = ""
			m = m.flash("queue selesai")
			return m, tickCmd()
		}
		m.currentTrack = nil
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

	case "t":
		item, ok := m.list.SelectedItem().(trackItem)
		if !ok {
			break
		}
		m.pendingTrack = item.track
		m.mode = modePlaylistPicker
		m.status = "pilih playlist untuk '" + item.track.Title + "'"
		return m, loadPlaylistsCmd(m.db)

	case "enter":
		selected, ok := m.list.SelectedItem().(trackItem)
		if !ok {
			break
		}

		// Sumber pemutaran = Library.
		m.playbackSourceID = 0
		m.playbackSourceName = "Library"

		// Bangun queue dari item yang SEDANG TAMPIL (hasil filter).
		visible := m.list.VisibleItems()
		tracks := make([]library.Track, 0, len(visible))
		for _, it := range visible {
			if ti, ok := it.(trackItem); ok {
				tracks = append(tracks, ti.track)
			}
		}

		idx := -1
		for i, t := range tracks {
			if t.Path == selected.track.Path {
				idx = i
				break
			}
		}
		if idx < 0 {
			break
		}

		m.queue = tracks
		m.savedQueue = nil
		// Reset filter setelah play supaya user lihat seluruh library.
		m.list.ResetFilter()

		// Kalau shuffle aktif, acak queue sebelum menentukan index.
		// Queue baru — reset savedQueue lama.
		if m.shuffle {
			m.savedQueue = make([]library.Track, len(tracks))
			copy(m.savedQueue, tracks)
			m.queue = queue.Shuffle(tracks)
			idx = queue.FindIndex(m.queue, selected.track.Path)
			if idx < 0 {
				idx = 0
			}
		}

		return m.playAt(idx)

	case "n", ">":
		next, cmd := m.playNext()
		return next, cmd

	case "p", "<":
		if m.currentTrack == nil {
			break
		}
		// Standar music player: kalau posisi sudah lewat 3 detik,
		// "prev" restart lagu ini. Kalau masih di awal, baru pindah
		// ke track sebelumnya di queue.
		const restartThreshold = 3 * time.Second
		if m.player.Position() > restartThreshold {
			m.player.Restart()
			m = m.flash("⏮ " + m.currentTrack.Title)
			break
		}
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
		m.cfg.RepeatMode = m.repeat.Key()
		m.cfg.Save()
		m = m.flash(m.repeat.String())

	case "+", "=":
		pct := m.player.Volume() + 5
		m.player.SetVolume(pct)
		m.cfg.Volume = m.player.Volume()
		m.cfg.Save()
		m = m.flash(fmt.Sprintf("🔊 %d%%", m.player.Volume()))

	case "-", "_":
		pct := m.player.Volume() - 5
		m.player.SetVolume(pct)
		m.cfg.Volume = m.player.Volume()
		m.cfg.Save()
		m = m.flash(fmt.Sprintf("🔊 %d%%", m.player.Volume()))

	case "s":
		m.shuffle = !m.shuffle
		m.cfg.Shuffle = m.shuffle
		m.cfg.Save()

		if len(m.queue) > 0 && m.currentTrack != nil {
			if m.shuffle {
				// Simpan urutan asli, lalu acak.
				m.savedQueue = make([]library.Track, len(m.queue))
				copy(m.savedQueue, m.queue)
				m.queue = queue.Shuffle(m.queue)

				// Cari posisi current track di queue yang baru.
				idx := queue.FindIndex(m.queue, m.currentTrack.Path)
				if idx >= 0 {
					m.queueIndex = idx
				}
			} else {
				// Restore urutan asli.
				if len(m.savedQueue) == len(m.queue) {
					m.queue = m.savedQueue
					idx := queue.FindIndex(m.queue, m.currentTrack.Path)
					if idx >= 0 {
						m.queueIndex = idx
					}
				}
				m.savedQueue = nil
			}
		} else if !m.shuffle {
			// Kalau di-off saat belum ada queue, bersihkan saved.
			m.savedQueue = nil
		}

		if m.shuffle {
			m = m.flash("🔀 shuffle: on")
		} else {
			m = m.flash("shuffle: off")
		}

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

// updateInput menangani keystroke saat textinput aktif (create/rename playlist).
func (m Model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.inputMode = inputNone
		m.textInput.Blur()
		m.textInput.SetValue("")
		return m, nil

	case "enter":
		name := strings.TrimSpace(m.textInput.Value())
		if name == "" {
			m.status = "nama tidak boleh kosong"
			return m, nil
		}

		switch m.inputMode {
		case inputCreatePlaylist:
			if _, err := m.db.CreatePlaylist(name); err != nil {
				m.status = "gagal buat playlist: " + err.Error()
				return m, nil
			}
			m = m.flash("playlist dibuat: " + name)

		case inputRenamePlaylist:
			if err := m.db.RenamePlaylist(m.editingPlaylistID, name); err != nil {
				m.status = "gagal rename: " + err.Error()
				return m, nil
			}
			m = m.flash("rename: " + name)
		}

		m.inputMode = inputNone
		m.editingPlaylistID = 0
		m.textInput.Blur()
		m.textInput.SetValue("")
		return m, loadPlaylistsCmd(m.db)
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) updatePlaylists(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Filter guard (list playlist punya filter juga).
	if m.playlistList.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.playlistList, cmd = m.playlistList.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c":
		m.saveResume()
		return m, tea.Quit

	case "esc", "q":
		m.mode = modeLibrary
		return m, loadTracksCmd(m.db)

	case "n":
		m.inputMode = inputCreatePlaylist
		m.textInput.Placeholder = "nama playlist baru..."
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, textinput.Blink

	case "r":
		item, ok := m.playlistList.SelectedItem().(playlistItem)
		if !ok {
			return m, nil
		}
		m.inputMode = inputRenamePlaylist
		m.editingPlaylistID = item.playlist.ID
		m.textInput.Placeholder = "nama baru..."
		m.textInput.SetValue(item.playlist.Name)
		m.textInput.Focus()
		return m, textinput.Blink

	case "d", "delete", "backspace":
		item, ok := m.playlistList.SelectedItem().(playlistItem)
		if !ok {
			return m, nil
		}
		if err := m.db.DeletePlaylist(item.playlist.ID); err != nil {
			m.status = "gagal hapus: " + err.Error()
			return m, nil
		}
		m = m.flash("dihapus: " + item.playlist.Name)
		return m, loadPlaylistsCmd(m.db)

	case "enter":
		item, ok := m.playlistList.SelectedItem().(playlistItem)
		if !ok {
			return m, nil
		}
		m.mode = modePlaylistDetail
		m.status = "memuat " + item.playlist.Name + "..."
		return m, loadPlaylistDetailCmd(m.db, item.playlist.ID, item.playlist.Name)
	}

	var cmd tea.Cmd
	m.playlistList, cmd = m.playlistList.Update(msg)
	return m, cmd
}

func (m Model) updatePlaylistPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.playlistPicker.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.playlistPicker, cmd = m.playlistPicker.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c":
		m.saveResume()
		return m, tea.Quit

	case "esc", "q":
		m.mode = modeLibrary
		m.pendingTrack = library.Track{}
		return m, nil

	case "enter":
		item, ok := m.playlistPicker.SelectedItem().(playlistItem)
		if !ok {
			return m, nil
		}
		playlist := item.playlist
		track := m.pendingTrack

		exists, err := m.db.TrackInPlaylist(playlist.ID, track.Path)
		if err != nil {
			m.status = "gagal cek playlist: " + err.Error()
			return m, nil
		}

		if exists {
			// Tampilkan popup konfirmasi.
			m.confirmDuplicate = true
			m.confirmPlaylistID = playlist.ID
			m.confirmPlaylistName = playlist.Name
			return m, nil
		}

		if err := m.db.AddTrackToPlaylist(playlist.ID, track.Path); err != nil {
			m.status = "gagal tambah: " + err.Error()
			return m, nil
		}
		m = m.flash(fmt.Sprintf("+ '%s' → '%s'", track.Title, playlist.Name))
		m.mode = modeLibrary
		m.pendingTrack = library.Track{}
		return m, nil
	}

	var cmd tea.Cmd
	m.playlistPicker, cmd = m.playlistPicker.Update(msg)
	return m, cmd
}

func (m Model) updatePlaylistDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.playlistTrackList.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.playlistTrackList, cmd = m.playlistTrackList.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c":
		m.saveResume()
		return m, tea.Quit

	case "esc", "backspace", "left":
		m.mode = modePlaylists
		m.currentPlaylistID = 0
		m.currentPlaylistName = ""
		m.status = "memuat playlist..."
		return m, loadPlaylistsCmd(m.db)

	case "enter":
		selected, ok := m.playlistTrackList.SelectedItem().(trackItem)
		if !ok {
			break
		}

		// Sumber pemutaran = playlist ini.
		m.playbackSourceID = m.currentPlaylistID
		m.playbackSourceName = m.currentPlaylistName
		tracks := m.playlistTracksSnapshot()
		idx := -1
		for i, t := range tracks {
			if t.Path == selected.track.Path {
				idx = i
				break
			}
		}
		if idx < 0 {
			break
		}

		m.queue = tracks
		m.savedQueue = nil
		if m.shuffle {
			m.savedQueue = make([]library.Track, len(tracks))
			copy(m.savedQueue, tracks)
			m.queue = queue.Shuffle(tracks)
			idx = queue.FindIndex(m.queue, selected.track.Path)
			if idx < 0 {
				idx = 0
			}
		}
		return m.playAt(idx)

	case "n", ">":
		next, cmd := m.playNext()
		return next, cmd

	case "p", "<":
		if m.currentTrack == nil {
			break
		}
		const restartThreshold = 3 * time.Second
		if m.player.Position() > restartThreshold {
			m.player.Restart()
			m = m.flash("⏮ " + m.currentTrack.Title)
			break
		}
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

	case " ":
		if m.currentTrack == nil {
			break
		}
		m.player.TogglePause()
		if m.player.IsPaused() {
			m = m.flash("⏸ paused")
		} else {
			m = m.flash("▶ playing")
		}

	case "m":
		m.repeat = (m.repeat + 1) % 3
		m.cfg.RepeatMode = m.repeat.Key()
		m.cfg.Save()
		m = m.flash(m.repeat.String())

	case "d", "delete":
		idx := m.playlistTrackList.Index()
		if idx < 0 {
			break
		}
		if err := m.db.RemoveTrackFromPlaylistAt(m.currentPlaylistID, idx); err != nil {
			m.status = "gagal hapus track: " + err.Error()
			return m, nil
		}
		m = m.flash("track dihapus dari playlist")
		return m, loadPlaylistDetailCmd(m.db, m.currentPlaylistID, m.currentPlaylistName)
	}

	var cmd tea.Cmd
	m.playlistTrackList, cmd = m.playlistTrackList.Update(msg)
	return m, cmd
}

// playlistTracksSnapshot mengembalikan track-track di playlistTrackList
// sebagai slice library.Track. Dipakai untuk membangun queue.
func (m Model) playlistTracksSnapshot() []library.Track {
	items := m.playlistTrackList.Items()
	tracks := make([]library.Track, 0, len(items))
	for _, it := range items {
		if ti, ok := it.(trackItem); ok {
			tracks = append(tracks, ti.track)
		}
	}
	return tracks
}

// handleConfirmDuplicateYes menangani user menekan 'y' di popup konfirmasi.
// Menambahkan track sebagai duplikat, lalu kembali ke library.
func (m Model) handleConfirmDuplicateYes() (tea.Model, tea.Cmd) {
	track := m.pendingTrack
	playlistID := m.confirmPlaylistID
	playlistName := m.confirmPlaylistName

	// Reset state popup dulu supaya tidak ada race kalau error.
	m.confirmDuplicate = false
	m.confirmPlaylistID = 0
	m.confirmPlaylistName = ""
	m.pendingTrack = library.Track{}
	m.mode = modeLibrary

	if err := m.db.AddTrackToPlaylist(playlistID, track.Path); err != nil {
		m.status = "gagal tambah: " + err.Error()
		return m, nil
	}
	m = m.flash(fmt.Sprintf("+ '%s' (duplikat) → '%s'", track.Title, playlistName))
	return m, nil
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
	next := m.queueIndex + 1
	if next >= len(m.queue) {
		// Queue sudah di akhir.
		if m.shuffle && len(m.queue) > 0 {
			// Shuffle aktif: reshuffle dan main dari awal.
			// savedQueue tidak disentuh — tetap urutan asli,
			// jadi toggle off nanti tetap restore dengan benar.
			m.queue = queue.Shuffle(m.queue)
			m.queueIndex = 0
			return m.playAt(0)
		}
		// Shuffle off: tidak ada aksi.
		return m, nil
	}
	return m.playAt(next)
}

func (m Model) playPrev() (Model, tea.Cmd) {
	prev := m.queueIndex - 1
	if prev < 0 {
		return m, nil
	}
	return m.playAt(prev)
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

// isFiltering mengembalikan true kalau list di mode aktif sedang
// dalam keadaan filter. Dipakai untuk mencegah hotkey global
// (L/P/B/F) di-intercept saat user sedang mengetik filter.
func (m Model) isFiltering() bool {
	switch m.mode {
	case modeLibrary:
		return m.list.FilterState() == list.Filtering
	case modeBrowser:
		return m.browser.FilterState() == list.Filtering
	case modePlaylists:
		return m.playlistList.FilterState() == list.Filtering
	case modePlaylistPicker:
		return m.playlistPicker.FilterState() == list.Filtering
	case modePlaylistDetail:
		return m.playlistTrackList.FilterState() == list.Filtering
	}
	return false
}

func (m Model) gotoLibrary() (tea.Model, tea.Cmd) {
	if m.mode == modeLibrary {
		return m, nil
	}
	m.mode = modeLibrary
	return m, loadTracksCmd(m.db)
}

func (m Model) clearActiveFilter() (tea.Model, tea.Cmd) {
	hadFilter := false

	switch m.mode {
	case modeLibrary:
		if m.list.FilterState() != list.Unfiltered {
			m.list.ResetFilter()
			m = m.rebuildLibraryList()
			hadFilter = true
		}
	case modeBrowser:
		if m.browser.FilterState() != list.Unfiltered {
			m.browser.ResetFilter()
			hadFilter = true
		}
	case modePlaylists:
		if m.playlistList.FilterState() != list.Unfiltered {
			m.playlistList.ResetFilter()
			hadFilter = true
		}
	case modePlaylistPicker:
		if m.playlistPicker.FilterState() != list.Unfiltered {
			m.playlistPicker.ResetFilter()
			hadFilter = true
		}
	case modePlaylistDetail:
		if m.playlistTrackList.FilterState() != list.Unfiltered {
			m.playlistTrackList.ResetFilter()
			m = m.rebuildPlaylistTrackList()
			hadFilter = true
		}
	}

	if hadFilter {
		m = m.flash("filter dibersihkan")
	}
	return m, nil
}


func (m Model) gotoPlaylists() (tea.Model, tea.Cmd) {
	if m.mode == modePlaylists {
		return m, nil
	}
	// Reset detail state kalau lagi di detail.
	m.currentPlaylistID = 0
	m.currentPlaylistName = ""
	m.mode = modePlaylists
	m.status = "memuat playlist..."
	return m, loadPlaylistsCmd(m.db)
}

func (m Model) gotoBrowser() (tea.Model, tea.Cmd) {
	if m.mode == modeBrowser {
		return m, nil
	}
	m.mode = modeBrowser
	start := m.browserPath
	if start == "" {
		start = homeDir()
	}
	m.status = "memuat " + start + "..."
	return m, loadDirCmd(start)
}

func (m Model) gotoFolders() (tea.Model, tea.Cmd) {
	if m.mode == modeFolders {
		return m, nil
	}
	m.mode = modeFolders
	items := make([]list.Item, 0, len(m.cfg.MusicFolders))
	for _, p := range m.cfg.MusicFolders {
		items = append(items, folderItem{path: p})
	}
	m.folderList.SetItems(items)
	return m, nil
}

func loadPlaylistDetailCmd(database *db.DB, id int64, name string) tea.Cmd {
	return func() tea.Msg {
		tracks, err := database.TracksInPlaylist(id)
		return playlistDetailMsg{
			playlistID:   id,
			playlistName: name,
			tracks:       tracks,
			err:          err,
		}
	}
}

func formatStatus(n int) string {
	if n == 0 {
		return "belum ada lagu • a tambah folder • r rescan"
	}
	return fmt.Sprintf("siap • %d lagu • a tambah folder • d kelola folder • r rescan", n)
}
