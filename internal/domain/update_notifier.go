package domain

import (
	"encoding/xml"
	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"io.github.binatory/budich-cli/metadata"
	"io/ioutil"
	"net/http"
	"strings"
)

type UpdateStatus struct {
	IsUpToDate       bool
	LatestVersion    string
	LatestVersionUrl string
}

type UpdateNotifier interface {
	Check() (UpdateStatus, error)
}

type checkResp struct {
	XMLName xml.Name         `xml:"feed"`
	Entries []checkRespEntry `xml:"entry"`
}

type checkRespEntry struct {
	XMLName xml.Name           `xml:"entry"`
	Id      string             `xml:"id"`
	Link    checkRespEntryLink `xml:"link"`
}

type checkRespEntryLink struct {
	Href string `xml:"href,attr"`
}

type updateNotifier struct {
	feedUrl        string
	idPrefix       string
	currentVersion semver.Version
	releasesOnly   bool
	httpclient     HttpClient
}

func NewUpdateNotifier(httpClient HttpClient, releasesOnly bool) UpdateNotifier {
	return &updateNotifier{
		feedUrl:        "https://github.com/binatory/budich-cli/releases.atom",
		idPrefix:       "tag:github.com,2008:Repository/387143722/",
		currentVersion: metadata.Version,
		releasesOnly:   releasesOnly,
		httpclient:     httpClient,
	}
}

func (u *updateNotifier) Check() (UpdateStatus, error) {
	req, err := http.NewRequest(http.MethodGet, u.feedUrl, nil)
	if err != nil {
		return UpdateStatus{}, errors.Wrap(err, "error creating request")
	}
	req.Header.Set("accept", "application/atom+xml")

	resp, err := u.httpclient.Do(req)
	if err != nil {
		return UpdateStatus{}, errors.Wrap(err, "error sending request")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return UpdateStatus{}, errors.Wrap(err, "error reading response body")
	}

	contentType := resp.Header.Get("content-type")
	if resp.StatusCode != http.StatusOK || strings.SplitN(contentType, ";", 2)[0] != "application/atom+xml" {
		return UpdateStatus{}, errors.Errorf("got unexpected response status=%d, contentType=%s, body=%s",
			resp.StatusCode, resp.Header.Get("content-type"), body)
	}

	var feed checkResp
	if err := xml.Unmarshal(body, &feed); err != nil {
		return UpdateStatus{}, errors.Wrapf(err, "error decoding response body=%s", body)
	}

	return u.parse(feed), nil
}

func (u *updateNotifier) parse(feed checkResp) UpdateStatus {
	latestVersion := u.currentVersion
	latestVersionUrl := ""

	for _, entry := range feed.Entries {
		trimmed := strings.TrimPrefix(entry.Id, u.idPrefix)
		if version, err := semver.ParseTolerant(trimmed); err == nil {
			isPR := len(version.Pre) > 0
			isFiltered := u.releasesOnly && isPR
			if !isFiltered && version.GT(latestVersion) {
				latestVersion = version
				latestVersionUrl = entry.Link.Href
			}
		}
	}

	return UpdateStatus{
		IsUpToDate:       u.currentVersion.Equals(latestVersion),
		LatestVersion:    latestVersion.String(),
		LatestVersionUrl: latestVersionUrl,
	}
}
