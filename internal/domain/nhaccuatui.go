package domain

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	httpClient *http.Client
	nowFn      func() time.Time
	deviceId   string
	token      string
}

func NewConnectorNhacCuaTui(httpClient *http.Client) *connectorNhacCuaTui {
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

func (c *connectorNhacCuaTui) Init() error {
	now := strconv.FormatInt(c.nowFn().UnixNano(), 10)
	hash := md5.New()
	_, _ = hash.Write(secretKey)
	_, _ = hash.Write([]byte(now))
	digest := hex.EncodeToString(hash.Sum(nil))

	u := url.URL{Scheme: "https", Host: "tvapi.nhaccuatui.com", Path: "/v1/commons/token"}

	form := url.Values{}
	form.Set("timestamp", now)
	form.Set("md5", digest)
	form.Set("deviceinfo", "{'DeviceName':'browser','DeviceID':'','OsName':'WebApp','OsVersion':'none','AppName':'NhacCuaTui','AppVersion':'2.0.0','UserName':'','LocationInfo':'','Adv':'0'}")

	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return errors.Wrap(err, "error creating auth request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-NCT-DESKTOP", "true")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error sending auth request")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("got non-ok response statusCode=%d body=%s", resp.StatusCode, body)
	}

	var decoded authResp
	if err := json.Unmarshal(body, &decoded); err != nil {
		return errors.Wrapf(err, "error decoding response body: %s", body)
	}

	if decoded.Code != 0 || decoded.Data.DeviceId == "" || decoded.Data.JwtToken == "" {
		return errors.Errorf("got invalid response %s", body)
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
	} `json:"data"`
}

func (c *connectorNhacCuaTui) Search(name string) ([]Song, error) {
	u := url.URL{Scheme: "https", Host: "tvapi.nhaccuatui.com", Path: "/v1/searchs/song"}

	form := url.Values{}
	form.Set("keyword", name)
	form.Set("pageindex", "1")
	form.Set("pagesize", "6")

	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, "error creating search request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-NCT-DESKTOP", "true")
	req.Header.Set("X-NCT-DEVICEID", c.deviceId)
	req.Header.Set("X-NCT-TOKEN", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error sending search request")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("got non-ok response statusCode=%d body=%s", resp.StatusCode, body)
	}

	var decoded nctSearchResp
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, errors.Wrapf(err, "error decoding response body: %s", body)
	}

	if decoded.Code != 0 {
		return nil, errors.Errorf("got invalid response %s", body)
	}

	result := make([]Song, len(decoded.Data))
	for idx, data := range decoded.Data {
		result[idx] = Song{
			Id:      data.SongKey,
			Name:    data.SongTitle,
			Artists: data.ArtistName,
		}
	}

	return result, nil
}

type nctSongResp struct {
	Code int `json:"code"`
	Data struct {
		StreamURL []struct {
			Type    string `json:"type"`
			Stream  string `json:"stream"`
			OnlyVIP bool   `json:"onlyVIP"`
		} `json:"streamURL"`
	} `json:"data"`
}

func (c *connectorNhacCuaTui) GetStreamingUrl(id string) (string, error) {
	u := url.URL{Scheme: "https", Host: "tvapi.nhaccuatui.com", Path: fmt.Sprintf("/v1/songs/%s", id)}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", errors.Wrap(err, "error creating search request")
	}
	req.Header.Set("X-NCT-DESKTOP", "true")
	req.Header.Set("X-NCT-DEVICEID", c.deviceId)
	req.Header.Set("X-NCT-TOKEN", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "error sending search request")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "error reading response body")
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("got non-ok response statusCode=%d body=%s", resp.StatusCode, body)
	}

	var decoded nctSongResp
	if err := json.Unmarshal(body, &decoded); err != nil {
		return "", errors.Wrapf(err, "error decoding response body: %s", body)
	}

	if decoded.Code != 0 {
		return "", errors.Errorf("got invalid response %s", body)
	}

	for _, data := range decoded.Data.StreamURL {
		if !data.OnlyVIP {
			return data.Stream, nil
		}
	}

	return "", errors.New("no playable stream has been found")
}
