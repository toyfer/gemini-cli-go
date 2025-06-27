package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"gemini-cli-go/internal/api"
	"gemini-cli-go/internal/filesystem"
	"gemini-cli-go/internal/auth"
	"gemini-cli-go/internal/errors"
	config_pkg "gemini-cli-go/internal/config"
	tool_pkg "gemini-cli-go/internal/tool"
	"gemini-cli-go/internal/telemetry"
	"gemini-cli-go/internal/shared"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid" // Import for sessionId
)

var globalCliConfig *config_pkg.CliConfig

var rootCmd = &cobra.Command{
	Use:   "gemini",
	Short: "Gemini CLI is a command-line tool for interacting with Gemini API",
	Long: `A fast and flexible command-line interface for Google Gemini API.
	
Gemini CLI allows you to interact with Gemini models,
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
		ctx := context.Background()
		var client *api.Client
		var err error

		// Use globalCliConfig for authentication and model
		if globalCliConfig.SelectedAuthType != nil && *globalCliConfig.SelectedAuthType == "oauth" {
			token, err := auth.LoadToken()
			if err == nil && token.Valid() {
				// Use OAuth2 client
				oauthConfig := auth.GetOAuth2Config(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET")) // Still using env for client ID/secret for now
				httpClient := auth.GetHTTPClient(ctx, oauthConfig, token)
				client, err = api.NewClient(ctx, "", httpClient, globalCliConfig.Model) // Pass modelName
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error creating API client: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("Using OAuth2 for authentication.")
			} else {
				fmt.Println("Error: OAuth2 selected but no valid token found. Please run 'gemini auth'.")
				os.Exit(1)
			}
		} else { // Fallback to API key if no auth type selected or not oauth
			apiKey := os.Getenv("GEMINI_API_KEY") // Still using env for API key for now
			if apiKey == "" {
				fmt.Println("Error: GEMINI_API_KEY environment variable not set and no valid OAuth2 token found.")
				fmt.Println("Please get your API key from https://aistudio.google.com/apikey or run 'gemini auth'.")
				os.Exit(1)
			}
			client, err = api.NewClient(ctx, apiKey, nil, globalCliConfig.Model) // Pass modelName
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating API client: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Using API key for authentication.")
		}

		prompt := args[0]

		toolRegistry := tool_pkg.NewToolRegistry()
		toolRegistry.RegisterTool(&tool_pkg.ReadTool{})
		toolRegistry.RegisterTool(&tool_pkg.ListFilesTool{})
		toolRegistry.RegisterTool(&tool_pkg.WriteFileTool{}) // WriteFileToolを登録

		currentPrompt := prompt
		for { // 無限ループで対話を続ける
			fmt.Printf("Sending prompt to Gemini: \"%s\"\n", currentPrompt)
			stream, err := client.GenerateContentStream(ctx, currentPrompt, &shared.Tools{FunctionDeclarations: toolRegistry.GetFunctionDeclarations()})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating content: %v\n", err)
				os.Exit(1)
			}

			var fullTextResponse string
			var functionCalls []shared.FunctionCall

			// ストリームから応答を読み込むループ
			for {
				resp, err := stream.Next()
				if err == io.EOF {
					break // ストリームの終端に達したらループを抜ける
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error streaming response: %v\n", err)
					os.Exit(1)
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

			fmt.Println() // ストリーム応答の後に改行を追加

			if len(functionCalls) > 0 {
				fc := functionCalls[0] // 簡略化のため、一度に1つの関数呼び出しを想定
				fmt.Printf("\nGemini called tool: %s with args: %v\n", fc.Name, fc.Args)

				calledTool, ok := toolRegistry.GetTool(fc.Name)
				if !ok {
					fmt.Fprintf(os.Stderr, "Error: Tool %s not found.\n", fc.Name)
					os.Exit(1)
				}

				toolOutput, err := calledTool.Execute(context.Background(), fc.Args)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error executing tool %s: %v\n", fc.Name, err)
					os.Exit(1)
				}

				fmt.Printf("Tool %s output:\n%s\n", fc.Name, toolOutput)
				currentPrompt = fmt.Sprintf("Tool %s returned: %s", fc.Name, toolOutput) // ツール出力をGeminiに送り返す
			} else if fullTextResponse != "" {
				break // Geminiがテキストで応答したら対話ループを終了
			} else {
				fmt.Println("No text or function call in Gemini's response.")
				break // テキストも関数呼び出しもなければループを終了
			}
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
		ctx := context.Background()
		var client *api.Client
		var err error
		// Use globalCliConfig for authentication and model
		if globalCliConfig.SelectedAuthType != nil && *globalCliConfig.SelectedAuthType == "oauth" {
			token, err := auth.LoadToken()
			if err == nil && token.Valid() {
				// Use OAuth2 client
				oauthConfig := auth.GetOAuth2Config(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET")) // Still using env for client ID/secret for now
				httpClient := auth.GetHTTPClient(ctx, oauthConfig, token)
				client, err = api.NewClient(ctx, "", httpClient, globalCliConfig.Model) // Pass modelName
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error creating API client: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("Using OAuth2 for authentication.")
			} else {
				fmt.Println("Error: OAuth2 selected but no valid token found. Please run 'gemini auth'.")
				os.Exit(1)
			}
		} else { // Fallback to API key if no auth type selected or not oauth
			apiKey := os.Getenv("GEMINI_API_KEY") // Still using env for API key for now
			if apiKey == "" {
				fmt.Println("Error: GEMINI_API_KEY environment variable not set and no valid OAuth2 token found.")
				fmt.Println("Please get your API key from https://aistudio.google.com/apikey or run 'gemini auth'.")
				os.Exit(1)
			}
			client, err = api.NewClient(ctx, apiKey, nil, globalCliConfig.Model) // Pass modelName
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating API client: %v\n", err)
				os.Exit(1)
			}
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
        
        // GenerateContentStream を使用
        stream, err := client.GenerateContentStream(ctx, fullPrompt, nil)
        
        var generatedContent string
        for {
            resp, err := stream.Next()
            if err == io.EOF {
                break
            }
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error streaming generated code: %v\n", err)
                os.Exit(1)
            }
            if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
                generatedContent += string(resp.Candidates[0].Content.Parts[0].(genai.Text))
            }
        }

        if generatedContent != "" {
            fmt.Println("\nGenerated Code:")
            fmt.Println(generatedContent)
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
		if err := auth.SaveToken(token); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving token: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Authentication successful! Token saved.")
		fmt.Printf("Access Token: %s\n", token.AccessToken)
		fmt.Printf("Refresh Token: %s\n", token.RefreshToken)
	},
}

var writeFileCmd = &cobra.Command{
	Use:   "write-file [filePath] [content]",
	Short: "Writes content to a specified file",
	Args:  cobra.ExactArgs(2), // ファイルパスとコンテンツの2つが必要
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		content := []byte(args[1])

		err := filesystem.WriteFile(filePath, content);
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully wrote to %s\n", filePath)
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(listFilesCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(generateCodeCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(writeFileCmd)

	// Add global flags from config.ts to rootCmd
	rootCmd.PersistentFlags().StringP("model", "m", os.Getenv("GEMINI_MODEL"), "Model") // Default from env or config.go
	rootCmd.PersistentFlags().StringP("prompt", "p", "", "Prompt. Appended to input on stdin (if any).")
	rootCmd.PersistentFlags().BoolP("sandbox", "s", false, "Run in sandbox?") // Default to false, will be overridden by settings
	rootCmd.PersistentFlags().String("sandbox-image", "", "Sandbox image URI.")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Run in debug mode?")
	rootCmd.PersistentFlags().BoolP("all_files", "a", false, "Include ALL files in context?")
	rootCmd.PersistentFlags().Bool("show_memory_usage", false, "Show memory usage in status bar")
	rootCmd.PersistentFlags().BoolP("yolo", "y", false, "Automatically accept all actions (aka YOLO mode).")
	rootCmd.PersistentFlags().Bool("telemetry", false, "Enable telemetry?") // Default to false, will be overridden by settings
	rootCmd.PersistentFlags().String("telemetry-target", "", "Set the telemetry target (local or gcp). Overrides settings files.")
	rootCmd.PersistentFlags().String("telemetry-otlp-endpoint", "", "Set the OTLP endpoint for telemetry. Overrides environment variables and settings files.")
	rootCmd.PersistentFlags().Bool("telemetry-log-prompts", false, "Enable or disable logging of user prompts for telemetry. Overrides settings files.")
	rootCmd.PersistentFlags().BoolP("checkpointing", "c", false, "Enables checkpointing of file edits")


	// list-files コマンドに --ext フラグを追加
	listFilesCmd.Flags().StringSliceP("ext", "e", []string{}, "Comma-separated list of file extensions to filter (e.g., .go,.txt)")

	// context コマンドに --ext フラグを追加
	contextCmd.Flags().StringSliceP("ext", "e", []string{}, "Comma-separated list of file extensions to filter (e.g., .go,.txt)")

	// generate-code コマンドに --context-dir と --ext フラグを追加
	generateCodeCmd.Flags().StringP("context-dir", "c", "", "Directory to use as context for code generation")
	generateCodeCmd.Flags().StringSliceP("ext", "e", []string{}, "Comma-separated list of file extensions to filter in context directory (e.g., .go,.txt)")
}

func main() {
	// Generate a session ID
	sessionId := uuid.New().String()

	// Load CLI configuration
	var configErrors []errors.SettingError
	globalCliConfig, configErrors = config_pkg.LoadCliConfig(os.Getenv("PWD"), sessionId, rootCmd)

	if len(configErrors) > 0 {
		for _, err := range configErrors {
			fmt.Fprintf(os.Stderr, "Error in %s: %s\n", err.Path, err.Message)
		}
		fmt.Fprintf(os.Stderr, "Please fix the errors and try again.\n")
        os.Exit(1)
    }

    // Initialize Telemetry
    telemetry.InitializeTelemetry(globalCliConfig)
    defer telemetry.ShutdownTelemetry(context.Background())

    // Now, globalCliConfig contains all merged settings and command-line arguments.
    // It can be accessed by other command Run functions.

	// Handle sandbox logic
	if os.Getenv("SANDBOX") == "" && globalCliConfig.Sandbox != nil {
		// Check if sandbox is enabled (boolean true or non-empty string)
		sandboxEnabled := false
		sandboxImage := ""

		switch s := globalCliConfig.Sandbox.(type) {
		case bool:
			sandboxEnabled = s
		case string:
			sandboxEnabled = (s != "")
			sandboxImage = s
		}

		if sandboxEnabled {
			fmt.Println("Entering sandbox...")

			// Validate authentication before entering sandbox if OAuth is selected
			var err error
			if globalCliConfig.SelectedAuthType != nil && *globalCliConfig.SelectedAuthType == "oauth" {
								err = auth.ValidateAuthMethod(*globalCliConfig.SelectedAuthType)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error validating auth method before sandbox: %v\n", err)
					os.Exit(1)
				}
				// Refresh auth (equivalent to config.refreshAuth in JS)
				// This part is tricky as refreshAuth is usually tied to API client creation.
				// For now, we'll assume the token is valid or will be refreshed upon API client creation.
				// A more robust solution might involve a dedicated auth refresh function.
				fmt.Println("OAuth2 authentication validated for sandbox.")
			}

			cmdArgs := os.Args[1:]
			cmdPath := os.Args[0]

			// Add SANDBOX=true to environment variables
			env := os.Environ()
			env = append(env, "SANDBOX=true")

			// If a sandbox image is specified, pass it as an argument or environment variable
			if sandboxImage != "" {
				env = append(env, fmt.Sprintf("GEMINI_SANDBOX_IMAGE=%s", sandboxImage))
			}

			cmd := exec.Command(cmdPath, cmdArgs...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = env

			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error running in sandbox: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0) // Exit current process after sandbox finishes
		}
	}

	// Check if stdin is a TTY
	isTTY := isStdinTTY()

	// Get prompt from command-line arguments
	input := globalCliConfig.Prompt

	if isTTY && input == "" {
		// Interactive mode: If no prompt is provided and it's a TTY,
		// let Cobra handle the default behavior (e.g., showing help or welcome message).
		// The actual interactive UI (like Ink) is not directly ported here.
	} else {
		// Non-interactive mode
		// If not a TTY, read from stdin
		if !isTTY {
			stdinBytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				os.Exit(1)
			}
			if len(stdinBytes) > 0 {
				input += string(stdinBytes)
			}
		}

		if input == "" {
			fmt.Fprintf(os.Stderr, "No input provided via stdin or command line.\n")
			os.Exit(1)
		}

		ctx := context.Background()
		var client *api.Client
		var err error

		// Use globalCliConfig for authentication and model
		if globalCliConfig.SelectedAuthType != nil && *globalCliConfig.SelectedAuthType == "oauth" {
			token, err := auth.LoadToken()
			if err == nil && token.Valid() {
				// Use OAuth2 client
				oauthConfig := auth.GetOAuth2Config(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"))
				httpClient := auth.GetHTTPClient(ctx, oauthConfig, token)
				client, err = api.NewClient(ctx, "", httpClient, globalCliConfig.Model)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error creating API client: %v\n", err)
					os.Exit(1)
				}
			} else {
				fmt.Println("Error: OAuth2 selected but no valid token found. Please run 'gemini auth'.")
				os.Exit(1)
			}
		} else { // Fallback to API key if no auth type selected or not oauth
			apiKey := os.Getenv("GEMINI_API_KEY")
			if apiKey == "" {
				fmt.Println("Error: GEMINI_API_KEY environment variable not set and no valid OAuth2 token found.")
				fmt.Println("Please get your API key from https://aistudio.google.com/apikey or run 'gemini auth'.")
				os.Exit(1)
			}
			client, err = api.NewClient(ctx, apiKey, nil, globalCliConfig.Model)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating API client: %v\n", err)
				os.Exit(1)
			}
		}

		toolRegistry := tool_pkg.NewToolRegistry()
		toolRegistry.RegisterTool(&tool_pkg.ReadTool{})
		toolRegistry.RegisterTool(&tool_pkg.ListFilesTool{})
		toolRegistry.RegisterTool(&tool_pkg.WriteFileTool{})

		// Run non-interactive mode
		if err := api.RunNonInteractive(ctx, globalCliConfig, client, toolRegistry, input); err != nil {
			fmt.Fprintf(os.Stderr, "Non-interactive execution failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0) // Exit after non-interactive execution
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// isStdinTTY checks if os.Stdin is connected to a terminal.
func isStdinTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}