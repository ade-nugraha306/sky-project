package player

import (
	"math"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/speaker"
)

const targetSampleRate = beep.SampleRate(44100)
const resampleQuality = 4

type Player struct {
	ctrl      *beep.Ctrl
	stream    beep.StreamSeekCloser
	resampler *beep.Resampler
	volume    *effects.Volume
	volPct    int
	format    beep.Format
	loaded    bool
}

func New() *Player {
	return &Player{volPct: 100}
}

func (p *Player) Load(path string) error {
	return p.LoadAt(path, 0, false)
}

func (p *Player) LoadAt(path string, start time.Duration, paused bool) error {
	streamer, format, err := decodeFile(path)
	if err != nil {
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
			streamer.Close()
			return err
		}
	}

	var s beep.Streamer = streamer
	var resampler *beep.Resampler
	if format.SampleRate != targetSampleRate {
		resampler = beep.Resample(
			resampleQuality,
			format.SampleRate,
			targetSampleRate,
			streamer,
		)
		s = resampler
	}

	// Bungkus dengan volume wrapper. Kalau volPct == 100,
	// Base^0 = 1 → tidak ada perubahan amplitudo.
	vol := &effects.Volume{
		Streamer: s,
		Base:     2,
		Volume:   pctToLog2(p.volPct),
		Silent:   p.volPct == 0,
	}

	p.stream = streamer
	p.resampler = resampler
	p.volume = vol
	p.format = format
	p.ctrl = &beep.Ctrl{Streamer: vol, Paused: paused}
	speaker.Play(p.ctrl)
	return nil
}

// pctToLog2 mengubah persentase linear (0..100) jadi eksponen
// log2 untuk effects.Volume. 100% → 0 (normal), 50% → -1, 25% → -2.
func pctToLog2(pct int) float64 {
	if pct <= 0 {
		return 0 // Silent flag yang urus
	}
	return math.Log2(float64(pct) / 100.0)
}

// SetVolume mengubah volume secara live tanpa restart track.
// pct di-clamp ke [0, 100]. Efeknya langsung terdengar.
func (p *Player) SetVolume(pct int) {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	p.volPct = pct

	if p.volume == nil {
		return
	}

	speaker.Lock()
	p.volume.Volume = pctToLog2(pct)
	p.volume.Silent = pct == 0
	speaker.Unlock()
}

func (p *Player) Volume() int {
	return p.volPct
}

func (p *Player) TogglePause() {
	if p.ctrl == nil {
		return
	}
	speaker.Lock()
	p.ctrl.Paused = !p.ctrl.Paused
	speaker.Unlock()
}

func (p *Player) IsPaused() bool {
	if p.ctrl == nil {
		return false
	}
	speaker.Lock()
	defer speaker.Unlock()
	return p.ctrl.Paused
}

func (p *Player) Stop() {
	speaker.Clear()
	if p.stream != nil {
		p.stream.Close()
		p.stream = nil
	}
	p.resampler = nil
	p.volume = nil
	p.ctrl = nil
}

func (p *Player) Seek(delta time.Duration) {
	if p.stream == nil || p.ctrl == nil {
		return
	}
	speaker.Lock()
	defer speaker.Unlock()

	deltaSamples := p.format.SampleRate.N(delta)
	target := p.stream.Position() + deltaSamples

	if target < 0 {
		target = 0
	}
	if target >= p.stream.Len() {
		if p.stream.Len() > 0 {
			target = p.stream.Len() - 1
		} else {
			target = 0
		}
	}

	if err := p.stream.Seek(target); err != nil {
		return
	}

	if p.resampler != nil {
		p.resampler = beep.Resample(
			resampleQuality,
			p.format.SampleRate,
			targetSampleRate,
			p.stream,
		)
		// Kita perlu wrap ulang karena streamer di dalam Volume
		// sudah berubah. Update field Streamer di Volume.
		p.volume.Streamer = p.resampler
	}
}

// Restart memutar dari awal track yang sedang aktif.
// Tidak mengubah state paused.
func (p *Player) Restart() {
	if p.stream == nil {
		return
	}
	p.Seek(-p.Position())
}

func (p *Player) Position() time.Duration {
	if p.stream == nil {
		return 0
	}
	pos := p.stream.Position()
	return p.format.SampleRate.D(pos)
}

func (p *Player) Duration() time.Duration {
	if p.stream == nil {
		return 0
	}
	length := p.stream.Len()
	return p.format.SampleRate.D(length)
}

func (p *Player) IsFinished() bool {
	if p.stream == nil {
		return false
	}
	return p.stream.Position() >= p.stream.Len()
}