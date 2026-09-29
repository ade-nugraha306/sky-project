package tui

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type dirItem struct {
	name     string
	path     string
	isParent bool
}

func (d dirItem) Title() string       { return d.name }
func (d dirItem) Description() string { return "" }
func (d dirItem) FilterValue() string { return d.name }

type dirLoadedMsg struct {
	path  string
	items []list.Item
	err   error
}

func loadDirCmd(path string) tea.Cmd {
	return func() tea.Msg {
		abs, err := filepath.Abs(path)
		if err != nil {
			return dirLoadedMsg{path: path, err: err}
		}

		entries, err := os.ReadDir(abs)
		if err != nil {
			return dirLoadedMsg{path: abs, err: err}
		}

		var dirs []dirItem
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			if len(name) > 0 && name[0] == '.' {
				continue // skip hidden
			}
			dirs = append(dirs, dirItem{
				name: name,
				path: filepath.Join(abs, name),
			})
		}
		sort.Slice(dirs, func(i, j int) bool {
			return dirs[i].name < dirs[j].name
		})

		items := make([]list.Item, 0, len(dirs)+1)

		// Sisipkan ".." kalau bukan root.
		parent := filepath.Dir(abs)
		if parent != abs {
			items = append(items, dirItem{
				name:     ".. (naik)",
				path:     parent,
				isParent: true,
			})
		}
		for _, d := range dirs {
			items = append(items, d)
		}

		return dirLoadedMsg{path: abs, items: items, err: nil}
	}
}

func homeDir() string {
	if h, err := os.UserHomeDir(); err == nil && h != "" {
		return h
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}