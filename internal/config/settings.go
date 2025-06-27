package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"

	"gemini-cli-go/internal/errors"
)

const (
	SettingsDirectoryName = ".gemini"
	SettingsFileName      = "settings.json"
)

// SettingScope defines the scope of a setting.
type SettingScope string

const (
	SettingScopeUser      SettingScope = "User"
	SettingScopeWorkspace SettingScope = "Workspace"
)

// CheckpointingSettings defines settings related to checkpointing.
type CheckpointingSettings struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// AccessibilitySettings defines settings related to accessibility.
type AccessibilitySettings struct {
	DisableLoadingPhrases *bool `json:"disableLoadingPhrases,omitempty"`
}

// TelemetrySettings defines settings related to telemetry.
type TelemetrySettings struct {
	Enabled      *bool   `json:"enabled,omitempty"`
	Target       *string `json:"target,omitempty"`
	OtlpEndpoint *string `json:"otlpEndpoint,omitempty"`
	LogPrompts   *bool   `json:"logPrompts,omitempty"`
}

// BugCommandSettings defines settings for the bug command.
type BugCommandSettings struct {
	// Add fields as needed based on the JS/TS reference
}

// FileFilteringSettings defines settings for git-aware file filtering.
type FileFilteringSettings struct {
	RespectGitIgnore      *bool `json:"respectGitIgnore,omitempty"`
	EnableRecursiveFileSearch *bool `json:"enableRecursiveFileSearch,omitempty"`
}

// Settings defines the structure of the settings.json file.
type Settings struct {
	Theme                        *string                `json:"theme,omitempty"`
	SelectedAuthType             *string                `json:"selectedAuthType,omitempty"` // Corresponds to AuthType in JS/TS
	Sandbox                      interface{}            `json:"sandbox,omitempty"`         // boolean or string
	CoreTools                    []string               `json:"coreTools,omitempty"`
	ExcludeTools                 []string               `json:"excludeTools,omitempty"`
	ToolDiscoveryCommand         *string                `json:"toolDiscoveryCommand,omitempty"`
	ToolCallCommand              *string                `json:"toolCallCommand,omitempty"`
	McpServerCommand             *string                `json:"mcpServerCommand,omitempty"`
	McpServers                   map[string]interface{} `json:"mcpServers,omitempty"` // Corresponds to Record<string, MCPServerConfig>
	ShowMemoryUsage              *bool                  `json:"showMemoryUsage,omitempty"`
	ContextFileName              interface{}            `json:"contextFileName,omitempty"` // string or []string
	Accessibility                *AccessibilitySettings `json:"accessibility,omitempty"`
	Telemetry                    *TelemetrySettings     `json:"telemetry,omitempty"`
	UsageStatisticsEnabled       *bool                  `json:"usageStatisticsEnabled,omitempty"`
	PreferredEditor              *string                `json:"preferredEditor,omitempty"`
	BugCommand                   *BugCommandSettings    `json:"bugCommand,omitempty"`
	Checkpointing                *CheckpointingSettings `json:"checkpointing,omitempty"`
	AutoConfigureMaxOldSpaceSize *bool                  `json:"autoConfigureMaxOldSpaceSize,omitempty"`
	FileFiltering                *FileFilteringSettings `json:"fileFiltering,omitempty"`
	HideWindowTitle              *bool                  `json:"hideWindowTitle,omitempty"`
}

// SettingsFile represents a loaded settings file with its path.
type SettingsFile struct {
	Settings Settings
	Path     string
}

// LoadedSettings holds the user, workspace, and merged settings.
type LoadedSettings struct {
	User      SettingsFile
	Workspace SettingsFile
	Errors    []errors.SettingError
	Merged    Settings
}

