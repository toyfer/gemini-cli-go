package filesystem

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestWalkDir(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := ioutil.TempDir("", "testdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) // clean up

	// Create some files
	filesToCreate := map[string]string{
		"file1.txt": "content1",
		"file2.go":  "package main",
		"subdir/file3.txt": "content3",
		"subdir/file4.js":  "console.log('hello');",
		"anotherdir/file5.py": "print('python')",
	}

	for path, content := range filesToCreate {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := ioutil.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name       string
		root       string
		extensions []string
		expected   []FileContent
		expectErr  bool
	}{
		{
			name:       "No filter",
			root:       tmpDir,
			extensions: []string{},
			expected: []FileContent{
				{Path: filepath.Join(tmpDir, "anotherdir/file5.py"), Content: []byte("print('python')")},
				{Path: filepath.Join(tmpDir, "file1.txt"), Content: []byte("content1")},
				{Path: filepath.Join(tmpDir, "file2.go"), Content: []byte("package main")},
				{Path: filepath.Join(tmpDir, "subdir/file3.txt"), Content: []byte("content3")},
				{Path: filepath.Join(tmpDir, "subdir/file4.js"), Content: []byte("console.log('hello');")},
			},
			expectErr: false,
		},
		{
			name:       "Filter .txt files",
			root:       tmpDir,
			extensions: []string{".txt"},
			expected: []FileContent{
				{Path: filepath.Join(tmpDir, "file1.txt"), Content: []byte("content1")},
				{Path: filepath.Join(tmpDir, "subdir/file3.txt"), Content: []byte("content3")},
			},
			expectErr: false,
		},
		{
			name:       "Filter .go and .js files",
			root:       tmpDir,
			extensions: []string{".go", ".js"},
			expected: []FileContent{
				{Path: filepath.Join(tmpDir, "file2.go"), Content: []byte("package main")},
				{Path: filepath.Join(tmpDir, "subdir/file4.js"), Content: []byte("console.log('hello');")},
			},
			expectErr: false,
		},
		{
			name:       "Non-existent directory",
			root:       filepath.Join(tmpDir, "nonexistent"),
			extensions: []string{},
			expected:   nil,
			expectErr:  true,
		},
		{
			name:       "Empty directory",
			root:       filepath.Join(tmpDir, "emptydir"),
			extensions: []string{},
			expected:   []FileContent{},
			expectErr:  false,
		},
	}

	// Create empty directory for "Empty directory" test case
	if err := os.MkdirAll(filepath.Join(tmpDir, "emptydir"), 0755); err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := WalkDir(tt.root, tt.extensions)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected an error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("WalkDir returned an unexpected error: %v", err)
			}

			// Sort results for consistent comparison
			sort.Slice(result, func(i, j int) bool {
				return result[i].Path < result[j].Path
			})
			sort.Slice(tt.expected, func(i, j int) bool {
				return tt.expected[i].Path < tt.expected[j].Path
			})

			if len(result) != len(tt.expected) {
				t.Fatalf("Mismatched number of files: got %d, want %d", len(result), len(tt.expected))
			}

			for i := range result {
				if result[i].Path != tt.expected[i].Path {
					t.Errorf("Mismatched file path at index %d: got %q, want %q", i, result[i].Path, tt.expected[i].Path)
				}
				if string(result[i].Content) != string(tt.expected[i].Content) {
					t.Errorf("Mismatched file content at index %d for path %q: got %q, want %q", i, result[i].Path, result[i].Content, tt.expected[i].Content)
				}
			}
		})
	}
}
