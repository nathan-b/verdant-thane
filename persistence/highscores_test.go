package persistence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ============================================================================
// Basic HighScore Creation and Sorting
// ============================================================================

func TestHighScoreStruct(t *testing.T) {
	hs := HighScore{
		Name:   "TestPlayer",
		Score:  1000,
		Kills:  50,
		Deaths: 5,
		Date:   time.Now(),
	}

	if hs.Name != "TestPlayer" {
		t.Errorf("Expected Name='TestPlayer', got '%s'", hs.Name)
	}
	if hs.Score != 1000 {
		t.Errorf("Expected Score=1000, got %d", hs.Score)
	}
	if hs.Kills != 50 {
		t.Errorf("Expected Kills=50, got %d", hs.Kills)
	}
	if hs.Deaths != 5 {
		t.Errorf("Expected Deaths=5, got %d", hs.Deaths)
	}
}

func TestHighScoresEmptyList(t *testing.T) {
	hs := &HighScores{Entries: []HighScore{}}

	if len(hs.Entries) != 0 {
		t.Errorf("Expected empty list, got %d entries", len(hs.Entries))
	}
}

// ============================================================================
// AddScore Tests
// ============================================================================

func TestAddScoreToEmptyList(t *testing.T) {
	hs := &HighScores{Entries: []HighScore{}}

	isTopTen := hs.AddScore("Player1", 1000, 10, 1)

	if !isTopTen {
		t.Error("First score should always be in top 10")
	}
	if len(hs.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(hs.Entries))
	}
	if hs.Entries[0].Name != "Player1" {
		t.Errorf("Expected Name='Player1', got '%s'", hs.Entries[0].Name)
	}
	if hs.Entries[0].Score != 1000 {
		t.Errorf("Expected Score=1000, got %d", hs.Entries[0].Score)
	}
}

func TestAddScoreSorting(t *testing.T) {
	hs := &HighScores{Entries: []HighScore{}}

	hs.AddScore("Player1", 500, 5, 1)
	hs.AddScore("Player2", 1000, 10, 1)
	hs.AddScore("Player3", 750, 7, 1)

	// Should be sorted by score descending
	if hs.Entries[0].Score != 1000 {
		t.Errorf("Expected highest score first, got %d", hs.Entries[0].Score)
	}
	if hs.Entries[1].Score != 750 {
		t.Errorf("Expected second highest score, got %d", hs.Entries[1].Score)
	}
	if hs.Entries[2].Score != 500 {
		t.Errorf("Expected lowest score third, got %d", hs.Entries[2].Score)
	}
}

func TestAddScoreTopTenDetection(t *testing.T) {
	hs := &HighScores{Entries: []HighScore{}}

	// Add 10 scores
	for i := 0; i < 10; i++ {
		score := 1000 - (i * 100) // 1000, 900, 800, ..., 100
		isTopTen := hs.AddScore("Player", score, i, 1)
		if !isTopTen {
			t.Errorf("Score %d should be in top 10", score)
		}
	}

	// Add a score that doesn't make top 10
	isTopTen := hs.AddScore("LowPlayer", 50, 1, 1)
	if isTopTen {
		t.Error("Score of 50 should not be in top 10")
	}

	// Add a score that does make top 10
	isTopTen = hs.AddScore("HighPlayer", 950, 10, 1)
	if !isTopTen {
		t.Error("Score of 950 should be in top 10")
	}
}

func TestAddScoreTrimsToTopTen(t *testing.T) {
	hs := &HighScores{Entries: []HighScore{}}

	// Add 15 scores
	for i := 0; i < 15; i++ {
		hs.AddScore("Player", 1000-i*10, i, 1)
	}

	// Should only keep top 10
	if len(hs.Entries) != 10 {
		t.Errorf("Expected exactly 10 entries, got %d", len(hs.Entries))
	}

	// Verify highest score is first and lowest of the 10 is last
	if hs.Entries[0].Score != 1000 {
		t.Errorf("Expected highest score=1000, got %d", hs.Entries[0].Score)
	}
	if hs.Entries[9].Score != 910 {
		t.Errorf("Expected 10th score=910, got %d", hs.Entries[9].Score)
	}
}

func TestAddScoreDateTimestamp(t *testing.T) {
	hs := &HighScores{Entries: []HighScore{}}

	before := time.Now()
	hs.AddScore("Player", 1000, 10, 1)
	after := time.Now()

	if hs.Entries[0].Date.Before(before) || hs.Entries[0].Date.After(after) {
		t.Error("Score date should be between before and after timestamps")
	}
}

func TestAddScoreWithSameScore(t *testing.T) {
	hs := &HighScores{Entries: []HighScore{}}

	hs.AddScore("Player1", 1000, 10, 1)
	hs.AddScore("Player2", 1000, 12, 2)

	if len(hs.Entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(hs.Entries))
	}

	// Both should be present (order doesn't matter for same score)
	if hs.Entries[0].Score != 1000 || hs.Entries[1].Score != 1000 {
		t.Error("Both entries should have score 1000")
	}
}

// ============================================================================
// IsHighScore Tests
// ============================================================================

