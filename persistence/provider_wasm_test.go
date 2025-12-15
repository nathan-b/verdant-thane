//go:build js && wasm

package persistence

import (
	"encoding/json"
	"testing"
	"time"
)

// TestHighScoresJSONSerialization tests that HighScores can be properly marshaled and unmarshaled
// This is important for WASM localStorage which stores data as JSON strings
func TestHighScoresJSONSerialization(t *testing.T) {
	original := &HighScores{
		Entries: []HighScore{
			{Name: "Player1", Score: 1000, Kills: 50, Deaths: 10, Date: time.Now()},
			{Name: "Player2", Score: 800, Kills: 40, Deaths: 15, Date: time.Now()},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal HighScores: %v", err)
	}

	// Unmarshal back
	var restored HighScores
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal HighScores: %v", err)
	}

	// Verify
	if len(restored.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(restored.Entries))
	}
	if restored.Entries[0].Name != "Player1" {
		t.Errorf("Expected name 'Player1', got '%s'", restored.Entries[0].Name)
	}
	if restored.Entries[0].Score != 1000 {
		t.Errorf("Expected score 1000, got %d", restored.Entries[0].Score)
	}
}

// TestSettingsJSONSerialization tests that Settings can be properly marshaled and unmarshaled
func TestSettingsJSONSerialization(t *testing.T) {
	original := &Settings{
		SoundMuted:  true,
		MusicMuted:  false,
		SoundVolume: 0.75,
		MusicVolume: 0.5,
		ChatEnabled: true,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal Settings: %v", err)
	}

	// Unmarshal back
	var restored Settings
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal Settings: %v", err)
	}

	// Verify
	if restored.SoundMuted != true {
		t.Error("SoundMuted not preserved")
	}
	if restored.MusicMuted != false {
		t.Error("MusicMuted not preserved")
	}
	if restored.SoundVolume != 0.75 {
		t.Errorf("Expected SoundVolume 0.75, got %f", restored.SoundVolume)
	}
	if restored.MusicVolume != 0.5 {
		t.Errorf("Expected MusicVolume 0.5, got %f", restored.MusicVolume)
	}
	if restored.ChatEnabled != true {
		t.Error("ChatEnabled not preserved")
	}
}

// TestEmptyHighScoresJSON tests that empty HighScores serializes correctly
func TestEmptyHighScoresJSON(t *testing.T) {
	original := &HighScores{Entries: []HighScore{}}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal empty HighScores: %v", err)
	}

	var restored HighScores
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal empty HighScores: %v", err)
	}

	if len(restored.Entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(restored.Entries))
	}
}

// TestDefaultSettings tests that default settings are properly structured
func TestDefaultSettings(t *testing.T) {
	defaults := &Settings{
		SoundMuted:  false,
		MusicMuted:  false,
		SoundVolume: 1.0,
		MusicVolume: 1.0,
		ChatEnabled: true,
	}

	data, err := json.Marshal(defaults)
	if err != nil {
		t.Fatalf("Failed to marshal default Settings: %v", err)
	}

	var restored Settings
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal default Settings: %v", err)
	}

	if restored.SoundVolume != 1.0 || restored.MusicVolume != 1.0 {
		t.Error("Default volumes not correct")
	}
}
