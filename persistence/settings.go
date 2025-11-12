package persistence

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings represents game settings that persist across sessions
type Settings struct {
	SoundMuted  bool    `json:"sound_muted"`
	MusicMuted  bool    `json:"music_muted"`
	SoundVolume float64 `json:"sound_volume"` // 0.0 to 1.0
	MusicVolume float64 `json:"music_volume"` // 0.0 to 1.0
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

// LoadSettings loads game settings from disk
// Returns default settings if file doesn't exist
func LoadSettings() (*Settings, error) {
	settingsPath, err := GetSettingsPath()
	if err != nil {
		return &Settings{
			SoundMuted:  false,
			MusicMuted:  false,
			SoundVolume: 1.0,
			MusicVolume: 1.0,
		}, err
	}

	// If file doesn't exist, return default settings
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return &Settings{
			SoundMuted:  false,
			MusicMuted:  false,
			SoundVolume: 1.0,
			MusicVolume: 1.0,
		}, nil
	}

	// Read file
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return &Settings{
			SoundMuted:  false,
			MusicMuted:  false,
			SoundVolume: 1.0,
			MusicVolume: 1.0,
		}, err
	}

	// Parse JSON
	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return &Settings{
			SoundMuted:  false,
			MusicMuted:  false,
			SoundVolume: 1.0,
			MusicVolume: 1.0,
		}, err
	}

	// Apply defaults for any missing fields (for backwards compatibility)
	if settings.SoundVolume == 0 {
		settings.SoundVolume = 1.0
	}
	if settings.MusicVolume == 0 {
		settings.MusicVolume = 1.0
	}

	return &settings, nil
}

// SaveSettings saves game settings to disk
func SaveSettings(settings *Settings) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	settingsPath, err := GetSettingsPath()
	if err != nil {
		return err
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(settingsPath, data, 0644)
}
