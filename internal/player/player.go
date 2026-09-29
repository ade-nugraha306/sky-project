package player

import (
	"os"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

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
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		p.loaded = true
	}

	speaker.Clear()
	p.stream = streamer
	p.format = format
	p.ctrl = &beep.Ctrl{Streamer: streamer}
	speaker.Play(p.ctrl)
	return nil
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
}