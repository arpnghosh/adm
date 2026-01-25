package parser

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/arpnghosh/adm/internal/downloader/http"
)

var (
	proxyType               string
	ErrEmptyURL             = errors.New("URL cannot be empty")
	ErrMissingScheme        = errors.New("URL must start with http:// or https://")
	ErrMissingHost          = errors.New("URL is missing host")
	ErrMissingPath          = errors.New("URL is missing file path")
	ErrUnsupportedProtocol  = errors.New("unsupported protocol")
	ErrMissingProxyScheme   = errors.New("When proxy address doesn't include scheme (e.g., http://), --proxy-type is required")
	ErrUnsupportedProxyType = errors.New("invalid proxy type. Valid types: http, https, socks4, socks5")
)

func getProxyType(proxyAddr string) string {
	if strings.Contains(proxyAddr, "://") {
		parts := strings.Split(proxyAddr, "://")
		return parts[0]
	}
	return "Unknown"
}

func validateURL(u *url.URL) error {
	if u.Scheme == "" {
		return ErrMissingScheme
	}
	if u.Host == "" {
		return ErrMissingHost
	}
	if u.Path == "" {
		return ErrMissingPath
	}
	return nil
}

func validateProxyURL(proxyAddr string) error {
	if proxyAddr != "" {
		if !strings.Contains(proxyAddr, "://") {
			return ErrMissingProxyScheme
		}
	}

	proxyType = getProxyType(proxyAddr)

	validProxyTypes := map[string]bool{
		"http":   true,
		"https":  true,
		"socks5": true,
	}
	_, exists := validProxyTypes[proxyType]

	if !exists {
		return ErrUnsupportedProxyType
	}
	return nil
}

func ParseProtocol(rawURL string, segment int, filename string, rawProxyURL string) error {
	if rawURL == "" {
		return ErrEmptyURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	err = validateURL(parsedURL)
	if err != nil {
		return err
	}

	if rawProxyURL != "" {
		err = validateProxyURL(rawProxyURL)
		if err != nil {
			return err
		}
	}

	// parsedProxyURL, err := url.Parse(rawProxyURL)
	// if err != nil {
	// 	return err
	// }

	if filename == "" {
		filename = extractFileName(parsedURL.Path)
	}

	switch parsedURL.Scheme {
	case "https", "http":
		return httpdownload.DownloadFile(rawURL, rawProxyURL, segment, filename)
	default:
		return ErrUnsupportedProtocol
	}
}

func extractFileName(f string) string {
	pathSlice := strings.Split(f, "/")
	filename := pathSlice[len(pathSlice)-1]
	for {
		ext := filepath.Ext(filename)
		if ext == "" {
			break
		}
		filename = strings.TrimSuffix(filename, ext)
	}
	return filename
}
