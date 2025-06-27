package tool

import (
	"context"
	"fmt"
	"gemini-cli-go/internal/filesystem"
	"gemini-cli-go/internal/shared"
)

// WriteFileTool implements the Tool interface for writing content to a file.
type WriteFileTool struct{}

// Name returns the name of the tool.
func (t *WriteFileTool) Name() string {
	return "write_file"
}

// Description returns a description of the tool.
func (t *WriteFileTool) Description() string {
	return "Writes content to a specified file. If the file does not exist, it will be created. If it exists, its content will be truncated."
}

// FunctionDeclaration returns the function declaration for the tool.
func (t *WriteFileTool) FunctionDeclaration() shared.FunctionDeclaration {
	return shared.FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: shared.Schema{
			Type: shared.TypeObject,
			Properties: map[string]shared.Schema{
				"filePath": {
					Type:        shared.TypeString,
					Description: "The absolute path to the file to write to.",
				},
				"content": {
					Type:        shared.TypeString,
					Description: "The content to write to the file.",
				},
			},
			Required: []string{"filePath", "content"},
		},
	}
}

// Execute executes the write_file tool.
func (t *WriteFileTool) Execute(_ context.Context, args map[string]interface{}) (string, error) {
	filePath, ok := args["filePath"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid 'filePath' argument")
	}
	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid 'content' argument")
	}

	err := filesystem.WriteFile(filePath, []byte(content))
	if err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return fmt.Sprintf("Successfully wrote to %s", filePath), nil
}
