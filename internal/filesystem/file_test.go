package filesystem

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"
)

func TestReadFile(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := ioutil.TempFile("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	content := []byte("Hello, world!")
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test case 1: Read existing file
	readContent, err := ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(readContent) != string(content) {
		t.Errorf("ReadFile returned unexpected content: got %q, want %q", readContent, content)
	}

	// Test case 2: Read non-existent file
	_, err = ReadFile("non_existent_file.txt")
	if err == nil {
		t.Error("ReadFile did not return an error for a non-existent file")
	}
			if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("ReadFile returned unexpected error type for non-existent file: %v", err)
	}
}