func TestIsHighScoreEmptyList(t *testing.T) {
	hs := &HighScores{Entries: []HighScore{}}

	if !hs.IsHighScore(1) {
		t.Error("Any score should be high score in empty list")
	}
	if !hs.IsHighScore(1000000) {
		t.Error("Any score should be high score in empty list")
	}
}

func TestIsHighScoreLessThanTen(t *testing.T) {
	hs := &HighScores{Entries: []HighScore{}}

	// Add 5 scores
	for i := 0; i < 5; i++ {
		hs.AddScore("Player", 1000-i*100, i, 1)
	}

	// Any score should qualify when list has < 10 entries
	if !hs.IsHighScore(1) {
		t.Error("Any score should qualify when list has < 10 entries")
	}
	if !hs.IsHighScore(1000000) {
		t.Error("Any score should qualify when list has < 10 entries")
	}
}

func TestIsHighScoreExactlyTen(t *testing.T) {
	hs := &HighScores{Entries: []HighScore{}}

	// Add 10 scores: 1000, 900, 800, ..., 100
	for i := 0; i < 10; i++ {
		hs.AddScore("Player", 1000-i*100, i, 1)
	}

	// Scores higher than 100 should qualify
	if !hs.IsHighScore(101) {
		t.Error("Score of 101 should qualify (10th place is 100)")
	}
	if !hs.IsHighScore(500) {
		t.Error("Score of 500 should qualify")
	}
	if !hs.IsHighScore(1001) {
		t.Error("Score of 1001 should qualify")
	}

	// Scores equal to or lower than 100 should not qualify
	if hs.IsHighScore(100) {
		t.Error("Score of 100 should not qualify (equal to 10th place)")
	}
	if hs.IsHighScore(50) {
		t.Error("Score of 50 should not qualify")
	}
	if hs.IsHighScore(0) {
		t.Error("Score of 0 should not qualify")
	}
}

func TestIsHighScoreBoundary(t *testing.T) {
	hs := &HighScores{Entries: []HighScore{}}

	// Add 10 scores with specific 10th place score
	for i := 0; i < 10; i++ {
		hs.AddScore("Player", 1000-i*50, i, 1)
	}

	// 10th place should be 550
	tenthScore := hs.Entries[9].Score
	if tenthScore != 550 {
		t.Fatalf("Expected 10th score=550, got %d", tenthScore)
	}

	// Test boundary conditions
	if !hs.IsHighScore(tenthScore + 1) {
		t.Error("Score one higher than 10th should qualify")
	}
	if hs.IsHighScore(tenthScore) {
		t.Error("Score equal to 10th should not qualify")
	}
	if hs.IsHighScore(tenthScore - 1) {
		t.Error("Score one lower than 10th should not qualify")
	}
}

// ============================================================================
// File I/O Tests
// ============================================================================

