package httpdownload

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

type Client struct {
	http *http.Client
}

func baseTransport() *http.Transport {
	return &http.Transport{
		TLSHandshakeTimeout: 20 * time.Second,
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
	}
}

func createTransport(proxyURL string) *http.Transport {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return baseTransport()
	}

	switch parsedURL.Scheme {
	case "http", "https":
		return setupHTTPProxy(parsedURL)
	case "socks5":
		t, err := setupSOCKSProxy(parsedURL)
		if err != nil {
			return baseTransport()
		}
		return t
	default:
		return baseTransport()
	}
}

func setupHTTPProxy(p *url.URL) *http.Transport {
	transport := baseTransport()
	transport.Proxy = http.ProxyURL(p)
	return transport

	// return &http.Transport{
	// 	TLSHandshakeTimeout: 20 * time.Second,
	// 	MaxIdleConns:        100,
	// 	IdleConnTimeout:     90 * time.Second,
	// 	Proxy:               http.ProxyURL(p),
	// }
}

func setupSOCKSProxy(proxyURL *url.URL) (*http.Transport, error) {
	var auth *proxy.Auth

	if proxyURL.User != nil {
		password, _ := proxyURL.User.Password()
		auth = &proxy.Auth{
			User:     proxyURL.User.Username(),
			Password: password,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
	if err != nil {
		return nil, err
	}

	transport := baseTransport()
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.Dial(network, addr)
	}

	return transport, nil
	// return &http.Transport{
	// 	DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
	// 		return dialer.Dial(network, addr)
	// 	},
	// 	TLSHandshakeTimeout: 20 * time.Second,
	// 	MaxIdleConns:        100,
	// 	IdleConnTimeout:     90 * time.Second,
	// }
}

func newClient(proxy string) *Client {
	return &Client{
		http: &http.Client{
			Transport: createTransport(proxy),
		},
	}
}

func (c *Client) Head(url string) (*http.Response, error) {
	return c.http.Head(url)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.http.Do(req)
}
