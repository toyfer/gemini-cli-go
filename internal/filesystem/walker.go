package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileContent represents the content of a file with its path.
type FileContent struct {
	Path    string
	Content []byte
}

// WalkDir recursively walks a directory and returns the content of files
// that match the given file extensions.
func WalkDir(root string, extensions []string) ([]FileContent, error) {
	var files []FileContent

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if the file extension matches any of the desired extensions
		if len(extensions) > 0 {
			match := false
			for _, ext := range extensions {
				if strings.HasSuffix(info.Name(), ext) {
					match = true
					break
				}
			}
			if !match {
				return nil
			}
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		files = append(files, FileContent{
			Path:    path,
			Content: content,
		})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", root, err)
	}

	return files, nil
}
