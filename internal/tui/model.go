package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nshekhawat/lighttr/internal/request"
)

var (
	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	blurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(1, 2)
)

type inputField struct {
	textinput textinput.Model
	label     string
}

type screen int

const (
	screenRequest screen = iota
	screenPreview
	screenResponse
)

type Model struct {
	inputs      []inputField
	activeInput int
	requestData *request.RequestData
	response    *request.ResponseData
	screen      screen
	viewport    viewport.Model
	err         error
	authType    request.AuthType
}

func NewModel() Model {
	inputs := []inputField{
		{label: "URL", textinput: textinput.New()},
		{label: "Method", textinput: textinput.New()},
		{label: "Auth Type (none/basic/apikey/mtls)", textinput: textinput.New()},
		{label: "Auth Username", textinput: textinput.New()},
		{label: "Auth Password", textinput: textinput.New()},
		{label: "API Key", textinput: textinput.New()},
		{label: "TLS Cert File", textinput: textinput.New()},
		{label: "TLS Key File", textinput: textinput.New()},
		{label: "Headers (key:value,key2:value2)", textinput: textinput.New()},
		{label: "Query Params (key=value&key2=value2)", textinput: textinput.New()},
		{label: "Body", textinput: textinput.New()},
	}

	// Configure inputs
	for i := range inputs {
		if i == 0 { // Only focus the URL field
			inputs[i].textinput.Focus()
			inputs[i].textinput.PromptStyle = focusedStyle
			inputs[i].textinput.TextStyle = focusedStyle
		} else {
			inputs[i].textinput.Blur()
			inputs[i].textinput.PromptStyle = blurredStyle
			inputs[i].textinput.TextStyle = blurredStyle
		}
	}

	// Configure input placeholders
	inputs[0].textinput.Placeholder = "https://api.example.com/path"
	inputs[1].textinput.Placeholder = "GET"
	inputs[1].textinput.SetValue("GET")
	inputs[2].textinput.Placeholder = "none"
	inputs[2].textinput.SetValue("none")
	inputs[3].textinput.Placeholder = "username"
	inputs[4].textinput.Placeholder = "password"
	inputs[4].textinput.EchoMode = textinput.EchoPassword
	inputs[5].textinput.Placeholder = "your-api-key"
	inputs[6].textinput.Placeholder = "/path/to/cert.pem"
	inputs[7].textinput.Placeholder = "/path/to/key.pem"
	inputs[8].textinput.Placeholder = "Content-Type:application/json"
	inputs[9].textinput.Placeholder = "key=value&key2=value2"
	inputs[10].textinput.Placeholder = "{\"key\": \"value\"}"

	return Model{
		inputs:      inputs,
		activeInput: 0,
		requestData: request.NewRequestData(),
		screen:      screenRequest,
		viewport:    viewport.New(0, 0),
		authType:    request.NoAuth,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case error:
		// Handle error messages
		m.err = msg
		return m, nil
	case *request.ResponseData:
		// Handle the response from request execution
		m.response = msg
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab", "shift+tab", "up", "down":
			// Handle navigation between inputs
			if m.screen == screenRequest {
				s := msg.String()

				if s == "up" || s == "shift+tab" {
					m.activeInput--
				} else {
					m.activeInput++
				}

				if m.activeInput >= len(m.inputs) {
					m.activeInput = 0
				} else if m.activeInput < 0 {
					m.activeInput = len(m.inputs) - 1
				}

				for i := range m.inputs {
					if i == m.activeInput {
						m.inputs[i].textinput.Focus()
						continue
					}
					m.inputs[i].textinput.Blur()
				}

				return m, nil
			}

		case "esc":
			if m.screen != screenRequest {
				m.screen = screenRequest
				m.response = nil // Clear the response when going back
				m.err = nil      // Clear any errors
				return m, nil
			}

		case "enter":
			switch m.screen {
			case screenRequest:
				// Build request data
				m.buildRequestData()
				m.screen = screenPreview
				return m, nil
			case screenPreview:
				// Execute request
				m.screen = screenResponse
				m.response = nil // Clear previous response
				m.err = nil      // Clear previous errors
				return m, m.executeRequest
			}
		}
	}

	// Handle viewport updates
	if m.screen == screenResponse {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Handle input updates
	if m.screen == screenRequest {
		for i := range m.inputs {
			m.inputs[i].textinput, cmd = m.inputs[i].textinput.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) buildRequestData() {
	m.requestData = request.NewRequestData()
	m.requestData.URL = m.inputs[0].textinput.Value()
	m.requestData.Method = m.inputs[1].textinput.Value()

	// Handle authentication
	authType := request.AuthType(m.inputs[2].textinput.Value())
	m.requestData.Auth = request.AuthData{
		Type: authType,
	}

	switch authType {
	case request.BasicAuth:
		m.requestData.Auth.Username = m.inputs[3].textinput.Value()
		m.requestData.Auth.Password = m.inputs[4].textinput.Value()
	case request.APIKeyAuth:
		m.requestData.Auth.APIKey = m.inputs[5].textinput.Value()
	case request.MutualTLSAuth:
		m.requestData.Auth.CertFile = m.inputs[6].textinput.Value()
		m.requestData.Auth.KeyFile = m.inputs[7].textinput.Value()
	}

	// Parse headers
	if headers := m.inputs[8].textinput.Value(); headers != "" {
		for _, header := range strings.Split(headers, ",") {
			parts := strings.SplitN(header, ":", 2)
			if len(parts) == 2 {
				m.requestData.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Parse query params
	if params := m.inputs[9].textinput.Value(); params != "" {
		for _, param := range strings.Split(params, "&") {
			parts := strings.SplitN(param, "=", 2)
			if len(parts) == 2 {
				m.requestData.QueryParams[parts[0]] = parts[1]
			}
		}
	}

	m.requestData.Body = m.inputs[10].textinput.Value()
}

func (m Model) executeRequest() tea.Msg {
	// Validate request data first
	if err := m.requestData.Validate(); err != nil {
		return fmt.Errorf("invalid request: %v", err)
	}

	resp, err := m.requestData.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute request: %v", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("request error: %s", resp.Error)
	}

	return resp
}

func (m Model) View() string {
	switch m.screen {
	case screenRequest:
		return m.renderRequestScreen()
	case screenPreview:
		return m.renderPreviewScreen()
	case screenResponse:
		return m.renderResponseScreen()
	default:
		return "Unknown screen"
	}
}

func (m Model) renderRequestScreen() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Lighttr - HTTP Request Builder"))
	b.WriteString("\n\n")

	// Get current auth type
	currentAuthType := request.AuthType(m.inputs[2].textinput.Value())

	for i, input := range m.inputs {
		// Skip auth fields that aren't relevant for the current auth type
		if shouldSkipAuthField(i, currentAuthType) {
			continue
		}

		style := blurredStyle
		if i == m.activeInput {
			style = focusedStyle
		}
		b.WriteString(style.Render(input.label) + "\n")
		b.WriteString(input.textinput.View() + "\n\n")
	}

	b.WriteString("\nPress Enter to preview request • ESC to go back • Ctrl+C to quit\n")
	return b.String()
}

// shouldSkipAuthField determines if an auth-related field should be shown based on the current auth type
func shouldSkipAuthField(fieldIndex int, authType request.AuthType) bool {
	switch authType {
	case request.NoAuth:
		// Hide all auth fields except the auth type selector
		return fieldIndex >= 3 && fieldIndex <= 7
	case request.BasicAuth:
		// Show only username and password fields
		return (fieldIndex >= 5 && fieldIndex <= 7)
	case request.APIKeyAuth:
		// Show only API key field
		return (fieldIndex >= 3 && fieldIndex <= 4) || (fieldIndex >= 6 && fieldIndex <= 7)
	case request.MutualTLSAuth:
		// Show only cert and key file fields
		return (fieldIndex >= 3 && fieldIndex <= 5)
	default:
		return false
	}
}

func (m Model) renderPreviewScreen() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Request Preview"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("%s %s\n", m.requestData.Method, m.requestData.URL))

	// Show authentication details
	b.WriteString(fmt.Sprintf("\nAuthentication: %s\n", m.requestData.Auth.Type))
	switch m.requestData.Auth.Type {
	case request.BasicAuth:
		b.WriteString(fmt.Sprintf("Username: %s\n", m.requestData.Auth.Username))
		b.WriteString("Password: ********\n")
	case request.APIKeyAuth:
		b.WriteString("API Key: ********\n")
	case request.MutualTLSAuth:
		b.WriteString(fmt.Sprintf("Certificate File: %s\n", m.requestData.Auth.CertFile))
		b.WriteString(fmt.Sprintf("Key File: %s\n", m.requestData.Auth.KeyFile))
	}

	if len(m.requestData.Headers) > 0 {
		b.WriteString("\nHeaders:\n")
		for k, v := range m.requestData.Headers {
			b.WriteString(fmt.Sprintf("%s: %s\n", k, v))
		}
	}

	if len(m.requestData.QueryParams) > 0 {
		b.WriteString("\nQuery Parameters:\n")
		for k, v := range m.requestData.QueryParams {
			b.WriteString(fmt.Sprintf("%s=%s\n", k, v))
		}
	}

	if m.requestData.Body != "" {
		b.WriteString("\nBody:\n")
		b.WriteString(m.requestData.Body)
	}

	b.WriteString("\n\nPress Enter to send request • ESC to go back • Ctrl+C to quit\n")
	return b.String()
}

func (m Model) renderResponseScreen() string {
	var b strings.Builder

	if m.err != nil {
		b.WriteString(titleStyle.Render("Error"))
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("Error: %v\n", m.err))
		return b.String()
	}

	if m.response == nil {
		return "Loading..."
	}

	b.WriteString(titleStyle.Render("Response"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Status: %d\n", m.response.StatusCode))
	b.WriteString(fmt.Sprintf("Time: %v\n", m.response.ResponseTime))

	if len(m.response.Headers) > 0 {
		b.WriteString("\nHeaders:\n")
		for k, v := range m.response.Headers {
			b.WriteString(fmt.Sprintf("%s: %s\n", k, v))
		}
	}

	if m.response.Body != "" {
		b.WriteString("\nBody:\n")
		b.WriteString(m.response.Body)
	}

	b.WriteString("\n\nESC to go back • Ctrl+C to quit\n")
	return b.String()
}
