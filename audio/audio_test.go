package audio

import (
	"bytes"
	"encoding/binary"
	"testing"
	"testing/fstest"
)

// Global manager for testing (Ebiten only allows one audio context per process)
var testManager Manager

// TestMain sets up the shared audio manager for all tests
func TestMain(m *testing.M) {
	testManager = NewManager()
	m.Run()
}

// createMinimalWAV creates a minimal valid WAV file for testing
// Creates a 44.1kHz mono 16-bit PCM WAV with a few samples
func createMinimalWAV() []byte {
	var buf bytes.Buffer

	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(36)) // chunk size (will be wrong, but sufficient for testing)
	buf.WriteString("WAVE")

	// fmt subchunk
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16))    // subchunk1 size
	binary.Write(&buf, binary.LittleEndian, uint16(1))     // audio format (PCM)
	binary.Write(&buf, binary.LittleEndian, uint16(1))     // num channels (mono)
	binary.Write(&buf, binary.LittleEndian, uint32(44100)) // sample rate
	binary.Write(&buf, binary.LittleEndian, uint32(88200)) // byte rate
	binary.Write(&buf, binary.LittleEndian, uint16(2))     // block align
	binary.Write(&buf, binary.LittleEndian, uint16(16))    // bits per sample

	// data subchunk
	buf.WriteString("data")
	samples := []int16{0, 100, 200, 100, 0, -100, -200, -100} // Simple waveform
	binary.Write(&buf, binary.LittleEndian, uint32(len(samples)*2))
	for _, sample := range samples {
		binary.Write(&buf, binary.LittleEndian, sample)
	}

	return buf.Bytes()
}

// createInvalidWAV creates invalid data that will fail to decode
func createInvalidWAV() []byte {
	return []byte("NOT A VALID WAV FILE")
}

// createMinimalMP3 creates minimal MP3-like data
// Note: This is not a fully valid MP3, but sufficient for testing file loading
// Real MP3 decoding in tests would require actual MP3 files or complex generation
func createMinimalMP3() []byte {
	// MP3 frame header: 0xFFE3 (MPEG1 Layer3)
	// This is a minimal valid-looking MP3 header, though it won't decode properly
	// For real testing, we'd need actual MP3 data, but this is sufficient for
	// testing the loading path
	return []byte{
		0xFF, 0xE3, 0x18, 0xC4, // MP3 sync word + header
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
}

// TestNewManager tests audio manager creation
func TestNewManager(t *testing.T) {
	// Use the shared test manager
	mgr := testManager

	// Check default values
	if mgr.soundVolume != 1.0 {
		t.Errorf("Default sound volume = %f, want 1.0", mgr.soundVolume)
	}
	if mgr.musicVolume != 1.0 {
		t.Errorf("Default music volume = %f, want 1.0", mgr.musicVolume)
	}
	if mgr.soundMuted {
		t.Error("Sound should not be muted by default")
	}
	if mgr.musicMuted {
		t.Error("Music should not be muted by default")
	}

	// If initialized, check that maps are created
	if mgr.initialized {
		if mgr.sounds == nil {
			t.Error("Initialized manager should have sounds map")
		}
		if mgr.musicFiles == nil {
			t.Error("Initialized manager should have musicFiles map")
		}
		if mgr.loopingSounds == nil {
			t.Error("Initialized manager should have loopingSounds map")
		}
		if mgr.context == nil {
			t.Error("Initialized manager should have audio context")
		}
	}
}

// TestVolumeControl tests sound volume getter/setter
func TestVolumeControl(t *testing.T) {
	// Create a separate uninitialized manager for volume testing
	// (Volume controls should work even when uninitialized)
	mgr := Manager{soundVolume: 1.0, musicVolume: 1.0}

	testCases := []struct {
		name          string
		setValue      float64
		expectedValue float64
	}{
		{"minimum volume", 0.0, 0.0},
		{"half volume", 0.5, 0.5},
		{"maximum volume", 1.0, 1.0},
		{"below minimum (clamped)", -0.5, 0.0},
		{"above maximum (clamped)", 1.5, 1.0},
		{"normal value", 0.75, 0.75},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mgr.SetSoundVolume(tc.setValue)
			got := mgr.GetSoundVolume()
			if got != tc.expectedValue {
				t.Errorf("SetSoundVolume(%f) -> GetSoundVolume() = %f, want %f",
					tc.setValue, got, tc.expectedValue)
			}
		})
	}
}

