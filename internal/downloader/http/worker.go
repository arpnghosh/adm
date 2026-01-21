package httpdownload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const maxRetries = 5

func downloadSegment(ctx context.Context, client *Client, seg *Segment, url string, useRange bool, pw *ProgressWriter) error {
	var lastErr error

	for attempt := range maxRetries {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}

		if attempt > 0 {
			pw.Reset()
			backoff := time.Duration(1<<attempt) * time.Second
			select {
			case <-ctx.Done():
				return context.Cause(ctx)
			case <-time.After(backoff):
			}
		}

		lastErr = doDownload(ctx, client, seg, url, useRange, pw)
		if lastErr == nil {
			return nil
		}

		if errors.Is(lastErr, context.Canceled) {
			return lastErr
		}
	}

	return fmt.Errorf("segment %d failed after %d retries: %w", seg.Index, maxRetries, lastErr)
}

func doDownload(ctx context.Context, client *Client, seg *Segment, url string, useRange bool, pw *ProgressWriter) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	if useRange {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", seg.Start, seg.End))
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if useRange && resp.StatusCode != http.StatusPartialContent {
		return ErrNoRangeSupport
	}
	if !useRange && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	f, err := os.Create(seg.TempFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(io.MultiWriter(f, pw), resp.Body)
	return err
}
