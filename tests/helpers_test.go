package tests

import (
	"os"
	"path/filepath"
	"testing"
)

// setTempHome mengarahkan os.UserHomeDir() ke temp dir untuk isolasi.
// Di Unix env-nya HOME, di Windows USERPROFILE. Kita set dua-duanya
// supaya jalan di semua OS. Dipakai oleh config tests.
func setTempHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	return tmp
}

// createFile membuat file kosong di dir/name.
func createFile(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatalf("createFile %s: %v", path, err)
	}
}