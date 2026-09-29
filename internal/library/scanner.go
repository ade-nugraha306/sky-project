package library

import (
	"io/fs"
	"path/filepath"
	"strings"
)

var SupportedExts = map[string]bool{
	".mp3":  true,
	".flac": true,
	".m4a":  true,
	".ogg":  true,
	".wav":  true,
}

type Track struct {
	Path       string
	Title      string
	Artist     string
	Album      string
	DurationMs int64
}

func ScanFolder(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if SupportedExts[ext] {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}