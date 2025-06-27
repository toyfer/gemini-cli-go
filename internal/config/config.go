package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv" // For .env file loading
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"gemini-cli-go/internal/errors"
)

const (
	DEFAULT_GEMINI_MODEL         = "gemini-2.5-pro" // From JS/TS reference
	DEFAULT_GEMINI_EMBEDDING_MODEL = "embedding-001" // From JS/TS reference
	DEFAULT_OTLP_ENDPOINT        = "http://localhost:4317" // From JS/TS reference
)

// CliConfig holds the merged configuration from settings files and command-line arguments.
type CliConfig struct {
	// Fields from settings.json
	Settings

	// Fields from command-line arguments (parsed by Cobra)
	Model                        string
	Prompt                       string
	Sandbox                      interface{} // boolean or string
	SandboxImage                 string
	DebugMode                    bool
	AllFiles                     bool
	ShowMemoryUsage              bool
	Yolo                         bool
	TelemetryEnabled             *bool
	TelemetryTarget              string
	TelemetryOtlpEndpoint        string
	TelemetryLogPrompts          *bool
	CheckpointingEnabled         bool

	// Other runtime configurations
	SessionID string
	TargetDir string
	CWD       string
	Proxy     string
	// FileDiscoveryService // This will be an interface, not directly in config struct
	// GitService // This will be an interface, not directly in config struct
	// BugCommand // This is already in Settings
	// ExtensionContextFilePaths []string // This will be passed separately
}

// LoadCliConfig loads the hierarchical settings and merges them with command-line arguments.
// It also handles .env file loading.
func LoadCliConfig(workspaceDir string, sessionId string, cmd *cobra.Command) (*CliConfig, []errors.SettingError) {
	// 1. Load .env files
	loadEnvironment(workspaceDir)

	// 2. Load settings from .gemini/settings.json
	loadedSettings := LoadSettings(workspaceDir)
	if len(loadedSettings.Errors) > 0 {
		return nil, loadedSettings.Errors
	}

	// 3. Parse command-line arguments using Cobra flags
	// These flags need to be defined on the cobra.Command (e.g., rootCmd)
	// and then retrieved here. This function will be called *after* cobra has parsed the flags.

	model, _ := cmd.Flags().GetString("model")
	prompt, _ := cmd.Flags().GetString("prompt")
	sandboxFlag, _ := cmd.Flags().GetBool("sandbox") // Assuming bool for now, will need to handle string later
	sandboxImage, _ := cmd.Flags().GetString("sandbox-image")
	debugMode, _ := cmd.Flags().GetBool("debug")
	allFiles, _ := cmd.Flags().GetBool("all_files")
	showMemoryUsage, _ := cmd.Flags().GetBool("show_memory_usage")
	yolo, _ := cmd.Flags().GetBool("yolo")
	telemetryEnabled, _ := cmd.Flags().GetBool("telemetry")
	telemetryTarget, _ := cmd.Flags().GetString("telemetry-target")
	telemetryOtlpEndpoint, _ := cmd.Flags().GetString("telemetry-otlp-endpoint")
	telemetryLogPrompts, _ := cmd.Flags().GetBool("telemetry-log-prompts")
	checkpointingEnabled, _ := cmd.Flags().GetBool("checkpointing")


	// 4. Merge settings and command-line arguments
	// Command-line arguments generally override settings.
	cliConfig := &CliConfig{
		Settings: loadedSettings.Merged, // Start with merged settings

		// Override with command-line arguments if provided
		Model: model,
		Prompt: prompt,
		DebugMode: debugMode,
		AllFiles: allFiles,
		ShowMemoryUsage: showMemoryUsage,
		Yolo: yolo,
		CheckpointingEnabled: checkpointingEnabled,

		// Telemetry flags
		TelemetryTarget: telemetryTarget,

		// Sandbox flags
		SandboxImage: sandboxImage,

		SessionID: sessionId,
		TargetDir: workspaceDir,
		CWD:       os.Getenv("PWD"), // Current working directory
		Proxy:     getProxyEnv(),
	}

	// Handle TelemetryOtlpEndpoint separately due to its specific merging logic
	if cmd.Flags().Changed("telemetry-otlp-endpoint") {
		cliConfig.TelemetryOtlpEndpoint = telemetryOtlpEndpoint
	} else if loadedSettings.Merged.Telemetry != nil && loadedSettings.Merged.Telemetry.OtlpEndpoint != nil {
		cliConfig.TelemetryOtlpEndpoint = *loadedSettings.Merged.Telemetry.OtlpEndpoint
	} else {
		cliConfig.TelemetryOtlpEndpoint = DEFAULT_OTLP_ENDPOINT
	}

	// Handle boolean flags that can be explicitly set to false
	if cmd.Flags().Changed("telemetry") {
		cliConfig.TelemetryEnabled = &telemetryEnabled
	} else if loadedSettings.Merged.Telemetry != nil && loadedSettings.Merged.Telemetry.Enabled != nil {
		cliConfig.TelemetryEnabled = loadedSettings.Merged.Telemetry.Enabled
	} else {
		// Default to true if not specified anywhere
		defaultTelemetry := true
		cliConfig.TelemetryEnabled = &defaultTelemetry
	}

	if cmd.Flags().Changed("telemetry-log-prompts") {
		cliConfig.TelemetryLogPrompts = &telemetryLogPrompts
	} else if loadedSettings.Merged.Telemetry != nil && loadedSettings.Merged.Telemetry.LogPrompts != nil {
		cliConfig.TelemetryLogPrompts = loadedSettings.Merged.Telemetry.LogPrompts
	} else {
		// Default to true if not specified anywhere
		defaultLogPrompts := true
		cliConfig.TelemetryLogPrompts = &defaultLogPrompts
	}

	// Handle sandbox flag: can be boolean or string
	if cmd.Flags().Changed("sandbox") {
		cliConfig.Sandbox = sandboxFlag
	} else if loadedSettings.Merged.Sandbox != nil {
		cliConfig.Sandbox = loadedSettings.Merged.Sandbox
	}


	// If model is not set by flag or settings, use default
	if cliConfig.Model == "" {
		cliConfig.Model = DEFAULT_GEMINI_MODEL
	}

	return cliConfig, nil
}

