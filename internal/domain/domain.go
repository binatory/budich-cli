package domain

type Connector interface {
	Search(name string) []Song
	GetStreamingUrl(id string) string
}

type Song struct {
	Id      string
	Name    string
	Artists string
}
