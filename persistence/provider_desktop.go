//go:build !js && !wasm

package persistence

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
)

// desktopProvider implements PersistenceProvider using file-based storage
type desktopProvider struct{}

// newDesktopProvider creates a new desktop file-based persistence provider
func newDesktopProvider() PersistenceProvider {
	return &desktopProvider{}
}

// getConfigPath returns the full path to the config directory
func (p *desktopProvider) getConfigPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, configDir), nil
}

// getScoresPath returns the full path to the high scores file
func (p *desktopProvider) getScoresPath() (string, error) {
	configPath, err := p.getConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, scoresFile), nil
}

// getSettingsPath returns the full path to the settings file
func (p *desktopProvider) getSettingsPath() (string, error) {
	configPath, err := p.getConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, settingsFile), nil
}

// LoadHighScores loads high scores from disk
// Returns an empty list if file doesn't exist
func (p *desktopProvider) LoadHighScores() (*HighScores, error) {
	scoresPath, err := p.getScoresPath()
	if err != nil {
		return &HighScores{Entries: []HighScore{}}, err
	}

	// If file doesn't exist, return empty list
	if _, err := os.Stat(scoresPath); os.IsNotExist(err) {
		return &HighScores{Entries: []HighScore{}}, nil
	}

	// Read file
	data, err := os.ReadFile(scoresPath)
	if err != nil {
		return &HighScores{Entries: []HighScore{}}, err
	}

	// Parse JSON
	var scores HighScores
	if err := json.Unmarshal(data, &scores); err != nil {
		return &HighScores{Entries: []HighScore{}}, err
	}

	return &scores, nil
}

// SaveHighScores saves high scores to disk
func (p *desktopProvider) SaveHighScores(scores *HighScores) error {
	configPath, err := p.getConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	scoresPath, err := p.getScoresPath()
	if err != nil {
		return err
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(scores, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(scoresPath, data, 0644)
}

// LoadSettings loads game settings from disk
// Returns default settings if file doesn't exist
func (p *desktopProvider) LoadSettings() (*Settings, error) {
	settingsPath, err := p.getSettingsPath()
	if err != nil {
		return &Settings{
			SoundMuted:  false,
			MusicMuted:  false,
			SoundVolume: 1.0,
			MusicVolume: 1.0,
			ChatEnabled: true,
		}, err
	}

	// If file doesn't exist, return default settings
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return &Settings{
			SoundMuted:  false,
			MusicMuted:  false,
			SoundVolume: 1.0,
			MusicVolume: 1.0,
			ChatEnabled: true,
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
			ChatEnabled: true,
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
			ChatEnabled: true,
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
func (p *desktopProvider) SaveSettings(settings *Settings) error {
	configPath, err := p.getConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	settingsPath, err := p.getSettingsPath()
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
