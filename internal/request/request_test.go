package request

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

	if req.Auth.Type != NoAuth {
		t.Errorf("Expected default auth type to be none, got %s", req.Auth.Type)
	}
}

func TestRequestData_Validate(t *testing.T) {
	// Create temporary cert and key files for mTLS tests
	certFile, err := os.CreateTemp("", "cert*.pem")
	if err != nil {
		t.Fatalf("Failed to create temp cert file: %v", err)
	}
	defer os.Remove(certFile.Name())

	keyFile, err := os.CreateTemp("", "key*.pem")
	if err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}
	defer os.Remove(keyFile.Name())

	tests := []struct {
		name    string
		req     *RequestData
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with no auth",
			req: &RequestData{
				Method: "GET",
				URL:    "https://api.example.com",
				Auth:   AuthData{Type: NoAuth},
			},
			wantErr: false,
		},
		{
			name: "empty method",
			req: &RequestData{
				URL: "https://api.example.com",
			},
			wantErr: true,
			errMsg:  "method cannot be empty",
		},
		{
			name: "empty URL",
			req: &RequestData{
				Method: "GET",
			},
			wantErr: true,
			errMsg:  "URL cannot be empty",
		},
		{
			name: "invalid URL",
			req: &RequestData{
				Method: "GET",
				URL:    "not-a-url",
			},
			wantErr: true,
			errMsg:  "invalid URL: must include scheme and host",
		},
		{
			name: "valid basic auth",
			req: &RequestData{
				Method: "GET",
				URL:    "https://api.example.com",
				Auth: AuthData{
					Type:     BasicAuth,
					Username: "testuser",
					Password: "testpass",
				},
			},
			wantErr: false,
		},
		{
			name: "basic auth missing username",
			req: &RequestData{
				Method: "GET",
				URL:    "https://api.example.com",
				Auth: AuthData{
					Type:     BasicAuth,
					Password: "testpass",
				},
			},
			wantErr: true,
			errMsg:  "username is required for basic authentication",
		},
		{
			name: "basic auth missing password",
			req: &RequestData{
				Method: "GET",
				URL:    "https://api.example.com",
				Auth: AuthData{
					Type:     BasicAuth,
					Username: "testuser",
				},
			},
			wantErr: true,
			errMsg:  "password is required for basic authentication",
		},
		{
			name: "valid api key auth",
			req: &RequestData{
				Method: "GET",
				URL:    "https://api.example.com",
				Auth: AuthData{
					Type:   APIKeyAuth,
					APIKey: "test-api-key",
				},
			},
			wantErr: false,
		},
		{
			name: "api key auth missing key",
			req: &RequestData{
				Method: "GET",
				URL:    "https://api.example.com",
				Auth: AuthData{
					Type: APIKeyAuth,
				},
			},
			wantErr: true,
			errMsg:  "API key is required for API key authentication",
		},
		{
			name: "valid mutual TLS auth",
			req: &RequestData{
				Method: "GET",
				URL:    "https://api.example.com",
				Auth: AuthData{
					Type:     MutualTLSAuth,
					CertFile: certFile.Name(),
					KeyFile:  keyFile.Name(),
				},
			},
			wantErr: false,
		},
		{
			name: "mutual TLS auth missing cert file",
			req: &RequestData{
				Method: "GET",
				URL:    "https://api.example.com",
				Auth: AuthData{
					Type:    MutualTLSAuth,
					KeyFile: keyFile.Name(),
				},
			},
			wantErr: true,
			errMsg:  "certificate file is required for mutual TLS authentication",
		},
		{
			name: "mutual TLS auth missing key file",
			req: &RequestData{
				Method: "GET",
				URL:    "https://api.example.com",
				Auth: AuthData{
					Type:     MutualTLSAuth,
					CertFile: certFile.Name(),
				},
			},
			wantErr: true,
			errMsg:  "key file is required for mutual TLS authentication",
		},
		{
			name: "mutual TLS auth with non-existent cert file",
			req: &RequestData{
				Method: "GET",
				URL:    "https://api.example.com",
				Auth: AuthData{
					Type:     MutualTLSAuth,
					CertFile: "/non/existent/cert.pem",
					KeyFile:  keyFile.Name(),
				},
			},
			wantErr: true,
			errMsg:  "certificate file does not exist",
		},
		{
			name: "mutual TLS auth with non-existent key file",
			req: &RequestData{
				Method: "GET",
				URL:    "https://api.example.com",
				Auth: AuthData{
					Type:     MutualTLSAuth,
					CertFile: certFile.Name(),
					KeyFile:  "/non/existent/key.pem",
				},
			},
			wantErr: true,
			errMsg:  "key file does not exist",
		},
		{
			name: "invalid auth type",
			req: &RequestData{
				Method: "GET",
				URL:    "https://api.example.com",
				Auth: AuthData{
					Type: "invalid",
				},
			},
			wantErr: true,
			errMsg:  "invalid authentication type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("RequestData.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("RequestData.Validate() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestRequestData_Execute(t *testing.T) {
	// Create test servers for different auth methods
	basicAuthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "testuser" || password != "testpass" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"auth":"success"}`))
	}))
	defer basicAuthServer.Close()

	apiKeyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"auth":"success"}`))
	}))
	defer apiKeyServer.Close()

	standardServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	defer standardServer.Close()

	tests := []struct {
		name        string
		requestData *RequestData
		wantStatus  int
		wantErr     bool
	}{
		{
			name: "standard request",
			requestData: &RequestData{
				Method:  "POST",
				URL:     standardServer.URL,
				Headers: map[string]string{"Content-Type": "application/json"},
				QueryParams: map[string]string{
					"key": "value",
				},
				Body: `{"test":"data"}`,
				Auth: AuthData{Type: NoAuth},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "basic auth success",
			requestData: &RequestData{
				Method: "GET",
				URL:    basicAuthServer.URL,
				Auth: AuthData{
					Type:     BasicAuth,
					Username: "testuser",
					Password: "testpass",
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "basic auth failure",
			requestData: &RequestData{
				Method: "GET",
				URL:    basicAuthServer.URL,
				Auth: AuthData{
					Type:     BasicAuth,
					Username: "wronguser",
					Password: "wrongpass",
				},
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    false,
		},
		{
			name: "api key auth success",
			requestData: &RequestData{
				Method: "GET",
				URL:    apiKeyServer.URL,
				Auth: AuthData{
					Type:   APIKeyAuth,
					APIKey: "test-api-key",
				},
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "api key auth failure",
			requestData: &RequestData{
				Method: "GET",
				URL:    apiKeyServer.URL,
				Auth: AuthData{
					Type:   APIKeyAuth,
					APIKey: "wrong-api-key",
				},
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.requestData.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Execute() status = %v, want %v", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestRequestData_Execute_Error(t *testing.T) {
	// Test with invalid URL
	req := &RequestData{
		Method: "GET",
		URL:    "not-a-url",
		Auth:   AuthData{Type: NoAuth},
	}

	_, err := req.Execute()
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	// Test with non-existent server
	req = &RequestData{
		Method: "GET",
		URL:    "http://localhost:12345",
		Auth:   AuthData{Type: NoAuth},
	}

	resp, err := req.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp.Error == "" {
		t.Error("Expected error response for non-existent server")
	}
}