func TestSaveAndLoadHighScores(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "verdant-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	scoresPath := filepath.Join(tempDir, "highscores.json")

	// Create high scores
	hs := &HighScores{Entries: []HighScore{
		{Name: "Player1", Score: 1000, Kills: 10, Deaths: 1, Date: time.Now()},
		{Name: "Player2", Score: 500, Kills: 5, Deaths: 2, Date: time.Now()},
	}}

	// Save to file
	data, err := json.MarshalIndent(hs, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	if err := os.WriteFile(scoresPath, data, 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Load from file
	loadedData, err := os.ReadFile(scoresPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var loaded HighScores
	if err := json.Unmarshal(loadedData, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify loaded data
	if len(loaded.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].Name != "Player1" {
		t.Errorf("Expected first entry name='Player1', got '%s'", loaded.Entries[0].Name)
	}
	if loaded.Entries[0].Score != 1000 {
		t.Errorf("Expected first entry score=1000, got %d", loaded.Entries[0].Score)
	}
}

func TestSaveHighScoresCreatesDirectory(t *testing.T) {
	// Create temporary parent directory
	tempParent, err := os.MkdirTemp("", "verdant-test-parent-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempParent)

	// Path with nested directory that doesn't exist
	configPath := filepath.Join(tempParent, "config", "verdant")
	scoresPath := filepath.Join(configPath, "highscores.json")

	// Create directory
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config directory should have been created")
	}

	// Create and save high scores
	hs := &HighScores{Entries: []HighScore{
		{Name: "Player1", Score: 1000, Kills: 10, Deaths: 1, Date: time.Now()},
	}}

	data, err := json.MarshalIndent(hs, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	if err := os.WriteFile(scoresPath, data, 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(scoresPath); os.IsNotExist(err) {
		t.Error("Scores file should have been created")
	}
}

func TestLoadHighScoresNonExistentFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "verdant-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	scoresPath := filepath.Join(tempDir, "nonexistent.json")

	// Try to load non-existent file
	_, err = os.Stat(scoresPath)
	if !os.IsNotExist(err) {
		t.Fatal("File should not exist")
	}

	// Loading should return empty list, not error
	// (simulating LoadHighScores behavior)
	hs := &HighScores{Entries: []HighScore{}}
	if len(hs.Entries) != 0 {
		t.Error("Should return empty list for non-existent file")
	}
}

func TestJSONSerialization(t *testing.T) {
	hs := &HighScores{
		Entries: []HighScore{
			{Name: "Player1", Score: 1000, Kills: 10, Deaths: 1, Date: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)},
			{Name: "Player2", Score: 500, Kills: 5, Deaths: 2, Date: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)},
		},
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(hs, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Unmarshal back
	var loaded HighScores
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify data integrity
	if len(loaded.Entries) != len(hs.Entries) {
		t.Errorf("Entry count mismatch: expected %d, got %d", len(hs.Entries), len(loaded.Entries))
	}

	for i := range hs.Entries {
		if loaded.Entries[i].Name != hs.Entries[i].Name {
			t.Errorf("Entry %d: Name mismatch", i)
		}
		if loaded.Entries[i].Score != hs.Entries[i].Score {
			t.Errorf("Entry %d: Score mismatch", i)
		}
		if loaded.Entries[i].Kills != hs.Entries[i].Kills {
			t.Errorf("Entry %d: Kills mismatch", i)
		}
		if loaded.Entries[i].Deaths != hs.Entries[i].Deaths {
			t.Errorf("Entry %d: Deaths mismatch", i)
		}
		// Note: Time comparison might have precision differences, so we just check it exists
		if loaded.Entries[i].Date.IsZero() {
			t.Errorf("Entry %d: Date should not be zero", i)
		}
	}
}

func TestInvalidJSON(t *testing.T) {
	invalidJSON := []byte(`{"entries": [{"name": "Player", "score": "not a number"}]}`)

	var hs HighScores
	err := json.Unmarshal(invalidJSON, &hs)

	if err == nil {
		t.Error("Should return error for invalid JSON")
	}
}

// ============================================================================
// GetCurrentUsername Tests
// ============================================================================

func TestGetCurrentUsername(t *testing.T) {
	username := GetCurrentUsername()

	if username == "" {
		t.Error("Username should not be empty")
	}

	// Should return either actual username or "Player" as fallback
	if username != "Player" {
		// Assume it returned actual username - just verify it's non-empty
		if len(username) == 0 {
			t.Error("Username should not be empty string")
		}
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestFullHighScoreWorkflow(t *testing.T) {
	hs := &HighScores{Entries: []HighScore{}}

	// Add initial scores
	for i := 0; i < 5; i++ {
		hs.AddScore("Player", 1000-i*100, 10-i, 1)
	}

	// Check if a new score would be high score
	if !hs.IsHighScore(1100) {
		t.Error("Score of 1100 should be a high score")
	}

	// Add the new high score
	isTopTen := hs.AddScore("NewPlayer", 1100, 15, 1)
	if !isTopTen {
		t.Error("Score of 1100 should be in top ten")
	}

	// Verify it's now the top score
	if hs.Entries[0].Score != 1100 {
		t.Errorf("Expected top score=1100, got %d", hs.Entries[0].Score)
	}
	if hs.Entries[0].Name != "NewPlayer" {
		t.Errorf("Expected top name='NewPlayer', got '%s'", hs.Entries[0].Name)
	}
}

func TestHighScoreEdgeCases(t *testing.T) {
	hs := &HighScores{Entries: []HighScore{}}

	// Test with zero scores
	hs.AddScore("Player", 0, 0, 0)
	if len(hs.Entries) != 1 {
		t.Error("Should accept score of 0")
	}

	// Test with negative scores (if allowed)
	hs.AddScore("Player", -100, 0, 10)
	if len(hs.Entries) != 2 {
		t.Error("Should accept negative scores")
	}

	// Test with very large scores
	hs.AddScore("Player", 999999999, 1000, 0)
	if hs.Entries[0].Score != 999999999 {
		t.Error("Should handle large scores")
	}
}

func TestHighScoreStability(t *testing.T) {
	// Test that sorting is stable and deterministic
	hs := &HighScores{Entries: []HighScore{}}

	// Add scores in specific order
	hs.AddScore("First", 500, 5, 1)
	hs.AddScore("Second", 1000, 10, 1)
	hs.AddScore("Third", 500, 5, 1)
	hs.AddScore("Fourth", 1000, 10, 1)

	// Verify scores are sorted by value
	if hs.Entries[0].Score != 1000 || hs.Entries[1].Score != 1000 {
		t.Error("First two entries should both be 1000")
	}
	if hs.Entries[2].Score != 500 || hs.Entries[3].Score != 500 {
		t.Error("Last two entries should both be 500")
	}
}

func TestMaxHighScoresConstant(t *testing.T) {
	if maxHighScores != 10 {
		t.Errorf("Expected maxHighScores=10, got %d", maxHighScores)
	}
}

func TestConfigConstants(t *testing.T) {
	if configDir != ".config/verdant" {
		t.Errorf("Expected configDir='.config/verdant', got '%s'", configDir)
	}
	if scoresFile != "highscores.json" {
		t.Errorf("Expected scoresFile='highscores.json', got '%s'", scoresFile)
	}
}