// computeMergedSettings merges user and workspace settings, with workspace settings overriding user settings.
func computeMergedSettings(user, workspace Settings) Settings {
	merged := user

	// Simple merge for now. For complex types (maps, slices), deeper merge logic might be needed.
	// This assumes that if a field is present in workspace, it completely replaces the user's value.
	// For pointers, we check if the workspace value is not nil.

	if workspace.Theme != nil {
		merged.Theme = workspace.Theme
	}
	if workspace.SelectedAuthType != nil {
		merged.SelectedAuthType = workspace.SelectedAuthType
	}
	if workspace.Sandbox != nil {
		merged.Sandbox = workspace.Sandbox
	}
	if workspace.CoreTools != nil {
		merged.CoreTools = workspace.CoreTools
	}
	if workspace.ExcludeTools != nil {
		merged.ExcludeTools = workspace.ExcludeTools
	}
	if workspace.ToolDiscoveryCommand != nil {
		merged.ToolDiscoveryCommand = workspace.ToolDiscoveryCommand
	}
	if workspace.ToolCallCommand != nil {
		merged.ToolCallCommand = workspace.ToolCallCommand
	}
	if workspace.McpServerCommand != nil {
		merged.McpServerCommand = workspace.McpServerCommand
	}
	if workspace.McpServers != nil {
		merged.McpServers = workspace.McpServers
	}
	if workspace.ShowMemoryUsage != nil {
		merged.ShowMemoryUsage = workspace.ShowMemoryUsage
	}
	if workspace.ContextFileName != nil {
		merged.ContextFileName = workspace.ContextFileName
	}
	if workspace.Accessibility != nil {
		merged.Accessibility = workspace.Accessibility
	}
	if workspace.Telemetry != nil {
		merged.Telemetry = workspace.Telemetry
	}
	if workspace.UsageStatisticsEnabled != nil {
		merged.UsageStatisticsEnabled = workspace.UsageStatisticsEnabled
	}
	if workspace.PreferredEditor != nil {
		merged.PreferredEditor = workspace.PreferredEditor
	}
	if workspace.BugCommand != nil {
		merged.BugCommand = workspace.BugCommand
	}
	if workspace.Checkpointing != nil {
		merged.Checkpointing = workspace.Checkpointing
	}
	if workspace.AutoConfigureMaxOldSpaceSize != nil {
		merged.AutoConfigureMaxOldSpaceSize = workspace.AutoConfigureMaxOldSpaceSize
	}
	if workspace.FileFiltering != nil {
		merged.FileFiltering = workspace.FileFiltering
	}
	if workspace.HideWindowTitle != nil {
		merged.HideWindowTitle = workspace.HideWindowTitle
	}

	return merged
}

// resolveStringPtrEnvVars resolves environment variables in a string pointer.
func resolveStringPtrEnvVars(s **string) {
	if s != nil && *s != nil {
		**s = os.ExpandEnv(**s)
	}
}

// resolveSettingsEnvVars recursively resolves environment variables in a Settings object.
func resolveSettingsEnvVars(s *Settings) {
	if s == nil {
		return
	}

	resolveStringPtrEnvVars(&s.Theme)
	resolveStringPtrEnvVars(&s.SelectedAuthType)
	resolveStringPtrEnvVars(&s.ToolDiscoveryCommand)
	resolveStringPtrEnvVars(&s.ToolCallCommand)
	resolveStringPtrEnvVars(&s.McpServerCommand)
	resolveStringPtrEnvVars(&s.PreferredEditor)

	// Handle ContextFileName (string or []string)
	if s.ContextFileName != nil {
		switch v := s.ContextFileName.(type) {
		case string:
			s.ContextFileName = os.ExpandEnv(v)
		case []interface{}: // JSON unmarshals array of strings to []interface{}
			resolvedSlice := make([]string, len(v))
			for i, item := range v {
				if str, ok := item.(string); ok {
					resolvedSlice[i] = os.ExpandEnv(str)
				} else {
					// Handle non-string elements in the array if necessary, or log a warning
					resolvedSlice[i] = fmt.Sprintf("%v", item) // Convert to string as fallback
				}
			}
			s.ContextFileName = resolvedSlice
		}
	}

	// Recursively resolve for nested structs
	if s.Accessibility != nil {
		// No string fields in AccessibilitySettings currently, but if there were:
		// resolveStringPtrEnvVars(&s.Accessibility.SomeStringField)
	}
	if s.Telemetry != nil {
		resolveStringPtrEnvVars(&s.Telemetry.Target)
		resolveStringPtrEnvVars(&s.Telemetry.OtlpEndpoint)
	}
	if s.BugCommand != nil {
		// No string fields in BugCommandSettings currently, but if there were:
		// resolveStringPtrEnvVars(&s.BugCommand.SomeStringField)
	}
	if s.Checkpointing != nil {
		// No string fields in CheckpointingSettings currently, but if there were:
		// resolveStringPtrEnvVars(&s.Checkpointing.SomeStringField)
	}
	if s.FileFiltering != nil {
		// No string fields in FileFilteringSettings currently, but if there were:
		// resolveStringPtrEnvVars(&s.FileFiltering.SomeStringField)
	}

	// For map[string]interface{} like McpServers, we would need to iterate and resolve strings
	// This is more complex and might require reflection or specific type assertions if values are nested strings. 
	// For now, assuming direct string values in McpServers are not expected to contain env vars,
	// or that they are handled by the core logic that consumes them.
}

