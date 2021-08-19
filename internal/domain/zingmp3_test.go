package domain

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var emptyReader = ioutil.NopCloser(strings.NewReader(""))

func Test_makeSig(t *testing.T) {
	type args struct {
		path    string
		queries url.Values
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"happy case", args{"/api/path", url.Values{
			"count":     []string{"1"},
			"ctime":     []string{"123"},
			"id":        []string{"abcd"},
			"page":      []string{"10"},
			"type":      []string{"song"},
			"version":   []string{"v1.0"},
			"ignore_me": []string{"ok"},
		}}, "d28721493f78acc5fba49faaffe982e3eeaef2b2df3fe8424036295b1244d7e502ab2e3874e50a68c390be5be050fdcc27183b6832ecd9bbdcdc10a72ba43d0f"},
		{"real world get streaming", args{"/api/v2/song/get/streaming", url.Values{
			"ctime":   []string{"1633853055"},
			"id":      []string{"ZU77WA8Z"},
			"version": []string{"1.4.0"},
		}}, "5086edb33643aa49a4ef257380d04b129809380d55b7f08ad3891be261ce58204d2adacf0ba0c3053173d88ad3d8a30ff7c768c00951bb4fc5fad83d8814b68d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, makeSig(tt.args.path, tt.args.queries))
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
		{"real world get streaming", func() time.Time { return time.Unix(1633853055, 0) }, args{
			path:    "/api/v2/song/get/streaming",
			queries: url.Values{"id": []string{"ZU77WA8Z"}},
		}, url.URL{
			Scheme:   "https",
			Host:     "zingmp3.vn",
			Path:     "/api/v2/song/get/streaming",
			RawQuery: "apiKey=88265e23d4284f25963e6eedac8fbfa3&ctime=1633853055&id=ZU77WA8Z&sig=5086edb33643aa49a4ef257380d04b129809380d55b7f08ad3891be261ce58204d2adacf0ba0c3053173d88ad3d8a30ff7c768c00951bb4fc5fad83d8814b68d&version=1.4.0",
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

func Test_connectorZingMp3_getCookie(t *testing.T) {
	tests := []struct {
		name       string
		doResp     *http.Response
		doErr      error
		wantCookie string
		wantErr    bool
	}{
		{"happy case", &http.Response{
			StatusCode: http.StatusOK,
			Body:       emptyReader,
			Header: http.Header{
				"Set-Cookie": []string{zmp3Rqid + "=my_id;"},
			},
		}, nil, zmp3Rqid + "=my_id", false},
		{"error sending request", nil, errors.New("unexpected"), "", true},
		{"non 200 response", &http.Response{
			StatusCode: http.StatusTeapot,
			Body:       emptyReader,
		}, nil, "", true},
		{"no rqid cookie found", &http.Response{
			StatusCode: http.StatusOK,
			Body:       emptyReader,
		}, nil, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, zmp3Url, nil)
			require.NoError(t, err)

			mhc := new(mockHttpClient)
			mhc.On("Do", req).Return(tt.doResp, tt.doErr)

			c := NewConnectorZingMp3(mhc)

			err = c.getCookie()
			require.Equal(t, tt.wantCookie, c.cookie)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mhc.AssertExpectations(t)
		})
	}
}
