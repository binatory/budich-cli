package domain

import (
	"io.github.binatory/busich-cli/internal/utils"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type App interface {
	Init() error
	ConnectorNames() []string
	Search(cName, term string) ([]Song, error)
	Play(id, connectorName string) (Player, error)
	CheckForUpdate() (UpdateStatus, error)
}

type app struct {
	connectors     map[string]Connector
	updateNotifier UpdateNotifier
}

var (
	defaultHttpClient = &http.Client{Timeout: 30 * time.Second}
	defaultApp        = NewApp(
		NewUpdateNotifier(defaultHttpClient, true), // TODO use releasesOnly from user preferences
		NewConnectorZingMp3(defaultHttpClient),
		NewConnectorNhacCuaTui(defaultHttpClient),
	)
)

func DefaultApp() App {
	return defaultApp
}

func NewApp(updateNotifier UpdateNotifier, connectors ...Connector) App {
	c := make(map[string]Connector, len(connectors))
	for _, conn := range connectors {
		c[conn.Name()] = conn
	}

	return &app{connectors: c, updateNotifier: updateNotifier}
}

func (a *app) Init() error {
	for name, c := range a.connectors {
		if err := c.Init(); err != nil {
			return errors.Wrapf(err, "error initializing connector %s", name)
		}
	}
	return nil
}

func (a *app) ConnectorNames() []string {
	return utils.GetMapKeys(a.connectors)
}

func (a *app) Search(cName, term string) ([]Song, error) {
	c, foundConnector := a.connectors[cName]
	if !foundConnector {
		return nil, errors.Errorf("connector %s not recognized", cName)
	}

	return c.Search(term)
}

func (a *app) Play(id, connectorName string) (Player, error) {
	c, foundConnector := a.connectors[connectorName]
	if !foundConnector {
		return nil, errors.Errorf("connector %s not recognized", connectorName)
	}

	song, err := c.GetStreamingUrl(id)
	if err != nil {
		return nil, errors.Wrapf(err, "error playing song id=%s", id)
	}

	return NewPlayer(song), nil
}

func (a *app) CheckForUpdate() (UpdateStatus, error) {
	return a.updateNotifier.Check()
}
