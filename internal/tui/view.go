package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

func (m Model) View() string {
	view := m.list.View()

	status := statusStyle.Render(m.status)
	help := helpStyle.Render("↑/↓ navigasi • enter play • spasi pause • / filter • r rescan • q keluar")

	return view + "\n" + status + "\n" + help
}