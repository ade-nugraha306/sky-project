package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ade-nugraha306/sky-project/internal/config"
	"github.com/ade-nugraha306/sky-project/internal/db"
	"github.com/ade-nugraha306/sky-project/internal/library"
	"github.com/ade-nugraha306/sky-project/internal/player"
)

type inputMode int

const (
	inputNone inputMode = iota
	inputCreatePlaylist
	inputRenamePlaylist
)

type confirmKind int

const (
	confirmNone confirmKind = iota
	confirmDuplicate
	confirmDeleteFolder
	confirmDeletePlaylist
)

type mode int

const (
	modeLibrary mode = iota
	modeBrowser
	modeFolders
	modePlaylists
	modePlaylistPicker
	modePlaylistDetail
)

type Model struct {
	cfg         *config.Config
	db          *db.DB
	player      *player.Player
	mode        mode
	repeat      RepeatMode
	shuffle     bool
	list        list.Model
	browser     list.Model
	browserPath string
	folderList  list.Model
	playlistList   list.Model
	playlistPicker list.Model

	playlistTrackList   list.Model
	currentPlaylistID   int64
	currentPlaylistName string

	playbackSourceID   int64
	playbackSourceName string

	// Track yang sedang menunggu ditambahkan ke playlist.
	pendingTrack library.Track

	// Overlay konfirmasi duplikat.
	confirmKind         confirmKind
	confirmFolderPath   string
	confirmPlaylistID   int64
	confirmPlaylistName string

	// Textinput untuk create/rename playlist.
	textInput         textinput.Model
	inputMode         inputMode
	editingPlaylistID int64

	preview      []previewEntry
	previewFor   string
	previewTotal int
	previewErr   string

	queue        []library.Track
	queueIndex   int
	currentTrack *library.Track
	savedQueue   []library.Track

	resumeChecked bool

	toast      string
	toastUntil time.Time

	showHelp bool

	width  int
	height int
	status string
}


// RepeatMode mengatur perilaku auto-next saat lagu selesai.
type RepeatMode int

const (
	RepeatOff RepeatMode = iota // lagu terakhir selesai → stop
	RepeatOne                   // lagu yang sama diulang terus
	RepeatAll                   // queue berputar, kembali ke index 0
)

func (r RepeatMode) String() string {
	switch r {
	case RepeatOne:
		return "repeat: one"
	case RepeatAll:
		return "repeat: all"
	default:
		return "repeat: off"
	}
}

// Key mengembalikan representasi singkat untuk persistensi.
// Berbeda dari String() yang untuk display.
func (r RepeatMode) Key() string {
	switch r {
	case RepeatOne:
		return "one"
	case RepeatAll:
		return "all"
	default:
		return "off"
	}
}

func parseRepeatMode(s string) RepeatMode {
	switch s {
	case "one":
		return RepeatOne
	case "all":
		return RepeatAll
	default:
		return RepeatOff
	}
}


func NewModel(cfg *config.Config, database *db.DB) Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Sky — Library"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	b := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	b.Title = "Pilih Folder"
	b.SetShowStatusBar(false)
	b.SetFilteringEnabled(true)

	f := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	f.Title = "Kelola Folder Musik"
	f.SetShowStatusBar(false)
	f.SetFilteringEnabled(false)

	pl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	pl.Title = "Playlists"
	pl.SetShowStatusBar(false)
	pl.SetFilteringEnabled(true)

	pp := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	pp.Title = "Pilih Playlist"
	pp.SetShowStatusBar(false)
	pp.SetFilteringEnabled(true)

	plDetail := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	plDetail.Title = "Playlist"
	plDetail.SetShowStatusBar(false)
	plDetail.SetFilteringEnabled(true)

	ti := textinput.New()
	ti.Placeholder = "nama playlist..."
	ti.CharLimit = 100
	ti.Width = 50

	playerInst := player.New()
	playerInst.SetVolume(cfg.Volume)

	return Model{
		cfg:          cfg,
		db:           database,
		player:       playerInst,
		mode:         modeLibrary,
		repeat:       parseRepeatMode(cfg.RepeatMode),
		shuffle:      cfg.Shuffle,
		list:         l,
		browser:      b,
		folderList:   f,
		playlistList: pl,
		playlistPicker: pp,
		playlistTrackList: plDetail,
		textInput:    ti,
		status:       "memuat...",
	}
}

func (m Model) rebuildPlaylistTrackList() Model {
	items := m.playlistTrackList.Items()
	if len(items) == 0 {
		return m
	}
	cursor := m.playlistTrackList.Index()

	newItems := make([]list.Item, len(items))
	for i, it := range items {
		if ti, ok := it.(trackItem); ok {
			// Prefix hanya kalau sumbernya playlist INI.
			ti.playing = m.currentTrack != nil &&
				ti.track.Path == m.currentTrack.Path &&
				m.playbackSourceID != 0 &&
				m.playbackSourceID == m.currentPlaylistID
			newItems[i] = ti
		} else {
			newItems[i] = it
		}
	}
	m.playlistTrackList.SetItems(newItems)
	if cursor >= 0 && cursor < len(newItems) {
		m.playlistTrackList.Select(cursor)
	}
	return m
}

func (m Model) rebuildLibraryList() Model {
	if m.list.FilterState() != list.Unfiltered {
		return m
	}
	items := m.list.Items()
	if len(items) == 0 {
		return m
	}
	cursor := m.list.Index()

	newItems := make([]list.Item, len(items))
	for i, it := range items {
		if ti, ok := it.(trackItem); ok {
			// Prefix hanya muncul kalau track ini sedang diputar
			// DAN sumber pemutaran adalah library (bukan playlist).
			ti.playing = m.currentTrack != nil &&
				ti.track.Path == m.currentTrack.Path &&
				m.playbackSourceID == 0
			newItems[i] = ti
		} else {
			newItems[i] = it
		}
	}
	m.list.SetItems(newItems)
	if cursor >= 0 && cursor < len(newItems) {
		m.list.Select(cursor)
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(scanCmd(m.cfg, m.db), tickCmd())
}