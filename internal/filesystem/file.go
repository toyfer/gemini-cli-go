package filesystem

import (
	"fmt"
	"io/ioutil"
	"os"
)

// ReadFile reads the content of a file at the given path.
func ReadFile(filePath string) ([]byte, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		return nil, err // Return the original error
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return content, nil
}

// WriteFile writes content to a file at the given path.
// If the file does not exist, it will be created. If it exists, its content will be truncated.
func WriteFile(filePath string, content []byte) error {
	err := ioutil.WriteFile(filePath, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}
