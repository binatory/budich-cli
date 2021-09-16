package model

import (
	"io.github.binatory/busich-cli/internal/domain"
	"sync"
)

type Model struct {
	sync.RWMutex
	CurrentPage PageEnum
	Search      SearchModel
	SongsList   []domain.Song
	Player      PlayerModel
}

func New(connectorsName []string) *Model {

	return &Model{
		CurrentPage: PageSearch,
		Search: SearchModel{
			ConnectorNames: connectorsName,
		},
		SongsList: nil,
		Player:    PlayerModel{},
	}
}
