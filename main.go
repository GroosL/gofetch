package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
)

func downloadFile(filepath string, url string, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done() // Notify that this goroutine is done

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		errChan <- fmt.Errorf("failed to create file %s: %w", filepath, err)
		return
	}
	defer out.Close()

	// Make a HEAD request to get file size
	headResp, err := http.Head(url)
	size := -1
	if err == nil && headResp != nil {
		sizeStr := headResp.Header.Get("Content-Length")
		if parsedSize, parseErr := strconv.Atoi(sizeStr); parseErr == nil {
			size = parsedSize
		}
	}

	if size > 0 {
		fmt.Printf("%s (%d bytes)...\n", filepath, size)
	} else {
		fmt.Printf("%s (unknown size)...\n", filepath)
	}

	// Perform the actual download
	resp, err := http.Get(url)
	if err != nil {
		errChan <- fmt.Errorf("failed to download %s: %w", url, err)
		return
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("failed to download %s: HTTP %d %s", url, resp.StatusCode, resp.Status)
		return
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		errChan <- fmt.Errorf("failed to write file %s: %w", filepath, err)
		return
	}
}

func main() {
	var wg sync.WaitGroup
	errChan := make(chan error, len(os.Args)-1)

	// Check if any arguments provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: gofetch <url1> [url2] [url3] ...")
		fmt.Println("Example: gofetch example.com https://example.org http://example.net/foobar.zip")
		os.Exit(1)
	}

	for _, link := range os.Args[1:] {
		// Add http:// if no protocol is specified, but preserve https://
		if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
			link = "http://" + link
		}

		fileURL, err := url.Parse(link)
		if err != nil {
			log.Printf("Invalid URL %s: %v\n", link, err)
			continue
		}

		path := fileURL.Path
		segments := strings.Split(path, "/")
		fileName := segments[len(segments)-1]

		// Handle empty filename (URL ends with /)
		if fileName == "" {
			fileName = "index.html"
		}

		wg.Add(1)
		go downloadFile(fileName, link, &wg, errChan)
	}

	wg.Wait()
	close(errChan)

	// Process errors
	for err := range errChan {
		log.Printf("Error: %v\n", err)
	}
}
