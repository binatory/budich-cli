package tui

import (
	"io.github.binatory/busich-cli/internal/domain"
	"io.github.binatory/busich-cli/internal/tui/model"
	"time"
)

type controller struct {
	app   domain.App
	model *model.Model
	view  *view
}

func New(app domain.App) *controller {
	c := &controller{
		app:   app,
		model: model.New(app.ConnectorNames()),
	}

	v := NewView(c.model, c.onSelectSong, c.switchPage, c.onPauseOrResume, c.onSearch)
	c.view = v

	go c.WatchPlayer()

	return c
}

func (c *controller) Start() error {
	return c.view.StartView()
}

func (c *controller) PollPlayer() {
	c.model.Player.Lock()
	defer c.model.Player.Unlock()

	if c.model.Player.Underlying != nil {
		c.model.Player.Status = c.model.Player.Underlying.Report()
		c.view.updatePlayerView(true)
	}
}

func (c *controller) onSearch() {
	songs, err := c.app.Search(c.model.Search.SelectedConnector, c.model.Search.Term)
	if err != nil {
		// TODO show error modal
		return
	}

	c.model.SongsList = songs
	c.switchPage(model.PageList)
}

func (c *controller) onSelectSong(song domain.Song) {
	c.model.Player.Lock()
	defer c.model.Player.Unlock()

	// update view no matter what :yolo:
	defer c.view.updateViewsAsync()

	player := &c.model.Player

	if player.Underlying != nil {
		player.Underlying.Stop()
	}

	player.IsInitialized = true
	player.SongName = song.Name
	player.ArtistsName = song.Artists

	underlying, err := c.app.Play(song.Id, song.Connector)
	if err != nil {
		player.Status.State = domain.StateError
		return
	}
	player.Underlying = underlying
	go underlying.Start()
}

func (c *controller) switchPage(page model.PageEnum) {
	c.model.CurrentPage = page
	c.view.updateViewsAsync()
}

func (c *controller) onPauseOrResume() {
	c.model.Player.Lock()
	defer c.model.Player.Unlock()

	if c.model.Player.Underlying != nil {
		c.model.Player.Underlying.PauseOrResume()
	}
}

func (c *controller) WatchPlayer() {
	ticker := time.Tick(500 * time.Millisecond)

	for {
		select {
		case <-ticker:
			c.PollPlayer()
		}
	}
}
