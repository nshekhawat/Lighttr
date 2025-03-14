package request

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RequestData represents a complete HTTP request configuration
type RequestData struct {
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	QueryParams map[string]string `json:"query_params"`
	Body        string            `json:"body"`
	Timestamp   time.Time         `json:"timestamp"`
}

// ResponseData represents the HTTP response
type ResponseData struct {
	StatusCode   int               `json:"status_code"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	ResponseTime time.Duration     `json:"response_time"`
	Error        string            `json:"error,omitempty"`
}

// NewRequestData creates a new RequestData with initialized maps
func NewRequestData() *RequestData {
	return &RequestData{
		Method:      "GET",
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
		Timestamp:   time.Now(),
	}
}

// Execute sends the HTTP request and returns the response
func (r *RequestData) Execute() (*ResponseData, error) {
	// Validate request first
	if err := r.Validate(); err != nil {
		return nil, err
	}

	// Parse the base URL
	baseURL, err := url.Parse(r.URL)
	if err != nil {
		return nil, err
	}

	// Add query parameters
	q := baseURL.Query()
	for key, value := range r.QueryParams {
		q.Add(key, value)
	}
	baseURL.RawQuery = q.Encode()

	// Create the request
	req, err := http.NewRequest(r.Method, baseURL.String(), strings.NewReader(r.Body))
	if err != nil {
		return nil, err
	}

	// Add headers
	for key, value := range r.Headers {
		req.Header.Add(key, value)
	}

	// Execute the request
	start := time.Now()
	client := &http.Client{}
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return &ResponseData{
			Error:        err.Error(),
			ResponseTime: duration,
		}, nil
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Convert response headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	return &ResponseData{
		StatusCode:   resp.StatusCode,
		Headers:      headers,
		Body:         string(bodyBytes),
		ResponseTime: duration,
	}, nil
}

// Validate checks if the request data is valid
func (r *RequestData) Validate() error {
	if r.Method == "" {
		return fmt.Errorf("method cannot be empty")
	}
	if r.URL == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	parsedURL, err := url.Parse(r.URL)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("invalid URL: must include scheme and host")
	}
	return nil
}
