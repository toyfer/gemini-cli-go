package tool

import (
	"context"
	"fmt"
	"gemini-cli-go/internal/filesystem"
	"gemini-cli-go/internal/shared"
	"strings"
)

// ListFilesTool implements the Tool interface for listing files.
type ListFilesTool struct{}

// Name returns the name of the tool.
func (t *ListFilesTool) Name() string {
	return "list_files"
}

// Description returns the description of the tool.
func (t *ListFilesTool) Description() string {
	return "Lists files in a directory, optionally filtered by extension."
}

// FunctionDeclaration returns the function declaration for the tool.
func (t *ListFilesTool) FunctionDeclaration() shared.FunctionDeclaration {
	return shared.FunctionDeclaration{
		Name:		t.Name(),
		Description:	t.Description(),
		Parameters: shared.Schema{
            Type: shared.TypeObject,
            Properties: map[string]shared.Schema{
                "dir": {
                    Type:        shared.TypeString,
                    Description: "The directory to list files from.",
                },
                "ext": {
                    Type:        shared.TypeArray,
                    Description: "Optional: List of file extensions to filter (e.g., \"go\", \"txt\").",
                    Items: &shared.Schema{ // Add Items field for array type
                        Type: shared.TypeString,
                    },
                },
            },
            Required: []string{"dir"},
        },
	}
}


// Execute executes the list_files tool.
func (t *ListFilesTool) Execute(_ context.Context, args map[string]interface{}) (string, error) {
	dir, ok := args["dir"].(string)
	if !ok || dir == "" {
		return "", fmt.Errorf("missing or invalid 'dir' argument for list_files tool")
	}

	var extensions []string
	if extArg, ok := args["ext"]; ok {
		if extList, isList := extArg.([]interface{}); isList {
			for _, item := range extList {
				if ext, isString := item.(string); isString {
					extensions = append(extensions, ext)
				}
			}
		}
	}

	files, err := filesystem.WalkDir(dir, extensions)
	if err != nil {
		return "", fmt.Errorf("failed to list files in %s: %w", dir, err)
	}

	if len(files) == 0 {
		return fmt.Sprintf("No files found in %s with extensions %v", dir, extensions), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Files in %s (filtered by %v):\n", dir, extensions))
	for _, file := range files {
		result.WriteString(fmt.Sprintf("- %s (size: %d bytes)\n", file.Path, len(file.Content)))
	}
	return result.String(), nil
}
