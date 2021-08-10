package domain

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
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
	zVersion    = "1.2.10"
	zPrivateKey = []byte("882QcNXV4tUZbvAsjmFOHqNC1LpcBRKW")
	zApiKey     = "kI44ARvPwaqL7v0KuDSM0rGORtdY1nnw"
)

type connectorZingMp3 struct {
	httpClient *http.Client
	cookie     string
	nowFn      func() time.Time
}

func NewConnectorZingMp3(httpClient *http.Client) *connectorZingMp3 {
	return &connectorZingMp3{httpClient: httpClient, nowFn: time.Now}
}

func (c *connectorZingMp3) Name() string {
	return "zmp3"
}

func (c *connectorZingMp3) Init() error {
	return errors.Wrap(c.getCookie(), "error initializing ConnectorZingMp3")
}

func makeUrl(path string, queries url.Values) url.URL {
	// filter query params then concatenate for hashing
	sb := strings.Builder{}
	for _, key := range []string{"count", "ctime", "id", "page", "type", "version"} {
		if val := queries.Get(key); val != "" {
			sb.WriteString(key)
			sb.WriteRune('=')
			sb.WriteString(val)
		}
	}

	// Hash queries
	p1 := sha256.New()
	p1.Write([]byte(sb.String()))
	p1Hex := hex.EncodeToString(p1.Sum(nil))

	// Sign queries hash
	p2 := hmac.New(sha512.New, zPrivateKey)
	p2.Write([]byte(path))
	p2.Write([]byte(p1Hex))
	sig := hex.EncodeToString(p2.Sum(nil))

	// Add the signature into the queries
	queries.Add("sig", sig)

	// Append apiKey
	queries.Set("apiKey", zApiKey)

	return url.URL{
		Scheme:   "https",
		Host:     "zingmp3.vn",
		Path:     path,
		RawQuery: queries.Encode(),
	}
}

func (c *connectorZingMp3) api(u url.URL, respStruct interface{}) error {
	// make the request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return errors.Wrapf(err, "error creating request %s", u.String())
	}
	req.Header.Add("Cookie", c.cookie)

	// send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "error sending request %s", u.String())
	}
	defer resp.Body.Close()

	// read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "error reading response")
	}

	// stop if the status code is unexpected
	if resp.StatusCode != http.StatusOK {
		return errors.Wrapf(err, "got unexpected status code %d for url %s. Response %s", resp.StatusCode, u.String(), body)
	}

	// decode response
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(respStruct); err != nil {
		return errors.Wrapf(err, "error decoding response for url %s. Response %s", u.String(), body)
	}

	return nil
}

type searchResp struct {
	Err  int    `json:"err"`
	Msg  string `json:"msg"`
	Data struct {
		Items []struct {
			EncodeId     string `json:"encodeId"`
			Title        string `json:"title"`
			ArtistsNames string `json:"artistsNames"`
		} `json:"items"`
		Total int `json:"total"`
	} `json:"data"`
	Timestamp int64 `json:"timestamp"`
}

func (c *connectorZingMp3) Search(name string) ([]Song, error) {
	// build the url containing query params and sig
	q := make(url.Values)
	q.Set("ctime", strconv.FormatInt(c.nowFn().Unix(), 10))
	q.Set("count", strconv.FormatInt(18, 10))
	q.Set("page", strconv.FormatInt(1, 10))
	q.Set("type", "song")
	q.Set("version", zVersion)
	q.Set("q", name)
	u := makeUrl("/api/v2/search", q)

	// send request then decode response
	var resp searchResp
	if err := c.api(u, &resp); err != nil {
		return nil, errors.WithStack(err)
	}

	// validate response
	if resp.Err != 0 || resp.Data.Items == nil {
		return nil, errors.Errorf("got unexpected response for url %s. Response %+v", u.String(), resp)
	}

	// build result
	res := make([]Song, len(resp.Data.Items))
	for idx, item := range resp.Data.Items {
		res[idx] = Song{
			Id:      item.EncodeId,
			Name:    item.Title,
			Artists: item.ArtistsNames,
		}
	}
	return res, nil
}

type getStreamingResp struct {
	Err       int               `json:"err"`
	Msg       string            `json:"msg"`
	Data      map[string]string `json:"data"`
	Timestamp int64             `json:"timestamp"`
}

func (c *connectorZingMp3) GetStreamingUrl(id string) (string, error) {
	// build the url containing query params and sig
	q := make(url.Values)
	q.Set("ctime", strconv.FormatInt(c.nowFn().Unix(), 10))
	q.Set("id", id)
	q.Set("version", zVersion)
	u := makeUrl("/api/v2/song/getStreaming", q)

	// send request then decode response
	var resp getStreamingResp
	if err := c.api(u, &resp); err != nil {
		return "", errors.WithStack(err)
	}

	// validate response
	if resp.Err != 0 || resp.Data == nil || resp.Data["128"] == "" {
		return "", errors.Errorf("got unexpected response for url %s: %+v", u.String(), resp)
	}

	return resp.Data["128"], nil
}

func (c *connectorZingMp3) getCookie() error {
	resp, err := http.Get("https://zingmp3.vn")
	if err != nil {
		return errors.Wrap(err, "error getting cookie")
	}
	defer resp.Body.Close()

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "zmp3_rqid" {
			c.cookie = fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
			return nil
		}
	}

	return errors.New("required cookie not found")
}