// TestMusicVolumeControl tests music volume getter/setter
func TestMusicVolumeControl(t *testing.T) {
	mgr := Manager{soundVolume: 1.0, musicVolume: 1.0}

	testCases := []struct {
		name          string
		setValue      float64
		expectedValue float64
	}{
		{"minimum volume", 0.0, 0.0},
		{"half volume", 0.5, 0.5},
		{"maximum volume", 1.0, 1.0},
		{"below minimum (clamped)", -0.5, 0.0},
		{"above maximum (clamped)", 1.5, 1.0},
		{"normal value", 0.75, 0.75},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mgr.SetMusicVolume(tc.setValue)
			got := mgr.GetMusicVolume()
			if got != tc.expectedValue {
				t.Errorf("SetMusicVolume(%f) -> GetMusicVolume() = %f, want %f",
					tc.setValue, got, tc.expectedValue)
			}
		})
	}
}

// TestSoundMuteControl tests sound mute state
func TestSoundMuteControl(t *testing.T) {
	mgr := Manager{soundVolume: 1.0, musicVolume: 1.0}

	// Initial state: not muted
	if mgr.IsSoundMuted() {
		t.Error("Sound should not be muted initially")
	}

	// Test SetSoundMuted
	mgr.SetSoundMuted(true)
	if !mgr.IsSoundMuted() {
		t.Error("Sound should be muted after SetSoundMuted(true)")
	}

	mgr.SetSoundMuted(false)
	if mgr.IsSoundMuted() {
		t.Error("Sound should not be muted after SetSoundMuted(false)")
	}

	// Test ToggleSoundMute
	initialState := mgr.IsSoundMuted()
	mgr.ToggleSoundMute()
	if mgr.IsSoundMuted() == initialState {
		t.Error("ToggleSoundMute should change mute state")
	}

	mgr.ToggleSoundMute()
	if mgr.IsSoundMuted() != initialState {
		t.Error("ToggleSoundMute twice should return to original state")
	}
}

// TestMusicMuteControl tests music mute state
func TestMusicMuteControl(t *testing.T) {
	mgr := Manager{soundVolume: 1.0, musicVolume: 1.0}

	// Initial state: not muted
	if mgr.IsMusicMuted() {
		t.Error("Music should not be muted initially")
	}

	// Test SetMusicMuted
	mgr.SetMusicMuted(true)
	if !mgr.IsMusicMuted() {
		t.Error("Music should be muted after SetMusicMuted(true)")
	}

	mgr.SetMusicMuted(false)
	if mgr.IsMusicMuted() {
		t.Error("Music should not be muted after SetMusicMuted(false)")
	}

	// Test ToggleMusicMute
	initialState := mgr.IsMusicMuted()
	mgr.ToggleMusicMute()
	if mgr.IsMusicMuted() == initialState {
		t.Error("ToggleMusicMute should change mute state")
	}

	mgr.ToggleMusicMute()
	if mgr.IsMusicMuted() != initialState {
		t.Error("ToggleMusicMute twice should return to original state")
	}
}

// TestLegacyMuteAPI tests backward compatibility methods
func TestLegacyMuteAPI(t *testing.T) {
	mgr := Manager{soundVolume: 1.0, musicVolume: 1.0}

	// Test legacy SetMuted (should control sound mute)
	mgr.SetMuted(true)
	if !mgr.IsMuted() {
		t.Error("IsMuted should return true after SetMuted(true)")
	}
	if !mgr.IsSoundMuted() {
		t.Error("IsSoundMuted should return true after legacy SetMuted(true)")
	}

	mgr.SetMuted(false)
	if mgr.IsMuted() {
		t.Error("IsMuted should return false after SetMuted(false)")
	}

	// Test legacy ToggleMute
	initialState := mgr.IsMuted()
	mgr.ToggleMute()
	if mgr.IsMuted() == initialState {
		t.Error("ToggleMute should change state")
	}
}

// TestLoadSound_ValidWAV tests loading a valid WAV file
func TestLoadSound_ValidWAV(t *testing.T) {
	mgr := testManager

	if !mgr.initialized {
		t.Skip("Audio manager not initialized (no audio device available)")
	}

	// Create mock filesystem with valid WAV
	mockFS := fstest.MapFS{
		"test.wav": &fstest.MapFile{
			Data: createMinimalWAV(),
		},
	}

	err := mgr.LoadSound("test-sound-valid", mockFS, "test.wav")
	if err != nil {
		t.Errorf("LoadSound with valid WAV failed: %v", err)
	}

	// Verify sound was loaded
	if _, exists := mgr.sounds["test-sound-valid"]; !exists {
		t.Error("Sound was not stored in sounds map")
	}
}

