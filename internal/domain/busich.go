package domain

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type busich struct {
	connector Connector
	writer    io.Writer
}

func NewBusich(connector Connector, writer io.Writer) *busich {
	return &busich{connector, writer}
}

func (b *busich) Search(name string) {
	for _, s := range b.connector.Search(name) {
		fmt.Fprintf(b.writer, "%s\t%s\t%s\n", s.Id, s.Name, s.Artists)
	}
}

func (b *busich) Play(id string) {
	streamingUrl := b.connector.GetStreamingUrl(id)
	resp, err := http.DefaultClient.Get(streamingUrl)
	if err != nil {
		panic(err)
	}

	streamer, format, err := mp3.Decode(resp.Body)
	if err != nil {
		panic(err)
	}
	defer streamer.Close()

	sr := format.SampleRate * 2
	speaker.Init(sr, sr.N(time.Second/10))
	resampled := beep.Resample(4, format.SampleRate, sr, streamer)

	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		done <- true
	})))

	<-done
}
