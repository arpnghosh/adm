package httpdownload

import (
	"net/http"
)

type ProbeResult struct {
	ContentLength int64
	SupportsRange bool
}

func Probe(client *Client, url string) (*ProbeResult, error) {
	resp, err := client.Head(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrServerUnreachable
	}

	result := &ProbeResult{
		ContentLength: resp.ContentLength,
	}

	if resp.ContentLength <= 0 {
		return nil, ErrInvalidContentLength
	}

	result.SupportsRange = probeRangeSupport(client, url)
	return result, nil
}

func probeRangeSupport(client *Client, url string) bool {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Range", "bytes=0-0")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusPartialContent
}