// TestLoadSound_InvalidFile tests loading an invalid file
func TestLoadSound_InvalidFile(t *testing.T) {
	mgr := testManager

	if !mgr.initialized {
		t.Skip("Audio manager not initialized (no audio device available)")
	}

	// Create mock filesystem with invalid data
	mockFS := fstest.MapFS{
		"invalid.wav": &fstest.MapFile{
			Data: createInvalidWAV(),
		},
	}

	err := mgr.LoadSound("bad-sound", mockFS, "invalid.wav")
	if err == nil {
		t.Error("LoadSound should fail with invalid WAV data")
	}
}

// TestLoadSound_MissingFile tests loading a non-existent file
func TestLoadSound_MissingFile(t *testing.T) {
	mgr := testManager

	if !mgr.initialized {
		t.Skip("Audio manager not initialized (no audio device available)")
	}

	mockFS := fstest.MapFS{}

	err := mgr.LoadSound("missing", mockFS, "nonexistent.wav")
	if err == nil {
		t.Error("LoadSound should fail with missing file")
	}
}

// TestLoadSound_Uninitialized tests that LoadSound is a no-op when uninitialized
func TestLoadSound_Uninitialized(t *testing.T) {
	// Create an uninitialized manager
	mgr := Manager{
		initialized: false,
		soundVolume: 1.0,
		musicVolume: 1.0,
	}

	mockFS := fstest.MapFS{
		"test.wav": &fstest.MapFile{
			Data: createMinimalWAV(),
		},
	}

	err := mgr.LoadSound("test", mockFS, "test.wav")
	if err != nil {
		t.Errorf("LoadSound on uninitialized manager should be no-op, got error: %v", err)
	}
}

// TestPlaySound_NotLoaded tests playing a sound that hasn't been loaded
func TestPlaySound_NotLoaded(t *testing.T) {
	mgr := testManager

	if !mgr.initialized {
		t.Skip("Audio manager not initialized (no audio device available)")
	}

	err := mgr.PlaySound("definitely-nonexistent-sound-12345")
	if err == nil {
		t.Error("PlaySound should return error for non-existent sound")
	}
}

// TestPlaySound_AfterLoading tests playing a loaded sound
func TestPlaySound_AfterLoading(t *testing.T) {
	mgr := testManager

	if !mgr.initialized {
		t.Skip("Audio manager not initialized (no audio device available)")
	}

	// Load a sound
	mockFS := fstest.MapFS{
		"play.wav": &fstest.MapFile{
			Data: createMinimalWAV(),
		},
	}
	err := mgr.LoadSound("playable", mockFS, "play.wav")
	if err != nil {
		t.Fatalf("LoadSound failed: %v", err)
	}

	// Ensure unmuted
	mgr.SetSoundMuted(false)

	// Play the sound
	err = mgr.PlaySound("playable")
	if err != nil {
		t.Errorf("PlaySound failed: %v", err)
	}

	// Play it again (should allow overlapping)
	err = mgr.PlaySound("playable")
	if err != nil {
		t.Errorf("PlaySound second time failed: %v", err)
	}
}

// TestPlaySound_Muted tests that muted sounds don't error
func TestPlaySound_Muted(t *testing.T) {
	mgr := testManager

	if !mgr.initialized {
		t.Skip("Audio manager not initialized (no audio device available)")
	}

	// Load a sound
	mockFS := fstest.MapFS{
		"mute.wav": &fstest.MapFile{
			Data: createMinimalWAV(),
		},
	}
	mgr.LoadSound("mute-test", mockFS, "mute.wav")

	// Mute sounds
	mgr.SetSoundMuted(true)

	// Should not error, but won't actually play
	err := mgr.PlaySound("mute-test")
	if err != nil {
		t.Errorf("PlaySound when muted should not error: %v", err)
	}

	// Unmute for other tests
	mgr.SetSoundMuted(false)
}

// TestPlaySound_Uninitialized tests that PlaySound is a no-op when uninitialized
func TestPlaySound_Uninitialized(t *testing.T) {
	mgr := Manager{
		initialized: false,
		soundVolume: 1.0,
		musicVolume: 1.0,
	}

	err := mgr.PlaySound("anything")
	if err != nil {
		t.Errorf("PlaySound on uninitialized manager should be no-op, got error: %v", err)
	}
}

