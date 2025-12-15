package main

import "path/filepath"
import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func TestLoadKillstreakText(t *testing.T) {
	// Test loading actual killtext.json file
	fsys := os.DirFS(".")
	kt, err := LoadKillstreakText(fsys, "assets/killtext.json")
	if err != nil {
		t.Fatalf("Failed to load killtext.json: %v", err)
	}

	// Verify some expected entries
	tests := []struct {
		killCount int
		expected  string
	}{
		{3, "Multikill!"},
		{4, "Quadroslay!"},
		{5, "Killing Spree!"},
		{9, "Green Baron!"},
		{19, "OK, you win!"},
	}

	for _, tt := range tests {
		got := kt.GetMessage(tt.killCount)
		if got != tt.expected {
			t.Errorf("GetMessage(%d) = %q, want %q", tt.killCount, got, tt.expected)
		}
	}

	// Test non-existent kill count
	if msg := kt.GetMessage(999); msg != "" {
		t.Errorf("GetMessage(999) = %q, want empty string", msg)
	}
}

func TestLoadKillstreakTextInvalidFile(t *testing.T) {
	fsys := os.DirFS(".")
	_, err := LoadKillstreakText(fsys, "nonexistent.json")
	if err == nil {
		t.Error("Expected error loading nonexistent file, got nil")
	}
}

