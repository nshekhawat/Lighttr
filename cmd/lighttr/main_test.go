package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestExecuteDirectRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check method
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Write response
		w.Header().Set("X-Test", "test-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute request
	executeDirectRequest(
		"POST",
		server.URL,
		"Content-Type:application/json",
		`{"test":"data"}`,
	)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Verify output contains expected information
	expectedStrings := []string{
		"Status: 200",
		"X-Test: test-value",
		`{"status":"ok"}`,
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains(buf.Bytes(), []byte(expected)) {
			t.Errorf("Expected output to contain %q", expected)
		}
	}
}

func TestExecuteDirectRequest_Error(t *testing.T) {
	// Mock os.Exit
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	osExit = func(code int) {
		panic("os.Exit called")
	}

	defer func() {
		if r := recover(); r != nil {
			if r != "os.Exit called" {
				t.Errorf("unexpected panic: %v", r)
			}
		}
	}()

	// This should fail with invalid URL
	executeDirectRequest(
		"GET",
		"not-a-url",
		"",
		"",
	)

	// If we get here, executeDirectRequest didn't call os.Exit
	t.Error("Expected executeDirectRequest to exit")
}
