package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/oauth2"
)

func TestGetOAuth2Config(t *testing.T) {
	clientID := "test-client-id"
	clientSecret := "test-client-secret"

	config := GetOAuth2Config(clientID, clientSecret)

	if config.ClientID != clientID {
		t.Errorf("Expected ClientID %s, got %s", clientID, config.ClientID)
	}
	if config.ClientSecret != clientSecret {
		t.Errorf("Expected ClientSecret %s, got %s", clientSecret, config.ClientSecret)
	}
	if config.RedirectURL != "http://localhost:8080/oauth/callback" {
		t.Errorf("Expected RedirectURL %s, got %s", "http://localhost:8080/oauth/callback", config.RedirectURL)
	}
	if len(config.Scopes) != 1 || config.Scopes[0] != geminiAPIScope {
		t.Errorf("Expected scopes [%s], got %v", geminiAPIScope, config.Scopes)
	}
}

func TestGetAuthCodeURL(t *testing.T) {
	config := GetOAuth2Config("test-client-id", "test-client-secret")
	authURL := GetAuthCodeURL(config)

	if authURL == "" {
		t.Error("Expected non-empty auth URL")
	}
	// Basic check that it contains expected elements
	if !contains(authURL, "oauth2") {
		t.Error("Expected auth URL to contain 'oauth2'")
	}
}

func TestSaveAndLoadToken(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "auth_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create a test token
	testToken := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
	}

	// Test saving the token
	err = SaveToken(testToken)
	if err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	// Verify the token file was created
	tokenPath := filepath.Join(tmpDir, geminiDirName, tokenFileName)
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		t.Fatal("Token file was not created")
	}

	// Test loading the token
	loadedToken, err := LoadToken()
	if err != nil {
		t.Fatalf("LoadToken failed: %v", err)
	}

	// Verify the loaded token matches the saved token
	if loadedToken.AccessToken != testToken.AccessToken {
		t.Errorf("Expected AccessToken %s, got %s", testToken.AccessToken, loadedToken.AccessToken)
	}
	if loadedToken.RefreshToken != testToken.RefreshToken {
		t.Errorf("Expected RefreshToken %s, got %s", testToken.RefreshToken, loadedToken.RefreshToken)
	}
	if loadedToken.TokenType != testToken.TokenType {
		t.Errorf("Expected TokenType %s, got %s", testToken.TokenType, loadedToken.TokenType)
	}
}

func TestLoadTokenNotFound(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "auth_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Test loading a non-existent token
	_, err = LoadToken()
	if err == nil {
		t.Error("Expected an error when loading non-existent token, but got none")
	}
}

func TestValidateAuthMethod(t *testing.T) {
	tests := []struct {
		name        string
		authType    string
		setupEnv    func()
		cleanupEnv  func()
		expectError bool
	}{
		{
			name:        "Invalid auth type",
			authType:    "invalid",
			setupEnv:    func() {},
			cleanupEnv:  func() {},
			expectError: true,
		},
		{
			name:     "API key auth with missing env var",
			authType: "api_key",
			setupEnv: func() {
				os.Unsetenv("GEMINI_API_KEY")
			},
			cleanupEnv:  func() {},
			expectError: true,
		},
		{
			name:     "API key auth with valid env var",
			authType: "api_key",
			setupEnv: func() {
				os.Setenv("GEMINI_API_KEY", "test-api-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("GEMINI_API_KEY")
			},
			expectError: false,
		},
		{
			name:     "OAuth auth without token",
			authType: "oauth",
			setupEnv: func() {
				// Create a temporary directory for testing
				tmpDir, _ := os.MkdirTemp("", "auth_test")
				os.Setenv("HOME", tmpDir)
			},
			cleanupEnv: func() {
				tmpDir := os.Getenv("HOME")
				os.Unsetenv("HOME")
				os.RemoveAll(tmpDir)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			err := ValidateAuthMethod(tt.authType)
			if tt.expectError && err == nil {
				t.Error("Expected an error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}
		})
	}
}

func TestGetHTTPClient(t *testing.T) {
	config := GetOAuth2Config("test-client-id", "test-client-secret")
	token := &oauth2.Token{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
	}

	client := GetHTTPClient(context.Background(), config, token)
	if client == nil {
		t.Error("Expected non-nil HTTP client")
	}
}

func TestExchangeCodeForToken(t *testing.T) {
	// This test would require a mock HTTP server to simulate the OAuth2 exchange
	// For now, we'll just test that the function exists and can be called
	config := GetOAuth2Config("test-client-id", "test-client-secret")
	
	// This will fail with a real OAuth2 endpoint, but tests the function signature
	_, err := ExchangeCodeForToken(config, "invalid-code")
	if err == nil {
		t.Error("Expected an error when exchanging invalid code, but got none")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}