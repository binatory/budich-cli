package domain

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	privateKey = "b476ff4e291877c1b83b052905d6f229ae290cc1"
	publicKey1 = "ce46f2337ef14943fb51ca558229163e7444f93b"
	publicKey2 = "3bf21d8608473090625e102f5bcdb026"

	defaultParameters = map[string]string{
		"appVersion": "2107015",
		"os":         "android",
		"osVersion":  "30",
		"publicKey":  publicKey1,
		"deviceId":   "1869ac5379584a14",
		"zDeviceId":  "2002.SSZ-wu8DGTLrXQddcmj2d261jxEN6qwRCyx-iPOTIPamXlspbrP8c321_R4rCZG.1",
	}
	parameterKeysForSigning []string
)

func init() {
	// default parameters
	parameterKeysForSigning = append(parameterKeysForSigning, "appVersion", "os", "deviceId", "osVersion", "zDeviceId")
	// search songs by keyword
	parameterKeysForSigning = append(parameterKeysForSigning, "length", "cTime", "lastIndex", "keyword", "searchSessionId")
	// get song details
	parameterKeysForSigning = append(parameterKeysForSigning, "id") // duplicates: cTime
	// sort keys
	sort.Strings(parameterKeysForSigning)
}

type connectorZingMp3 struct {
	httpClient      httpClient
	searchSessionId string
	nowFn           func() time.Time
}

func NewConnectorZingMp3(httpClient httpClient) *connectorZingMp3 {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}

	return &connectorZingMp3{
		httpClient:      httpClient,
		searchSessionId: hex.EncodeToString(buf),
		nowFn:           time.Now}
}

func (c *connectorZingMp3) Name() string {
	return "zmp3"
}

func (c *connectorZingMp3) Init() error {
	return nil
}

func makeSig(queries url.Values) string {
	// filter query params then concatenate for hashing
	sb := strings.Builder{}
	for _, key := range parameterKeysForSigning {
		val, found := defaultParameters[key]
		if !found {
			val = queries.Get(key)
		}

		if val != "" {
			sb.WriteString(key)
			sb.WriteRune('=')
			sb.WriteString(val)
			sb.WriteRune('&')
		}
	}
	sb.WriteString(privateKey)

	// Hash
	digest := md5.New()
	digest.Write([]byte(sb.String()))
	return hex.EncodeToString(digest.Sum(nil))
}

func (c *connectorZingMp3) makeUrl(path string, queries url.Values) url.URL {
	// Append required parameters
	queries.Set("cTime", strconv.FormatInt(c.nowFn().Unix()*1000, 10))

	// Sign the queries
	sig := makeSig(queries)
	queries.Add("sig", sig)

	// Append publicKey
	queries.Set("publicKey", publicKey2)

	return url.URL{
		Scheme:   "https",
		Host:     "api.zingmp3.app",
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
	for k, v := range defaultParameters {
		req.Header[strings.ToLower(k)] = []string{v}
	}

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
			Id      int64  `json:"id"`
			Title   string `json:"title"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
			PlayStatus int `json:"playStatus"`
		} `json:"items"`
		LastIndex int  `json:"lastIndex"`
		IsMore    bool `json:"isMore"`
	} `json:"data"`
	STime int64 `json:"sTime"`
}

func (c *connectorZingMp3) Search(name string) ([]Song, error) {
	// build the url containing query params and sig
	q := make(url.Values)
	q.Set("length", strconv.FormatInt(20, 10))
	q.Set("lastIndex", strconv.FormatInt(0, 10))
	q.Set("keyword", name)
	q.Set("searchSessionId", c.searchSessionId)
	u := c.makeUrl("/v1/search/core/get/list-song", q)

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
		artists := make([]string, len(item.Artists))
		for idx, artist := range item.Artists {
			artists[idx] = artist.Name
		}

		res[idx] = Song{
			Id:      strconv.FormatInt(item.Id, 10),
			Name:    item.Title,
			Artists: strings.Join(artists, ", "),
		}
	}
	return res, nil
}

type getStreamingResp struct {
	Err  int    `json:"err"`
	Msg  string `json:"msg"`
	Data struct {
		Src map[string]string `json:"src"`
	} `json:"data"`
	STime int64 `json:"sTime"`
}

func (c *connectorZingMp3) GetStreamingUrl(id string) (string, error) {
	// build the url containing query params and sig
	q := make(url.Values)
	q.Set("id", id)
	u := c.makeUrl("/v1/song/core/get/detail", q)

	// send request then decode response
	var resp getStreamingResp
	if err := c.api(u, &resp); err != nil {
		return "", errors.WithStack(err)
	}

	// validate response
	if resp.Err != 0 || resp.Data.Src == nil || resp.Data.Src["128"] == "" {
		return "", errors.Errorf("got unexpected response for url %s: %+v", u.String(), resp)
	}

	return resp.Data.Src["128"], nil
}
