package tui

import (
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
)

type previewEntry struct {
	name  string
	isDir bool
}

type previewLoadedMsg struct {
	path    string
	entries []previewEntry
	total   int
	err     error
}

// previewLimit membatasi jumlah baris yang kita simpan,
// supaya folder dengan 10.000 file tidak menghabiskan memori.
const previewLimit = 500

func loadPreviewCmd(path string) tea.Cmd {
	return func() tea.Msg {
		entries, err := os.ReadDir(path)
		if err != nil {
			return previewLoadedMsg{path: path, err: err}
		}

		var dirs, files []previewEntry
		for _, e := range entries {
			name := e.Name()
			if len(name) > 0 && name[0] == '.' {
				continue // skip hidden
			}
			entry := previewEntry{name: name, isDir: e.IsDir()}
			if e.IsDir() {
				dirs = append(dirs, entry)
			} else {
				files = append(files, entry)
			}
		}
		sort.Slice(dirs, func(i, j int) bool { return dirs[i].name < dirs[j].name })
		sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })

		// Folder dulu, lalu file. Pakai make supaya slice selalu non-nil.
		all := make([]previewEntry, 0, len(dirs)+len(files))
		all = append(all, dirs...)
		all = append(all, files...)

		total := len(all)
		if len(all) > previewLimit {
			all = all[:previewLimit]
		}
		return previewLoadedMsg{path: path, entries: all, total: total}
	}
}