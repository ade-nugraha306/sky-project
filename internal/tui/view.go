package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	previewHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true)
	previewDirStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	previewFileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	previewDimStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	previewErrStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	progressFilledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	progressEmptyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	progressTimeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
)

func (m Model) View() string {
	var body, help string

	switch m.mode {
	case modeBrowser:
		body = m.viewBrowser()
		help = helpStyle.Render(
			"↑/↓ navigasi • enter masuk • backspace/← naik • s pilih folder ini • / filter • esc batal",
		)
	case modeFolders:
		body = m.folderList.View()
		help = helpStyle.Render(
			"↑/↓ navigasi • d/backspace hapus folder • esc kembali",
		)
	default:
		body = m.viewLibrary()
		help = helpStyle.Render(
			"↑/↓ navigasi • enter play • spasi pause • n/p next/prev • a tambah folder • d kelola folder • / filter • r rescan • q keluar",
		)
	}

	// Toast ambil prioritas kalau masih aktif; kalau tidak, tampil status biasa.
	statusText := m.status
	if m.toast != "" {
		statusText = m.toast
	}
	status := statusStyle.Render(statusText)

	return body + "\n" + status + "\n" + help
}

func (m Model) viewLibrary() string {
	listView := m.list.View()
	progress := m.viewProgress()
	return listView + "\n" + progress
}

func (m Model) viewProgress() string {
	if m.currentTrack == nil {
		bar := progressEmptyStyle.Render(strings.Repeat("─", m.progressWidth()))
		return progressTimeStyle.Render("  --:-- ") + bar + progressTimeStyle.Render(" --:--  ")
	}

	pos := m.player.Position()
	dur := m.player.Duration()

	if dur <= 0 {
		dur = 0
	}

	posStr := formatDuration(pos)
	durStr := formatDuration(dur)

	width := m.progressWidth()

	var bar string
	if dur <= 0 {
		bar = progressEmptyStyle.Render(strings.Repeat("─", width))
	} else {
		ratio := float64(pos) / float64(dur)
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}
		filled := int(float64(width) * ratio)
		bar = progressFilledStyle.Render(strings.Repeat("━", filled)) +
			progressEmptyStyle.Render(strings.Repeat("─", width-filled))
	}

	return "  " +
		progressTimeStyle.Render(posStr) + " " +
		bar + " " +
		progressTimeStyle.Render(durStr) + "  "
}

func (m Model) progressWidth() int {
	w := m.width - 20
	if w < 10 {
		w = 10
	}
	if w > 80 {
		w = 80
	}
	return w
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Seconds())
	h := total / 3600
	mnt := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, mnt, s)
	}
	return fmt.Sprintf("%d:%02d", mnt, s)
}

func (m Model) viewBrowser() string {
	listWidth := m.width * 40 / 100
	if listWidth < 20 {
		listWidth = 20
	}
	if listWidth > 60 {
		listWidth = 60
	}

	previewWidth := m.width - listWidth - 2
	if previewWidth < 20 {
		previewWidth = 20
	}

	left := m.browser.View()
	right := m.renderPreview(previewWidth)

	leftStyle := lipgloss.NewStyle().Width(listWidth)
	rightStyle := lipgloss.NewStyle().Width(previewWidth).PaddingLeft(2)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.Render(left),
		rightStyle.Render(right),
	)
}

func (m Model) renderPreview(width int) string {
	var b strings.Builder

	header := "Isi Folder"
	if m.previewFor != "" {
		header = m.previewFor
	}
	b.WriteString(previewHeaderStyle.Render(header))
	b.WriteString("\n\n")

	if m.previewFor == "" {
		b.WriteString(previewDimStyle.Render("(belum ada folder dipilih)"))
		return b.String()
	}

	if m.previewErr != "" {
		b.WriteString(previewErrStyle.Render("⚠ " + m.previewErr))
		return b.String()
	}

	if len(m.preview) == 0 {
		b.WriteString(previewDimStyle.Render("(kosong)"))
		return b.String()
	}

	maxLines := m.height - 10
	if maxLines < 5 {
		maxLines = 5
	}

	shown := 0
	for _, e := range m.preview {
		if shown >= maxLines {
			break
		}
		var line string
		if e.isDir {
			line = previewDirStyle.Render("▸ " + e.name)
		} else {
			line = previewFileStyle.Render("  " + e.name)
		}
		line = lipgloss.NewStyle().MaxWidth(width).Render(line)
		b.WriteString(line)
		b.WriteString("\n")
		shown++
	}

	if m.previewTotal > shown {
		b.WriteString(previewDimStyle.Render(
			fmt.Sprintf("... dan %d lainnya", m.previewTotal-shown),
		))
	}

	return b.String()
}