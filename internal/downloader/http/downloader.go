package httpdownload

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/h2non/filetype"
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

	fileExtension, err := inferFiletypeFromSegment("segment_0")

	if err != nil {
		log.Fatal("failed to infer file type")
	}

	log.Printf(fileExtension)

	err = mergeTempFiles(tempFiles, fmt.Sprintf("output.%s", fileExtension))
	if err != nil {
		log.Printf("Failed to merge temporary files: %v", err)
		return
	}
}

func validateResponse(res http.Response) bool {
	return res.StatusCode == http.StatusOK || res.Header.Get("Accept-Ranges") == "bytes"
}

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

func inferFiletypeFromSegment(segmentPath string) (string, error) {
	file, err := os.Open(segmentPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buff := make([]byte, 250)
	_, err = io.ReadFull(file, buff)
	if err != nil {
		return "", err
	}

	kind, _ := filetype.Match(buff)
	if kind == filetype.Unknown {
		return "", fmt.Errorf("unknown filetype")
	}
	return kind.Extension, nil
}
