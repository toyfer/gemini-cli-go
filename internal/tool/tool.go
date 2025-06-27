package tool

import (
	"gemini-cli-go/internal/shared"
)

// ToolRegistry manages available tools.
type ToolRegistry struct {
	tools map[string]shared.Tool
}

// NewToolRegistry creates a new ToolRegistry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]shared.Tool),
	}
}

// RegisterTool registers a tool with the registry.
func (tr *ToolRegistry) RegisterTool(tool shared.Tool) {
	tr.tools[tool.Name()] = tool
}

// GetTool retrieves a tool by its name.
func (tr *ToolRegistry) GetTool(name string) (shared.Tool, bool) {
	tool, ok := tr.tools[name]
	return tool, ok
}

// GetFunctionDeclarations returns all registered tool's function declarations.
func (tr *ToolRegistry) GetFunctionDeclarations() []shared.FunctionDeclaration {
	declarations := make([]shared.FunctionDeclaration, 0, len(tr.tools))
	for _, tool := range tr.tools {
		declarations = append(declarations, tool.FunctionDeclaration())
	}
	return declarations
}