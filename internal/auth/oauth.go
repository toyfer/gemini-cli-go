package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	// These are the scopes required for accessing Gemini API.
	// See https://developers.google.com/identity/protocols/oauth2/scopes#generativelanguage
	geminiAPIScope = "https://www.googleapis.com/auth/generativelanguage"

	tokenFileName = "token.json"
	geminiDirName = ".gemini"
)

// GetOAuth2Config returns the OAuth2 configuration for Google API.
func GetOAuth2Config(clientID, clientSecret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/oauth/callback", // This should be configurable
		Scopes:       []string{geminiAPIScope},
		Endpoint:     google.Endpoint,
	}
}

// GetAuthCodeURL generates the URL for user authorization.
func GetAuthCodeURL(config *oauth2.Config) string {
	return config.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCodeForToken exchanges the authorization code for an OAuth2 token.
func ExchangeCodeForToken(config *oauth2.Config, code string) (*oauth2.Token, error) {
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}
	return token, nil
}

// GetHTTPClient returns an HTTP client with the OAuth2 token.
func GetHTTPClient(ctx context.Context, config *oauth2.Config, token *oauth2.Token) *http.Client {
	return config.Client(ctx, token)
}

// SaveToken saves the OAuth2 token to a file in the user's home directory.
func SaveToken(token *oauth2.Token) error {
	home, err := homedir.Dir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	geminiDirPath := filepath.Join(home, geminiDirName)
	if _, err := os.Stat(geminiDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(geminiDirPath, 0700); err != nil {
			return fmt.Errorf("failed to create .gemini directory: %w", err)
		}
	}

	tokenPath := filepath.Join(geminiDirPath, tokenFileName)
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// LoadToken loads the OAuth2 token from a file in the user's home directory.
func LoadToken() (*oauth2.Token, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	tokenPath := filepath.Join(home, geminiDirName, tokenFileName)
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("token file not found at %s", tokenPath)
	}

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}

// ValidateAuthMethod validates the selected authentication method.
func ValidateAuthMethod(authType string) error {
	if authType == "oauth" {
		token, err := LoadToken()
		if err != nil {
			return fmt.Errorf("OAuth2 token not found or invalid: %w", err)
		}
		if !token.Valid() {
			return fmt.Errorf("OAuth2 token is expired or invalid")
		}
		return nil
	} else if authType == "api_key" {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("GEMINI_API_KEY environment variable not set")
		}
		return nil
	}
	return fmt.Errorf("unsupported authentication type: %s", authType)
}