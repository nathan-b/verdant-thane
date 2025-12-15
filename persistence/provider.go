package persistence

// PersistenceProvider is an interface for loading and saving game data
// It abstracts the storage mechanism to support both desktop (file-based)
// and WASM (browser localStorage) implementations
type PersistenceProvider interface {
	// LoadHighScores loads the high scores table
	LoadHighScores() (*HighScores, error)

	// SaveHighScores saves the high scores table
	SaveHighScores(scores *HighScores) error

	// LoadSettings loads the game settings
	LoadSettings() (*Settings, error)

	// SaveSettings saves the game settings
	SaveSettings(settings *Settings) error
}

// NewProvider creates the appropriate PersistenceProvider for the current platform
// On desktop, this returns a file-based provider
// On WASM, this returns a localStorage-based provider
func NewProvider() PersistenceProvider {
	return newDesktopProvider()
}
