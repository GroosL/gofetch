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
	respa, _ := http.Head(url)
	sizeStr := respa.Header.Get("Content-Length")
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		size = -1 // Failed to get the size
	}

	if size > 0 {
		fmt.Printf("%s (%d bytes)...\n", filepath, size)
	} else {
		fmt.Printf("%s (Unknown size)...\n", filepath)
	}

	// Perform the actual download
	resp, err := http.Get(url)
	if err != nil {
		errChan <- fmt.Errorf("Failed to download %s: %w", url, err)
		return
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)

}

func main() {
	var wg sync.WaitGroup
	errChan := make(chan error, len(os.Args)-1)

	for _, link := range os.Args[1:] {
    if strings.HasPrefix(link, "https://") {
      link = strings.TrimPrefix(link, "https://") 
    }
		if strings.HasPrefix(link, "http://") != true {
			link = "http://" + link
		}
		fileURL, err := url.Parse(link)
		if err != nil {
			log.Printf("Link invalido %s: %v\n", link, err)
			continue
		}

		path := fileURL.Path
		segments := strings.Split(path, "/")
		fileName := segments[len(segments)-1]

		wg.Add(1)
		go downloadFile(fileName, link, &wg, errChan)
	}

	wg.Wait()
	close(errChan)

	// Process errors
	for err := range errChan {
		log.Printf("Erro: %v\n", err)
	}
}
