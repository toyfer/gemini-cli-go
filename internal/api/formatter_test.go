package api

import (
	"testing"

	"gemini-cli-go/internal/filesystem"
)

func TestFormatFilesForGemini(t *testing.T) {
	tests := []struct {
		name  string
		files []filesystem.FileContent
		want  string
	}{
		{
			name:  "Empty slice",
			files: []filesystem.FileContent{},
			want:  "",
		},
		{
			name: "Single file",
			files: []filesystem.FileContent{
				{Path: "/path/to/file1.txt", Content: []byte("Content of file1.")},
			},
			want: "--- File: /path/to/file1.txt ---\nContent of file1.\n--- End of File: /path/to/file1.txt ---\n\n",
		},
		{
			name: "Multiple files",
			files: []filesystem.FileContent{
				{Path: "/path/to/file1.txt", Content: []byte("Content of file1.")},
				{Path: "/path/to/file2.go", Content: []byte("package main\n\nfunc main(){}")},
			},
			want: "--- File: /path/to/file1.txt ---\nContent of file1.\n--- End of File: /path/to/file1.txt ---\n\n" +
				"--- File: /path/to/file2.go ---\npackage main\n\nfunc main(){}\n--- End of File: /path/to/file2.go ---\n\n",
		},
		{
			name: "File with empty content",
			files: []filesystem.FileContent{
				{Path: "/path/to/empty.txt", Content: []byte("")},
			},
			want: "--- File: /path/to/empty.txt ---\n\n--- End of File: /path/to/empty.txt ---\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatFilesForGemini(tt.files)
			if got != tt.want {
				t.Errorf("FormatFilesForGemini() got = %q, want %q", got, tt.want)
			}
		})
	}
}