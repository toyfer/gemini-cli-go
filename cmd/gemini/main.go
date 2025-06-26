package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gemini-cli-go/internal/api"
	"gemini-cli-go/internal/filesystem"
	"gemini-cli-go/internal/auth"
	config_pkg "gemini-cli-go/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "gemini",
	Short: "Gemini CLI is a command-line tool for interacting with Gemini API",
	Long: `A fast and flexible command-line interface for Google Gemini API.
	
Gemini CLI allows you to interact with Gemini models, manage your projects,
and automate AI-powered workflows directly from your terminal.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior when no subcommand is provided
		fmt.Println("Welcome to Gemini CLI! Use 'gemini --help' for more information.")
	},
}

var chatCmd = &cobra.Command{
	Use:   "chat [prompt]",
	Short: "Interact with Gemini API to generate content",
	Args:  cobra.ExactArgs(1), // プロンプトが1つだけ必要
	Run: func(cmd *cobra.Command, args []string) {
		var client *api.Client
		token, err := config_pkg.LoadToken()
		if err == nil && token.Valid() {
			// Use OAuth2 client
			oauthConfig := auth.GetOAuth2Config(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"))
			httpClient := auth.GetHTTPClient(context.Background(), oauthConfig, token)
			client = api.NewClient("", httpClient) // No API key needed if using OAuth2 client
			fmt.Println("Using OAuth2 for authentication.")
		} else {
			// Fallback to API key
			apiKey := os.Getenv("GEMINI_API_KEY")
			if apiKey == "" {
				fmt.Println("Error: GEMINI_API_KEY environment variable not set and no valid OAuth2 token found.")
				fmt.Println("Please get your API key from https://aistudio.google.com/apikey or run 'gemini auth'.")
				os.Exit(1)
			}
			client = api.NewClient(apiKey, nil)
			fmt.Println("Using API key for authentication.")
		}

		prompt := args[0]

		fmt.Printf("Sending prompt to Gemini: \"%s\"\n", prompt)
		resp, err := client.GenerateContent(prompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating content: %v\n", err)
			os.Exit(1)
		}

		if len(resp.Candidates) > 0 {
			fmt.Println("\nGemini's response:")
			fmt.Println(resp.Candidates[0].Content.Parts[0].Text)
		} else {
			fmt.Println("No response candidates from Gemini.")
		}
	},
}

var readCmd = &cobra.Command{
	Use:   "read [filePath]",
	Short: "Reads the content of a specified file",
	Args:  cobra.ExactArgs(1), // ファイルパスが1つだけ必要
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		content, err := filesystem.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Content of %s:\n", filePath)
		fmt.Println(string(content))
	},
}

var listFilesCmd = &cobra.Command{
	Use:   "list-files [directory]",
	Short: "Lists files in a directory, optionally filtered by extension",
	Args:  cobra.ExactArgs(1), // ディレクトリパスが1つだけ必要
	Run: func(cmd *cobra.Command, args []string) {
		dirPath := args[0]
		extensions, _ := cmd.Flags().GetStringSlice("ext")

		files, err := filesystem.WalkDir(dirPath, extensions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing files: %v\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Printf("No files found in %s with extensions %v\n", dirPath, extensions)
			return
		}

		fmt.Printf("Files in %s (filtered by %v):\n", dirPath, extensions)
		for _, file := range files {
			fmt.Printf("- %s (size: %d bytes)\n", file.Path, len(file.Content))
		}
	},
}

var contextCmd = &cobra.Command{
	Use:   "context [directory]",
	Short: "Generates a formatted context string from files in a directory",
	Args:  cobra.ExactArgs(1), // ディレクトリパスが1つだけ必要
	Run: func(cmd *cobra.Command, args []string) {
		dirPath := args[0]
		extensions, _ := cmd.Flags().GetStringSlice("ext")

		files, err := filesystem.WalkDir(dirPath, extensions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Printf("No files found in %s with extensions %v to generate context.\n", dirPath, extensions)
			return
		}

		formattedContext := api.FormatFilesForGemini(files)
		fmt.Println(formattedContext)
	},
}

var generateCodeCmd = &cobra.Command{
	Use:   "generate-code [prompt]",
	Short: "Generates code based on a prompt and optional context",
	Args:  cobra.ExactArgs(1), // プロンプトが1つだけ必要
	Run: func(cmd *cobra.Command, args []string) {
		var client *api.Client
		token, err := config_pkg.LoadToken()
		if err == nil && token.Valid() {
			// Use OAuth2 client
			oauthConfig := auth.GetOAuth2Config(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"))
			httpClient := auth.GetHTTPClient(context.Background(), oauthConfig, token)
			client = api.NewClient("", httpClient) // No API key needed if using OAuth2 client
			fmt.Println("Using OAuth2 for authentication.")
		} else {
			// Fallback to API key
			apiKey := os.Getenv("GEMINI_API_KEY")
			if apiKey == "" {
				fmt.Println("Error: GEMINI_API_KEY environment variable not set and no valid OAuth2 token found.")
				fmt.Println("Please get your API key from https://aistudio.google.com/apikey or run 'gemini auth'.")
				os.Exit(1)
			}
			client = api.NewClient(apiKey, nil)
			fmt.Println("Using API key for authentication.")
		}

		prompt := args[0]
		contextDir, _ := cmd.Flags().GetString("context-dir")
		extensions, _ := cmd.Flags().GetStringSlice("ext")

		fullPrompt := prompt
		if contextDir != "" {
			files, err := filesystem.WalkDir(contextDir, extensions)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading context directory: %v\n", err)
				os.Exit(1)
			}
			if len(files) > 0 {
				formattedContext := api.FormatFilesForGemini(files)
				fullPrompt = fmt.Sprintf("%s\n\nHere is the context:\n%s", prompt, formattedContext)
			}
		}

		fmt.Printf("Sending prompt to Gemini:\n---\n%s\n---\n", fullPrompt)
		resp, err := client.GenerateContent(fullPrompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
			os.Exit(1)
		}

		if len(resp.Candidates) > 0 {
			fmt.Println("\nGenerated Code:")
			fmt.Println(resp.Candidates[0].Content.Parts[0].Text)
		} else {
			fmt.Println("No code generated by Gemini.")
		}
	},
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticates with Google Account for Gemini API access",
	Long: `This command initiates the OAuth2 flow to authenticate your Google Account
