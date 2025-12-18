//go:build !js && !wasm

package persistence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// testDesktopProvider is a test-friendly version that uses a custom config path
type testDesktopProvider struct {
	configPath string
}

func (p *testDesktopProvider) getConfigPath() (string, error) {
	return p.configPath, nil
}

func (p *testDesktopProvider) getScoresPath() (string, error) {
	return filepath.Join(p.configPath, scoresFile), nil
}

func (p *testDesktopProvider) getSettingsPath() (string, error) {
	return filepath.Join(p.configPath, settingsFile), nil
}

func (p *testDesktopProvider) LoadHighScores() (*HighScores, error) {
	scoresPath, err := p.getScoresPath()
	if err != nil {
		return &HighScores{Entries: []HighScore{}}, err
	}

	if _, err := os.Stat(scoresPath); os.IsNotExist(err) {
		return &HighScores{Entries: []HighScore{}}, nil
	}

	data, err := os.ReadFile(scoresPath)
	if err != nil {
		return &HighScores{Entries: []HighScore{}}, err
	}

	var scores HighScores
	if err := json.Unmarshal(data, &scores); err != nil {
		return &HighScores{Entries: []HighScore{}}, err
	}

	return &scores, nil
}

func (p *testDesktopProvider) SaveHighScores(scores *HighScores) error {
	configPath, err := p.getConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	scoresPath, err := p.getScoresPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(scores, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(scoresPath, data, 0644)
}

func (p *testDesktopProvider) LoadSettings() (*Settings, error) {
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

	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return &Settings{
			SoundMuted:  false,
			MusicMuted:  false,
			SoundVolume: 1.0,
			MusicVolume: 1.0,
			ChatEnabled: true,
		}, nil
	}

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

	// Apply defaults for backwards compatibility
	if settings.SoundVolume == 0 {
		settings.SoundVolume = 1.0
	}
	if settings.MusicVolume == 0 {
		settings.MusicVolume = 1.0
	}

	return &settings, nil
}

func (p *testDesktopProvider) SaveSettings(settings *Settings) error {
	configPath, err := p.getConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	settingsPath, err := p.getSettingsPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, data, 0644)
}

// TestDesktopProvider_LoadHighScores_EmptyFile tests loading when file doesn't exist
func TestDesktopProvider_LoadHighScores_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	provider := &testDesktopProvider{configPath: tempDir}

	scores, err := provider.LoadHighScores()
	if err != nil {
		t.Errorf("LoadHighScores should not error on missing file: %v", err)
	}

	if scores == nil {
		t.Fatal("LoadHighScores should return non-nil scores")
	}

	if len(scores.Entries) != 0 {
		t.Errorf("Empty file should return empty entries, got %d entries", len(scores.Entries))
	}
}

// TestDesktopProvider_SaveAndLoadHighScores tests round-trip save/load
func TestDesktopProvider_SaveAndLoadHighScores(t *testing.T) {
	tempDir := t.TempDir()
	provider := &testDesktopProvider{configPath: tempDir}

	testScores := &HighScores{
		Entries: []HighScore{
			{
				Name:   "Alice",
				Score:  1000,
				Kills:  50,
				Deaths: 5,
				Date:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				Name:   "Bob",
				Score:  800,
				Kills:  40,
				Deaths: 10,
				Date:   time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC),
			},
		},
	}

	err := provider.SaveHighScores(testScores)
	if err != nil {
		t.Fatalf("SaveHighScores failed: %v", err)
	}

	scoresPath := filepath.Join(tempDir, scoresFile)
	if _, err := os.Stat(scoresPath); os.IsNotExist(err) {
		t.Error("Scores file should be created")
	}

	loaded, err := provider.LoadHighScores()
	if err != nil {
		t.Fatalf("LoadHighScores failed: %v", err)
	}

	if len(loaded.Entries) != len(testScores.Entries) {
		t.Errorf("Loaded %d entries, expected %d", len(loaded.Entries), len(testScores.Entries))
	}

	for i, entry := range loaded.Entries {
		expected := testScores.Entries[i]
		if entry.Name != expected.Name {
			t.Errorf("Entry %d: Name = %q, want %q", i, entry.Name, expected.Name)
		}
		if entry.Score != expected.Score {
			t.Errorf("Entry %d: Score = %d, want %d", i, entry.Score, expected.Score)
		}
		if entry.Kills != expected.Kills {
			t.Errorf("Entry %d: Kills = %d, want %d", i, entry.Kills, expected.Kills)
		}
		if entry.Deaths != expected.Deaths {
			t.Errorf("Entry %d: Deaths = %d, want %d", i, entry.Deaths, expected.Deaths)
		}
		if !entry.Date.Equal(expected.Date) {
			t.Errorf("Entry %d: Date = %v, want %v", i, entry.Date, expected.Date)
		}
	}
}

