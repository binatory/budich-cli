package domain

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

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
		}}, "dea008424c04795415f6b345436e5d0d3d9b701721cdcad2e93c76d3b1b9df4e2a2f6a76c0d427a7296cf15991d4cfa241e7c6d8f4b33032be53cdfc81dd300c"},
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
		{"error sending request", &http.Response{}, errors.New("unexpected"), "", true},
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
