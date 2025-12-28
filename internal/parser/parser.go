package parser

import (
	"fmt"
	"net/url"

	"github.com/arpnghosh/adm/internal/downloader/http"
)

func ParseProtocol(rawURL string, segment int) error {
	if rawURL == "" {
		return fmt.Errorf("URL can not be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL, %w", err)
	}
	if parsedURL.Scheme == "" {
		return fmt.Errorf("invalid URL, must start with http:// or https://")
	}
	if parsedURL.Host == "" {
		return fmt.Errorf("invalid URL, missing Host")
	}

	switch parsedURL.Scheme {
	case "https", "http":
		return httpdownload.DownloadFile(rawURL, segment)
	default:
		return fmt.Errorf("Unsupported network protocol")
	}
}
