package audio

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 44100

// EnableProfiling controls whether PlayMusic outputs timing information
var EnableProfiling = false

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
}

// NewManager creates a new audio manager
func NewManager() (*Manager, error) {
	ctx := audio.NewContext(sampleRate)

	return &Manager{
		context:     ctx,
		sounds:      make(map[string]*SoundEffect),
		soundMuted:  false,
		soundVolume: 1.0,
		musicMuted:  false,
		musicVolume: 1.0,
		musicPlayer: nil,
		musicFiles:  make(map[string][]byte),
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
	funcStart := time.Now()
	var t time.Time

	// Stop current music if playing
	t = time.Now()
	m.StopMusic()
	stopMusicTime := time.Since(t)

	if m.musicMuted {
		return nil
	}

	t = time.Now()
	compressedData, exists := m.musicFiles[name]
	if !exists {
		return fmt.Errorf("music %s not loaded", name)
	}
	lookupTime := time.Since(t)

	// Decode MP3 stream (this happens quickly since it's streaming, not loading all at once)
	t = time.Now()
	stream, err := mp3.DecodeWithoutResampling(bytes.NewReader(compressedData))
	if err != nil {
		return fmt.Errorf("failed to decode MP3: %w", err)
	}
	decodeTime := time.Since(t)

	// Get the stream length for looping
	// We need to read the entire stream once to get its length
	t = time.Now()
	data, err := io.ReadAll(stream)
	if err != nil {
		return fmt.Errorf("failed to read decoded stream: %w", err)
	}
	readAllTime := time.Since(t)

	// Create infinite loop from the decoded data
	t = time.Now()
	infiniteLoop := audio.NewInfiniteLoop(bytes.NewReader(data), int64(len(data)))
	loopTime := time.Since(t)

	// Create player from infinite loop
	t = time.Now()
	player, err := m.context.NewPlayer(infiniteLoop)
	if err != nil {
		return fmt.Errorf("failed to create music player: %w", err)
	}
	newPlayerTime := time.Since(t)

	t = time.Now()
	m.musicPlayer = player
	player.SetVolume(m.musicVolume)
	player.Play()
	playTime := time.Since(t)

	totalTime := time.Since(funcStart)

	if EnableProfiling {
		fmt.Printf("\n=== PlayMusic(\"%s\") Profiling ===\n", name)
		fmt.Printf("  StopMusic():                %6.2f ms\n", float64(stopMusicTime.Microseconds())/1000.0)
		fmt.Printf("  Lookup music data:          %6.2f ms\n", float64(lookupTime.Microseconds())/1000.0)
		fmt.Printf("  mp3.DecodeWithoutResampling:%6.2f ms\n", float64(decodeTime.Microseconds())/1000.0)
		fmt.Printf("  io.ReadAll(stream):         %6.2f ms (decoded %d bytes)\n", float64(readAllTime.Microseconds())/1000.0, len(data))
		fmt.Printf("  audio.NewInfiniteLoop():    %6.2f ms\n", float64(loopTime.Microseconds())/1000.0)
		fmt.Printf("  context.NewPlayer():        %6.2f ms (!)\n", float64(newPlayerTime.Microseconds())/1000.0)
		fmt.Printf("  SetVolume() + Play():       %6.2f ms\n", float64(playTime.Microseconds())/1000.0)
		fmt.Printf("  ---\n")
		fmt.Printf("  TOTAL PlayMusic():          %6.2f ms\n", float64(totalTime.Microseconds())/1000.0)
		fmt.Printf("======================================\n\n")
	}

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
