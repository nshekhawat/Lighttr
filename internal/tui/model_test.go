package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nshekhawat/lighttr/internal/request"
)

func TestNewModel(t *testing.T) {
	model := NewModel()

	// Check initial state
	if model.screen != screenRequest {
		t.Errorf("Expected initial screen to be screenRequest, got %v", model.screen)
	}

	if model.activeInput != 0 {
		t.Errorf("Expected initial active input to be 0, got %d", model.activeInput)
	}

	if len(model.inputs) != 11 {
		t.Errorf("Expected 11 input fields, got %d", len(model.inputs))
	}

	// Check input field configuration
	expectedFields := []struct {
		label       string
		placeholder string
		value       string
	}{
		{label: "URL", placeholder: "https://api.example.com/path", value: ""},
		{label: "Method", placeholder: "GET", value: "GET"},
		{label: "Auth Type (none/basic/apikey/mtls)", placeholder: "none", value: "none"},
		{label: "Auth Username", placeholder: "username", value: ""},
		{label: "Auth Password", placeholder: "password", value: ""},
		{label: "API Key", placeholder: "your-api-key", value: ""},
		{label: "TLS Cert File", placeholder: "/path/to/cert.pem", value: ""},
		{label: "TLS Key File", placeholder: "/path/to/key.pem", value: ""},
		{label: "Headers (key:value,key2:value2)", placeholder: "Content-Type:application/json", value: ""},
		{label: "Query Params (key=value&key2=value2)", placeholder: "key=value&key2=value2", value: ""},
		{label: "Body", placeholder: "{\"key\": \"value\"}", value: ""},
	}

	for i, expected := range expectedFields {
		if model.inputs[i].label != expected.label {
			t.Errorf("Expected input %d label to be %s, got %s", i, expected.label, model.inputs[i].label)
		}
		if model.inputs[i].textinput.Placeholder != expected.placeholder {
			t.Errorf("Expected input %d placeholder to be %s, got %s", i, expected.placeholder, model.inputs[i].textinput.Placeholder)
		}
		if model.inputs[i].textinput.Value() != expected.value {
			t.Errorf("Expected input %d value to be %s, got %s", i, expected.value, model.inputs[i].textinput.Value())
		}
	}

	// Check that only URL field is focused initially
	if !model.inputs[0].textinput.Focused() {
		t.Error("Expected URL field to be focused")
	}
	for i := 1; i < len(model.inputs); i++ {
		if model.inputs[i].textinput.Focused() {
			t.Errorf("Expected input %d to be blurred", i)
		}
	}
}

func TestModel_buildRequestData(t *testing.T) {
	tests := []struct {
		name     string
		inputs   map[int]string
		wantAuth request.AuthData
	}{
		{
			name: "no auth",
			inputs: map[int]string{
				0: "https://api.example.com",
				1: "POST",
				2: "none",
			},
			wantAuth: request.AuthData{
				Type: request.NoAuth,
			},
		},
		{
			name: "basic auth",
			inputs: map[int]string{
				0: "https://api.example.com",
				1: "GET",
				2: "basic",
				3: "testuser",
				4: "testpass",
			},
			wantAuth: request.AuthData{
				Type:     request.BasicAuth,
				Username: "testuser",
				Password: "testpass",
			},
		},
		{
			name: "api key auth",
			inputs: map[int]string{
				0: "https://api.example.com",
				1: "GET",
				2: "apikey",
				5: "test-api-key",
			},
			wantAuth: request.AuthData{
				Type:   request.APIKeyAuth,
				APIKey: "test-api-key",
			},
		},
		{
			name: "mutual TLS auth",
			inputs: map[int]string{
				0: "https://api.example.com",
				1: "GET",
				2: "mtls",
				6: "/path/to/cert.pem",
				7: "/path/to/key.pem",
			},
			wantAuth: request.AuthData{
				Type:     request.MutualTLSAuth,
				CertFile: "/path/to/cert.pem",
				KeyFile:  "/path/to/key.pem",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()

			// Set input values
			for i, value := range tt.inputs {
				model.inputs[i].textinput.SetValue(value)
			}

			// Build request data
			model.buildRequestData()

			// Check auth configuration
			if model.requestData.Auth.Type != tt.wantAuth.Type {
				t.Errorf("Expected auth type %s, got %s", tt.wantAuth.Type, model.requestData.Auth.Type)
			}
			if model.requestData.Auth.Username != tt.wantAuth.Username {
				t.Errorf("Expected username %s, got %s", tt.wantAuth.Username, model.requestData.Auth.Username)
			}
			if model.requestData.Auth.Password != tt.wantAuth.Password {
				t.Errorf("Expected password %s, got %s", tt.wantAuth.Password, model.requestData.Auth.Password)
			}
			if model.requestData.Auth.APIKey != tt.wantAuth.APIKey {
				t.Errorf("Expected API key %s, got %s", tt.wantAuth.APIKey, model.requestData.Auth.APIKey)
			}
			if model.requestData.Auth.CertFile != tt.wantAuth.CertFile {
				t.Errorf("Expected cert file %s, got %s", tt.wantAuth.CertFile, model.requestData.Auth.CertFile)
			}
			if model.requestData.Auth.KeyFile != tt.wantAuth.KeyFile {
				t.Errorf("Expected key file %s, got %s", tt.wantAuth.KeyFile, model.requestData.Auth.KeyFile)
			}
		})
	}
}

