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
