package request

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRequestData(t *testing.T) {
	req := NewRequestData()

	if req.Method != "GET" {
		t.Errorf("Expected default method to be GET, got %s", req.Method)
	}

	if req.Headers == nil {
		t.Error("Expected headers map to be initialized")
	}

	if req.QueryParams == nil {
		t.Error("Expected query params map to be initialized")
	}

	if req.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestRequestData_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *RequestData
		wantErr bool
	}{
		{
			name: "valid request",
			req: &RequestData{
				Method: "GET",
				URL:    "https://api.example.com",
			},
			wantErr: false,
		},
		{
			name: "empty method",
			req: &RequestData{
				URL: "https://api.example.com",
			},
			wantErr: true,
		},
		{
			name: "empty URL",
			req: &RequestData{
				Method: "GET",
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			req: &RequestData{
				Method: "GET",
				URL:    "not-a-url",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("RequestData.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequestData_Execute(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Test request method
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		// Test headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Test query parameters
		if r.URL.Query().Get("key") != "value" {
			t.Errorf("Expected query param key=value, got %s", r.URL.Query().Get("key"))
		}

		// Write response
		w.Header().Set("X-Test", "test-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Create request data
	req := &RequestData{
		Method:  "POST",
		URL:     server.URL,
		Headers: map[string]string{"Content-Type": "application/json"},
		QueryParams: map[string]string{
			"key": "value",
		},
		Body: `{"test":"data"}`,
	}

	// Execute request
	resp, err := req.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if resp.Headers["X-Test"] != "test-value" {
		t.Errorf("Expected X-Test header test-value, got %s", resp.Headers["X-Test"])
	}

	if resp.Body != `{"status":"ok"}` {
		t.Errorf("Expected body {\"status\":\"ok\"}, got %s", resp.Body)
	}

	if resp.ResponseTime <= 0 {
		t.Error("Expected response time to be greater than 0")
	}
}

func TestRequestData_Execute_Error(t *testing.T) {
	// Test with invalid URL
	req := &RequestData{
		Method: "GET",
		URL:    "not-a-url",
	}

	_, err := req.Execute()
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	// Test with non-existent server
	req = &RequestData{
		Method: "GET",
		URL:    "http://localhost:12345",
	}

	resp, err := req.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp.Error == "" {
		t.Error("Expected error response for non-existent server")
	}
}
