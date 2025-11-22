package audio

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
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
	context     *audio.Context
	sounds      map[string]*SoundEffect
	soundMuted  bool
	soundVolume float64 // 0.0 to 1.0
	musicMuted  bool
	musicVolume float64           // 0.0 to 1.0
	musicPlayer *audio.Player     // Current music player (only one plays at a time)
	musicFiles  map[string][]byte // Compressed MP3 data (not decoded)

	// Looping sound effects (e.g., beam weapon)
	loopingSounds map[int]*audio.Player // ID -> looping player
}

// NewManager creates a new audio manager
func NewManager() (*Manager, error) {
	ctx := audio.NewContext(sampleRate)

	return &Manager{
		context:       ctx,
		sounds:        make(map[string]*SoundEffect),
		soundMuted:    false,
		soundVolume:   1.0,
		musicMuted:    false,
		musicVolume:   1.0,
		musicPlayer:   nil,
		musicFiles:    make(map[string][]byte),
		loopingSounds: make(map[int]*audio.Player),
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
	if m.soundMuted {
		return nil
	}

	sound, exists := m.sounds[name]
	if !exists {
		return fmt.Errorf("sound %s not loaded", name)
	}

	// Create a new player from the stored data
	// This allows the same sound to play multiple times simultaneously
	player := m.context.NewPlayerFromBytes(sound.data)

	// Apply volume
	player.SetVolume(m.soundVolume)

	// Play the sound
	player.Play()

	return nil
}

// LoadMusic loads an MP3 file into memory (keeps it compressed)
func (m *Manager) LoadMusic(name string, filepath string) error {
	// Read the compressed MP3 file
	fileData, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read music file %s: %w", filepath, err)
	}

	// Store the compressed MP3 data (we'll decode on-the-fly during playback)
	m.musicFiles[name] = fileData

	return nil
}

// PlayMusic plays a music track by name (loops infinitely)
// Stops any currently playing music
func (m *Manager) PlayMusic(name string) error {
	// Stop current music if playing
	m.StopMusic()

	if m.musicMuted {
		return nil
	}

	compressedData, exists := m.musicFiles[name]
	if !exists {
		return fmt.Errorf("music %s not loaded", name)
	}

	// Decode MP3 stream
	stream, err := mp3.DecodeWithoutResampling(bytes.NewReader(compressedData))
	if err != nil {
		return fmt.Errorf("failed to decode MP3: %w", err)
	}

	// Get the stream length for looping
	streamLength := stream.Length()

	// Create infinite loop from the stream
	infiniteLoop := audio.NewInfiniteLoop(stream, streamLength)

	// Create player from infinite loop
	player, err := m.context.NewPlayer(infiniteLoop)
	if err != nil {
		return fmt.Errorf("failed to create music player: %w", err)
	}

	m.musicPlayer = player
	player.SetVolume(m.musicVolume)
	player.Play()

	return nil
}

// StopMusic stops the currently playing music
func (m *Manager) StopMusic() {
	if m.musicPlayer != nil {
		m.musicPlayer.Pause()
		m.musicPlayer.Close()
		m.musicPlayer = nil
	}
}

// SetSoundMuted sets whether sound effects are muted
func (m *Manager) SetSoundMuted(muted bool) {
	m.soundMuted = muted
}

// IsSoundMuted returns whether sound effects are currently muted
func (m *Manager) IsSoundMuted() bool {
	return m.soundMuted
}

// ToggleSoundMute toggles the sound effects mute state
func (m *Manager) ToggleSoundMute() {
	m.soundMuted = !m.soundMuted
}

// SetMusicMuted sets whether music is muted
func (m *Manager) SetMusicMuted(muted bool) {
	m.musicMuted = muted
	if muted && m.musicPlayer != nil {
		m.musicPlayer.Pause()
	} else if !muted && m.musicPlayer != nil {
		m.musicPlayer.Play()
	}
}

// IsMusicMuted returns whether music is currently muted
func (m *Manager) IsMusicMuted() bool {
	return m.musicMuted
}

// ToggleMusicMute toggles the music mute state
func (m *Manager) ToggleMusicMute() {
	m.SetMusicMuted(!m.musicMuted)
}

// SetMuted sets whether sounds are muted (legacy compatibility)
func (m *Manager) SetMuted(muted bool) {
	m.SetSoundMuted(muted)
}

// IsMuted returns whether sounds are currently muted (legacy compatibility)
func (m *Manager) IsMuted() bool {
	return m.IsSoundMuted()
}

// ToggleMute toggles the sound mute state (legacy compatibility)
func (m *Manager) ToggleMute() {
	m.ToggleSoundMute()
}

// SetSoundVolume sets the sound effects volume (0.0 to 1.0)
func (m *Manager) SetSoundVolume(volume float64) {
	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}
	m.soundVolume = volume
}

// GetSoundVolume returns the current sound effects volume (0.0 to 1.0)
func (m *Manager) GetSoundVolume() float64 {
	return m.soundVolume
}

// SetMusicVolume sets the music volume (0.0 to 1.0)
func (m *Manager) SetMusicVolume(volume float64) {
	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}
	m.musicVolume = volume
	// Apply volume to currently playing music
	if m.musicPlayer != nil {
		m.musicPlayer.SetVolume(volume)
	}
}

// GetMusicVolume returns the current music volume (0.0 to 1.0)
func (m *Manager) GetMusicVolume() float64 {
	return m.musicVolume
}

// StartLoopingSound starts playing a looping sound effect for a given entity ID
// If the sound is already playing for this ID, does nothing
func (m *Manager) StartLoopingSound(id int, name string) error {
	if m.soundMuted {
		return nil
	}

	// Check if already playing for this ID
	if _, exists := m.loopingSounds[id]; exists {
		return nil
	}

	sound, exists := m.sounds[name]
	if !exists {
		return fmt.Errorf("sound %s not loaded", name)
	}

	// Create infinite loop from the sound data
	infiniteLoop := audio.NewInfiniteLoopWithIntro(bytes.NewReader(sound.data), 0, int64(len(sound.data)))

	// Create player from infinite loop
	player, err := m.context.NewPlayer(infiniteLoop)
	if err != nil {
		return fmt.Errorf("failed to create looping player: %w", err)
	}

	player.SetVolume(m.soundVolume)
	player.Play()

	m.loopingSounds[id] = player
	return nil
}

// StopLoopingSound stops a looping sound for a given entity ID
func (m *Manager) StopLoopingSound(id int) {
	if player, exists := m.loopingSounds[id]; exists {
		player.Pause()
		player.Close()
		delete(m.loopingSounds, id)
	}
}

// IsLoopingSoundPlaying returns whether a looping sound is currently playing for an ID
func (m *Manager) IsLoopingSoundPlaying(id int) bool {
	_, exists := m.loopingSounds[id]
	return exists
}

// StopAllLoopingSounds stops all currently playing looping sounds
func (m *Manager) StopAllLoopingSounds() {
	for id, player := range m.loopingSounds {
		player.Pause()
		player.Close()
		delete(m.loopingSounds, id)
	}
}
