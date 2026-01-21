package httpdownload

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func DownloadFile(url string, segmentCount int, filename string) error {
	client := newClient()

	probe, err := Probe(client, url)
	if err != nil {
		return fmt.Errorf("failed to probe URL: %w", err)
	}

	if segmentCount <= 0 {
		return fmt.Errorf("segment count must be >= 1, got %d", segmentCount)
	}

	if !probe.SupportsRange {
		segmentCount = 1
	}

	contentLength := probe.ContentLength
	if contentLength <= 0 {
		return fmt.Errorf("content length is invalid or missing")
	}

	if int(contentLength) < segmentCount {
		segmentCount = int(contentLength)
	}

	segments := NewSegments(filename, contentLength, segmentCount)
	defer func() {
		for _, seg := range segments {
			seg.Cleanup()
		}
	}()

	progress := NewProgress(contentLength)
	defer progress.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	var wg sync.WaitGroup
	results := make(chan *SegmentResult, segmentCount)

	for _, seg := range segments {
		wg.Add(1)
		go func(s *Segment) {
			defer wg.Done()
			pw := NewProgressWriter(progress)
			err := downloadSegment(ctx, client, s, url, probe.SupportsRange, pw)
			results <- &SegmentResult{Segment: s, Err: err}
		}(seg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var firstErr error
	for res := range results {
		if res.Err != nil && firstErr == nil {
			firstErr = fmt.Errorf("segment %d failed: %w", res.Segment.Index, res.Err)
			cancel()
		}
	}

	if firstErr != nil {
		return firstErr
	}

	ext, err := DetectExtension(segments[0].TempFile)
	if err != nil {
		return fmt.Errorf("failed to detect file type: %w", err)
	}

	outputFile := fmt.Sprintf("%s.%s", filename, ext)
	if err := MergeSegments(segments, outputFile); err != nil {
		return fmt.Errorf("failed to merge segments: %w", err)
	}

	fmt.Printf("\nDownload complete! File saved as: %s\n", outputFile)
	return nil
}
