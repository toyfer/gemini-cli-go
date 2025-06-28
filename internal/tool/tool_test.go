package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gemini-cli-go/internal/shared"
)

func TestNewToolRegistry(t *testing.T) {
	registry := NewToolRegistry()
	
	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}
	if registry.tools == nil {
		t.Fatal("Expected non-nil tools map")
	}
	if len(registry.tools) != 0 {
		t.Errorf("Expected empty tools map, got %d tools", len(registry.tools))
	}
}

func TestToolRegistryRegisterTool(t *testing.T) {
	registry := NewToolRegistry()
	tool := &MockTool{name: "test_tool"}

	registry.RegisterTool(tool)

	if len(registry.tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(registry.tools))
	}
	
	registeredTool, exists := registry.tools["test_tool"]
	if !exists {
		t.Error("Expected tool to be registered")
	}
	if registeredTool != tool {
		t.Error("Expected registered tool to be the same instance")
	}
}

func TestToolRegistryGetTool(t *testing.T) {
	registry := NewToolRegistry()
	tool := &MockTool{name: "test_tool"}
	registry.RegisterTool(tool)

	// Test getting existing tool
	retrievedTool, ok := registry.GetTool("test_tool")
	if !ok {
		t.Error("Expected to find registered tool")
	}
	if retrievedTool != tool {
		t.Error("Expected retrieved tool to be the same instance")
	}

	// Test getting non-existent tool
	_, ok = registry.GetTool("non_existent")
	if ok {
		t.Error("Expected not to find non-existent tool")
	}
}

func TestToolRegistryGetFunctionDeclarations(t *testing.T) {
	registry := NewToolRegistry()
	
	tool1 := &MockTool{name: "tool1", description: "First tool"}
	tool2 := &MockTool{name: "tool2", description: "Second tool"}
	
	registry.RegisterTool(tool1)
	registry.RegisterTool(tool2)

	declarations := registry.GetFunctionDeclarations()
	
	if len(declarations) != 2 {
		t.Errorf("Expected 2 function declarations, got %d", len(declarations))
	}

	// Check that both tools are represented (order doesn't matter)
	foundTool1, foundTool2 := false, false
	for _, decl := range declarations {
		if decl.Name == "tool1" {
			foundTool1 = true
		}
		if decl.Name == "tool2" {
			foundTool2 = true
		}
	}
	
	if !foundTool1 {
		t.Error("Expected to find tool1 in function declarations")
	}
	if !foundTool2 {
		t.Error("Expected to find tool2 in function declarations")
	}
}

func TestReadTool(t *testing.T) {
	tool := &ReadTool{}

	// Test Name
	if tool.Name() != "read_file" {
		t.Errorf("Expected name 'read_file', got '%s'", tool.Name())
	}

	// Test Description
	expectedDesc := "Reads the content of a specified file from the local filesystem."
	if tool.Description() != expectedDesc {
		t.Errorf("Expected description '%s', got '%s'", expectedDesc, tool.Description())
	}

	// Test FunctionDeclaration
	fd := tool.FunctionDeclaration()
	if fd.Name != "read_file" {
		t.Errorf("Expected function declaration name 'read_file', got '%s'", fd.Name)
	}
	if fd.Parameters.Type != shared.TypeObject {
		t.Errorf("Expected parameters type TypeObject, got %v", fd.Parameters.Type)
	}
	if len(fd.Parameters.Required) != 1 || fd.Parameters.Required[0] != "path" {
		t.Errorf("Expected required parameters ['path'], got %v", fd.Parameters.Required)
	}
}

func TestReadToolExecute(t *testing.T) {
	tool := &ReadTool{}

	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "read_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := "test file content"
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Test successful read
	args := map[string]interface{}{
		"path": tmpfile.Name(),
	}
	output, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !containsString(output, content) {
		t.Errorf("Expected output to contain '%s', got '%s'", content, output)
	}

	// Test with missing path argument
	args = map[string]interface{}{}
	_, err = tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("Expected error for missing path argument")
	}

	// Test with invalid path argument type
	args = map[string]interface{}{
		"path": 123,
	}
	_, err = tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("Expected error for invalid path argument type")
	}

	// Test with non-existent file
	args = map[string]interface{}{
		"path": "/non/existent/file",
	}
	_, err = tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestWriteFileTool(t *testing.T) {
	tool := &WriteFileTool{}

	// Test Name
	if tool.Name() != "write_file" {
		t.Errorf("Expected name 'write_file', got '%s'", tool.Name())
	}

	// Test Description
	expectedDesc := "Writes content to a specified file. If the file does not exist, it will be created. If it exists, its content will be truncated."
	if tool.Description() != expectedDesc {
		t.Errorf("Expected description '%s', got '%s'", expectedDesc, tool.Description())
	}

	// Test FunctionDeclaration
	fd := tool.FunctionDeclaration()
	if fd.Name != "write_file" {
		t.Errorf("Expected function declaration name 'write_file', got '%s'", fd.Name)
	}
	if fd.Parameters.Type != shared.TypeObject {
		t.Errorf("Expected parameters type TypeObject, got %v", fd.Parameters.Type)
	}
	if len(fd.Parameters.Required) != 2 {
		t.Errorf("Expected 2 required parameters, got %d", len(fd.Parameters.Required))
	}
}

