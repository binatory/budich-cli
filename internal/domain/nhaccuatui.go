package domain

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io.github.binatory/busich-cli/internal/utils"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	secretKey = []byte("#syntcjd!@3d^ff")
)

type connectorNhacCuaTui struct {
	httpClient HttpClient
	nowFn      func() time.Time
	deviceId   string
	token      string
}

func NewConnectorNhacCuaTui(httpClient HttpClient) *connectorNhacCuaTui {
	return &connectorNhacCuaTui{httpClient: httpClient, nowFn: time.Now}
}

type authResp struct {
	Code int `json:"code"`
	Data struct {
		DeviceId string `json:"deviceId"`
		JwtToken string `json:"jwtToken"`
	} `json:"data"`
}

func (c *connectorNhacCuaTui) Name() string {
	return "nct"
}

func (c *connectorNhacCuaTui) api(method, path, contentType string, reqBody io.Reader, respDecoded interface{}) error {
	// create request
	u := url.URL{Scheme: "https", Host: "tvapi.nhaccuatui.com", Path: path}
	req, err := http.NewRequest(method, u.String(), reqBody)
	if err != nil {
		return errors.Wrap(err, "error creating request")
	}

	// set headers
	req.Header.Set("X-NCT-DESKTOP", "true")
	if c.deviceId != "" {
		req.Header.Set("X-NCT-DEVICEID", c.deviceId)
	}
	if c.token != "" {
		req.Header.Set("X-NCT-TOKEN", c.token)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error sending request")
	}
	defer resp.Body.Close()

	// read response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	// validate response status
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("got non-ok response statusCode=%d body=%s", resp.StatusCode, respBody)
	}

	// decode response
	if err := json.Unmarshal(respBody, respDecoded); err != nil {
		return errors.Wrapf(err, "error decoding response body: %s", respBody)
	}

	return nil
}

func (c *connectorNhacCuaTui) Init() error {
	now := strconv.FormatInt(c.nowFn().UnixNano(), 10)
	hash := md5.New()
	_, _ = hash.Write(secretKey)
	_, _ = hash.Write([]byte(now))
	digest := hex.EncodeToString(hash.Sum(nil))

	form := url.Values{}
	form.Set("timestamp", now)
	form.Set("md5", digest)
	form.Set("deviceinfo", "{'DeviceName':'browser','DeviceID':'','OsName':'WebApp','OsVersion':'none','AppName':'NhacCuaTui','AppVersion':'2.0.0','UserName':'','LocationInfo':'','Adv':'0'}")

	var decoded authResp
	if err := c.api(http.MethodPost, "/v1/commons/token",
		"application/x-www-form-urlencoded", strings.NewReader(form.Encode()),
		&decoded,
	); err != nil {
		return errors.Wrap(err, "error authenticating with nct")
	}

	if decoded.Code != 0 || decoded.Data.DeviceId == "" || decoded.Data.JwtToken == "" {
		return errors.Errorf("got invalid response %+v", decoded)
	}

	c.deviceId = decoded.Data.DeviceId
	c.token = decoded.Data.JwtToken

	return nil
}

type nctSearchResp struct {
	Code int `json:"code"`
	Data []struct {
		ArtistName string `json:"artistName"`
		SongTitle  string `json:"songTitle"`
		SongKey    string `json:"songKey"`
		Duration   int64  `json:"duration"`
	} `json:"data"`
}

func (c *connectorNhacCuaTui) Search(name string) ([]Song, error) {
	form := url.Values{}
	form.Set("keyword", name)
	form.Set("pageindex", "1")
	form.Set("pagesize", "6")

	var decoded nctSearchResp
	if err := c.api(
		http.MethodPost, "/v1/searchs/song",
		"application/x-www-form-urlencoded", strings.NewReader(form.Encode()),
		&decoded,
	); err != nil {
		return nil, errors.Wrapf(err, "error searching for song name=%s", name)
	}

	if decoded.Code != 0 {
		return nil, errors.Errorf("got invalid response %+v", decoded)
	}

	result := make([]Song, len(decoded.Data))
	for idx, data := range decoded.Data {
		result[idx] = Song{
			Id:        data.SongKey,
			Name:      data.SongTitle,
			Artists:   data.ArtistName,
			Duration:  time.Duration(data.Duration) * time.Second,
			Connector: c.Name(),
		}
	}

	return result, nil
}

type nctSongResp struct {
	Code int `json:"code"`
	Data struct {
		SongKey    string `json:"songKey"`
		SongTitle  string `json:"songTitle"`
		ArtistName string `json:"artistName"`
		Duration   int64  `json:"duration"`
		StreamURL  []struct {
			Type    string `json:"type"`
			Stream  string `json:"stream"`
			OnlyVIP bool   `json:"onlyVIP"`
		} `json:"streamURL"`
	} `json:"data"`
}

func (c *connectorNhacCuaTui) GetStreamingUrl(id string) (StreamableSong, error) {
	var decoded nctSongResp
	if err := c.api(http.MethodGet, fmt.Sprintf("/v1/songs/%s", id), "", nil, &decoded); err != nil {
		return StreamableSong{}, errors.Wrapf(err, "error getting streamingUrl for id=%s", id)
	}

	if decoded.Code != 0 {
		return StreamableSong{}, errors.Errorf("got invalid response %+v", decoded)
	}

	for _, stream := range decoded.Data.StreamURL {
		if !stream.OnlyVIP {
			return StreamableSong{
				Song: Song{
					Id:        decoded.Data.SongKey,
					Name:      decoded.Data.SongTitle,
					Artists:   decoded.Data.ArtistName,
					Duration:  utils.SecondsToDuration(decoded.Data.Duration),
					Connector: c.Name(),
				},
				StreamingUrl: stream.Stream,
			}, nil
		}
	}

	return StreamableSong{}, errors.New("no playable stream has been found")
}
