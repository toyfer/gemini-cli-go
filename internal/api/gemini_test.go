package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateContent(t *testing.T) {
	// Mock server for Gemini API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta/models/gemini-pro:generateContent" {
			t.Errorf("Expected to request '/v1beta/models/gemini-pro:generateContent', got: %s", r.URL.Path)
		}
		if r.Header.Get("x-goog-api-key") != "test-api-key" {
			t.Errorf("Expected API key 'test-api-key', got: %s", r.Header.Get("x-goog-api-key"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got: %s", r.Header.Get("Content-Type"))
		}

		// Simulate a successful response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"Generated response."}]}}]}`))
	}))
	defer server.Close()

	client := NewClient("test-api-key")
	client.APIURL = server.URL + "/v1beta/models/gemini-pro:generateContent" // Use mock server URL

	resp, err := client.GenerateContent("test prompt")
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	if len(resp.Candidates) == 0 {
		t.Fatal("No candidates in response")
	}
	if resp.Candidates[0].Content.Parts[0].Text != "Generated response." {
		t.Errorf("Expected 'Generated response.', got: %s", resp.Candidates[0].Content.Parts[0].Text)
	}

	// Test error response from API
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"message":"Internal server error"}}`))
	}))
	defer errorServer.Close()

	client.APIURL = errorServer.URL + "/v1beta/models/gemini-pro:generateContent"
	_, err = client.GenerateContent("test prompt")
	if err == nil {
		t.Error("Expected an error for internal server error, but got none")
	}
	expectedErrorMsg := "API request failed with status 500: {\"error\":{\"message\":\"Internal server error\"}}"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message %q, got %q", expectedErrorMsg, err.Error())
	}
}