// loadEnvironment loads environment variables from .env files.
// It searches for .env files hierarchically, similar to the JS/TS reference.
func loadEnvironment(startDir string) {
	envFilePath := findEnvFile(startDir)
	if envFilePath != "" {
		err := godotenv.Load(envFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error loading .env file %s: %v\n", envFilePath, err)
		}
	}
}

// findEnvFile searches for .env files hierarchically.
func findEnvFile(startDir string) string {
	currentDir := startDir
	for {
		// Prefer gemini-specific .env under .gemini directory
		geminiEnvPath := filepath.Join(currentDir, SettingsDirectoryName, ".env")
		if _, err := os.Stat(geminiEnvPath); err == nil {
			return geminiEnvPath
		}

		// Check .env in current directory
		envPath := filepath.Join(currentDir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir { // Reached root
			break
		}
		currentDir = parentDir
	}

	// Check .env under home as fallback, again preferring gemini-specific .env
	homeDir, err := homedir.Dir()
	if err == nil {
		homeGeminiEnvPath := filepath.Join(homeDir, SettingsDirectoryName, ".env")
		if _, err := os.Stat(homeGeminiEnvPath); err == nil {
			return homeGeminiEnvPath
		}
		homeEnvPath := filepath.Join(homeDir, ".env")
		if _, err := os.Stat(homeEnvPath); err == nil {
			return homeEnvPath
		}
	}
	return ""
}

// getProxyEnv retrieves proxy environment variables.
func getProxyEnv() string {
	if proxy := os.Getenv("HTTPS_PROXY"); proxy != "" {
		return proxy
	}
	if proxy := os.Getenv("https_proxy"); proxy != "" {
		return proxy
	}
	if proxy := os.Getenv("HTTP_PROXY"); proxy != "" {
		return proxy
	}
	if proxy := os.Getenv("http_proxy"); proxy != "" {
		return proxy
	}
	return ""
}