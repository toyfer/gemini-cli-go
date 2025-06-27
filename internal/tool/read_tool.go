package tool

import (
	"context"
	"fmt"
	"gemini-cli-go/internal/filesystem"
	"gemini-cli-go/internal/shared"
)

// ReadTool implements the Tool interface for reading file content.
type ReadTool struct{}

// Name returns the name of the tool.
func (t *ReadTool) Name() string {
	return "read_file"
}

// Description returns the description of the tool.
func (t *ReadTool) Description() string {
	return "Reads the content of a specified file from the local filesystem."
}

// FunctionDeclaration returns the function declaration for the tool.
func (t *ReadTool) FunctionDeclaration() shared.FunctionDeclaration {
	return shared.FunctionDeclaration{
		Name:		t.Name(),
		Description:	t.Description(),
		Parameters: shared.Schema{
			Type: shared.TypeObject,
			Properties: map[string]shared.Schema{
				"path": {
					Type:        shared.TypeString,
					Description: "The absolute path to the file to read.",
				},
			},
			Required: []string{"path"},
		},
	}
}

// Execute executes the read_file tool.
func (t *ReadTool) Execute(_ context.Context, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("missing or invalid 'path' argument for read_file tool")
	}

	content, err := filesystem.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return fmt.Sprintf("File content of %s:\n%s", path, content), nil
}
