package player

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/vorbis"
	"github.com/gopxl/beep/v2/wav"
)

// decodeFile membuka file dan mengembalikan streamer sesuai ekstensinya.
// Streamer yang dikembalikan bertanggung jawab menutup file saat Close()
// dipanggil oleh player.
func decodeFile(path string) (beep.StreamSeekCloser, beep.Format, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, beep.Format{}, err
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp3":
		s, format, err := mp3.Decode(f)
		if err != nil {
			f.Close()
			return nil, beep.Format{}, err
		}
		return s, format, nil

	case ".flac":
		// flac.Decode menerima io.Reader, bukan io.ReadCloser,
		// jadi dia tidak menutup file. Kita bungkus supaya
		// Close() juga menutup file.
		s, format, err := flac.Decode(f)
		if err != nil {
			f.Close()
			return nil, beep.Format{}, err
		}
		return &fileClosingStream{StreamSeekCloser: s, f: f}, format, nil

	case ".ogg":
		s, format, err := vorbis.Decode(f)
		if err != nil {
			f.Close()
			return nil, beep.Format{}, err
		}
		return s, format, nil

	case ".wav":
		s, format, err := wav.Decode(f)
		if err != nil {
			f.Close()
			return nil, beep.Format{}, err
		}
		return s, format, nil

	default:
		f.Close()
		return nil, beep.Format{}, fmt.Errorf("format %s belum didukung", ext)
	}
}

// fileClosingStream membungkus StreamSeekCloser dan menutup file
// saat Close() dipanggil. Dipakai khusus untuk FLAC karena
// flac.Decode tidak mengelola siklus hidup file.
type fileClosingStream struct {
	beep.StreamSeekCloser
	f *os.File
}

func (c *fileClosingStream) Close() error {
	err := c.StreamSeekCloser.Close()
	if cerr := c.f.Close(); err == nil {
		err = cerr
	}
	return err
}