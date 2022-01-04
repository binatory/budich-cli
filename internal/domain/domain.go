package domain

import (
	"net/http"
	"time"
)

type Connector interface {
	Name() string
	Init() error
	Search(name string) ([]Song, error)
	GetStreamingUrl(id string) (StreamableSong, error)
}

type Song struct {
	Id        string
	Name      string
	Artists   string
	Duration  time.Duration
	Connector string
}

type StreamableSong struct {
	Song
	StreamingUrl string
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}
