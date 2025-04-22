package httpdownload

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	//	"github.com/schollz/progressbar/v3"
)

func DownloadFile(url string, segment int) {
	var tempFiles []string
	defer func() {
		cleanupTempFiles(tempFiles)
	}()

	resp, err := http.Head(url)
	if err != nil {
		log.Fatal("Error while making a Head request", err)
		return
	}
	defer resp.Body.Close()
	isValid := validateResponse(*resp)
	if !isValid {
		log.Fatal("Server does not support partial content download")
		return
	}
	log.Printf("Server supports partial content download")

	contentLength := resp.ContentLength
	if contentLength <= 0 {
		log.Fatal("content length is invalid or missing")
		return
	}

	//	bar := progressbar.DefaultBytes(
	//		contentLength,
	//		"Downloading",
	//	)
	//	var mu sync.Mutex // Mutex to protect the progress bar

	contentType := resp.Header.Get("Content-Type")
	log.Printf(contentType)
	fileExtension, err := findFileExtension(contentType)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	log.Printf("content type: %v", contentType)
	log.Printf("content length: %v", contentLength)

	var start int64
	var wg sync.WaitGroup
	segmentSize := contentLength / int64(segment)
	for i := range segment {
		start = int64(i) * segmentSize
		end := start + segmentSize - 1
		if i == segment-1 {
			end = contentLength - 1
		}
		tempFile := fmt.Sprintf("segment_%d", i)
		tempFiles = append(tempFiles, tempFile)
		wg.Add(1)
		go workerFunc(&wg, tempFile, start, end, url)
	}
	wg.Wait()

	err = mergeTempFiles(tempFiles, fmt.Sprintf("output.%s", fileExtension))
	if err != nil {
		log.Printf("Failed to merge temporary files: %v", err)
		return
	}
}

func validateResponse(res http.Response) bool {
	return res.StatusCode == http.StatusOK || res.Header.Get("Accept-Ranges") == "bytes"
}

//type ProgressBarWriter struct {
//	bar *progressbar.ProgressBar
//	mu  *sync.Mutex
//}

//func (pw *ProgressBarWriter) Write(p []byte) (n int, err error) {
//	pw.mu.Lock()
//	defer pw.mu.Unlock()
//	n, err = pw.bar.Write(p)
//	return n, err
//}

func workerFunc(wg *sync.WaitGroup, tempFile string, start int64, end int64, url string) {
	defer wg.Done()
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %v", err)
		return
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to make HTTP request to %s: %v", url, err)
		return
	}
	defer res.Body.Close()

	file, err := os.Create(tempFile)
	if err != nil {
		log.Fatalf("Error to create temp file %s: %v", tempFile, err)
		return
	}
	defer file.Close()

	//	progressWriter := io.MultiWriter(file, &ProgressBarWriter{
	//		bar: bar,
	//		mu:  mu,
	//	})

	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Printf("Error writing to temp file %s: %v", tempFile, err)
	}
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
			return fmt.Errorf("error opening temp file %s: %v", file, err)
		}
		_, err = io.Copy(out, f)
		if err != nil {
			return fmt.Errorf("error writing to output file from %s: %v", file, err)
		}
		f.Close()
	}
	return nil
}

func cleanupTempFiles(tempFiles []string) {
	for _, f := range tempFiles {
		err := os.Remove(f)
		if err != nil {
			log.Printf("Failed to remove temp file %s: %v", f, err)
		}
	}
}

func findFileExtension(contentType string) (string, error) {
	parts := strings.Split(contentType, "/")
	if len(parts) != 2 {
		return "", errors.New("invalid content type format")
	}
	return parts[1], nil
}
