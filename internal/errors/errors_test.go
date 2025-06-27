package errors

import (
	"testing"
)

func TestSettingError(t *testing.T) {
	// Test creating a SettingError
	err := SettingError{
		Message: "Test error message",
		Path:    "/path/to/settings.json",
	}

	// Verify the fields are set correctly
	if err.Message != "Test error message" {
		t.Errorf("Expected message 'Test error message', got '%s'", err.Message)
	}
	if err.Path != "/path/to/settings.json" {
		t.Errorf("Expected path '/path/to/settings.json', got '%s'", err.Path)
	}
}

func TestSettingErrorEmpty(t *testing.T) {
	// Test creating an empty SettingError
	err := SettingError{}

	// Verify empty values
	if err.Message != "" {
		t.Errorf("Expected empty message, got '%s'", err.Message)
	}
	if err.Path != "" {
		t.Errorf("Expected empty path, got '%s'", err.Path)
	}
}

func TestSettingErrorSlice(t *testing.T) {
	// Test working with a slice of SettingError
	errors := []SettingError{
		{Message: "Error 1", Path: "/path1"},
		{Message: "Error 2", Path: "/path2"},
	}

	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}

	if errors[0].Message != "Error 1" {
		t.Errorf("Expected first error message 'Error 1', got '%s'", errors[0].Message)
	}
	if errors[1].Path != "/path2" {
		t.Errorf("Expected second error path '/path2', got '%s'", errors[1].Path)
	}
}