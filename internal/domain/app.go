package domain

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type App struct {
	connectors map[string]Connector
}

var defaultApp = NewApp(
	NewConnectorZingMp3(http.DefaultClient),
	NewConnectorNhacCuaTui(http.DefaultClient),
)

func DefaultApp() *App {
	return defaultApp
}

func NewApp(connectors ...Connector) *App {
	c := make(map[string]Connector, len(connectors))
	for _, conn := range connectors {
		c[conn.Name()] = conn
	}

	return &App{c}
}

func (a *App) Init() error {
	for name, c := range a.connectors {
		if err := c.Init(); err != nil {
			return errors.Wrapf(err, "error initializing connector %s", name)
		}
	}
	return nil
}

func (a *App) Search(cName, term string) error {
	c, foundConnector := a.connectors[cName]
	if !foundConnector {
		return errors.Errorf("connector %s not recognized", cName)
	}

	songs, err := c.Search(term)
	if err != nil {
		return errors.Wrapf(err, "error searching term=%s", term)
	}

	tw := tabwriter.NewWriter(os.Stdout, 1, 1, 5, ' ', 0)
	defer tw.Flush()

	fmt.Fprint(tw, "Id\tBài hát\tCa sĩ\n")
	fmt.Fprint(tw, "----------\t----------\t----------\n")
	for _, s := range songs {
		fmt.Fprintf(tw, "%s.%s\t%s\t%s\n", cName, s.Id, s.Name, s.Artists)
	}

	return nil
}

func (a *App) Play(id string) error {
	parts := strings.SplitN(id, ".", 2)
	if len(parts) != 2 {
		return errors.Errorf("invalid id %s", id)
	}
	cName, id := parts[0], parts[1]

	c, foundConnector := a.connectors[cName]
	if !foundConnector {
		return errors.Errorf("connector %s not recognized", cName)
	}

	streamingUrl, err := c.GetStreamingUrl(id)
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
	speaker.Init(sr, sr.N(time.Second/10))
	resampled := beep.Resample(4, format.SampleRate, sr, streamer)

	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		close(done)
	})))

	for {
		select {
		case <-time.After(time.Second):
			speaker.Lock()
			log.Info().Msgf(
				"Đang phát nhạc: %s / %s",
				format.SampleRate.D(streamer.Position()).Round(time.Second),
				format.SampleRate.D(streamer.Len()).Round(time.Second),
			)
			speaker.Unlock()
		case <-done:
			return nil
		}
	}
}
