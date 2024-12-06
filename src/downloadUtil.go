package download

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"sync"
)

func DownloadUtil(url string, segment int) {
	resp, err := http.Head(url)
	if err != nil {
		fmt.Printf("Error while making a Head request, %v \n", err)
		return
	}
	defer resp.Body.Close()
	isValid := validateResponse(*resp)
	if !isValid {
		fmt.Printf("Server does not support partial content download \n")
		return
	}
	fmt.Printf("Server supports partial content download \n")

	contentLenth := resp.ContentLength

	contentType := resp.Header.Get("Content-Type")
	mimeType, _ := mime.ExtensionsByType(contentType)
	fileExtension := mimeType[0]

	fmt.Printf("content type: %v \n", contentType)
	fmt.Printf("content length: %v \n", contentLenth)

	calByteRange(contentLenth, segment, url, fileExtension)
}

func validateResponse(res http.Response) bool {
	if res.StatusCode != http.StatusOK || res.Header.Get("Accept-Ranges") != "bytes" {
		return false
	}
	return true
}

func calByteRange(contentLength int64, segment int, url string, fileExtension string) {
	var start int64
	var wg sync.WaitGroup
	tempFiles := make([]string, segment)
	segmentSize := contentLength / int64(segment)
	for i := 0; i < segment; i++ {
		start = int64(i) * segmentSize
		end := start + segmentSize - 1
		if i == segment-1 {
			end = contentLength - 1
		}
		tempFile := fmt.Sprintf("segment_%d", i)
		tempFiles[i] = tempFile
		wg.Add(1)
		go workerFunc(&wg, tempFile, start, end, url)
	}
	wg.Wait()

	mergeTempFiles(tempFiles, fmt.Sprintf("output.%s", fileExtension))
	for _, f := range tempFiles {
		os.Remove(f)
	}
}

func workerFunc(wg *sync.WaitGroup, tempFile string, start int64, end int64, url string) {
	defer wg.Done()
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end)) // set header for byte ranges
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	file, _ := os.Create(tempFile)
	_, err = io.Copy(file, res.Body)
	defer file.Close()
}

func mergeTempFiles(tempFiles []string, outputFile string) {
	out, _ := os.Create(outputFile)
	for _, file := range tempFiles {
		f, _ := os.Open(file)
		_, err := io.Copy(out, f)
		if err != nil {
		}
	}
}
