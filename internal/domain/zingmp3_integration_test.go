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

	url, err := c.GetStreamingUrl("ZU77WA8Z")
	require.NoError(t, err)
	require.NotEmpty(t, url)
}

func Test_connectorZingMp3_Search_realworld(t *testing.T) {
	c := NewConnectorZingMp3(&http.Client{Timeout: 30 * time.Second})
	require.NoError(t, c.Init())

	songs, err := c.Search("lam tinh nguyen het minh")
	require.NoError(t, err)
	require.NotEmpty(t, songs)
	require.EqualValues(t, songs[0], Song{Id: "IW9DCA08", Name: "Làm tình nguyện hết mình", Artists: "Ba Con Sói"})
}
