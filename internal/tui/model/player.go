package model

import (
	"io.github.binatory/budich-cli/internal/domain"
	"sync"
)

type PlayerModel struct {
	sync.RWMutex
	IsInitialized bool
	SongName      string
	ArtistsName   string
	Underlying    domain.Player
	Status        domain.PlayerStatus
}
