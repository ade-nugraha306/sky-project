package tui

import (
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
)

type Model struct {
	cfg         *config.Config
	db          *db.DB
	player      *player.Player
	mode        mode
	list        list.Model
	browser     list.Model
	browserPath string

	// Preview
	preview      []previewEntry
	previewFor   string
	previewTotal int
	previewErr   string

	// Queue & playback
	queue        []library.Track
	queueIndex   int
	currentTrack *library.Track

	width      int
	height     int
	status     string
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

	return Model{
		cfg:    cfg,
		db:     database,
		player: player.New(),
		mode:   modeLibrary,
		list:   l,
		browser: b,
		status: "memuat...",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(scanCmd(m.cfg, m.db), tickCmd())
}