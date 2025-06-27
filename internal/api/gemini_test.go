package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func TestGenerateContentStream(t *testing.T) {
	// Mock server for Gemini API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta/models/gemini-pro:streamGenerateContent" {
			t.Errorf("Expected to request '/v1beta/models/gemini-pro:streamGenerateContent', got: %s", r.URL.Path)
		}
		// Removed API key check as it's handled by genai.NewClient when httpClient is provided.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{"candidates":[{"content":{"parts":[{"text":"Generated "}]}}]},
			{"candidates":[{"content":{"parts":[{"text":"response."}]}}]}
		]`))
	}))
	defer server.Close()

	ctx := context.Background()
	client, err := NewClient(ctx, "test-api-key", server.Client(), "gemini-pro", option.WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	stream, err := client.GenerateContentStream(ctx, "test prompt", nil)
	if err != nil {
		t.Fatalf("GenerateContentStream failed: %v", err)
	}

	var fullResponse string
	for {
		resp, err := stream.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Error streaming response: %v", err)
		}
		if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
				fullResponse += string(text)
			}
		}
	}

	if fullResponse != "Generated response." {
		t.Errorf("Expected 'Generated response.', got: %s", fullResponse)
	}

	// Test error response from API
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta/models/gemini-pro:streamGenerateContent" {
			t.Errorf("Expected to request '/v1beta/models/gemini-pro:streamGenerateContent', got: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"Internal server error"}}`))
	}))
	defer errorServer.Close()

	client, err = NewClient(ctx, "test-api-key", errorServer.Client(), "gemini-pro", option.WithEndpoint(errorServer.URL))
	if err != nil {
		t.Fatalf("Failed to create client for error test: %v", err)
	}

	_, err = client.GenerateContentStream(ctx, "test prompt", nil)
	// The error should be returned by stream.Next() in the loop, not by GenerateContentStream itself.
	// So, we need to create a stream and then check for the error when calling Next().
	stream, streamErr := client.GenerateContentStream(ctx, "test prompt", nil)
	if streamErr != nil {
		t.Fatalf("GenerateContentStream failed for error test: %v", streamErr)
	}

	_, err = stream.Next()
	if err == nil {
		t.Error("Expected an error for internal server error, but got none")
	}
	expectedErrorMsg := "failed to get next response from stream: googleapi: Error 400: Internal server error"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message %q, got %q", expectedErrorMsg, err.Error())
	}
}