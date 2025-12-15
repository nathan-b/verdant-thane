//go:build js && wasm

package persistence

import (
	"encoding/json"
	"syscall/js"
)

// wasmProvider implements PersistenceProvider using browser localStorage
type wasmProvider struct{}

// newWasmProvider creates a new WASM localStorage-based persistence provider
func newWasmProvider() PersistenceProvider {
	return &wasmProvider{}
}

// localStorage keys
const (
	highScoresKey = "verdant_highscores"
	settingsKey   = "verdant_settings"
)

// getLocalStorage returns the browser's localStorage object
func getLocalStorage() js.Value {
	return js.Global().Get("localStorage")
}

// LoadHighScores loads high scores from browser localStorage
// Returns an empty list if data doesn't exist
func (p *wasmProvider) LoadHighScores() (*HighScores, error) {
	localStorage := getLocalStorage()

	// Get item from localStorage
	item := localStorage.Call("getItem", highScoresKey)
	if item.IsNull() || item.IsUndefined() {
		// No data stored, return empty list
		return &HighScores{Entries: []HighScore{}}, nil
	}

	// Parse JSON
	jsonStr := item.String()
	var scores HighScores
	if err := json.Unmarshal([]byte(jsonStr), &scores); err != nil {
		return &HighScores{Entries: []HighScore{}}, err
	}

	return &scores, nil
}

// SaveHighScores saves high scores to browser localStorage
func (p *wasmProvider) SaveHighScores(scores *HighScores) error {
	// Marshal to JSON
	data, err := json.Marshal(scores)
	if err != nil {
		return err
	}

	// Store in localStorage
	localStorage := getLocalStorage()
	localStorage.Call("setItem", highScoresKey, string(data))

	return nil
}

// LoadSettings loads game settings from browser localStorage
// Returns default settings if data doesn't exist
func (p *wasmProvider) LoadSettings() (*Settings, error) {
	localStorage := getLocalStorage()

	// Get item from localStorage
	item := localStorage.Call("getItem", settingsKey)
	if item.IsNull() || item.IsUndefined() {
		// No data stored, return defaults
		return &Settings{
			SoundMuted:  false,
			MusicMuted:  false,
			SoundVolume: 1.0,
			MusicVolume: 1.0,
			ChatEnabled: true,
		}, nil
	}

	// Parse JSON
	jsonStr := item.String()
	var settings Settings
	if err := json.Unmarshal([]byte(jsonStr), &settings); err != nil {
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

// SaveSettings saves game settings to browser localStorage
func (p *wasmProvider) SaveSettings(settings *Settings) error {
	// Marshal to JSON
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	// Store in localStorage
	localStorage := getLocalStorage()
	localStorage.Call("setItem", settingsKey, string(data))

	return nil
}
