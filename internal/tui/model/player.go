package model

import (
	"io.github.binatory/busich-cli/internal/domain"
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
