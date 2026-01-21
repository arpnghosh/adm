package httpdownload

import (
	"net/http"
	"time"
)

type Client struct {
	http *http.Client
}

func newClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: 0 * time.Second,
			Transport: &http.Transport{
				TLSHandshakeTimeout: 20 * time.Second,
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (c *Client) Head(url string) (*http.Response, error) {
	return c.http.Head(url)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.http.Do(req)
}
