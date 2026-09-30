package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ade-nugraha306/sky-project/internal/config"
	"github.com/ade-nugraha306/sky-project/internal/db"
	"github.com/ade-nugraha306/sky-project/internal/library"
	"github.com/ade-nugraha306/sky-project/internal/player"
)

type mode int

const (
	modeLibrary mode = iota
	modeBrowser
	modeFolders
)

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

type Model struct {
	cfg         *config.Config
	db          *db.DB
	player      *player.Player
	mode        mode
	repeat      RepeatMode
	list        list.Model
	browser     list.Model
	browserPath string
	folderList  list.Model

	preview      []previewEntry
	previewFor   string
	previewTotal int
	previewErr   string

	queue        []library.Track
	queueIndex   int
	currentTrack *library.Track

	resumeChecked bool

	toast      string
	toastUntil time.Time

	showHelp bool // overlay panduan hotkey

	width  int
	height int
	status string
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

	pl := player.New()
	pl.SetVolume(cfg.Volume)

	return Model{
		cfg:        cfg,
		db:         database,
		player:     pl,
		mode:       modeLibrary,
		repeat:     parseRepeatMode(cfg.RepeatMode),
		list:       l,
		browser:    b,
		folderList: f,
		status:     "memuat...",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(scanCmd(m.cfg, m.db), tickCmd())
}