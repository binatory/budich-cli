package domain

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/pkg/errors"
)

type App struct {
	connector Connector
}

func NewApp(connector Connector) *App {
	return &App{connector}
}

func (a *App) Init() error {
	return errors.Wrap(a.connector.Init(), "error initializing app")
}

func (a *App) Search(name string) error {
	songs, err := a.connector.Search(name)
	if err != nil {
		return errors.Wrapf(err, "error searching name=%s", name)
	}

	tw := tabwriter.NewWriter(os.Stdout, 1, 1, 5, ' ', 0)
	defer tw.Flush()

	fmt.Fprint(tw, "Id\tBài hát\tCa sĩ\n")
	fmt.Fprint(tw, "----------\t----------\t----------\n")
	for _, s := range songs {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", s.Id, s.Name, s.Artists)
	}

	return nil
}

func (a *App) Play(id string) error {
	streamingUrl, err := a.connector.GetStreamingUrl(id)
	if err != nil {
		return errors.Wrapf(err, "error playing song id=%s", id)
	}

	resp, err := http.DefaultClient.Get(streamingUrl)
	if err != nil {
		return errors.Wrapf(err, "error requesting streamingUrl id=%s, url=%s", id, streamingUrl)
	}
	defer resp.Body.Close()

	streamer, format, err := mp3.Decode(io.NopCloser(bufio.NewReaderSize(resp.Body, 1024*1024)))
	if err != nil {
		return errors.Wrapf(err, "error decoding song id=%s, url=%s", id, streamingUrl)
	}
	defer streamer.Close()

	sr := format.SampleRate * 2
	speaker.Init(sr, sr.N(time.Second))
	resampled := beep.Resample(4, format.SampleRate, sr, streamer)

	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		done <- true
	})))

	for {
		select {
		case <-time.After(time.Second):
			fmt.Printf("\r%s", format.SampleRate.D(streamer.Position()).Round(time.Second))
		case <-done:
		}
	}
}
