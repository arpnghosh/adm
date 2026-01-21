package httpdownload

import (
	"io"
	"mime"
	"net/http"
	"os"
)

func DetectExtension(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := io.ReadAtLeast(f, buf, 1)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}

	mimeType := http.DetectContentType(buf[:n])
	exts, err := mime.ExtensionsByType(mimeType)
	if err != nil || len(exts) == 0 {
		return "", ErrUnknownFileType
	}

	return exts[0][1:], nil
}
