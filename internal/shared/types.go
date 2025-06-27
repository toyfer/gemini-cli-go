package shared

import (
	"context"
)

// FunctionDeclaration represents a function call made by the model.
type FunctionDeclaration struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  Schema `json:"parameters"`
}

// Tools represents a list of function declarations.
type Tools struct {
	FunctionDeclarations []FunctionDeclaration `json:"function_declarations"`
}

// FunctionCall represents a function call from the model.
type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// Schema represents the schema of a function's parameters.
type Schema struct {
	Type        Type              `json:"type"`
	Properties  map[string]Schema `json:"properties,omitempty"`
	Required    []string          `json:"required,omitempty"`
	Description string            `json:"description,omitempty"`
	Items       *Schema           `json:"items,omitempty"` // Add Items field for array type
}

// Type represents the data type of a schema property.
type Type string

const (
	TypeString  Type = "string"
	TypeNumber  Type = "number"
	TypeInteger Type = "integer"
	TypeBoolean Type = "boolean"
	TypeArray   Type = "array"
	TypeObject  Type = "object"
)

// ToolRegistryInterface defines the methods that RunNonInteractive needs from a ToolRegistry.
type ToolRegistryInterface interface {
	GetFunctionDeclarations() []FunctionDeclaration
	GetTool(name string) (Tool, bool)
}

// Tool represents a callable tool that the Gemini model can use.
type Tool interface {
	Name() string
	Description() string
	FunctionDeclaration() FunctionDeclaration
	Execute(ctx context.Context, args map[string]interface{}) (string, error)
}