and gain access to the Gemini API. It will open a browser window for you to
log in and grant permissions.`,
	Run: func(cmd *cobra.Command, args []string) {
		clientID := os.Getenv("GOOGLE_CLIENT_ID")
		clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

		if clientID == "" || clientSecret == "" {
			fmt.Println("Error: GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET environment variables must be set.")
			fmt.Println("Please create OAuth 2.0 Client IDs in Google Cloud Console.")
			os.Exit(1)
		}

		config := auth.GetOAuth2Config(clientID, clientSecret)
		authURL := auth.GetAuthCodeURL(config)

		fmt.Println("Opening your browser to complete authentication...")
		fmt.Printf("If your browser does not open automatically, please visit this URL:\n%s\n", authURL)

		code, err := auth.StartLocalServer(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error during authentication callback: %v\n", err)
			os.Exit(1)
		}

		token, err := auth.ExchangeCodeForToken(config, code)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error exchanging code for token: %v\n", err)
			os.Exit(1)
		}

		// Save the token
		if err := config_pkg.SaveToken(token); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving token: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Authentication successful! Token saved.")
		fmt.Printf("Access Token: %s\n", token.AccessToken)
		fmt.Printf("Refresh Token: %s\n", token.RefreshToken)
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(listFilesCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(generateCodeCmd)
	rootCmd.AddCommand(authCmd)

	// list-files コマンドに --ext フラグを追加
	listFilesCmd.Flags().StringSliceP("ext", "e", []string{}, "Comma-separated list of file extensions to filter (e.g., .go,.txt)")

	// context コマンドに --ext フラグを追加
	contextCmd.Flags().StringSliceP("ext", "e", []string{}, "Comma-separated list of file extensions to filter (e.g., .go,.txt)")

	// generate-code コマンドに --context-dir と --ext フラグを追加
	generateCodeCmd.Flags().StringP("context-dir", "c", "", "Directory to use as context for code generation")
	generateCodeCmd.Flags().StringSliceP("ext", "e", []string{}, "Comma-separated list of file extensions to filter in context directory (e.g., .go,.txt)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}