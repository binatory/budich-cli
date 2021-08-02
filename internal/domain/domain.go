package domain

type Connector interface {
	Init() error
	Search(name string) ([]Song, error)
	GetStreamingUrl(id string) (string, error)
}

type Song struct {
	Id      string
	Name    string
	Artists string
}
