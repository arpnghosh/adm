package httpdownload

import (
	"context"
	"errors"
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

	"github.com/schollz/progressbar/v3"
)

type workerChanInfo struct {
	segment string
	err     error
	start   int64
	end     int64
	ctx     context.Context
}

var devMode bool = false

func DownloadFile(url string, segment int, filename string) error {
	errorChannel := make(chan workerChanInfo, segment)

	resp, err := http.Head(url)
	if err != nil {
		return fmt.Errorf("error while making a Head request %w", err)
	}

	defer resp.Body.Close()

	if isValid := validateResponse(*resp); !isValid {
		return fmt.Errorf("server does not support partial content download")
	}
	if devMode {
		log.Printf("Server supports partial content download")
	}

	contentLength := resp.ContentLength
	if contentLength <= 0 {
		return fmt.Errorf("content length is invalid or missing")
	}

	if devMode {
		log.Printf("content length: %v", contentLength)
	}

	bar := progressbar.NewOptions64(
		contentLength,
		progressbar.OptionSetDescription("Download Progress"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	var start int64
	var wg sync.WaitGroup
	segmentSize := contentLength / int64(segment)
	tempFiles := make([]string, segment)

	defer func() {
		bar.Close()
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
		tempFile := fmt.Sprintf("%s_segment_%d",filename, i)
		tempFiles[i] = tempFile

		wg.Add(1)
		go workerFunc(&wg, ctx, tempFile, start, end, url, errorChannel, bar)
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
			if errors.Is(r.err, context.Canceled) {
				mu.Lock()
				if firstError == nil {
					firstError = fmt.Errorf("segment %s cancelled", r.segment)
				}
				mu.Unlock()
			} else if r.err != nil {
				if retryErr := retryWithBackoff(r.segment, r.ctx, r.start, r.end, url, 5, bar); retryErr != nil {
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

	extInferSegmentPath := fmt.Sprintf("%s_segment_%d", filename, 0)
	fileExtension, err := inferFiletypeFromSegment(extInferSegmentPath)
	if err != nil {
		return fmt.Errorf("failed to infer file type")
	}

	if devMode {
		log.Printf("File Extension: %v", fileExtension)
	}

	outputFileName := fmt.Sprintf("%s.%s",filename, fileExtension)
	err = mergeTempFiles(tempFiles, outputFileName)
	if err != nil {
		return fmt.Errorf("failed to merge temporary files: %v", err)
	}
	fmt.Printf("\n✓ Download complete! File saved as: %s\n", outputFileName)
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

func workerFunc(wg *sync.WaitGroup, ctx context.Context, tempFile string, start int64, end int64, url string, errorChannel chan workerChanInfo, bar *progressbar.ProgressBar) {
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
		sendError(fmt.Errorf("server does not support partial content download"))
		return
	}

	defer res.Body.Close()

	file, err := os.Create(tempFile)
	if err != nil {
		sendError(err)
		return
	}
	defer file.Close()

	multiWriter := io.MultiWriter(file, bar)
	_, err = io.Copy(multiWriter, res.Body)
	if err != nil {
		sendError(err)
		return
	}
}

func retryWithBackoff(temp string, ctx context.Context, start int64, end int64, url string, maxRetries int, bar *progressbar.ProgressBar) error {
	backoff := time.Duration(1) * time.Second

	for attempt := range maxRetries {

		select {
		case <-ctx.Done():
			return fmt.Errorf("segment %s cancelled", temp)
		default:
		}

		errorChannel := make(chan workerChanInfo, 1)
		if devMode {
			log.Printf("Retrying segment %s (attempt %d/%d)", temp, attempt+1, maxRetries)
		}
		workerFunc(nil, ctx, temp, start, end, url, errorChannel, bar)
		close(errorChannel)

		result := <-errorChannel

		if result.err == nil {
			return nil
		} else {
			if devMode {
				log.Printf("Segment %s failed, retrying in %v (attempt %d/%d)", temp, backoff, attempt+1, maxRetries)
			}
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
		return "", fmt.Errorf("%w", err)
	}
	defer file.Close()

	const magic = 512

	buff := make([]byte, magic)

	_, err = io.ReadAtLeast(file, buff, 1)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", fmt.Errorf("%w", err)
	}

	mimeType := http.DetectContentType(buff)

	kind, err := mime.ExtensionsByType(mimeType)
	if err != nil || len(kind) == 0 {
		return "", fmt.Errorf("unknown file type: %s", mimeType)
	}

	return kind[0][1:], nil
}
