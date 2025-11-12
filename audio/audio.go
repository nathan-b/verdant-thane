package audio

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 44100

// SoundEffect represents a loaded sound effect
type SoundEffect struct {
	data   []byte
	player *audio.Player
}

// Manager handles all audio playback for the game
type Manager struct {
	context *audio.Context
	sounds  map[string]*SoundEffect
	muted   bool
}

// NewManager creates a new audio manager
func NewManager() (*Manager, error) {
	ctx := audio.NewContext(sampleRate)

	return &Manager{
		context: ctx,
		sounds:  make(map[string]*SoundEffect),
		muted:   false,
	}, nil
}

// LoadSound loads a WAV file into memory
func (m *Manager) LoadSound(name string, filepath string) error {
	// Read the WAV file
	fileData, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read sound file %s: %w", filepath, err)
	}

	// Decode WAV data
	stream, err := wav.DecodeWithoutResampling(bytes.NewReader(fileData))
	if err != nil {
		return fmt.Errorf("failed to decode WAV file %s: %w", filepath, err)
	}

	// Read all data into memory
	data, err := io.ReadAll(stream)
	if err != nil {
		return fmt.Errorf("failed to read stream data %s: %w", filepath, err)
	}

	// Store the sound effect
	m.sounds[name] = &SoundEffect{
		data: data,
	}

	return nil
}

// PlaySound plays a sound effect by name
// Creates a new player each time to allow overlapping sounds
func (m *Manager) PlaySound(name string) error {
	if m.muted {
		return nil
	}

	sound, exists := m.sounds[name]
	if !exists {
		return fmt.Errorf("sound %s not loaded", name)
	}

	// Create a new player from the stored data
	// This allows the same sound to play multiple times simultaneously
	player := m.context.NewPlayerFromBytes(sound.data)

	// Play the sound
	player.Play()

	return nil
}

// SetMuted sets whether sounds are muted
func (m *Manager) SetMuted(muted bool) {
	m.muted = muted
}

// IsMuted returns whether sounds are currently muted
func (m *Manager) IsMuted() bool {
	return m.muted
}

// ToggleMute toggles the mute state
func (m *Manager) ToggleMute() {
	m.muted = !m.muted
}
