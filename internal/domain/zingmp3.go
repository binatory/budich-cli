package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type connectorZingMp3 struct{}

func NewConnectorZingMp3() *connectorZingMp3 {
	return &connectorZingMp3{}
}

func (c *connectorZingMp3) Search(name string) []Song {
	ctime := time.Now().Unix()
	input := fmt.Sprintf("count=18ctime=%dpage=1type=songversion=1.2.10", ctime)

	p1 := sha256.New()
	p1.Write([]byte(input))

	p2 := hmac.New(sha512.New, []byte("882QcNXV4tUZbvAsjmFOHqNC1LpcBRKW"))
	_, err := p2.Write([]byte("/api/v2/search" + hex.EncodeToString(p1.Sum(nil))))
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://zingmp3.vn/api/v2/search?ctime=%d&q=%s&version=1.2.10&apiKey=kI44ARvPwaqL7v0KuDSM0rGORtdY1nnw&sig=%s&type=song&page=1&count=18", ctime, name, hex.EncodeToString(p2.Sum(nil))), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Cookie", getCookie())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stderr, resp.Body)
		panic(fmt.Errorf("unexpected status code %d", resp.StatusCode))
	}

	respStruct := struct {
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
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&respStruct); err != nil {
		panic(err)
	}
	if respStruct.Err != 0 || respStruct.Data.Items == nil {
		panic(fmt.Errorf("got unexpected response: %+v", respStruct))
	}

	res := make([]Song, len(respStruct.Data.Items))
	for idx, item := range respStruct.Data.Items {
		res[idx] = Song{
			Id:      item.EncodeId,
			Name:    item.Title,
			Artists: item.ArtistsNames,
		}
	}
	return res
}

func (c *connectorZingMp3) GetStreamingUrl(id string) string {
	ctime := time.Now().Unix()
	input := fmt.Sprintf("ctime=%did=%sversion=1.2.10", ctime, id)

	p1 := sha256.New()
	p1.Write([]byte(input))

	p2 := hmac.New(sha512.New, []byte("882QcNXV4tUZbvAsjmFOHqNC1LpcBRKW"))
	_, err := p2.Write([]byte("/api/v2/song/getStreaming" + hex.EncodeToString(p1.Sum(nil))))
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://zingmp3.vn/api/v2/song/getStreaming?ctime=%d&id=%s&version=1.2.10&apiKey=kI44ARvPwaqL7v0KuDSM0rGORtdY1nnw&sig=%s", ctime, id, hex.EncodeToString(p2.Sum(nil))), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Cookie", getCookie())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}

	respStruct := struct {
		Err       int               `json:"err"`
		Msg       string            `json:"msg"`
		Data      map[string]string `json:"data"`
		Timestamp int64             `json:"timestamp"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&respStruct); err != nil {
		panic(err)
	}
	if respStruct.Err != 0 || respStruct.Data == nil || respStruct.Data["128"] == "" {
		panic(fmt.Errorf("got unexpected response: %+v", respStruct))
	}
	return respStruct.Data["128"]
}

func getCookie() string {
	resp, err := http.Get("https://zingmp3.vn")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	for _, c := range resp.Cookies() {
		if c.Name == "zmp3_rqid" {
			return fmt.Sprintf("%s=%s", c.Name, c.Value)
		}
	}

	return ""
}
