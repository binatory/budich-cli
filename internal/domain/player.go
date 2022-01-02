package domain

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/pkg/errors"
	"io.github.binatory/busich-cli/internal/domain/musicstream"
	"net/http"
	"time"
)

type State string

const (
	StateNotInitialized State = "StateNotInitialized"
	StateLoading              = "StateLoading"
	StatePlaying              = "StatePlaying"
	StatePaused               = "StatePaused"
	StateError                = "StateError"
)

type PlayerStatus struct {
	State State
	Err   error
	Pos   time.Duration
	Len   time.Duration
}

type Player interface {
	Start() error
	PauseOrResume()
	Stop()
	Report() PlayerStatus
}

type player struct {
	state        State
	streamingUrl string
	done         chan struct{}
	streamer     beep.StreamSeekCloser
	format       *beep.Format
	ctrl         *beep.Ctrl
}

const systemSampleRate = beep.SampleRate(48000)

func init() {
	if err := speaker.Init(systemSampleRate, systemSampleRate.N(time.Second/10)); err != nil {
		panic(err)
	}
}

func NewPlayer(streamingUrl string) Player {
	return &player{
		StateNotInitialized,
		streamingUrl,
		make(chan struct{}),
		nil, nil, nil,
	}
}

func (p *player) Start() (err error) {
	defer func() {
		if err != nil {
			p.state = StateError
		}
	}()

	// switch state to StateLoading
	p.state = StateLoading

	// create a stream
	ms, err := musicstream.New(p.streamingUrl, http.DefaultClient)
	if err != nil {
		err = errors.Wrapf(err, "error creating music stream url=%s", p.streamingUrl)
		return
	}
	defer ms.Close()

	// start decoding
	streamer, format, err := mp3.Decode(ms)
	if err != nil {
		err = errors.Wrapf(err, "error decoding song url=%s", p.streamingUrl)
		return
	}
	defer streamer.Close()

	// create beep streamers
	resampled := beep.Resample(4, format.SampleRate, systemSampleRate, streamer)
	ctrl := &beep.Ctrl{Streamer: resampled, Paused: false}

	// mutate the player
	p.streamer, p.format = streamer, &format
	p.ctrl = ctrl
	p.state = StatePlaying

	// start playing
	speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
		p.Stop()
	})))

	// wait until playing done
	for {
		select {
		case <-p.done:
			return
		}
	}
}

func (p *player) PauseOrResume() {
	if p.ctrl != nil {
		p.ctrl.Paused = !p.ctrl.Paused

		if p.ctrl.Paused {
			p.state = StatePaused
		} else {
			p.state = StatePlaying
		}
	}
}

func (p *player) Stop() {
	if p.ctrl != nil {
		p.ctrl.Streamer = nil // stop playing
	}

	select {
	case <-p.done:
		// already closed
	default:
		close(p.done)
	}
}

func (p *player) Report() PlayerStatus {
	status := PlayerStatus{State: p.state}
	if p.format == nil || p.streamer == nil {
		return status
	}

	speaker.Lock()
	defer speaker.Unlock()
	status.Err = nil
	status.Pos = p.format.SampleRate.D(p.streamer.Position()).Round(time.Second)
	status.Len = p.format.SampleRate.D(p.streamer.Len()).Round(time.Second)
	return status
}
