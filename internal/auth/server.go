package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// StartLocalServer starts a local HTTP server to handle the OAuth2 callback.
// It returns the authorization code received from the callback.
func StartLocalServer(config *oauth2.Config) (string, error) {
	codeCh := make(chan string)
	errorCh := make(chan error)

	server := &http.Server{
		Addr: "localhost:8080",
	}

	http.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errorCh <- fmt.Errorf("no authorization code received")
			http.Error(w, "No authorization code received", http.StatusBadRequest)
			return
		}
		codeCh <- code
		fmt.Fprintf(w, "Authentication successful! You can close this window.")
	})

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return "", fmt.Errorf("failed to listen on %s: %w", server.Addr, err)
	}

	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			errorCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case code := <-codeCh:
		// Shutdown the server gracefully
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		return code, nil
	case err := <-errorCh:
		// Shutdown the server gracefully
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		return "", err
	case <-time.After(5 * time.Minute): // Timeout after 5 minutes
		// Shutdown the server gracefully
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		return "", fmt.Errorf("authentication timed out")
	}
}
