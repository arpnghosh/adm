package parser

import (
	"log"
	"net/url"

	"github.com/arpnghosh/adm/internal/downloader/ftp"
	"github.com/arpnghosh/adm/internal/downloader/http"
)

func ParseProtocol(rawURL string, segment int) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	switch parsedURL.Scheme {
	case "https", "http":
		httpdownload.DownloadFile(rawURL, segment)
	case "ftp":
		ftpdownload.DownloadFile(rawURL, parsedURL, segment)
	default:
		log.Fatal("unsupported network protocol")
	}
	return nil
}
