package player

import (
	"os"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

const targetSampleRate = beep.SampleRate(44100)
const resampleQuality = 4

type Player struct {
	ctrl   *beep.Ctrl
	stream beep.StreamSeekCloser
	format beep.Format
	loaded bool
}

func New() *Player {
	return &Player{}
}

func (p *Player) Load(path string) error {
	return p.LoadAt(path, 0, false)
}

func (p *Player) LoadAt(path string, start time.Duration, paused bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		f.Close()
		return err
	}

	if !p.loaded {
		speaker.Init(targetSampleRate, targetSampleRate.N(time.Second/10))
		p.loaded = true
	}

	speaker.Clear()
	if p.stream != nil {
		p.stream.Close()
	}

	if start > 0 {
		sample := format.SampleRate.N(start)
		if sample < 0 {
			sample = 0
		}
		if sample >= streamer.Len() {
			if streamer.Len() > 0 {
				sample = streamer.Len() - 1
			} else {
				sample = 0
			}
		}
		if err := streamer.Seek(sample); err != nil {
			f.Close()
			return err
		}
	}

	var finalStreamer beep.Streamer = streamer
	if format.SampleRate != targetSampleRate {
		finalStreamer = beep.Resample(
			resampleQuality,
			format.SampleRate,
			targetSampleRate,
			streamer,
		)
	}

	p.stream = streamer
	p.format = format

	// Paused di-set SEBELUM speaker.Play, jadi tidak ada
	// beberapa milidetik audio yang bocor keluar.
	p.ctrl = &beep.Ctrl{Streamer: finalStreamer, Paused: paused}
	speaker.Play(p.ctrl)
	return nil
}

func (p *Player) IsPaused() bool {
	if p.ctrl == nil {
		return false
	}
	speaker.Lock()
	defer speaker.Unlock()
	return p.ctrl.Paused
}

func (p *Player) TogglePause() {
	if p.ctrl == nil {
		return
	}
	speaker.Lock()
	p.ctrl.Paused = !p.ctrl.Paused
	speaker.Unlock()
}

func (p *Player) Stop() {
	speaker.Clear()
	if p.stream != nil {
		p.stream.Close()
		p.stream = nil
	}
	p.ctrl = nil
}

// Position mengembalikan posisi playback saat ini.
func (p *Player) Position() time.Duration {
	if p.stream == nil {
		return 0
	}
	pos := p.stream.Position()
	return p.format.SampleRate.D(pos)
}

// Duration mengembalikan total durasi lagu.
func (p *Player) Duration() time.Duration {
	if p.stream == nil {
		return 0
	}
	length := p.stream.Len()
	return p.format.SampleRate.D(length)
}

// IsFinished true saat lagu sudah selesai diputar.
func (p *Player) IsFinished() bool {
	if p.stream == nil {
		return false
	}
	return p.stream.Position() >= p.stream.Len()
}