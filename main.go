package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ade-nugraha306/sky-project/internal/config"
	"github.com/ade-nugraha306/sky-project/internal/db"
	"github.com/ade-nugraha306/sky-project/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal load config: %v\n", err)
		os.Exit(1)
	}

	database, err := db.New(config.DefaultDBPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal buka database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	p := tea.NewProgram(
		tui.NewModel(cfg, database),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}