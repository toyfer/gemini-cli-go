package errors

// SettingsError represents an error encountered while loading settings.
type SettingError struct {
	Message string
	Path    string
}
