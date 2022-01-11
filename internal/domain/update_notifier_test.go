package domain

import (
	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"io.github.binatory/budich-cli/metadata"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestNewUpdateNotifier(t *testing.T) {
	mhc := &mockHttpClient{}
	notifier := NewUpdateNotifier(mhc, true)
	require.EqualValues(t, &updateNotifier{
		feedUrl:        "https://github.com/binatory/budich-cli/releases.atom",
		idPrefix:       "tag:github.com,2008:Repository/387143722/",
		currentVersion: metadata.Version,
		releasesOnly:   true,
		httpclient:     mhc,
	}, notifier)
}

func Test_updateNotifier_parse(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		releasesOnly   bool
		feed           checkResp
		want           UpdateStatus
	}{
		{"empty input", "1.0.0", false, checkResp{}, UpdateStatus{
			IsUpToDate:       true,
			LatestVersion:    "1.0.0",
			LatestVersionUrl: "",
		}},
		{"up-to-date", "1.0.0", true, checkResp{Entries: []checkRespEntry{
			{Id: "my:prefix/v1.0.0"},
		}}, UpdateStatus{IsUpToDate: true, LatestVersion: "1.0.0"}},
		{"up-to-date (ignore prerelease)", "1.0.0", true, checkResp{Entries: []checkRespEntry{
			{Id: "my:prefix/v1.0.1-beta"}, {Id: "my:prefix/v1.0.0"},
		}}, UpdateStatus{IsUpToDate: true, LatestVersion: "1.0.0"}},
		{"new pre release", "1.0.0", false, checkResp{Entries: []checkRespEntry{
			{Id: "my:prefix/v1.1.1-alpha", Link: checkRespEntryLink{Href: "https://link/to/v1.1.1-alpha"}},
			{Id: "my:prefix/v1.1.0"}, {Id: "my:prefix/v1.1.0-rc1"}, {Id: "my:prefix/v1.0.1"},
		}}, UpdateStatus{IsUpToDate: false, LatestVersion: "1.1.1-alpha", LatestVersionUrl: "https://link/to/v1.1.1-alpha"}},
		{"new release", "1.0.0", true, checkResp{Entries: []checkRespEntry{
			{Id: "my:prefix/v1.1.1-alpha"}, {Id: "my:prefix/v1.1.0", Link: checkRespEntryLink{Href: "https://link/to/v1.1.0"}},
			{Id: "my:prefix/v1.1.0-rc1"}, {Id: "my:prefix/v1.0.1"},
		}}, UpdateStatus{IsUpToDate: false, LatestVersion: "1.1.0", LatestVersionUrl: "https://link/to/v1.1.0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentVersion, err := semver.ParseTolerant(tt.currentVersion)
			require.NoError(t, err)

			u := updateNotifier{currentVersion: currentVersion, idPrefix: "my:prefix/", releasesOnly: tt.releasesOnly}
			got := u.parse(tt.feed)
			require.EqualValues(t, tt.want, got)
		})
	}
}

