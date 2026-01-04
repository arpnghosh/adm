package httpdownload

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type workerChanInfo struct {
	segment string
	err     error
	start   int64
	end     int64
	ctx     context.Context
}

func DownloadFile(url string, segment int) error {
	errorChannel := make(chan workerChanInfo, segment)

	resp, err := http.Head(url)
	if err != nil {
		return fmt.Errorf("Error while making a Head request %w", err)
	}

	defer resp.Body.Close()

	if isValid := validateResponse(*resp); !isValid {
		return fmt.Errorf("Server does not support partial content download")
	}
	log.Printf("Server supports partial content download")

	contentLength := resp.ContentLength
	if contentLength <= 0 {
		return fmt.Errorf("content length is invalid or missing")
	}

	log.Printf("content length: %v", contentLength)

	var start int64
	var wg sync.WaitGroup
	segmentSize := contentLength / int64(segment)
	tempFiles := make([]string, segment)

	defer func() {
		cleanupTempFiles(tempFiles)
	}()

	ctx, cancel := context.WithCancel(context.Background())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		cancel()
	}()

	for i := range segment {
		start = int64(i) * segmentSize
		end := start + segmentSize - 1
		if i == segment-1 {
			end = contentLength - 1
		}
		tempFile := fmt.Sprintf("segment_%d", i)
		tempFiles[i] = tempFile
		wg.Add(1)
		go workerFunc(&wg, ctx, tempFile, start, end, url, errorChannel)
	}
	wg.Wait()

	close(errorChannel)

	var retryWaitGroup sync.WaitGroup
	var mu sync.Mutex
	var firstError error

	for returnedError := range errorChannel {
		retryWaitGroup.Add(1)

		go func(r workerChanInfo) {
			defer retryWaitGroup.Done()
			if r.err != context.Canceled && r.err != nil {
				if retryErr := retryWithBackoff(r.segment, r.ctx, r.start, r.end, url, 5); retryErr != nil {
					mu.Lock()
					if firstError == nil {
						firstError = retryErr
					}
					mu.Unlock()
				}
			}
		}(returnedError)
	}

	retryWaitGroup.Wait()

	if firstError != nil {
		return firstError
	}

	fileExtension, err := inferFiletypeFromSegment("segment_0")
	if err != nil {
		return fmt.Errorf("failed to infer file type")
	}

	log.Printf("File Extension: %v", fileExtension)

	err = mergeTempFiles(tempFiles, fmt.Sprintf("output.%s", fileExtension))
	if err != nil {
		return fmt.Errorf("Failed to merge temporary files: %v", err)
	}
	log.Printf("File Downloaded")
	return nil
}

func validateResponse(res http.Response) bool {
	if res.StatusCode != http.StatusOK {
		return false
	}
	if res.Header.Get("Accept-Ranges") != "bytes" {
		return false
	}
	return true
}

func workerFunc(wg *sync.WaitGroup, ctx context.Context, tempFile string, start int64, end int64, url string, errorChannel chan workerChanInfo) {
	if wg != nil {
		defer wg.Done()
	}

	sendError := func(err error) {
		errorChannel <- workerChanInfo{
			segment: tempFile,
			err:     err,
			start:   start,
			end:     end,
			ctx:     ctx,
		}
	}

	select {
	case <-ctx.Done():
		sendError(ctx.Err())
		return
	default:
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		sendError(err)
		return
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	res, err := client.Do(req)
	if err != nil {
		sendError(err)
		return
	}

	if res.StatusCode != http.StatusPartialContent {
		sendError(fmt.Errorf("Server does not support partial content download"))
		return
	}

	defer res.Body.Close()

	file, err := os.Create(tempFile)
	if err != nil {
		sendError(err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		sendError(err)
		return
	}
}

func retryWithBackoff(temp string, ctx context.Context, start int64, end int64, url string, maxRetries int) error {
	backoff := time.Duration(1) * time.Second

	for attempt := range maxRetries {

		select {
		case <-ctx.Done():
			return fmt.Errorf("segment %s cancelled", temp)
		default:
		}

		errorChannel := make(chan workerChanInfo, 1)
		workerFunc(nil, ctx, temp, start, end, url, errorChannel)
		close(errorChannel)

		result := <-errorChannel

		if result.err == nil {
			return nil
		} else {
			log.Printf("Segment %s failed, retrying in %v (attempt %d/%d)", temp, backoff, attempt+1, maxRetries)
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	return fmt.Errorf("segment %s failed after %d retries", temp, maxRetries)
}

func mergeTempFiles(tempFiles []string, outputFile string) error {
	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file")
	}
	defer out.Close()
	for _, file := range tempFiles {
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("error opening temp file %s: %w", file, err)
		}
		_, err = io.Copy(out, f)
		f.Close()
		if err != nil {
			return fmt.Errorf("error writing to output file from %s: %w", file, err)
		}
	}
	return nil
}

func cleanupTempFiles(tempFiles []string) {
	for _, f := range tempFiles {
		_ = os.Remove(f)
	}
}

func inferFiletypeFromSegment(segmentPath string) (string, error) {
	file, err := os.Open(segmentPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	const magic = 512

	buff := make([]byte, magic)

	_, err = io.ReadAtLeast(file, buff, 1)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}

	mimeType := http.DetectContentType(buff)

	kind, err := mime.ExtensionsByType(mimeType)
	if err != nil || len(kind) == 0 {
		return "", fmt.Errorf("unknown file type: %s", mimeType)
	}

	return kind[0][1:], nil
}
