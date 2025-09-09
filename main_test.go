package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "13")
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	// Test download
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	
	wg.Add(1)
	go downloadFile("test_file.txt", server.URL, &wg, errChan)
	
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		t.Fatalf("Download failed: %v", err)
	}

	// Check if file was created
	if _, err := os.Stat("test_file.txt"); os.IsNotExist(err) {
		t.Fatal("Downloaded file does not exist")
	}

	// Check file content
	content, err := os.ReadFile("test_file.txt")
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != "Hello, World!" {
		t.Fatalf("File content mismatch. Expected 'Hello, World!', got '%s'", string(content))
	}

	// Cleanup
	os.Remove("test_file.txt")
}

func TestDownloadFileHTTPError(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Test download
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	
	wg.Add(1)
	go downloadFile("error_test.txt", server.URL, &wg, errChan)
	
	wg.Wait()
	close(errChan)

	// Check that we got an error
	errorReceived := false
	for err := range errChan {
		if err != nil {
			errorReceived = true
			t.Logf("Expected error received: %v", err)
		}
	}

	if !errorReceived {
		t.Fatal("Expected an error for 404 response, but got none")
	}

	// Cleanup (file might exist from the failed attempt)
	os.Remove("error_test.txt")
}