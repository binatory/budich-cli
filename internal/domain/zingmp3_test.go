package domain

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_makeSig(t *testing.T) {
	type args struct {
		queries url.Values
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"happy case", args{url.Values{
			"length":          []string{"123"},
			"cTime":           []string{"456"},
			"lastIndex":       []string{"99"},
			"keyword":         []string{"hihi"},
			"searchSessionId": []string{"abcdef1234567890"},
			"id":              []string{"987654321"},
			"ignore_me":       []string{"ok"},
		}}, "3bdd6120a72dd50a1743e7ebf3e5d217"},
		{"real world get song detail", args{url.Values{
			"cTime": []string{"1637095026330"},
			"id":    []string{"1075525434"},
		}}, "73dd2f09649a82badfe828c1e39b1238"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, makeSig(tt.args.queries))
		})
	}
}

type mockHttpClient struct {
	mock.Mock
}

func (mhc *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	args := mhc.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func Test_connectorZingMp3_makeUrl(t *testing.T) {
	type args struct {
		path    string
		queries url.Values
	}
	tests := []struct {
		name  string
		nowFn func() time.Time
		args  args
		want  url.URL
	}{
		{"real world get song detail", func() time.Time { return time.Unix(1633853055, 0) }, args{
			path:    "/v1/song/core/get/detail",
			queries: url.Values{"id": []string{"1075525434"}},
		}, url.URL{
			Scheme:   "https",
			Host:     "api.zingmp3.app",
			Path:     "/v1/song/core/get/detail",
			RawQuery: "cTime=1633853055000&id=1075525434&publicKey=3bf21d8608473090625e102f5bcdb026&sig=db91c4f6c05ca175090177883ba7c3ec",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &connectorZingMp3{
				nowFn: tt.nowFn,
			}
			require.EqualValues(t, tt.want, c.makeUrl(tt.args.path, tt.args.queries))
		})
	}
}