// TestLoadMusic tests loading music files
func TestLoadMusic(t *testing.T) {
	mgr := testManager

	if !mgr.initialized {
		t.Skip("Audio manager not initialized (no audio device available)")
	}

	mockFS := fstest.MapFS{
		"test.mp3": &fstest.MapFile{
			Data: createMinimalMP3(),
		},
	}

	err := mgr.LoadMusic("test-music", mockFS, "test.mp3")
	if err != nil {
		t.Errorf("LoadMusic failed: %v", err)
	}

	// Verify music was loaded
	if _, exists := mgr.musicFiles["test-music"]; !exists {
		t.Error("Music was not stored in musicFiles map")
	}
}

// TestLoadMusic_MissingFile tests loading non-existent music
func TestLoadMusic_MissingFile(t *testing.T) {
	mgr := testManager

	if !mgr.initialized {
		t.Skip("Audio manager not initialized (no audio device available)")
	}

	mockFS := fstest.MapFS{}

	err := mgr.LoadMusic("missing", mockFS, "nonexistent.mp3")
	if err == nil {
		t.Error("LoadMusic should fail with missing file")
	}
}

// TestStopMusic tests stopping music
func TestStopMusic(t *testing.T) {
	mgr := testManager

	if !mgr.initialized {
		t.Skip("Audio manager not initialized (no audio device available)")
	}

	// Should not panic even with no music playing
	mgr.StopMusic()

	// Verify musicPlayer is nil
	if mgr.musicPlayer != nil {
		t.Error("Music player should be nil after StopMusic")
	}
}

// TestStopMusic_Uninitialized tests that StopMusic is a no-op when uninitialized
func TestStopMusic_Uninitialized(t *testing.T) {
	mgr := Manager{
		initialized: false,
	}

	// Should not panic
	mgr.StopMusic()
}

// TestLoopingSound_StartStop tests starting and stopping looping sounds
func TestLoopingSound_StartStop(t *testing.T) {
	mgr := testManager

	if !mgr.initialized {
		t.Skip("Audio manager not initialized (no audio device available)")
	}

	// Load a sound
	mockFS := fstest.MapFS{
		"loop.wav": &fstest.MapFile{
			Data: createMinimalWAV(),
		},
	}
	err := mgr.LoadSound("loop-sound", mockFS, "loop.wav")
	if err != nil {
		t.Fatalf("LoadSound failed: %v", err)
	}

	entityID := 12300

	// Initially not playing
	if mgr.IsLoopingSoundPlaying(entityID) {
		t.Error("Looping sound should not be playing initially")
	}

	// Ensure unmuted
	mgr.SetSoundMuted(false)

	// Start looping sound
	err = mgr.StartLoopingSound(entityID, "loop-sound")
	if err != nil {
		t.Errorf("StartLoopingSound failed: %v", err)
	}

	// Should now be playing
	if !mgr.IsLoopingSoundPlaying(entityID) {
		t.Error("Looping sound should be playing after start")
	}

	// Stop looping sound
	mgr.StopLoopingSound(entityID)

	// Should no longer be playing
	if mgr.IsLoopingSoundPlaying(entityID) {
		t.Error("Looping sound should not be playing after stop")
	}
}

// TestLoopingSound_StartTwice tests that starting a looping sound twice is safe
func TestLoopingSound_StartTwice(t *testing.T) {
	mgr := testManager

	if !mgr.initialized {
		t.Skip("Audio manager not initialized (no audio device available)")
	}

	mockFS := fstest.MapFS{
		"loop2.wav": &fstest.MapFile{
			Data: createMinimalWAV(),
		},
	}
	mgr.LoadSound("loop-sound-2", mockFS, "loop2.wav")

	entityID := 45600

	// Ensure unmuted
	mgr.SetSoundMuted(false)

	// Start once
	err := mgr.StartLoopingSound(entityID, "loop-sound-2")
	if err != nil {
		t.Fatalf("First StartLoopingSound failed: %v", err)
	}

	// Start again (should be no-op)
	err = mgr.StartLoopingSound(entityID, "loop-sound-2")
	if err != nil {
		t.Errorf("Second StartLoopingSound should not error: %v", err)
	}

	// Should still be playing
	if !mgr.IsLoopingSoundPlaying(entityID) {
		t.Error("Looping sound should still be playing")
	}

	mgr.StopLoopingSound(entityID)
}

// TestLoopingSound_NotLoaded tests starting a looping sound that wasn't loaded
func TestLoopingSound_NotLoaded(t *testing.T) {
	mgr := testManager

	if !mgr.initialized {
		t.Skip("Audio manager not initialized (no audio device available)")
	}

	err := mgr.StartLoopingSound(78900, "totally-nonexistent-loop-sound")
	if err == nil {
		t.Error("StartLoopingSound should fail for non-existent sound")
	}
}

