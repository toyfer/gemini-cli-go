package api

import (
	"fmt"
	"strings"

	"gemini-cli-go/internal/filesystem"
)

// FormatFilesForGemini takes a slice of FileContent and formats it into a single string
// suitable for sending to the Gemini API as context.
// It includes file paths and their content, separated by clear markers.
func FormatFilesForGemini(files []filesystem.FileContent) string {
	var builder strings.Builder
	for _, file := range files {
		builder.WriteString(fmt.Sprintf("--- File: %s ---\n", file.Path))
		builder.Write(file.Content)
		builder.WriteString(fmt.Sprintf("\n--- End of File: %s ---\n\n", file.Path))
	}
	return builder.String()
}
