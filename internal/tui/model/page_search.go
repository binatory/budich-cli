package model

type SearchModel struct {
	// readonly
	ConnectorNames []string

	Term              string
	SelectedConnector string
	SelectedType      string
}
