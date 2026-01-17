package parser

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/arpnghosh/adm/internal/downloader/http"
)

func ParseProtocol(rawURL string, segment int, fnameFlag string) error {
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
	if parsedURL.Path == "" {
		return fmt.Errorf("invalid URL, missing file Path")
	}

	if fnameFlag == "" {
		pathSlice := strings.Split(parsedURL.Path, "/")
		fnameFlag = pathSlice[len(pathSlice)-1]
		for {
			ext := filepath.Ext(fnameFlag)
			if ext == "" {
				break
			}
			fnameFlag = strings.TrimSuffix(fnameFlag, ext)
		}
	}

	switch parsedURL.Scheme {
	case "https", "http":
		return httpdownload.DownloadFile(rawURL, segment, fnameFlag)
	default:
		return fmt.Errorf("Unsupported network protocol")
	}
}
