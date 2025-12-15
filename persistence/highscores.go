package persistence

import (
	"os/user"
	"path/filepath"
	"sort"
	"time"
)

// HighScore represents a single high score entry
type HighScore struct {
	Name   string    `json:"name"`
	Score  int       `json:"score"`
	Kills  int       `json:"kills"`
	Deaths int       `json:"deaths"`
	Date   time.Time `json:"date"`
}

// HighScores manages the list of high scores
type HighScores struct {
	Entries []HighScore `json:"entries"`
}

const (
	maxHighScores = 10
	configDir     = ".config/verdant"
	scoresFile    = "highscores.json"
)

// GetConfigPath returns the full path to the config directory
func GetConfigPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, configDir), nil
}

// GetScoresPath returns the full path to the high scores file
func GetScoresPath() (string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, scoresFile), nil
}

// LoadHighScores loads high scores using the default persistence provider
// Returns an empty list if file doesn't exist
// This function is kept for backwards compatibility with existing code and tests
func LoadHighScores() (*HighScores, error) {
	provider := NewProvider()
	return provider.LoadHighScores()
}

// SaveHighScores saves high scores using the default persistence provider
// This function is kept for backwards compatibility with existing code and tests
func SaveHighScores(scores *HighScores) error {
	provider := NewProvider()
	return provider.SaveHighScores(scores)
}

// AddScore adds a new high score and returns true if it made the top 10
func (hs *HighScores) AddScore(name string, score, kills, deaths int) bool {
	newScore := HighScore{
		Name:   name,
		Score:  score,
		Kills:  kills,
		Deaths: deaths,
		Date:   time.Now(),
	}

	// Add to list
	hs.Entries = append(hs.Entries, newScore)

	// Sort by score (descending)
	sort.Slice(hs.Entries, func(i, j int) bool {
		return hs.Entries[i].Score > hs.Entries[j].Score
	})

	// Check if new score is in top 10
	isTopTen := false
	for i, entry := range hs.Entries {
		if i >= maxHighScores {
			break
		}
		if entry.Name == name && entry.Score == score && entry.Date == newScore.Date {
			isTopTen = true
			break
		}
	}

	// Trim to top 10
	if len(hs.Entries) > maxHighScores {
		hs.Entries = hs.Entries[:maxHighScores]
	}

	return isTopTen
}

// IsHighScore returns true if the given score would make the top 10
func (hs *HighScores) IsHighScore(score int) bool {
	if len(hs.Entries) < maxHighScores {
		return true
	}
	return score > hs.Entries[maxHighScores-1].Score
}

// GetCurrentUsername returns the current OS username, or "Player" if unavailable
func GetCurrentUsername() string {
	usr, err := user.Current()
	if err != nil {
		return "Player"
	}
	return usr.Username
}
