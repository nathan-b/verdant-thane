package persistence

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// Basic Settings Creation
// ============================================================================

func TestSettingsStruct(t *testing.T) {
	settings := Settings{
		SoundMuted:  true,
		MusicMuted:  false,
		SoundVolume: 0.5,
		MusicVolume: 0.75,
		ChatEnabled: true,
	}

	if !settings.SoundMuted {
		t.Error("Expected SoundMuted=true")
	}
	if settings.MusicMuted {
		t.Error("Expected MusicMuted=false")
	}
	if settings.SoundVolume != 0.5 {
		t.Errorf("Expected SoundVolume=0.5, got %f", settings.SoundVolume)
	}
	if settings.MusicVolume != 0.75 {
		t.Errorf("Expected MusicVolume=0.75, got %f", settings.MusicVolume)
	}
	if !settings.ChatEnabled {
		t.Error("Expected ChatEnabled=true")
	}
}

// ============================================================================
// Path Functions
// ============================================================================

func TestGetSettingsPath(t *testing.T) {
	path, err := GetSettingsPath()
	if err != nil {
		t.Fatalf("GetSettingsPath returned error: %v", err)
	}

	if path == "" {
		t.Error("GetSettingsPath should return a non-empty path")
	}

	// Should end with the settings file name
	if filepath.Base(path) != settingsFile {
		t.Errorf("Expected path to end with '%s', got '%s'", settingsFile, filepath.Base(path))
	}

	// Should be in the verdant config directory
	dir := filepath.Dir(path)
	if filepath.Base(dir) != "verdant" {
		t.Errorf("Expected settings to be in 'verdant' directory, got '%s'", filepath.Base(dir))
	}
}

// ============================================================================
// Load/Save Integration Tests
// ============================================================================

func TestLoadSettingsNonExistent(t *testing.T) {
	// Backup existing file if present
	settingsPath, err := GetSettingsPath()
	if err != nil {
		t.Fatalf("GetSettingsPath returned error: %v", err)
	}

	var backupData []byte
	var hadBackup bool
	if data, err := os.ReadFile(settingsPath); err == nil {
		backupData = data
		hadBackup = true
		os.Remove(settingsPath)
	}

	// Restore backup after test
	defer func() {
		if hadBackup {
			configPath, _ := GetConfigPath()
			os.MkdirAll(configPath, 0755)
			os.WriteFile(settingsPath, backupData, 0644)
		}
	}()

	// Load from non-existent file should return defaults
	settings, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings returned error for non-existent file: %v", err)
	}

	if settings == nil {
		t.Fatal("LoadSettings should not return nil")
	}

	// Verify default values
	if settings.SoundMuted {
		t.Error("Default SoundMuted should be false")
	}
	if settings.MusicMuted {
		t.Error("Default MusicMuted should be false")
	}
	if settings.SoundVolume != 1.0 {
		t.Errorf("Default SoundVolume should be 1.0, got %f", settings.SoundVolume)
	}
	if settings.MusicVolume != 1.0 {
		t.Errorf("Default MusicVolume should be 1.0, got %f", settings.MusicVolume)
	}
	if !settings.ChatEnabled {
		t.Error("Default ChatEnabled should be true")
	}
}

