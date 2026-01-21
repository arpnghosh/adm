package parser

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/arpnghosh/adm/internal/downloader/http"
)

var (
	ErrEmptyURL            = errors.New("URL cannot be empty")
	ErrMissingScheme       = errors.New("URL must start with http:// or https://")
	ErrMissingHost         = errors.New("URL is missing host")
	ErrMissingPath         = errors.New("URL is missing file path")
	ErrUnsupportedProtocol = errors.New("unsupported protocol")
)

func ParseProtocol(rawURL string, segment int, filename string) error {
	if rawURL == "" {
		return ErrEmptyURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if parsedURL.Scheme == "" {
		return ErrMissingScheme
	}
	if parsedURL.Host == "" {
		return ErrMissingHost
	}
	if parsedURL.Path == "" {
		return ErrMissingPath
	}

	if filename == "" {
		filename = extractFileName(parsedURL.Path)
	}

	switch parsedURL.Scheme {
	case "https", "http":
		return httpdownload.DownloadFile(rawURL, segment, filename)
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