// TestDesktopProvider_LoadHighScores_CorruptJSON tests handling of corrupt JSON
func TestDesktopProvider_LoadHighScores_CorruptJSON(t *testing.T) {
	tempDir := t.TempDir()
	provider := &testDesktopProvider{configPath: tempDir}

	scoresPath := filepath.Join(tempDir, scoresFile)
	err := os.WriteFile(scoresPath, []byte("not valid JSON{{{"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	scores, err := provider.LoadHighScores()
	if err == nil {
		t.Error("LoadHighScores should return error for corrupt JSON")
	}

	if scores == nil {
		t.Error("LoadHighScores should return non-nil scores even on error")
	}
	if scores.Entries == nil {
		t.Error("LoadHighScores should return non-nil Entries even on error")
	}
}

// TestDesktopProvider_SaveHighScores_CreatesDirectory tests directory creation
func TestDesktopProvider_SaveHighScores_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nested", "config", "dir")
	provider := &testDesktopProvider{configPath: configPath}

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatal("Config directory should not exist yet")
	}

	testScores := &HighScores{
		Entries: []HighScore{
			{Name: "Test", Score: 100, Kills: 10, Deaths: 1, Date: time.Now()},
		},
	}

	err := provider.SaveHighScores(testScores)
	if err != nil {
		t.Fatalf("SaveHighScores failed: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("SaveHighScores should create config directory")
	}

	scoresPath := filepath.Join(configPath, scoresFile)
	if _, err := os.Stat(scoresPath); os.IsNotExist(err) {
		t.Error("SaveHighScores should create scores file")
	}
}

// TestDesktopProvider_LoadSettings_EmptyFile tests loading when file doesn't exist
func TestDesktopProvider_LoadSettings_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	provider := &testDesktopProvider{configPath: tempDir}

	settings, err := provider.LoadSettings()
	if err != nil {
		t.Errorf("LoadSettings should not error on missing file: %v", err)
	}

	if settings == nil {
		t.Fatal("LoadSettings should return non-nil settings")
	}

	if settings.SoundMuted {
		t.Error("Default SoundMuted should be false")
	}
	if settings.MusicMuted {
		t.Error("Default MusicMuted should be false")
	}
	if settings.SoundVolume != 1.0 {
		t.Errorf("Default SoundVolume = %f, want 1.0", settings.SoundVolume)
	}
	if settings.MusicVolume != 1.0 {
		t.Errorf("Default MusicVolume = %f, want 1.0", settings.MusicVolume)
	}
	if !settings.ChatEnabled {
		t.Error("Default ChatEnabled should be true")
	}
}

// TestDesktopProvider_SaveAndLoadSettings tests round-trip save/load
func TestDesktopProvider_SaveAndLoadSettings(t *testing.T) {
	tempDir := t.TempDir()
	provider := &testDesktopProvider{configPath: tempDir}

	testSettings := &Settings{
		SoundMuted:  true,
		MusicMuted:  false,
		SoundVolume: 0.75,
		MusicVolume: 0.5,
		ChatEnabled: false,
	}

	err := provider.SaveSettings(testSettings)
	if err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	settingsPath := filepath.Join(tempDir, settingsFile)
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("Settings file should be created")
	}

	loaded, err := provider.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	if loaded.SoundMuted != testSettings.SoundMuted {
		t.Errorf("SoundMuted = %v, want %v", loaded.SoundMuted, testSettings.SoundMuted)
	}
	if loaded.MusicMuted != testSettings.MusicMuted {
		t.Errorf("MusicMuted = %v, want %v", loaded.MusicMuted, testSettings.MusicMuted)
	}
	if loaded.SoundVolume != testSettings.SoundVolume {
		t.Errorf("SoundVolume = %f, want %f", loaded.SoundVolume, testSettings.SoundVolume)
	}
	if loaded.MusicVolume != testSettings.MusicVolume {
		t.Errorf("MusicVolume = %f, want %f", loaded.MusicVolume, testSettings.MusicVolume)
	}
	if loaded.ChatEnabled != testSettings.ChatEnabled {
		t.Errorf("ChatEnabled = %v, want %v", loaded.ChatEnabled, testSettings.ChatEnabled)
	}
}

