package auth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	// These are the scopes required for accessing Gemini API.
	// See https://developers.google.com/identity/protocols/oauth2/scopes#generativelanguage
	geminiAPIScope = "https://www.googleapis.com/auth/generativelanguage"
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