// LoadSettings loads settings from user and workspace directories.
// Project settings override user settings.
func LoadSettings(workspaceDir string) LoadedSettings {
	userSettings := Settings{}
	workspaceSettings := Settings{}
	settingsErrors := []errors.SettingError{}

	// Get user settings path
	home, err := homedir.Dir()
	if err != nil {
		settingsErrors = append(settingsErrors, errors.SettingError{
			Message: fmt.Sprintf("Failed to get home directory: %v", err),
			Path:    "", // No specific path for this error
		})
	}
	userSettingsPath := filepath.Join(home, SettingsDirectoryName, SettingsFileName)

	// Load user settings
	if _, err := os.Stat(userSettingsPath); err == nil {
		content, err := ioutil.ReadFile(userSettingsPath)
		if err != nil {
			settingsErrors = append(settingsErrors, errors.SettingError{
				Message: fmt.Sprintf("Failed to read user settings file: %v", err),
				Path:    userSettingsPath,
			})
		} else {
			// Remove comments before unmarshaling
			// A simple way to remove // comments is to split by newline and filter.
			// For more robust JSON comment stripping, a dedicated library might be better.
			lines := strings.Split(string(content), "\n")
			cleanContent := []string{}
			for _, line := range lines {
				trimmedLine := strings.TrimSpace(line)
				if strings.HasPrefix(trimmedLine, "//") {
					continue
				}
				cleanContent = append(cleanContent, line)
			}
			content = []byte(strings.Join(cleanContent, "\n"))

			if err := json.Unmarshal(content, &userSettings); err != nil {
				settingsErrors = append(settingsErrors, errors.SettingError{
					Message: fmt.Sprintf("Failed to parse user settings JSON: %v", err),
					Path:    userSettingsPath,
				})
			} else {
				resolveSettingsEnvVars(&userSettings) // Call the improved resolver
				// Handle legacy theme names
				if userSettings.Theme != nil {
					if *userSettings.Theme == "VS" {
						*userSettings.Theme = "Default Light" // Assuming DefaultLight.name is "Default Light"
					} else if *userSettings.Theme == "VS2015" {
						*userSettings.Theme = "Default Dark" // Assuming DefaultDark.name is "Default Dark"
					}
				}
			}
		}
	}

	// Get workspace settings path
	workspaceSettingsPath := filepath.Join(workspaceDir, SettingsDirectoryName, SettingsFileName)

	// Load workspace settings
	if _, err := os.Stat(workspaceSettingsPath); err == nil {
		content, err := ioutil.ReadFile(workspaceSettingsPath)
		if err != nil {
			settingsErrors = append(settingsErrors, errors.SettingError{
				Message: fmt.Sprintf("Failed to read workspace settings file: %v", err),
				Path:    workspaceSettingsPath,
			})
		} else {
			// Remove comments before unmarshaling
			lines := strings.Split(string(content), "\n")
			cleanContent := []string{}
			for _, line := range lines {
				trimmedLine := strings.TrimSpace(line)
				if strings.HasPrefix(trimmedLine, "//") {
					continue
				}
				cleanContent = append(cleanContent, line)
			}
			content = []byte(strings.Join(cleanContent, "\n"))

			if err := json.Unmarshal(content, &workspaceSettings); err != nil {
				settingsErrors = append(settingsErrors, errors.SettingError{
					Message: fmt.Sprintf("Failed to parse workspace settings JSON: %v", err),
					Path:    workspaceSettingsPath,
				})
			} else {
				resolveSettingsEnvVars(&workspaceSettings) // Call the improved resolver
				// Handle legacy theme names
				if workspaceSettings.Theme != nil {
					if *workspaceSettings.Theme == "VS" {
						*workspaceSettings.Theme = "Default Light"
					} else if *workspaceSettings.Theme == "VS2015" {
						*workspaceSettings.Theme = "Default Dark"
					}
				}
			}
		}
	}

	mergedSettings := computeMergedSettings(userSettings, workspaceSettings)

	return LoadedSettings{
		User:      SettingsFile{Path: userSettingsPath, Settings: userSettings},
		Workspace: SettingsFile{Path: workspaceSettingsPath, Settings: workspaceSettings},
		Errors:    settingsErrors,
		Merged:    mergedSettings,
	}
}

// SaveSettings saves the given settings file to disk.
func SaveSettings(settingsFile SettingsFile) error {
	dirPath := filepath.Dir(settingsFile.Path)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create settings directory %s: %w", dirPath, err)
		}
	}

	data, err := json.MarshalIndent(settingsFile.Settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings to JSON: %w", err)
	}

	if err := ioutil.WriteFile(settingsFile.Path, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file %s: %w", settingsFile.Path, err)
	}
	return nil
}