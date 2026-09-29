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

type Model struct {
	cfg         *config.Config
	db          *db.DB
	player      *player.Player
	mode        mode
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

	// Toast: pesan sementara yang otomatis hilang setelah beberapa detik.
	// Kalau kosong, `status` yang ditampilkan.
	toast      string
	toastUntil time.Time

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

	return Model{
		cfg:        cfg,
		db:         database,
		player:     player.New(),
		mode:       modeLibrary,
		list:       l,
		browser:    b,
		folderList: f,
		status:     "memuat...",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(scanCmd(m.cfg, m.db), tickCmd())
}