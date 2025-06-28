package shared

import (
	"context"
	"testing"
)

func TestFunctionDeclaration(t *testing.T) {
	// Test creating a FunctionDeclaration
	fd := FunctionDeclaration{
		Name:        "test_function",
		Description: "A test function",
		Parameters: Schema{
			Type: TypeObject,
			Properties: map[string]Schema{
				"param1": {
					Type:        TypeString,
					Description: "First parameter",
				},
			},
			Required: []string{"param1"},
		},
	}

	// Verify the fields are set correctly
	if fd.Name != "test_function" {
		t.Errorf("Expected name 'test_function', got '%s'", fd.Name)
	}
	if fd.Description != "A test function" {
		t.Errorf("Expected description 'A test function', got '%s'", fd.Description)
	}
	if fd.Parameters.Type != TypeObject {
		t.Errorf("Expected type TypeObject, got %v", fd.Parameters.Type)
	}
}

func TestTools(t *testing.T) {
	// Test creating Tools with function declarations
	tools := Tools{
		FunctionDeclarations: []FunctionDeclaration{
			{
				Name:        "function1",
				Description: "First function",
				Parameters: Schema{
					Type: TypeObject,
				},
			},
			{
				Name:        "function2",
				Description: "Second function",
				Parameters: Schema{
					Type: TypeObject,
				},
			},
		},
	}

	if len(tools.FunctionDeclarations) != 2 {
		t.Errorf("Expected 2 function declarations, got %d", len(tools.FunctionDeclarations))
	}
	if tools.FunctionDeclarations[0].Name != "function1" {
		t.Errorf("Expected first function name 'function1', got '%s'", tools.FunctionDeclarations[0].Name)
	}
}

func TestFunctionCall(t *testing.T) {
	// Test creating a FunctionCall
	fc := FunctionCall{
		Name: "test_call",
		Args: map[string]interface{}{
			"param1": "value1",
			"param2": 42,
		},
	}

	// Verify the fields are set correctly
	if fc.Name != "test_call" {
		t.Errorf("Expected name 'test_call', got '%s'", fc.Name)
	}
	if len(fc.Args) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(fc.Args))
	}
	if fc.Args["param1"] != "value1" {
		t.Errorf("Expected param1 to be 'value1', got %v", fc.Args["param1"])
	}
	if fc.Args["param2"] != 42 {
		t.Errorf("Expected param2 to be 42, got %v", fc.Args["param2"])
	}
}

func TestSchema(t *testing.T) {
	// Test creating a complex Schema
	schema := Schema{
		Type:        TypeObject,
		Description: "Test schema",
		Properties: map[string]Schema{
			"stringProp": {
				Type:        TypeString,
				Description: "A string property",
			},
			"arrayProp": {
				Type:        TypeArray,
				Description: "An array property",
				Items: &Schema{
					Type: TypeString,
				},
			},
		},
		Required: []string{"stringProp"},
	}

	// Verify the schema structure
	if schema.Type != TypeObject {
		t.Errorf("Expected type TypeObject, got %v", schema.Type)
	}
	if len(schema.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(schema.Properties))
	}
	if schema.Properties["stringProp"].Type != TypeString {
		t.Errorf("Expected stringProp type TypeString, got %v", schema.Properties["stringProp"].Type)
	}
	if schema.Properties["arrayProp"].Items == nil {
		t.Error("Expected arrayProp to have Items defined")
	} else if schema.Properties["arrayProp"].Items.Type != TypeString {
		t.Errorf("Expected arrayProp items type TypeString, got %v", schema.Properties["arrayProp"].Items.Type)
	}
	if len(schema.Required) != 1 || schema.Required[0] != "stringProp" {
		t.Errorf("Expected required fields ['stringProp'], got %v", schema.Required)
	}
}

