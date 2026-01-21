package httpdownload

import "errors"

var (
	ErrInvalidContentLength = errors.New("content length is invalid or missing")
	ErrServerUnreachable    = errors.New("server unreachable or returned error status")
	ErrNoRangeSupport       = errors.New("server does not support partial content")
	ErrDownloadCancelled    = errors.New("download cancelled")
	ErrUnknownFileType      = errors.New("unable to detect file type")
)