func TestSaveAndLoadSettingsIntegration(t *testing.T) {
	// Backup existing file if present
	settingsPath, err := GetSettingsPath()
	if err != nil {
		t.Fatalf("GetSettingsPath returned error: %v", err)
	}

	var backupData []byte
	var hadBackup bool
	if data, err := os.ReadFile(settingsPath); err == nil {
		backupData = data
		hadBackup = true
	}

	// Restore backup after test
	defer func() {
		if hadBackup {
			os.WriteFile(settingsPath, backupData, 0644)
		} else {
			os.Remove(settingsPath)
		}
	}()

	// Create test settings
	testSettings := &Settings{
		SoundMuted:  true,
		MusicMuted:  false,
		SoundVolume: 0.7,
		MusicVolume: 0.3,
		ChatEnabled: false,
	}

	// Save using actual function
	if err := SaveSettings(testSettings); err != nil {
		t.Fatalf("SaveSettings returned error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("SaveSettings should create the settings file")
	}

	// Load using actual function
	loadedSettings, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings returned error: %v", err)
	}

	// Verify loaded data matches saved data
	if loadedSettings.SoundMuted != testSettings.SoundMuted {
		t.Errorf("SoundMuted mismatch: expected %v, got %v", testSettings.SoundMuted, loadedSettings.SoundMuted)
	}
	if loadedSettings.MusicMuted != testSettings.MusicMuted {
		t.Errorf("MusicMuted mismatch: expected %v, got %v", testSettings.MusicMuted, loadedSettings.MusicMuted)
	}
	if loadedSettings.SoundVolume != testSettings.SoundVolume {
		t.Errorf("SoundVolume mismatch: expected %f, got %f", testSettings.SoundVolume, loadedSettings.SoundVolume)
	}
	if loadedSettings.MusicVolume != testSettings.MusicVolume {
		t.Errorf("MusicVolume mismatch: expected %f, got %f", testSettings.MusicVolume, loadedSettings.MusicVolume)
	}
	if loadedSettings.ChatEnabled != testSettings.ChatEnabled {
		t.Errorf("ChatEnabled mismatch: expected %v, got %v", testSettings.ChatEnabled, loadedSettings.ChatEnabled)
	}
}

func TestLoadSettingsInvalidJSON(t *testing.T) {
	// Backup existing file if present
	settingsPath, err := GetSettingsPath()
	if err != nil {
		t.Fatalf("GetSettingsPath returned error: %v", err)
	}

	var backupData []byte
	var hadBackup bool
	if data, err := os.ReadFile(settingsPath); err == nil {
		backupData = data
		hadBackup = true
	}

	// Restore backup after test
	defer func() {
		if hadBackup {
			os.WriteFile(settingsPath, backupData, 0644)
		} else {
			os.Remove(settingsPath)
		}
	}()

	// Create config directory and write invalid JSON
	configPath, _ := GetConfigPath()
	os.MkdirAll(configPath, 0755)
	os.WriteFile(settingsPath, []byte("this is not valid json{{{"), 0644)

	// Load should return defaults with error
	settings, err := LoadSettings()
	if err == nil {
		t.Error("LoadSettings should return error for invalid JSON")
	}

	// Should still return valid default struct
	if settings == nil {
		t.Fatal("LoadSettings should not return nil even on error")
	}

	// Verify defaults are returned
	if settings.SoundVolume != 1.0 {
		t.Errorf("Should return default SoundVolume on error, got %f", settings.SoundVolume)
	}
	if settings.MusicVolume != 1.0 {
		t.Errorf("Should return default MusicVolume on error, got %f", settings.MusicVolume)
	}
}

func TestLoadSettingsVolumeDefaults(t *testing.T) {
	// Backup existing file if present
	settingsPath, err := GetSettingsPath()
	if err != nil {
		t.Fatalf("GetSettingsPath returned error: %v", err)
	}

	var backupData []byte
	var hadBackup bool
	if data, err := os.ReadFile(settingsPath); err == nil {
		backupData = data
		hadBackup = true
	}

	// Restore backup after test
	defer func() {
		if hadBackup {
			os.WriteFile(settingsPath, backupData, 0644)
		} else {
			os.Remove(settingsPath)
		}
	}()

	// Create config directory and write JSON with zero volumes
	// (simulating old settings file without volume fields)
	configPath, _ := GetConfigPath()
	os.MkdirAll(configPath, 0755)
	os.WriteFile(settingsPath, []byte(`{"sound_muted": false, "music_muted": false}`), 0644)

	// Load should apply default volumes for missing/zero values
	settings, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings returned error: %v", err)
	}

	// Volumes should default to 1.0 when zero
	if settings.SoundVolume != 1.0 {
		t.Errorf("SoundVolume should default to 1.0 when zero, got %f", settings.SoundVolume)
	}
	if settings.MusicVolume != 1.0 {
		t.Errorf("MusicVolume should default to 1.0 when zero, got %f", settings.MusicVolume)
	}
}

// ============================================================================
// Constants Tests
// ============================================================================

func TestSettingsFileConstant(t *testing.T) {
	if settingsFile != "settings.json" {
		t.Errorf("Expected settingsFile='settings.json', got '%s'", settingsFile)
	}
}
