package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

const (
	configDir  = ".gemini-cli-go"
	tokenFile  = "token.json"
)

// TokenStore represents the structure to store OAuth2 tokens.
type TokenStore struct {
	Token *oauth2.Token `json:"token"`
}

// getConfigPath returns the full path to the configuration directory.
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, configDir), nil
}

// getTokenFilePath returns the full path to the token file.
func getTokenFilePath() (string, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, tokenFile), nil
}

// SaveToken saves the OAuth2 token to a file.
func SaveToken(token *oauth2.Token) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configPath, 0700); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configPath, err)
	}

	tokenFilePath, err := getTokenFilePath()
	if err != nil {
		return err
	}

	tokenStore := TokenStore{Token: token}
	data, err := json.MarshalIndent(tokenStore, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := ioutil.WriteFile(tokenFilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token to file %s: %w", tokenFilePath, err)
	}
	return nil
}

// LoadToken loads the OAuth2 token from a file.
func LoadToken() (*oauth2.Token, error) {
	tokenFilePath, err := getTokenFilePath()
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(tokenFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("token file not found: %s", tokenFilePath)
		}
		return nil, fmt.Errorf("failed to read token file %s: %w", tokenFilePath, err)
	}

	var tokenStore TokenStore
	if err := json.Unmarshal(data, &tokenStore); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}
	return tokenStore.Token, nil
}