func TestLoadKillstreakTextInvalidJSON(t *testing.T) {
	// Create temporary invalid JSON file
	tmpFile, err := os.CreateTemp("", "invalid-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write invalid JSON
	if _, err := tmpFile.WriteString("{invalid json}"); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Should fail to load
	fsys := os.DirFS(filepath.Dir(tmpFile.Name()))
	_, err = LoadKillstreakText(fsys, filepath.Base(tmpFile.Name()))
	if err == nil {
		t.Error("Expected error loading invalid JSON, got nil")
	}
}

func TestLoadKillstreakTextWithStringKeys(t *testing.T) {
	// Create temporary JSON file with string keys
	tmpFile, err := os.CreateTemp("", "killtext-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write JSON with string number keys
	jsonData := `{
		"3": "Test Message 3",
		"5": "Test Message 5",
		"10": "Test Message 10"
	}`
	if _, err := tmpFile.WriteString(jsonData); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Load and verify
	fsys := os.DirFS(filepath.Dir(tmpFile.Name()))
	kt, err := LoadKillstreakText(fsys, filepath.Base(tmpFile.Name()))
	if err != nil {
		t.Fatalf("Failed to load killtext: %v", err)
	}

	tests := []struct {
		killCount int
		expected  string
	}{
		{3, "Test Message 3"},
		{5, "Test Message 5"},
		{10, "Test Message 10"},
		{7, ""}, // Non-existent
	}

	for _, tt := range tests {
		got := kt.GetMessage(tt.killCount)
		if got != tt.expected {
			t.Errorf("GetMessage(%d) = %q, want %q", tt.killCount, got, tt.expected)
		}
	}
}

func TestKillstreakDisplayLifecycle(t *testing.T) {
	// Create test font source
	fontSource := createTestFontSource(t)

	// Create test killtext
	kt := &KillstreakText{
		messages: map[int]string{
			3: "Multikill!",
			5: "Killing Spree!",
		},
	}

	// Create display with short duration for testing
	display := NewKillstreakDisplay(kt, fontSource)
	display.displayDuration = 100 * time.Millisecond

	// Initially not active
	if display.isActive {
		t.Error("Display should not be active initially")
	}

	// Trigger display for kill count 3
	display.OnKill(3)

	// Should now be active with correct message
	if !display.isActive {
		t.Error("Display should be active after OnKill")
	}
	if display.currentMessage != "Multikill!" {
		t.Errorf("currentMessage = %q, want %q", display.currentMessage, "Multikill!")
	}

	// Update should keep it active (within duration)
	display.Update()
	if !display.isActive {
		t.Error("Display should still be active before duration expires")
	}

	// Wait for duration to expire
	time.Sleep(150 * time.Millisecond)
	display.Update()

	// Should no longer be active
	if display.isActive {
		t.Error("Display should not be active after duration expires")
	}
	if display.currentMessage != "" {
		t.Errorf("currentMessage = %q, want empty string after expiry", display.currentMessage)
	}
}

func TestKillstreakDisplayOverwrite(t *testing.T) {
	// Create test font source
	fontSource := createTestFontSource(t)

	// Create test killtext
	kt := &KillstreakText{
		messages: map[int]string{
			3: "Multikill!",
			4: "Quadroslay!",
		},
	}

	display := NewKillstreakDisplay(kt, fontSource)
	display.displayDuration = 1 * time.Second

	// Trigger first message
	display.OnKill(3)
	firstStart := display.displayStart

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Trigger second message (should overwrite)
	display.OnKill(4)

	// Should have new message and reset timer
	if display.currentMessage != "Quadroslay!" {
		t.Errorf("currentMessage = %q, want %q", display.currentMessage, "Quadroslay!")
	}
	if !display.displayStart.After(firstStart) {
		t.Error("displayStart should be reset when message is overwritten")
	}
	if !display.isActive {
		t.Error("Display should still be active after overwrite")
	}
}

func TestKillstreakDisplayNoMessage(t *testing.T) {
	// Create test font source
	fontSource := createTestFontSource(t)

	// Create test killtext with no entry for kill count 2
	kt := &KillstreakText{
		messages: map[int]string{
			3: "Multikill!",
		},
	}

	display := NewKillstreakDisplay(kt, fontSource)

	// Trigger for kill count with no message
	display.OnKill(2)

	// Should not activate
	if display.isActive {
		t.Error("Display should not be active for kill count with no message")
	}
	if display.currentMessage != "" {
		t.Errorf("currentMessage = %q, want empty string", display.currentMessage)
	}
}

func TestKillstreakDisplayRender(t *testing.T) {
	// Create test font source
	fontSource := createTestFontSource(t)

	// Create test killtext
	kt := &KillstreakText{
		messages: map[int]string{
			3: "Test!",
		},
	}

	display := NewKillstreakDisplay(kt, fontSource)

	// Create test screen
	screen := ebiten.NewImage(800, 600)

	// Render when not active (should not panic)
	display.Render(screen)

	// Activate display
	display.OnKill(3)

	// Render when active (should not panic)
	display.Render(screen)

	// Verify still active after render
	if !display.isActive {
		t.Error("Display should still be active after render")
	}
}

func TestHSVToRGB(t *testing.T) {
	tests := []struct {
		h, s, v     float64
		expectedR   uint8
		expectedG   uint8
		expectedB   uint8
		description string
	}{
		{0, 1, 1, 255, 0, 0, "Pure red"},
		{120, 1, 1, 0, 255, 0, "Pure green"},
		{240, 1, 1, 0, 0, 255, "Pure blue"},
		{0, 0, 1, 255, 255, 255, "White (no saturation)"},
		{0, 0, 0, 0, 0, 0, "Black (no value)"},
		{60, 1, 1, 255, 255, 0, "Yellow"},
		{180, 1, 1, 0, 255, 255, "Cyan"},
		{300, 1, 1, 255, 0, 255, "Magenta"},
	}

	for _, tt := range tests {
		r, g, b := hsvToRGB(tt.h, tt.s, tt.v)

		// Allow small tolerance for floating point arithmetic
		tolerance := uint8(2)

		if !colorClose(r, tt.expectedR, tolerance) ||
			!colorClose(g, tt.expectedG, tolerance) ||
			!colorClose(b, tt.expectedB, tolerance) {
			t.Errorf("%s: hsvToRGB(%v, %v, %v) = (%d, %d, %d), want (%d, %d, %d)",
				tt.description, tt.h, tt.s, tt.v, r, g, b, tt.expectedR, tt.expectedG, tt.expectedB)
		}
	}
}

// Helper function to check if two color values are close within tolerance
func colorClose(a, b, tolerance uint8) bool {
	diff := int(a) - int(b)
	if diff < 0 {
		diff = -diff
	}
	return diff <= int(tolerance)
}

// Helper function to create a test font source
func createTestFontSource(t *testing.T) *text.GoTextFaceSource {
	t.Helper()

	// Load actual font for testing
	fontBytes, err := os.ReadFile("assets/orbitron.ttf")
	if err != nil {
		t.Fatalf("Failed to load test font: %v", err)
	}

	fontSource, err := text.NewGoTextFaceSource(bytes.NewReader(fontBytes))
	if err != nil {
		t.Fatalf("Failed to create font source: %v", err)
	}

	return fontSource
}
