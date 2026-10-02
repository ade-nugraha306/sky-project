// Package queue menyediakan operasi untuk mengelola urutan playback:
// shuffle, cari index, dan (nanti) operasi playlist.
package queue

import (
	"math/rand"

	"github.com/ade-nugraha306/sky-project/internal/library"
)

// Shuffle mengembalikan salinan slice yang sudah diacak dengan
// algoritma Fisher-Yates. Slice asli tidak dimodifikasi.
func Shuffle(tracks []library.Track) []library.Track {
	out := make([]library.Track, len(tracks))
	copy(out, tracks)
	rand.Shuffle(len(out), func(i, j int) {
		out[i], out[j] = out[j], out[i]
	})
	return out
}

// FindIndex mencari index track berdasarkan path.
// Mengembalikan -1 kalau tidak ditemukan.
func FindIndex(tracks []library.Track, path string) int {
	for i, t := range tracks {
		if t.Path == path {
			return i
		}
	}
	return -1
}