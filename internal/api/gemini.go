package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gemini-cli-go/internal/config"
	"gemini-cli-go/internal/shared"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client is a client for the Gemini API.
type Client struct {
	model *genai.GenerativeModel
}

// NewClient creates a new Gemini API client.
// It can be initialized with an existing http.Client (e.g., for OAuth2) or an API key.
// The modelName specifies which Gemini model to use (e.g., "gemini-pro", "gemini-2.5-pro").
func NewClient(ctx context.Context, apiKey string, httpClient *http.Client, modelName string, opts ...option.ClientOption) (*Client, error) {
	var clientOpts []option.ClientOption
	if httpClient != nil {
		clientOpts = append(clientOpts, option.WithHTTPClient(httpClient), option.WithoutAuthentication())
	} else if apiKey != "" {
		clientOpts = append(clientOpts, option.WithAPIKey(apiKey))
	} else {
		return nil, fmt.Errorf("either httpClient or apiKey must be provided")
	}

	clientOpts = append(clientOpts, opts...)

	genaiClient, err := genai.NewClient(ctx, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := genaiClient.GenerativeModel(modelName)
	return &Client{model: model}, nil
}

// GenerateContentStream sends a request to the Gemini API to generate content and streams the response.
func (c *Client) GenerateContentStream(ctx context.Context, prompt string, tools *shared.Tools) (*ResponseStream, error) {
	var genaiTools []*genai.Tool

	if tools != nil && len(tools.FunctionDeclarations) > 0 {
		genaiTools = make([]*genai.Tool, len(tools.FunctionDeclarations))
		for i, fd := range tools.FunctionDeclarations {
			// Convert map[string]interface{} to *genai.Schema
			paramsJSON, err := json.Marshal(fd.Parameters)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal parameters: %w", err)
			}
			var schema genai.Schema
			if err := json.Unmarshal(paramsJSON, &schema); err != nil {
				return nil, fmt.Errorf("failed to unmarshal parameters to schema: %w", err)
			}

			genaiTools[i] = &genai.Tool{
				FunctionDeclarations: []*genai.FunctionDeclaration{
					{
						Name:        fd.Name,
						Description: fd.Description,
						Parameters:  &schema,
					},
				},
			}
		}
		// Set tools on the model
		c.model.Tools = genaiTools
	}

	iter := c.model.GenerateContentStream(ctx, genai.Text(prompt))
	return &ResponseStream{iter: iter}, nil
}

// ResponseStream wraps the genai.GenerateContentResponseIterator.
type ResponseStream struct {
	iter *genai.GenerateContentResponseIterator
}

// Next returns the next part of the streamed response.
func (rs *ResponseStream) Next() (*genai.GenerateContentResponse, error) {
	resp, err := rs.iter.Next()
	if err == iterator.Done {
		return nil, io.EOF
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get next response from stream: %w", err)
	}
	return resp, nil
}

// RunNonInteractive handles the non-interactive CLI interaction with Gemini.
func RunNonInteractive(ctx context.Context, cfg *config.CliConfig, client *Client, toolRegistry shared.ToolRegistryInterface, initialPrompt string) error {
	currentPrompt := initialPrompt
	for {
		fmt.Printf("Sending prompt to Gemini: \"%s\"\n", currentPrompt)
		stream, err := client.GenerateContentStream(ctx, currentPrompt, &shared.Tools{FunctionDeclarations: toolRegistry.GetFunctionDeclarations()})
		if err != nil {
			return fmt.Errorf("error generating content: %w", err)
		}

		var fullTextResponse string
		var functionCalls []shared.FunctionCall

		for {
			resp, err := stream.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("error streaming response: %w", err)
			}

			if len(resp.Candidates) == 0 {
				continue
			}

			candidate := resp.Candidates[0]
			for _, part := range candidate.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					fmt.Print(string(text))
					fullTextResponse += string(text)
				} else if functionCall, ok := part.(genai.FunctionCall); ok {
					functionCalls = append(functionCalls, shared.FunctionCall{
						Name: functionCall.Name,
						Args: functionCall.Args,
					})
				}
			}
		}

		fmt.Println() // Newline after streamed response

		if len(functionCalls) > 0 {
			fc := functionCalls[0] // Simplified: assuming one function call at a time
			fmt.Printf("\nGemini called tool: %s with args: %v\n", fc.Name, fc.Args)

			calledTool, ok := toolRegistry.GetTool(fc.Name)
			if !ok {
				return fmt.Errorf("tool %s not found", fc.Name)
			}

			toolOutput, err := calledTool.Execute(ctx, fc.Args)
			if err != nil {
				return fmt.Errorf("error executing tool %s: %w", fc.Name, err)
			}

			fmt.Printf("Tool %s output:\n%s\n", fc.Name, toolOutput)
			currentPrompt = fmt.Sprintf("Tool %s returned: %s", fc.Name, toolOutput) // Feed tool output back to Gemini
		} else if fullTextResponse != "" {
			return nil // Gemini responded with text, end interaction
		} else {
			return fmt.Errorf("no text or function call in Gemini's response")
		}
	}
}
