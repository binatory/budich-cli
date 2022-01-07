package cli

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io.github.binatory/busich-cli/internal/domain"
	"testing"
	"time"
)

type mockApp struct {
	mock.Mock
}

func (m *mockApp) Init() error {
	return m.Called().Error(0)
}

func (m *mockApp) ConnectorNames() []string {
	return m.Called().Get(0).([]string)
}

func (m *mockApp) Search(cName, term string) ([]domain.Song, error) {
	called := m.Called(cName, term)
	return called.Get(0).([]domain.Song), called.Error(1)
}

func (m *mockApp) Play(id, connectorName string) (domain.Player, error) {
	called := m.Called(id, connectorName)
	return called.Get(0).(domain.Player), called.Error(1)
}

func (m *mockApp) CheckForUpdate() (domain.UpdateStatus, error) {
	called := m.Called()
	return called.Get(0).(domain.UpdateStatus), called.Error(1)
}

type mockPlayer struct {
	mock.Mock
}

func (p *mockPlayer) Start() error {
	return p.Called().Error(0)
}

func (p *mockPlayer) PauseOrResume() {
	p.Called()
}

func (p *mockPlayer) Stop() {
	p.Called()
}

func (p *mockPlayer) Report() domain.PlayerStatus {
	return p.Called().Get(0).(domain.PlayerStatus)
}

func TestCLI_Search_should_return_the_underlying_error(t *testing.T) {
	var out bytes.Buffer
	ma := &mockApp{}
	ma.On("Search", "toto", "tata").Return([]domain.Song{}, errors.New("unexpected"))

	cli := New(&out, ma)
	got := cli.Search("toto", "tata")
	require.Error(t, got)
	require.Empty(t, out.String())

	ma.AssertExpectations(t)
}

func TestCLI_Search_should_output_formatted_result(t *testing.T) {
	var out bytes.Buffer
	ma := &mockApp{}
	ma.On("Search", "toto", "tata").Return([]domain.Song{
		{
			Id:        "id1",
			Name:      "tata1",
			Artists:   "artist1",
			Duration:  123,
			Connector: "toto",
		},
		{
			Id:        "id2",
			Name:      "tata2",
			Artists:   "artist2",
			Duration:  456,
			Connector: "toto",
		},
	}, nil)

	cli := New(&out, ma)
	got := cli.Search("toto", "tata")
	require.NoError(t, got)

	require.Equal(t, `Id             Bài hát        Ca sĩ
----------     ----------     ----------
toto.id1       tata1          artist1
toto.id2       tata2          artist2
`, out.String())

	ma.AssertExpectations(t)
}

func TestCLI_Play(t *testing.T) {
	var out bytes.Buffer

	mp := &mockPlayer{}
	song := domain.StreamableSong{
		Song: domain.Song{
			Id:        "toto",
			Name:      "My Song",
			Artists:   "Artist1, Artist2",
			Duration:  2*time.Minute + 30*time.Second,
			Connector: "playme",
		},
	}
	mp.On("Report").Return(domain.PlayerStatus{State: domain.StateNotInitialized, Err: nil, Pos: 0, Len: 0, Song: song}).Once()
	mp.On("Report").Return(domain.PlayerStatus{State: domain.StateLoading, Err: nil, Pos: 0, Len: 0, Song: song}).Once()
	mp.On("Report").Return(domain.PlayerStatus{State: domain.StateLoading, Err: nil, Pos: 0, Len: 0, Song: song}).Once()
	mp.On("Report").Return(domain.PlayerStatus{State: domain.StatePlaying, Err: nil, Pos: 1, Len: 2, Song: song}).Once()
	mp.On("Report").Return(domain.PlayerStatus{State: domain.StatePlaying, Err: nil, Pos: 3, Len: 4, Song: song}).Once()
	mp.On("Report").Return(domain.PlayerStatus{State: domain.StateStopped, Err: nil, Pos: 5, Len: 6, Song: song}).Once()
	mp.On("Start").Return(errors.New("error start")).After(time.Second)

	ma := &mockApp{}
	ma.On("Play", "playme", "toto").Return(mp, nil)

	cli := New(&out, ma)
	cli.reportInterval = 150 * time.Millisecond
	got := cli.Play("toto.playme")
	require.EqualError(t, got, "error start")

	require.Equal(t, `Playing My Song (Artist1, Artist2), duration 2m30s
Loading...
Playing: 1ns/2ns
Playing: 3ns/4ns
`, out.String())

	mp.AssertExpectations(t)
	ma.AssertExpectations(t)
}

func TestCLI_Play_on_error(t *testing.T) {
	tests := []struct {
		name  string
		input string
		setup func(ma *mockApp)
	}{
		{"invalid input", "invalid", nil},
		{"player.Play error", "valid.input", func(ma *mockApp) {
			ma.On("Play", mock.Anything, mock.Anything).Return(&mockPlayer{}, errors.New("unexpected"))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			ma := &mockApp{}
			if tt.setup != nil {
				tt.setup(ma)
			}
			c := New(&out, ma)
			err := c.Play(tt.input)
			require.Error(t, err)
			require.Empty(t, out.String())
			ma.AssertExpectations(t)
		})
	}
}
