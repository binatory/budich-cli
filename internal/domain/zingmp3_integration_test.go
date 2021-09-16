// +build it

package domain

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_connectorZingMp3_GetStreamingUrl_realworld(t *testing.T) {
	c := NewConnectorZingMp3(&http.Client{Timeout: 30 * time.Second})
	require.NoError(t, c.Init())

	url, err := c.GetStreamingUrl("1075525434")
	require.NoError(t, err)
	require.NotEmpty(t, url)
}

func Test_connectorZingMp3_Search_realworld(t *testing.T) {
	c := NewConnectorZingMp3(&http.Client{Timeout: 30 * time.Second})
	require.NoError(t, c.Init())

	songs, err := c.Search("yeu voi vang")
	require.NoError(t, err)
	require.NotEmpty(t, songs)
	expectedSong := Song{
		Id:        "1075525434",
		Name:      "Yêu Vội Vàng",
		Artists:   "Lê Bảo Bình",
		Duration:  317000000000,
		Connector: "zmp3",
	}
	require.EqualValues(t, expectedSong, songs[0])
}
