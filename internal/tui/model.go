package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ade-nugraha306/sky-project/internal/config"
	"github.com/ade-nugraha306/sky-project/internal/db"
	"github.com/ade-nugraha306/sky-project/internal/library"
	"github.com/ade-nugraha306/sky-project/internal/player"
)

type Model struct {
	cfg        *config.Config
	db         *db.DB
	player     *player.Player
	list       list.Model
	width      int
	height     int
	status     string
	nowPlaying *library.Track
}

func NewModel(cfg *config.Config, database *db.DB) Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Sky — Library"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	return Model{
		cfg:    cfg,
		db:     database,
		player: player.New(),
		list:   l,
		status: "memuat...",
	}
}

func (m Model) Init() tea.Cmd {
	return scanCmd(m.cfg, m.db)
}