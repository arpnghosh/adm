package parser

import (
	"log"
	"net/url"

	"github.com/arpnghosh/adm/src/ftpdownload"
	"github.com/arpnghosh/adm/src/httpdownload"
)

func ParseProtocol(rawURL string, segment int) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	// maybe use a switch-case ???
	if parsedURL.Scheme == "https" || parsedURL.Scheme == "http" {
		httpdownload.DownloadFile(rawURL, segment)
	} else if parsedURL.Scheme == "ftp" {
		ftpdownload.DownloadFile(rawURL, parsedURL, segment)
	} else {
		log.Fatal("unsupported network protocol")
	}
	return nil
}