// TestDesktopProvider_LoadSettings_CorruptJSON tests handling of corrupt JSON
func TestDesktopProvider_LoadSettings_CorruptJSON(t *testing.T) {
	tempDir := t.TempDir()
	provider := &testDesktopProvider{configPath: tempDir}

	settingsPath := filepath.Join(tempDir, settingsFile)
	err := os.WriteFile(settingsPath, []byte("invalid JSON}}}}"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	settings, err := provider.LoadSettings()
	if err == nil {
		t.Error("LoadSettings should return error for corrupt JSON")
	}

	if settings == nil {
		t.Error("LoadSettings should return non-nil settings even on error")
	}
	if settings.SoundVolume != 1.0 {
		t.Errorf("Error case should return default SoundVolume 1.0, got %f", settings.SoundVolume)
	}
}

// TestDesktopProvider_LoadSettings_BackwardsCompatibility tests volume defaults
func TestDesktopProvider_LoadSettings_BackwardsCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	provider := &testDesktopProvider{configPath: tempDir}

	oldSettings := map[string]interface{}{
		"sound_muted":  false,
		"music_muted":  false,
		"chat_enabled": true,
		// No volume fields (old format)
	}

	data, err := json.MarshalIndent(oldSettings, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	settingsPath := filepath.Join(tempDir, settingsFile)
	err = os.WriteFile(settingsPath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	loaded, err := provider.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Zero volumes should be defaulted to 1.0 for backwards compatibility
	if loaded.SoundVolume != 1.0 {
		t.Errorf("Zero SoundVolume should default to 1.0, got %f", loaded.SoundVolume)
	}
	if loaded.MusicVolume != 1.0 {
		t.Errorf("Zero MusicVolume should default to 1.0, got %f", loaded.MusicVolume)
	}
}

// TestDesktopProvider_SaveSettings_CreatesDirectory tests directory creation
func TestDesktopProvider_SaveSettings_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "deep", "nested", "config")
	provider := &testDesktopProvider{configPath: configPath}

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatal("Config directory should not exist yet")
	}

	testSettings := &Settings{
		SoundMuted:  false,
		MusicMuted:  false,
		SoundVolume: 1.0,
		MusicVolume: 1.0,
		ChatEnabled: true,
	}

	err := provider.SaveSettings(testSettings)
	if err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("SaveSettings should create config directory")
	}

	settingsPath := filepath.Join(configPath, settingsFile)
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("SaveSettings should create settings file")
	}
}

// TestDesktopProvider_JSONFormatting tests that JSON is properly formatted
func TestDesktopProvider_JSONFormatting(t *testing.T) {
	tempDir := t.TempDir()
	provider := &testDesktopProvider{configPath: tempDir}

	testScores := &HighScores{
		Entries: []HighScore{
			{Name: "Player", Score: 100, Kills: 10, Deaths: 1, Date: time.Now()},
		},
	}
	err := provider.SaveHighScores(testScores)
	if err != nil {
		t.Fatalf("SaveHighScores failed: %v", err)
	}

	scoresPath := filepath.Join(tempDir, scoresFile)
	data, err := os.ReadFile(scoresPath)
	if err != nil {
		t.Fatalf("Failed to read scores file: %v", err)
	}

	content := string(data)
	if !containsIndentation(content) {
		t.Error("JSON should be formatted with indentation for readability")
	}

	var parsed HighScores
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("Saved JSON should be valid: %v", err)
	}
}

// TestDesktopProvider_FilePermissions tests that files are created with correct permissions
func TestDesktopProvider_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	provider := &testDesktopProvider{configPath: tempDir}

	testSettings := &Settings{
		SoundVolume: 1.0,
		MusicVolume: 1.0,
	}
	err := provider.SaveSettings(testSettings)
	if err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	settingsPath := filepath.Join(tempDir, settingsFile)
	info, err := os.Stat(settingsPath)
	if err != nil {
		t.Fatalf("Failed to stat settings file: %v", err)
	}

	mode := info.Mode()
	if mode.Perm()&0600 != 0600 {
		t.Errorf("File should be readable and writable by owner, got permissions %o", mode.Perm())
	}
}

