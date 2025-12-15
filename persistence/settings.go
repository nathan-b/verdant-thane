package persistence

import (
	"path/filepath"
)

// Settings represents game settings that persist across sessions
type Settings struct {
	SoundMuted  bool    `json:"sound_muted"`
	MusicMuted  bool    `json:"music_muted"`
	SoundVolume float64 `json:"sound_volume"` // 0.0 to 1.0
	MusicVolume float64 `json:"music_volume"` // 0.0 to 1.0
	ChatEnabled bool    `json:"chat_enabled"` // Whether to show chat window
}

const settingsFile = "settings.json"

// GetSettingsPath returns the full path to the settings file
func GetSettingsPath() (string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, settingsFile), nil
}

// LoadSettings loads game settings using the default persistence provider
// Returns default settings if file doesn't exist
// This function is kept for backwards compatibility with existing code and tests
func LoadSettings() (*Settings, error) {
	provider := NewProvider()
	return provider.LoadSettings()
}

// SaveSettings saves game settings using the default persistence provider
// This function is kept for backwards compatibility with existing code and tests
func SaveSettings(settings *Settings) error {
	provider := NewProvider()
	return provider.SaveSettings(settings)
}
