package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	defaultAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent"
)

// GenerateContentRequest represents the request body for the generateContent API.
type GenerateContentRequest struct {
	Contents []Content `json:"contents"`
}

// Content represents a single content part in the request.
type Content struct {
	Parts []Part `json:"parts"`
}

// Part represents a single part of the content, e.g., text.
type Part struct {
	Text string `json:"text"`
}

// GenerateContentResponse represents the response body from the generateContent API.
type GenerateContentResponse struct {
	Candidates []Candidate `json:"candidates"`
}

// Candidate represents a generated content candidate.
type Candidate struct {
	Content Content `json:"content"`
}

// Client is a client for the Gemini API.
type Client struct {
	HTTPClient *http.Client // Use a generic HTTP client
	APIKey     string       // Still keep APIKey for fallback or direct use
	APIURL     string
}

// NewClient creates a new Gemini API client.
// It can be initialized with an existing http.Client (e.g., for OAuth2) or an API key.
func NewClient(apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &Client{
		HTTPClient: httpClient,
		APIKey:     apiKey,
		APIURL:     defaultAPIURL,
	}
}

// GenerateContent sends a request to the Gemini API to generate content.
func (c *Client) GenerateContent(prompt string) (*GenerateContentResponse, error) {
	reqBody := GenerateContentRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.APIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Only set API key header if HTTPClient is not already configured for OAuth2
	// (i.e., if APIKey is provided and HTTPClient is the default one)
	if c.APIKey != "" && c.HTTPClient == http.DefaultClient { // Simplified check, might need refinement
		req.Header.Set("x-goog-api-key", c.APIKey)
	} else if c.APIKey != "" { // If APIKey is set, but HTTPClient is custom, assume it's for OAuth2
		// Do nothing, OAuth2 client will handle auth
	}


	resp, err := c.HTTPClient.Do(req) // Use the injected HTTPClient
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var genResp GenerateContentResponse
	if err := json.Unmarshal(respBody, &genResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &genResp, nil
}