// TestDesktopProvider_MultipleOperations tests multiple save/load cycles
func TestDesktopProvider_MultipleOperations(t *testing.T) {
	tempDir := t.TempDir()
	provider := &testDesktopProvider{configPath: tempDir}

	for i := 1; i <= 5; i++ { // Start from 1 to avoid 0.0 volume (which defaults to 1.0)
		settings := &Settings{
			SoundVolume: float64(i) * 0.2,
			MusicVolume: float64(i) * 0.1,
			ChatEnabled: i%2 == 0,
		}

		err := provider.SaveSettings(settings)
		if err != nil {
			t.Fatalf("SaveSettings iteration %d failed: %v", i, err)
		}

		loaded, err := provider.LoadSettings()
		if err != nil {
			t.Fatalf("LoadSettings iteration %d failed: %v", i, err)
		}

		if loaded.SoundVolume != settings.SoundVolume {
			t.Errorf("Iteration %d: SoundVolume mismatch, got %f want %f", i, loaded.SoundVolume, settings.SoundVolume)
		}
	}
}

// TestDesktopProvider_EmptyScores tests saving and loading empty scores
func TestDesktopProvider_EmptyScores(t *testing.T) {
	tempDir := t.TempDir()
	provider := &testDesktopProvider{configPath: tempDir}

	emptyScores := &HighScores{
		Entries: []HighScore{},
	}

	err := provider.SaveHighScores(emptyScores)
	if err != nil {
		t.Fatalf("SaveHighScores with empty entries failed: %v", err)
	}

	loaded, err := provider.LoadHighScores()
	if err != nil {
		t.Fatalf("LoadHighScores failed: %v", err)
	}

	if len(loaded.Entries) != 0 {
		t.Errorf("Empty scores should remain empty, got %d entries", len(loaded.Entries))
	}
}

// TestDesktopProvider_LargeScoresList tests handling of many scores
func TestDesktopProvider_LargeScoresList(t *testing.T) {
	tempDir := t.TempDir()
	provider := &testDesktopProvider{configPath: tempDir}

	largeScores := &HighScores{
		Entries: make([]HighScore, 100),
	}

	for i := 0; i < 100; i++ {
		largeScores.Entries[i] = HighScore{
			Name:   "Player" + string(rune(i)),
			Score:  1000 - i*10,
			Kills:  50 - i,
			Deaths: i,
			Date:   time.Now().Add(time.Duration(-i) * time.Hour),
		}
	}

	err := provider.SaveHighScores(largeScores)
	if err != nil {
		t.Fatalf("SaveHighScores with large list failed: %v", err)
	}

	loaded, err := provider.LoadHighScores()
	if err != nil {
		t.Fatalf("LoadHighScores failed: %v", err)
	}

	if len(loaded.Entries) != 100 {
		t.Errorf("Large scores list: got %d entries, want 100", len(loaded.Entries))
	}
}

// TestDesktopProvider_SpecialCharactersInData tests handling of special characters
func TestDesktopProvider_SpecialCharactersInData(t *testing.T) {
	tempDir := t.TempDir()
	provider := &testDesktopProvider{configPath: tempDir}

	testScores := &HighScores{
		Entries: []HighScore{
			{
				Name:   "Player \"Quotes\" <Test>",
				Score:  100,
				Kills:  10,
				Deaths: 1,
				Date:   time.Now(),
			},
			{
				Name:   "Ñoño\t\nSpecial",
				Score:  200,
				Kills:  20,
				Deaths: 2,
				Date:   time.Now(),
			},
		},
	}

	err := provider.SaveHighScores(testScores)
	if err != nil {
		t.Fatalf("SaveHighScores with special chars failed: %v", err)
	}

	loaded, err := provider.LoadHighScores()
	if err != nil {
		t.Fatalf("LoadHighScores failed: %v", err)
	}

	if len(loaded.Entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(loaded.Entries))
	}

	for i, entry := range loaded.Entries {
		if entry.Name != testScores.Entries[i].Name {
			t.Errorf("Entry %d: Name = %q, want %q", i, entry.Name, testScores.Entries[i].Name)
		}
	}
}

// TestNewDesktopProvider tests the provider constructor
func TestNewDesktopProvider(t *testing.T) {
	provider := newDesktopProvider()

	if provider == nil {
		t.Fatal("newDesktopProvider should return non-nil provider")
	}

	// Should implement PersistenceProvider interface
	var _ PersistenceProvider = provider
}

// TestNewProvider tests the factory function
func TestNewProvider(t *testing.T) {
	provider := NewProvider()

	if provider == nil {
		t.Fatal("NewProvider should return non-nil provider")
	}

	// Should implement PersistenceProvider interface
	var _ PersistenceProvider = provider
}

// Helper function to check if content contains indentation
func containsIndentation(content string) bool {
	return len(content) > 0 && (content[0] == ' ' || content[0] == '\t' ||
		len(content) > 1 && (content[1] == ' ' || content[1] == '\n'))
}