func TestTypeConstants(t *testing.T) {
	// Test that all type constants are defined correctly
	expectedTypes := []Type{
		TypeString,
		TypeNumber,
		TypeInteger,
		TypeBoolean,
		TypeArray,
		TypeObject,
	}

	expectedValues := []string{
		"string",
		"number",
		"integer",
		"boolean",
		"array",
		"object",
	}

	for i, expectedType := range expectedTypes {
		if string(expectedType) != expectedValues[i] {
			t.Errorf("Expected type %v to have value '%s', got '%s'", expectedType, expectedValues[i], string(expectedType))
		}
	}
}

// MockTool implements the Tool interface for testing
type MockTool struct {
	name        string
	description string
	executed    bool
	output      string
	err         error
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Description() string {
	return m.description
}

func (m *MockTool) FunctionDeclaration() FunctionDeclaration {
	return FunctionDeclaration{
		Name:        m.name,
		Description: m.description,
		Parameters: Schema{
			Type: TypeObject,
		},
	}
}

func (m *MockTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	m.executed = true
	return m.output, m.err
}

func TestToolInterface(t *testing.T) {
	// Test that MockTool properly implements the Tool interface
	tool := &MockTool{
		name:        "mock_tool",
		description: "A mock tool for testing",
		output:      "test output",
		err:         nil,
	}

	// Test Tool interface methods
	if tool.Name() != "mock_tool" {
		t.Errorf("Expected name 'mock_tool', got '%s'", tool.Name())
	}
	if tool.Description() != "A mock tool for testing" {
		t.Errorf("Expected description 'A mock tool for testing', got '%s'", tool.Description())
	}

	fd := tool.FunctionDeclaration()
	if fd.Name != "mock_tool" {
		t.Errorf("Expected function declaration name 'mock_tool', got '%s'", fd.Name)
	}

	// Test Execute method
	output, err := tool.Execute(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if output != "test output" {
		t.Errorf("Expected output 'test output', got '%s'", output)
	}
	if !tool.executed {
		t.Error("Expected tool to be marked as executed")
	}
}

// MockToolRegistry implements ToolRegistryInterface for testing
type MockToolRegistry struct {
	tools map[string]Tool
}

func NewMockToolRegistry() *MockToolRegistry {
	return &MockToolRegistry{
		tools: make(map[string]Tool),
	}
}

func (m *MockToolRegistry) RegisterTool(tool Tool) {
	m.tools[tool.Name()] = tool
}

func (m *MockToolRegistry) GetTool(name string) (Tool, bool) {
	tool, ok := m.tools[name]
	return tool, ok
}

func (m *MockToolRegistry) GetFunctionDeclarations() []FunctionDeclaration {
	declarations := make([]FunctionDeclaration, 0, len(m.tools))
	for _, tool := range m.tools {
		declarations = append(declarations, tool.FunctionDeclaration())
	}
	return declarations
}

func TestToolRegistryInterface(t *testing.T) {
	// Test MockToolRegistry implements ToolRegistryInterface
	registry := NewMockToolRegistry()
	
	// Register a tool
	tool := &MockTool{
		name:        "test_tool",
		description: "Test tool",
	}
	registry.RegisterTool(tool)

	// Test GetTool
	retrievedTool, ok := registry.GetTool("test_tool")
	if !ok {
		t.Error("Expected to find registered tool")
	}
	if retrievedTool.Name() != "test_tool" {
		t.Errorf("Expected retrieved tool name 'test_tool', got '%s'", retrievedTool.Name())
	}

	// Test GetTool for non-existent tool
	_, ok = registry.GetTool("non_existent")
	if ok {
		t.Error("Expected not to find non-existent tool")
	}

	// Test GetFunctionDeclarations
	declarations := registry.GetFunctionDeclarations()
	if len(declarations) != 1 {
		t.Errorf("Expected 1 function declaration, got %d", len(declarations))
	}
	if declarations[0].Name != "test_tool" {
		t.Errorf("Expected function declaration name 'test_tool', got '%s'", declarations[0].Name)
	}
}