const responseRaw = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/" xml:lang="en-US">
    <id>tag:github.com,2008:https://github.com/hajimehoshi/oto/releases</id>
    <link type="text/html" rel="alternate" href="https://github.com/hajimehoshi/oto/releases"/>
    <link type="application/atom+xml" rel="self" href="https://github.com/hajimehoshi/oto/releases.atom"/>
    <title>Release notes from oto</title>
    <updated>2021-12-27T09:06:05Z</updated>
    <entry>
        <id>tag:github.com,2008:Repository/90259571/v2.1.0-alpha.5</id>
        <updated>2021-12-27T09:06:05Z</updated>
        <link rel="alternate" type="text/html" href="https://github.com/hajimehoshi/oto/releases/tag/v2.1.0-alpha.5"/>
        <title>v2.1.0-alpha.5</title>
        <content type="html">&lt;p&gt;internal/oboe: update Oboe to v1.6.1&lt;/p&gt;</content>
        <author>
            <name>hajimehoshi</name>
        </author>
        <media:thumbnail height="30" width="30" url="https://avatars.githubusercontent.com/u/16950?s=60&amp;v=4"/>
    </entry>
    <entry>
        <id>tag:github.com,2008:Repository/90259571/v2.1.0-alpha.4</id>
        <updated>2021-11-03T16:16:57Z</updated>
        <link rel="alternate" type="text/html" href="https://github.com/hajimehoshi/oto/releases/tag/v2.1.0-alpha.4"/>
        <title>v2.1.0-alpha.4</title>
        <content type="html">&lt;p&gt;js: Reduce allocations&lt;/p&gt;</content>
        <author>
            <name>hajimehoshi</name>
        </author>
        <media:thumbnail height="30" width="30" url="https://avatars.githubusercontent.com/u/16950?s=60&amp;v=4"/>
    </entry>
    <entry>
        <id>tag:github.com,2008:Repository/90259571/v2.1.0-alpha.3</id>
        <updated>2021-10-22T06:18:03Z</updated>
        <link rel="alternate" type="text/html" href="https://github.com/hajimehoshi/oto/releases/tag/v2.1.0-alpha.3"/>
        <title>v2.1.0-alpha.3: Add (*Context).Err</title>
        <content type="html">&lt;p&gt;Closes &lt;a class=&quot;issue-link js-issue-link&quot; data-error-text=&quot;Failed
            to load title&quot; data-id=&quot;1031800627&quot; data-permission-text=&quot;Title is private&quot;
            data-url=&quot;https://github.com/hajimehoshi/oto/issues/152&quot; data-hovercard-type=&quot;issue&quot;
            data-hovercard-url=&quot;/hajimehoshi/oto/issues/152/hovercard&quot; href=&quot;https://github.com/hajimehoshi/oto/issues/152&quot;&gt;#152&lt;/a&gt;&lt;/p&gt;
        </content>
        <author>
            <name>hajimehoshi</name>
        </author>
        <media:thumbnail height="30" width="30" url="https://avatars.githubusercontent.com/u/16950?s=60&amp;v=4"/>
    </entry>
    <entry>
        <id>tag:github.com,2008:Repository/90259571/v2.0.2</id>
        <updated>2021-10-22T06:07:03Z</updated>
        <link rel="alternate" type="text/html" href="https://github.com/hajimehoshi/oto/releases/tag/v2.0.2"/>
        <title>v2.0.2: darwin: Bug fix: Make Suspend and Resume concurrent-safe</title>
        <content type="html">&lt;p&gt;Closes &lt;a class=&quot;issue-link js-issue-link&quot; data-error-text=&quot;Failed
            to load title&quot; data-id=&quot;1033212791&quot; data-permission-text=&quot;Title is private&quot;
            data-url=&quot;https://github.com/hajimehoshi/oto/issues/153&quot; data-hovercard-type=&quot;issue&quot;
            data-hovercard-url=&quot;/hajimehoshi/oto/issues/153/hovercard&quot; href=&quot;https://github.com/hajimehoshi/oto/issues/153&quot;&gt;#153&lt;/a&gt;&lt;/p&gt;
        </content>
        <author>
            <name>hajimehoshi</name>
        </author>
        <media:thumbnail height="30" width="30" url="https://avatars.githubusercontent.com/u/16950?s=60&amp;v=4"/>
    </entry>
    <entry>
        <id>tag:github.com,2008:Repository/90259571/v2.1.0-alpha.2</id>
        <updated>2021-09-24T02:29:42Z</updated>
        <link rel="alternate" type="text/html" href="https://github.com/hajimehoshi/oto/releases/tag/v2.1.0-alpha.2"/>
        <title>v2.1.0-alpha.2</title>
        <content type="html">&lt;p&gt;windows: Update comments&lt;/p&gt;</content>
        <author>
            <name>hajimehoshi</name>
        </author>
        <media:thumbnail height="30" width="30" url="https://avatars.githubusercontent.com/u/16950?s=60&amp;v=4"/>
    </entry>
    <entry>
        <id>tag:github.com,2008:Repository/90259571/v2.0.1</id>
        <updated>2021-09-24T02:29:53Z</updated>
        <link rel="alternate" type="text/html" href="https://github.com/hajimehoshi/oto/releases/tag/v2.0.1"/>
        <title>v2.0.1</title>
        <content type="html">&lt;p&gt;windows: Update comments&lt;/p&gt;</content>
        <author>
            <name>hajimehoshi</name>
        </author>
        <media:thumbnail height="30" width="30" url="https://avatars.githubusercontent.com/u/16950?s=60&amp;v=4"/>
    </entry>
    <entry>
        <id>tag:github.com,2008:Repository/90259571/v2.1.0-alpha.1</id>
        <updated>2021-09-12T07:30:17Z</updated>
        <link rel="alternate" type="text/html" href="https://github.com/hajimehoshi/oto/releases/tag/v2.1.0-alpha.1"/>
        <title>v2.1.0-alpha.1</title>
        <content type="html">&lt;p&gt;Reland: Fix ring-buffer-like slice usages&lt;/p&gt;</content>
        <author>
            <name>hajimehoshi</name>
        </author>
        <media:thumbnail height="30" width="30" url="https://avatars.githubusercontent.com/u/16950?s=60&amp;v=4"/>
    </entry>
    <entry>
        <id>tag:github.com,2008:Repository/90259571/v2.0.0</id>
        <updated>2021-08-28T19:41:42Z</updated>
        <link rel="alternate" type="text/html" href="https://github.com/hajimehoshi/oto/releases/tag/v2.0.0"/>
        <title>v2.0.0</title>
        <content type="html">&lt;p&gt;&lt;a href=&quot;https://pkg.go.dev/github.com/hajimehoshi/oto/v2@v2.0.0&quot;
            rel=&quot;nofollow&quot;&gt;https://pkg.go.dev/github.com/hajimehoshi/oto/v2@v2.0.0&lt;/a&gt;&lt;/p&gt;
            &lt;ul&gt;
            &lt;li&gt;New APIs
            &lt;ul&gt;
            &lt;li&gt;Accepting &lt;code&gt;io.Reader&lt;/code&gt; instead of &lt;code&gt;io.Writer&lt;/code&gt; to
            create a player&lt;/li&gt;
            &lt;li&gt;Richer &lt;code&gt;Player&lt;/code&gt; interface&lt;/li&gt;
            &lt;/ul&gt;
            &lt;/li&gt;
            &lt;li&gt;New implementations
            &lt;ul&gt;
            &lt;li&gt;Much more buffers to avoid clicking noises&lt;/li&gt;
            &lt;/ul&gt;
            &lt;/li&gt;
            &lt;/ul&gt;
        </content>
        <author>
            <name>hajimehoshi</name>
        </author>
        <media:thumbnail height="30" width="30" url="https://avatars.githubusercontent.com/u/16950?s=60&amp;v=4"/>
    </entry>
    <entry>
        <id>tag:github.com,2008:Repository/90259571/v2.1.0-alpha</id>
        <updated>2021-08-25T14:56:22Z</updated>
        <link rel="alternate" type="text/html" href="https://github.com/hajimehoshi/oto/releases/tag/v2.1.0-alpha"/>
        <title>v2.1.0-alpha</title>
        <content type="html">&lt;p&gt;Update version to v2.1.0-alpha&lt;/p&gt;</content>
        <author>
            <name>hajimehoshi</name>
        </author>
        <media:thumbnail height="30" width="30" url="https://avatars.githubusercontent.com/u/16950?s=60&amp;v=4"/>
    </entry>
    <entry>
        <id>tag:github.com,2008:Repository/90259571/v2.0.0-rc.1</id>
        <updated>2021-08-25T14:55:44Z</updated>
        <link rel="alternate" type="text/html" href="https://github.com/hajimehoshi/oto/releases/tag/v2.0.0-rc.1"/>
        <title>v2.0.0-rc.1</title>
        <content type="html">&lt;p&gt;Update version to v2.0.0-rc.1&lt;/p&gt;</content>
        <author>
            <name>hajimehoshi</name>
        </author>
        <media:thumbnail height="30" width="30" url="https://avatars.githubusercontent.com/u/16950?s=60&amp;v=4"/>
    </entry>
</feed>`

func Test_updateNotifier_Check(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mhc *mockHttpClient)
		want    UpdateStatus
		wantErr bool
	}{
		{"happy case", func(mhc *mockHttpClient) {
			mhc.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.Method == http.MethodGet &&
					req.URL.String() == "https://feed_url" &&
					req.Header.Get("accept") == "application/atom+xml"
			})).Return(&http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/atom+xml"}},
				Body:       io.NopCloser(strings.NewReader(responseRaw)),
			}, nil)
		}, UpdateStatus{IsUpToDate: false, LatestVersion: "2.0.2", LatestVersionUrl: "https://github.com/hajimehoshi/oto/releases/tag/v2.0.2"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mhc := &mockHttpClient{}
			if tt.setup != nil {
				tt.setup(mhc)
			}

			u := &updateNotifier{
				feedUrl:        "https://feed_url",
				idPrefix:       "tag:github.com,2008:Repository/90259571/",
				currentVersion: semver.MustParse("1.0.0"),
				releasesOnly:   true,
				httpclient:     mhc,
			}

			got, err := u.Check()
			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Check() got = %v, want %v", got, tt.want)
			}
		})
	}
}
