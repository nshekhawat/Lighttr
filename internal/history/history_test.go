package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nshekhawat/lighttr/internal/request"
)

func TestNewManager(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "lighttr-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Check if .lighttr directory was created
	lighttrDir := filepath.Join(tmpDir, ".lighttr")
	if _, err := os.Stat(lighttrDir); os.IsNotExist(err) {
		t.Error("Expected .lighttr directory to be created")
	}

	// Check if history file path is set correctly
	expectedPath := filepath.Join(lighttrDir, "history.json")
	if manager.filePath != expectedPath {
		t.Errorf("Expected file path %s, got %s", expectedPath, manager.filePath)
	}

	// Check if history slice is initialized
	if manager.history == nil {
		t.Error("Expected history slice to be initialized")
	}
}

func TestManager_AddAndGetAll(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "lighttr-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Create test request data
	req1 := request.RequestData{
		Method:    "GET",
		URL:       "https://api.example.com/1",
		Timestamp: time.Now(),
	}
	req2 := request.RequestData{
		Method:    "POST",
		URL:       "https://api.example.com/2",
		Timestamp: time.Now(),
	}

	// Add requests to history
	if err := manager.Add(req1); err != nil {
		t.Errorf("Add() error = %v", err)
	}
	if err := manager.Add(req2); err != nil {
		t.Errorf("Add() error = %v", err)
	}

	// Get all requests
	history := manager.GetAll()
	if len(history) != 2 {
		t.Errorf("Expected 2 items in history, got %d", len(history))
	}

	// Verify the requests were saved correctly
	if history[0].Method != req1.Method || history[0].URL != req1.URL {
		t.Error("First request not saved correctly")
	}
	if history[1].Method != req2.Method || history[1].URL != req2.URL {
		t.Error("Second request not saved correctly")
	}

	// Verify the history was persisted to disk
	data, err := os.ReadFile(manager.filePath)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}

	var savedHistory []request.RequestData
	if err := json.Unmarshal(data, &savedHistory); err != nil {
		t.Fatalf("Failed to unmarshal history file: %v", err)
	}

	if len(savedHistory) != 2 {
		t.Errorf("Expected 2 items in saved history, got %d", len(savedHistory))
	}
}

func TestManager_Clear(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "lighttr-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Add a request to history
	req := request.RequestData{
		Method:    "GET",
		URL:       "https://api.example.com",
		Timestamp: time.Now(),
	}
	if err := manager.Add(req); err != nil {
		t.Errorf("Add() error = %v", err)
	}

	// Clear history
	if err := manager.Clear(); err != nil {
		t.Errorf("Clear() error = %v", err)
	}

	// Verify history is empty
	history := manager.GetAll()
	if len(history) != 0 {
		t.Errorf("Expected empty history after clear, got %d items", len(history))
	}

	// Verify history file is empty array
	data, err := os.ReadFile(manager.filePath)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}

	if string(data) != "[]" {
		t.Errorf("Expected empty array in history file, got %s", string(data))
	}
}
