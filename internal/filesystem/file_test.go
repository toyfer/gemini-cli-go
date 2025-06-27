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

func TestWriteFile(t *testing.T) {
	// Test case 1: Write to new file
	tmpfile, err := ioutil.TempFile("", "testwritefile")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name()) // clean up

	content := []byte("Test content for writing")
	err = WriteFile(tmpfile.Name(), content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify the content was written correctly
	readContent, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read back written file: %v", err)
	}
	if string(readContent) != string(content) {
		t.Errorf("WriteFile content mismatch: got %q, want %q", readContent, content)
	}

	// Test case 2: Overwrite existing file
	newContent := []byte("New content to overwrite")
	err = WriteFile(tmpfile.Name(), newContent)
	if err != nil {
		t.Fatalf("WriteFile failed on overwrite: %v", err)
	}

	// Verify the content was overwritten
	readContent, err = ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read back overwritten file: %v", err)
	}
	if string(readContent) != string(newContent) {
		t.Errorf("WriteFile overwrite content mismatch: got %q, want %q", readContent, newContent)
	}

	// Test case 3: Write to non-existent directory (should fail)
	err = WriteFile("/non/existent/directory/file.txt", content)
	if err == nil {
		t.Error("WriteFile did not return an error for non-existent directory")
	}
}