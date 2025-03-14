package request

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type AuthType string

const (
	NoAuth        AuthType = "none"
	BasicAuth     AuthType = "basic"
	APIKeyAuth    AuthType = "apikey"
	MutualTLSAuth AuthType = "mtls"
)

// AuthData represents authentication configuration
type AuthData struct {
	Type     AuthType `json:"type"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
	APIKey   string   `json:"api_key,omitempty"`
	CertFile string   `json:"cert_file,omitempty"`
	KeyFile  string   `json:"key_file,omitempty"`
}

// RequestData represents a complete HTTP request configuration
type RequestData struct {
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	QueryParams map[string]string `json:"query_params"`
	Body        string            `json:"body"`
	Timestamp   time.Time         `json:"timestamp"`
	Auth        AuthData          `json:"auth"`
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
		Auth:        AuthData{Type: NoAuth},
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

	// Configure client based on auth type
	client := &http.Client{}

	// Apply authentication
	switch r.Auth.Type {
	case BasicAuth:
		req.SetBasicAuth(r.Auth.Username, r.Auth.Password)

	case APIKeyAuth:
		if r.Auth.APIKey != "" {
			// Try to get header name from Headers map, default to "Authorization"
			headerName := "Authorization"
			req.Header.Add(headerName, "Bearer "+r.Auth.APIKey)
		}

	case MutualTLSAuth:
		// Load client certificate
		cert, err := tls.LoadX509KeyPair(r.Auth.CertFile, r.Auth.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %v", err)
		}

		// Create TLS config
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		// Create custom transport with TLS config
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	// Execute the request
	start := time.Now()
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

	// Validate authentication configuration
	switch r.Auth.Type {
	case BasicAuth:
		if r.Auth.Username == "" {
			return fmt.Errorf("username is required for basic authentication")
		}
		if r.Auth.Password == "" {
			return fmt.Errorf("password is required for basic authentication")
		}
	case APIKeyAuth:
		if r.Auth.APIKey == "" {
			return fmt.Errorf("API key is required for API key authentication")
		}
	case MutualTLSAuth:
		if r.Auth.CertFile == "" {
			return fmt.Errorf("certificate file is required for mutual TLS authentication")
		}
		if r.Auth.KeyFile == "" {
			return fmt.Errorf("key file is required for mutual TLS authentication")
		}
		// Check if cert and key files exist
		if _, err := os.Stat(r.Auth.CertFile); os.IsNotExist(err) {
			return fmt.Errorf("certificate file does not exist: %s", r.Auth.CertFile)
		}
		if _, err := os.Stat(r.Auth.KeyFile); os.IsNotExist(err) {
			return fmt.Errorf("key file does not exist: %s", r.Auth.KeyFile)
		}
	case NoAuth:
		// No validation needed for NoAuth
	default:
		return fmt.Errorf("invalid authentication type: %s", r.Auth.Type)
	}

	return nil
}
