package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nshekhawat/lighttr/internal/request"
)

// Manager handles the storage and retrieval of request history
type Manager struct {
	filePath string
	history  []request.RequestData
}

// NewManager creates a new history manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Create .lighttr directory if it doesn't exist
	lighttrDir := filepath.Join(homeDir, ".lighttr")
	if err := os.MkdirAll(lighttrDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(lighttrDir, "history.json")
	manager := &Manager{
		filePath: filePath,
		history:  make([]request.RequestData, 0),
	}

	// Load existing history if it exists
	if err := manager.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return manager, nil
}

// Add adds a new request to history
func (m *Manager) Add(req request.RequestData) error {
	m.history = append(m.history, req)
	return m.save()
}

// GetAll returns all historical requests
func (m *Manager) GetAll() []request.RequestData {
	return m.history
}

// Clear removes all history
func (m *Manager) Clear() error {
	m.history = make([]request.RequestData, 0)
	return m.save()
}

// load reads the history from disk
func (m *Manager) load() error {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &m.history)
}

// save writes the history to disk
func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %v", err)
	}

	return os.WriteFile(m.filePath, data, 0644)
}