func TestWriteFileToolExecute(t *testing.T) {
	tool := &WriteFileTool{}

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "write_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "test file content"

	// Test successful write
	args := map[string]interface{}{
		"filePath": testFile,
		"content":  testContent,
	}
	output, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !containsString(output, testFile) {
		t.Errorf("Expected output to contain file path '%s', got '%s'", testFile, output)
	}

	// Verify file was created and has correct content
	readContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}
	if string(readContent) != testContent {
		t.Errorf("Expected file content '%s', got '%s'", testContent, string(readContent))
	}

	// Test with missing filePath argument
	args = map[string]interface{}{
		"content": "some content",
	}
	_, err = tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("Expected error for missing filePath argument")
	}

	// Test with missing content argument
	args = map[string]interface{}{
		"filePath": testFile,
	}
	_, err = tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("Expected error for missing content argument")
	}

	// Test with invalid filePath argument type
	args = map[string]interface{}{
		"filePath": 123,
		"content":  "content",
	}
	_, err = tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("Expected error for invalid filePath argument type")
	}
}

func TestListFilesTool(t *testing.T) {
	tool := &ListFilesTool{}

	// Test Name
	if tool.Name() != "list_files" {
		t.Errorf("Expected name 'list_files', got '%s'", tool.Name())
	}

	// Test Description
	expectedDesc := "Lists files in a directory, optionally filtered by extension."
	if tool.Description() != expectedDesc {
		t.Errorf("Expected description '%s', got '%s'", expectedDesc, tool.Description())
	}

	// Test FunctionDeclaration
	fd := tool.FunctionDeclaration()
	if fd.Name != "list_files" {
		t.Errorf("Expected function declaration name 'list_files', got '%s'", fd.Name)
	}
	if fd.Parameters.Type != shared.TypeObject {
		t.Errorf("Expected parameters type TypeObject, got %v", fd.Parameters.Type)
	}
	if len(fd.Parameters.Required) != 1 || fd.Parameters.Required[0] != "dir" {
		t.Errorf("Expected required parameters ['dir'], got %v", fd.Parameters.Required)
	}
}

func TestListFilesToolExecute(t *testing.T) {
	tool := &ListFilesTool{}

	// Create a temporary directory with test files
	tmpDir, err := os.MkdirTemp("", "list_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := []string{"test.txt", "test.go", "readme.md"}
	for _, filename := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Test listing all files
	args := map[string]interface{}{
		"dir": tmpDir,
	}
	output, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	for _, filename := range testFiles {
		if !containsString(output, filename) {
			t.Errorf("Expected output to contain '%s', got '%s'", filename, output)
		}
	}

	// Test with extension filter
	args = map[string]interface{}{
		"dir": tmpDir,
		"ext": []interface{}{".go"},
	}
	output, err = tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !containsString(output, "test.go") {
		t.Errorf("Expected output to contain 'test.go', got '%s'", output)
	}

	// Test with missing dir argument
	args = map[string]interface{}{}
	_, err = tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("Expected error for missing dir argument")
	}

	// Test with non-existent directory
	args = map[string]interface{}{
		"dir": "/non/existent/directory",
	}
	_, err = tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

// MockTool implements the Tool interface for testing
type MockTool struct {
	name        string
	description string
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Description() string {
	return m.description
}

func (m *MockTool) FunctionDeclaration() shared.FunctionDeclaration {
	return shared.FunctionDeclaration{
		Name:        m.name,
		Description: m.description,
		Parameters: shared.Schema{
			Type: shared.TypeObject,
		},
	}
}

func (m *MockTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	return "mock output", nil
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}