// TestLoopingSound_Muted tests that muted looping sounds don't start
func TestLoopingSound_Muted(t *testing.T) {
	mgr := testManager

	if !mgr.initialized {
		t.Skip("Audio manager not initialized (no audio device available)")
	}

	mockFS := fstest.MapFS{
		"loop3.wav": &fstest.MapFile{
			Data: createMinimalWAV(),
		},
	}
	mgr.LoadSound("loop-sound-3", mockFS, "loop3.wav")

	// Mute sounds
	mgr.SetSoundMuted(true)

	entityID := 99900

	// Try to start looping sound while muted
	err := mgr.StartLoopingSound(entityID, "loop-sound-3")
	if err != nil {
		t.Errorf("StartLoopingSound when muted should not error: %v", err)
	}

	// Should not be playing (muted)
	if mgr.IsLoopingSoundPlaying(entityID) {
		t.Error("Looping sound should not start when muted")
	}

	// Unmute for other tests
	mgr.SetSoundMuted(false)
}

// TestStopAllLoopingSounds tests stopping all looping sounds at once
func TestStopAllLoopingSounds(t *testing.T) {
	mgr := testManager

	if !mgr.initialized {
		t.Skip("Audio manager not initialized (no audio device available)")
	}

	mockFS := fstest.MapFS{
		"loop-all.wav": &fstest.MapFile{
			Data: createMinimalWAV(),
		},
	}
	mgr.LoadSound("loop-all", mockFS, "loop-all.wav")

	// Ensure unmuted
	mgr.SetSoundMuted(false)

	// Start multiple looping sounds
	ids := []int{10001, 10002, 10003, 10004, 10005}
	for _, id := range ids {
		err := mgr.StartLoopingSound(id, "loop-all")
		if err != nil {
			t.Fatalf("StartLoopingSound(%d) failed: %v", id, err)
		}
	}

	// Verify all are playing
	for _, id := range ids {
		if !mgr.IsLoopingSoundPlaying(id) {
			t.Errorf("Looping sound %d should be playing", id)
		}
	}

	// Stop all
	mgr.StopAllLoopingSounds()

	// Verify none are playing
	for _, id := range ids {
		if mgr.IsLoopingSoundPlaying(id) {
			t.Errorf("Looping sound %d should not be playing after StopAll", id)
		}
	}
}

// TestStopAllLoopingSounds_Uninitialized tests no-op behavior
func TestStopAllLoopingSounds_Uninitialized(t *testing.T) {
	mgr := Manager{
		initialized: false,
	}

	// Should not panic
	mgr.StopAllLoopingSounds()
}

// TestIsLoopingSoundPlaying_Uninitialized tests return value when uninitialized
func TestIsLoopingSoundPlaying_Uninitialized(t *testing.T) {
	mgr := Manager{
		initialized: false,
	}

	// Should return false (not panic)
	if mgr.IsLoopingSoundPlaying(123) {
		t.Error("Uninitialized manager should report no looping sounds")
	}
}

// TestUninitializedManagerGracefulDegradation tests that an uninitialized manager
// handles all operations gracefully without panicking
func TestUninitializedManagerGracefulDegradation(t *testing.T) {
	mgr := Manager{
		initialized: false,
		soundVolume: 1.0,
		musicVolume: 1.0,
	}

	// All these operations should be no-ops (not panic)
	mockFS := fstest.MapFS{
		"test.wav": &fstest.MapFile{Data: createMinimalWAV()},
		"test.mp3": &fstest.MapFile{Data: createMinimalMP3()},
	}

	mgr.LoadSound("test", mockFS, "test.wav")
	mgr.PlaySound("test")
	mgr.LoadMusic("music", mockFS, "test.mp3")
	mgr.PlayMusic("music")
	mgr.StopMusic()
	mgr.StartLoopingSound(1, "test")
	mgr.StopLoopingSound(1)
	mgr.StopAllLoopingSounds()

	// Volume and mute controls should still work
	mgr.SetSoundVolume(0.5)
	if mgr.GetSoundVolume() != 0.5 {
		t.Error("Volume controls should work even when uninitialized")
	}

	mgr.SetMusicMuted(true)
	if !mgr.IsMusicMuted() {
		t.Error("Mute controls should work even when uninitialized")
	}
}

// TestAudioEnvironment logs whether audio is available in test environment
func TestAudioEnvironment(t *testing.T) {
	if testManager.initialized {
		t.Log("Audio device available - full audio tests will run")
	} else {
		t.Log("No audio device available - some tests will be skipped")
	}
}
