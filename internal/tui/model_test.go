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

	if len(model.inputs) != 5 {
		t.Errorf("Expected 5 input fields, got %d", len(model.inputs))
	}

	// Check input field configuration
	if model.inputs[0].label != "URL" {
		t.Errorf("Expected first input label to be URL, got %s", model.inputs[0].label)
	}

	if model.inputs[1].textinput.Value() != "GET" {
		t.Errorf("Expected method input default value to be GET, got %s", model.inputs[1].textinput.Value())
	}
}

func TestModel_buildRequestData(t *testing.T) {
	model := NewModel()

	// Set input values
	model.inputs[0].textinput.SetValue("https://api.example.com")
	model.inputs[1].textinput.SetValue("POST")
	model.inputs[2].textinput.SetValue("Content-Type:application/json,Authorization:Bearer token")
	model.inputs[3].textinput.SetValue("key1=value1&key2=value2")
	model.inputs[4].textinput.SetValue(`{"data":"test"}`)

	// Build request data
	model.buildRequestData()

	// Verify request data
	if model.requestData.URL != "https://api.example.com" {
		t.Errorf("Expected URL https://api.example.com, got %s", model.requestData.URL)
	}

	if model.requestData.Method != "POST" {
		t.Errorf("Expected method POST, got %s", model.requestData.Method)
	}

	// Check headers
	expectedHeaders := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer token",
	}
	for k, v := range expectedHeaders {
		if model.requestData.Headers[k] != v {
			t.Errorf("Expected header %s to be %s, got %s", k, v, model.requestData.Headers[k])
		}
	}

	// Check query params
	expectedParams := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	for k, v := range expectedParams {
		if model.requestData.QueryParams[k] != v {
			t.Errorf("Expected query param %s to be %s, got %s", k, v, model.requestData.QueryParams[k])
		}
	}

	if model.requestData.Body != `{"data":"test"}` {
		t.Errorf("Expected body {\"data\":\"test\"}, got %s", model.requestData.Body)
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
				if m.activeInput != 4 {
					t.Errorf("Expected active input to be 4, got %d", m.activeInput)
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
	}

	msg := model.executeRequest()
	if err, ok := msg.(error); !ok || err == nil {
		t.Error("Expected error message for invalid request")
	}

	// Test with valid request (using a mock would be better in a real implementation)
	model.requestData = &request.RequestData{
		Method: "GET",
		URL:    "http://example.com",
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