func TestModel_Update(t *testing.T) {
	model := NewModel()

	tests := []struct {
		name          string
		msg           tea.Msg
		expectedModel Model
		checkState    func(*testing.T, Model)
	}{
		{
			name: "handle tab key",
			msg:  tea.KeyMsg{Type: tea.KeyTab},
			checkState: func(t *testing.T, m Model) {
				if m.activeInput != 1 {
					t.Errorf("Expected active input to be 1, got %d", m.activeInput)
				}
			},
		},
		{
			name: "handle shift+tab key",
			msg:  tea.KeyMsg{Type: tea.KeyShiftTab},
			checkState: func(t *testing.T, m Model) {
				if m.activeInput != 10 {
					t.Errorf("Expected active input to be 10, got %d", m.activeInput)
				}
			},
		},
		{
			name: "handle enter key in request screen",
			msg:  tea.KeyMsg{Type: tea.KeyEnter},
			checkState: func(t *testing.T, m Model) {
				if m.screen != screenPreview {
					t.Errorf("Expected screen to be screenPreview, got %v", m.screen)
				}
			},
		},
		{
			name: "handle escape key",
			msg:  tea.KeyMsg{Type: tea.KeyEscape},
			checkState: func(t *testing.T, m Model) {
				if m.screen != screenRequest {
					t.Errorf("Expected screen to be screenRequest, got %v", m.screen)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newModel, _ := model.Update(tt.msg)
			updatedModel := newModel.(Model)
			if tt.checkState != nil {
				tt.checkState(t, updatedModel)
			}
		})
	}
}

func TestModel_executeRequest(t *testing.T) {
	model := NewModel()

	// Test with invalid request
	model.requestData = &request.RequestData{
		Method: "GET",
		URL:    "not-a-url",
		Auth:   request.AuthData{Type: request.NoAuth},
	}

	msg := model.executeRequest()
	if err, ok := msg.(error); !ok || err == nil {
		t.Error("Expected error message for invalid request")
	}

	// Test with valid request (using a mock would be better in a real implementation)
	model.requestData = &request.RequestData{
		Method: "GET",
		URL:    "http://example.com",
		Auth:   request.AuthData{Type: request.NoAuth},
	}

	msg = model.executeRequest()
	if _, ok := msg.(*request.ResponseData); !ok {
		t.Error("Expected response data for valid request")
	}
}

func TestModel_View(t *testing.T) {
	model := NewModel()

	// Test request screen
	view := model.View()
	if len(view) == 0 {
		t.Error("Expected non-empty view for request screen")
	}

	// Test preview screen
	model.screen = screenPreview
	model.requestData = &request.RequestData{
		Method: "GET",
		URL:    "https://api.example.com",
	}
	view = model.View()
	if len(view) == 0 {
		t.Error("Expected non-empty view for preview screen")
	}

	// Test response screen
	model.screen = screenResponse
	model.response = &request.ResponseData{
		StatusCode: 200,
		Body:       "test response",
	}
	view = model.View()
	if len(view) == 0 {
		t.Error("Expected non-empty view for response screen")
	}
}
