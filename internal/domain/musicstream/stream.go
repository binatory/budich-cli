package musicstream

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/pkg/errors"
)

type musicStream struct {
	sync.RWMutex

	// read only
	httpClient       *http.Client
	streamingUrl     string
	acceptByteRanges bool
	totalLength      int64

	// mutable
	respReader io.ReadCloser
	pos        int64
}

func New(streamingUrl string, httpClient *http.Client) (io.ReadCloser, error) {
	resp, err := httpClient.Head(streamingUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "error inspecting streamingUrl %s", streamingUrl)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("got unexpected status code %d", resp.StatusCode)
	}

	ms := &musicStream{
		streamingUrl:     streamingUrl,
		httpClient:       httpClient,
		acceptByteRanges: resp.Header.Get("Accept-Ranges") == "bytes",
		totalLength:      resp.ContentLength,
	}
	_, err = ms.Seek(0, io.SeekStart)

	if !ms.acceptByteRanges {
		return ms.respReader, nil
	}

	return ms, err
}

func (ms *musicStream) Read(p []byte) (int, error) {
	ms.Lock()
	defer ms.Unlock()

	n, err := ms.respReader.Read(p)
	ms.pos += int64(n)

	return n, err
}

func (ms *musicStream) Close() error {
	ms.RLock()
	defer ms.RUnlock()

	if ms.respReader != nil {
		return ms.respReader.Close()
	}
	return nil
}

func (ms *musicStream) Seek(offset int64, whence int) (int64, error) {
	ms.Lock()
	defer ms.Unlock()

	var start int64

	switch whence {
	case io.SeekStart:
		start = offset
	case io.SeekCurrent:
		start = ms.pos + offset
	default:
		return 0, errors.New("only io.SeekStart and io.SeekCurrent are supported")
	}

	// if ms is already initialized and the new pos is still the same then return immediately
	if start == ms.pos && ms.respReader != nil {
		return ms.pos, nil
	}

	req, err := http.NewRequest(http.MethodGet, ms.streamingUrl, nil)
	if err != nil {
		return 0, errors.Wrap(err, "error creating request for streaming")
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-", start))

	resp, err := ms.httpClient.Do(req)
	if err != nil {
		return 0, errors.Wrapf(err, "error requesting streamingUrl %s", ms.streamingUrl)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return 0, errors.Errorf("got unexpected status code %d", resp.StatusCode)
	}

	if ms.respReader != nil {
		ms.respReader.Close()
	}
	ms.respReader = resp.Body
	ms.pos = start

	return ms.pos, nil
}
