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
	repeatBadgeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	volBadgeStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("111"))

	nowPlayingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
	nowPlayingDimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	helpSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true).
				MarginTop(1)
	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Width(16)
	helpDescStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	helpBoxStyle  = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 3)

	shuffleBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("213")).
				Bold(true)
	nowPlayingSourceStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))
)

func (m Model) View() string {
	// Overlay help ambil alih seluruh layar.
	if m.showHelp {
		body := m.viewHelp()
		help := helpStyle.Render("? / esc / q tutup bantuan")
		status := statusStyle.Render(m.status)
		return body + "\n" + status + "\n" + help
	}

	var body, help string

	switch m.mode {
	case modeBrowser:
		body = m.viewBrowser()
		help = helpStyle.Render(
			"↑/↓ nav • enter masuk • ← naik • s pilih • L lib • P playlist • F folders • ? help • esc",
		)
	case modeFolders:
		body = m.folderList.View()
		help = helpStyle.Render(
			"↑/↓ nav • d hapus • L lib • P playlist • B browse • ? help • esc",
		)
	case modePlaylists:
		body = m.playlistList.View()
		if m.inputMode != inputNone {
			body += "\n\n  " + m.textInput.View()
			help = helpStyle.Render("enter simpan • esc batal")
		} else {
			help = helpStyle.Render(
				"↑/↓ nav • enter buka • n baru • r rename • d hapus • L lib • B browse • ? help • esc",
			)
		}
	case modePlaylistPicker:
		body = m.playlistPicker.View()
		help = helpStyle.Render(
			"↑/↓ nav • enter pilih • L lib • P playlist • ? help • esc batal",
		)
	case modePlaylistDetail:
		body = m.viewPlaylistDetail()
		help = helpStyle.Render(
			"↑/↓ • enter play • n/p • ,/. seek • spasi pause • m repeat • d hapus • / filter • ? help • esc",
		)
	default:
		body = m.viewLibrary()
		help = helpStyle.Render(
			"↑/↓ nav • enter play • spasi pause • L lib • P playlist • B browse • F folders • ? help • q keluar",
		)
	}

	statusText := m.status
	if m.toast != "" {
		statusText = m.toast
	}
	status := statusStyle.Render(statusText)

	return body + "\n" + status + "\n" + help
}

// viewHelp menampilkan overlay panduan hotkey lengkap, dikelompokkan
// per kategori. Dimensi di-clamp supaya tetap muat di terminal kecil.
func (m Model) viewHelp() string {
	if m.confirmDuplicate {
		body := m.renderDuplicateConfirm()
		help := helpStyle.Render("y tambahkan sebagai duplikat • n / esc batal")
		status := statusStyle.Render(m.status)
		return body + "\n" + status + "\n" + help
	}
	var b strings.Builder

	b.WriteString(helpSectionStyle.Render("Navigasi Cepat"))
	b.WriteString("\n")
	writeHelpRow(&b, "L", "Ke Library")
	writeHelpRow(&b, "P", "Ke Playlists")
	writeHelpRow(&b, "B", "Ke File Browser")
	writeHelpRow(&b, "F", "Ke Kelola Folder")
	writeHelpRow(&b, "a", "Tambah folder (dari Library)")
	writeHelpRow(&b, "x", "Bersihkan filter")

	b.WriteString(helpSectionStyle.Render("Pemutaran"))
	b.WriteString("\n")
	writeHelpRow(&b, "enter", "Putar track yang disorot")
	writeHelpRow(&b, "spasi", "Pause / resume")
	writeHelpRow(&b, "n / >", "Track berikutnya")
	writeHelpRow(&b, "p / <", "Track sebelumnya (restart kalau > 3 detik)")
	writeHelpRow(&b, ", / [", "Mundur 5 detik")
	writeHelpRow(&b, ". / ]", "Maju 5 detik")
	writeHelpRow(&b, "m", "Ganti repeat mode: off → one → all")
	writeHelpRow(&b, "s", "Toggle shuffle (acak urutan queue)")
	writeHelpRow(&b, "+ / =", "Volume naik 5%")
	writeHelpRow(&b, "- / _", "Volume turun 5%")

	b.WriteString(helpSectionStyle.Render("Navigasi & Filter"))
	b.WriteString("\n")
	writeHelpRow(&b, "↑/↓, j/k", "Navigasi list")
	writeHelpRow(&b, "/", "Aktifkan filter / cari")
	writeHelpRow(&b, "x", "Bersihkan filter")
	writeHelpRow(&b, "esc", "Batal filter / keluar")

	b.WriteString(helpSectionStyle.Render("Library"))
	b.WriteString("\n")
	writeHelpRow(&b, "a", "Tambah folder musik")
	writeHelpRow(&b, "d", "Kelola folder musik")
	writeHelpRow(&b, "r", "Rescan library")
	writeHelpRow(&b, "t", "Tambah track yang disorot ke playlist")

	b.WriteString(helpSectionStyle.Render("Browser Folder"))
	b.WriteString("\n")
	writeHelpRow(&b, "enter", "Masuk folder")
	writeHelpRow(&b, "backspace / ←", "Naik satu level")
	writeHelpRow(&b, "s", "Pilih folder ini sebagai music folder")
	writeHelpRow(&b, "esc", "Kembali ke library")

	b.WriteString(helpSectionStyle.Render("Kelola Folder"))
	b.WriteString("\n")
	writeHelpRow(&b, "d / backspace", "Hapus folder dari daftar")
	writeHelpRow(&b, "esc / q", "Kembali ke library")

	b.WriteString(helpSectionStyle.Render("Umum"))
	b.WriteString("\n")
	writeHelpRow(&b, "?", "Tampilkan / sembunyikan panduan ini")
	writeHelpRow(&b, "q / esc", "Keluar aplikasi (resume disimpan)")
	writeHelpRow(&b, "ctrl+c", "Keluar paksa")

	b.WriteString(helpSectionStyle.Render("Playlist"))
	b.WriteString("\n")
	writeHelpRow(&b, "enter", "Buka playlist / play track")
	writeHelpRow(&b, "n / >", "Track berikutnya (dari detail)")
	writeHelpRow(&b, "p / <", "Track sebelumnya (dari detail)")
	writeHelpRow(&b, ", / [", "Mundur 5 detik (dari detail)")
	writeHelpRow(&b, ". / ]", "Maju 5 detik (dari detail)")
	writeHelpRow(&b, "spasi", "Pause / resume (dari detail)")
	writeHelpRow(&b, "d", "Hapus track dari playlist (dari detail)")

	content := b.String()

	// Bungkus dalam border, clamp lebar.
	boxWidth := m.width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}
	if boxWidth > 80 {
		boxWidth = 80
	}

	return helpBoxStyle.Width(boxWidth).Render(content)
}

func writeHelpRow(b *strings.Builder, key, desc string) {
	b.WriteString("  ")
	b.WriteString(helpKeyStyle.Render(key))
	b.WriteString(helpDescStyle.Render(desc))
	b.WriteString("\n")
}

func (m Model) viewLibrary() string {
	listView := m.list.View()
	nowPlaying := m.viewNowPlaying()
	progress := m.viewProgress()
	return listView + "\n" + nowPlaying + "\n" + progress
}

func (m Model) viewPlaylistDetail() string {
	title := fmt.Sprintf("📋 %s  (%d lagu)", m.currentPlaylistName, len(m.playlistTrackList.Items()))
	if m.playbackSourceID == m.currentPlaylistID && m.currentPlaylistID != 0 {
		title += "  ▶ playing from here"
	}
	header := previewHeaderStyle.Render(title)
	listView := m.playlistTrackList.View()
	nowPlaying := m.viewNowPlaying()
	progress := m.viewProgress()
	return header + "\n\n" + listView + "\n" + nowPlaying + "\n" + progress
}

// viewNowPlaying menampilkan baris judul track yang sedang diputar,
// lengkap dengan ikon play/pause. Kalau tidak ada track, tampilkan
// placeholder abu-abu.
func (m Model) viewNowPlaying() string {
	if m.currentTrack == nil {
		return nowPlayingDimStyle.Render("  · tidak ada lagu yang diputar")
	}

	t := m.currentTrack
	display := t.Title
	if t.Artist != "" {
		display = t.Artist + " — " + t.Title
	}

	icon := "⏸"
	if !m.player.IsPaused() {
		icon = "▶"
	}

	titlePart := "  " + icon + " " + display

	// Kalau tidak ada source label (misal setelah queue habis),
	// tampilkan judul saja.
	if m.playbackSourceName == "" {
		return nowPlayingStyle.Render(
			lipgloss.NewStyle().MaxWidth(m.width).Render(titlePart),
		)
	}

	label := "[" + m.playbackSourceName + "]"
	labelWidth := lipgloss.Width(label)

	// Lebar untuk judul: sisa terminal dikurangi label dan 2 spasi pemisah.
	titleWidth := m.width - labelWidth - 2
	if titleWidth < 10 {
		titleWidth = 10
	}

	// Truncate judul kalau kepanjangan, lalu pad kanan supaya label
	// rata di ujung kanan terminal.
	titleTruncated := lipgloss.NewStyle().MaxWidth(titleWidth).Render(titlePart)
	titlePadded := lipgloss.NewStyle().Width(titleWidth).Render(titleTruncated)

	return nowPlayingStyle.Render(titlePadded) +
		nowPlayingSourceStyle.Render(label)
}

func (m Model) viewProgress() string {
	repeatBadge := repeatBadgeStyle.Render("[" + m.repeat.String() + "]")

	shuffleBadge := ""
	if m.shuffle {
		shuffleBadge = " " + shuffleBadgeStyle.Render("[🔀]")
	}

	vol := m.player.Volume()
	volBadge := volBadgeStyle.Render(fmt.Sprintf("🔊 %d%%", vol))
	if vol == 0 {
		volBadge = repeatBadgeStyle.Render("🔇 muted")
	}

	if m.currentTrack == nil {
		bar := progressEmptyStyle.Render(strings.Repeat("─", m.progressWidth()))
		return progressTimeStyle.Render("  --:-- ") + bar +
			progressTimeStyle.Render(" --:--  ") + repeatBadge + shuffleBadge + "  " + volBadge
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
		progressTimeStyle.Render(durStr) + "  " +
		repeatBadge + shuffleBadge + "  " + volBadge
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

func (m Model) renderDuplicateConfirm() string {
	msg := fmt.Sprintf(
		"  '%s'\n\n  sudah ada di playlist '%s'.\n\n  Tambah lagi sebagai duplikat?",
		m.pendingTrack.Title,
		m.confirmPlaylistName,
	)
	return helpBoxStyle.Render(msg)
}