package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test ChatWindow Creation
func TestNewChatWindow(t *testing.T) {
	// Create temporary test chatter file
	tempFile := createTestChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	if cw == nil {
		t.Fatal("ChatWindow should not be nil")
	}

	if cw.maxMessages != 10 {
		t.Errorf("Expected maxMessages=10, got %d", cw.maxMessages)
	}

	if cw.messageThrottle != 1*time.Second {
		t.Errorf("Expected messageThrottle=1s, got %v", cw.messageThrottle)
	}

	if len(cw.messages) != 0 {
		t.Errorf("Expected 0 initial messages, got %d", len(cw.messages))
	}

	if cw.templates == nil {
		t.Fatal("Templates should be loaded")
	}
}

// Test ChatWindow with Invalid File
func TestNewChatWindowInvalidFile(t *testing.T) {
	fsys := os.DirFS("/")
	_, err := NewChatWindow(fsys, "nonexistent/path/to/chatter.json")
	if err == nil {
		t.Error("Expected error when loading nonexistent file, got nil")
	}
}

// Test Message Throttling
func TestChatWindowMessageThrottling(t *testing.T) {
	tempFile := createTestChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	// First message should succeed
	added := cw.AddMessage("Green1", "First message", 0)
	if !added {
		t.Error("First message should be added")
	}

	// Second message immediately after should be throttled
	added = cw.AddMessage("Green2", "Second message", 0)
	if added {
		t.Error("Second message should be throttled")
	}

	// Verify only one message was added
	messages := cw.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message after throttle, got %d", len(messages))
	}
}

// Test Message Throttling with Time Delay
func TestChatWindowMessageThrottlingWithDelay(t *testing.T) {
	tempFile := createTestChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	// Reduce throttle for faster testing
	cw.messageThrottle = 100 * time.Millisecond

	// First message should succeed
	added := cw.AddMessage("Green1", "First message", 0)
	if !added {
		t.Error("First message should be added")
	}

	// Wait for throttle to expire
	time.Sleep(150 * time.Millisecond)

	// Second message should now succeed
	added = cw.AddMessage("Green2", "Second message", 0)
	if !added {
		t.Error("Second message should be added after throttle expires")
	}

	// Verify both messages were added
	messages := cw.GetMessages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages after delay, got %d", len(messages))
	}
}

// Test Max Messages Limit
func TestChatWindowMaxMessages(t *testing.T) {
	tempFile := createTestChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	// Reduce throttle and max messages for faster testing
	cw.messageThrottle = 0 // Disable throttle for this test
	cw.maxMessages = 3

	// Add 5 messages
	for i := 0; i < 5; i++ {
		cw.lastMessageTime = time.Time{} // Reset throttle
		added := cw.AddMessage("TestShip", "Message", 0)
		if !added {
			t.Errorf("Message %d should be added", i)
		}
	}

	// Should only keep last 3 messages
	messages := cw.GetMessages()
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages (max limit), got %d", len(messages))
	}
}

// Test Ship Name Generation
func TestChatWindowGenerateShipName(t *testing.T) {
	tempFile := createTestChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	tests := []struct {
		factionID int
		shipID    int
		expected  string
	}{
		{0, 42, "Green42"},
		{1, 18, "Blue18"},
		{2, 7, "Red7"},
		{3, 99, "Yellow99"},
		{0, 1, "Green1"},
		{1, 123, "Blue123"},
	}

	for _, tt := range tests {
		result := cw.GenerateShipName(tt.factionID, tt.shipID)
		if result != tt.expected {
			t.Errorf("GenerateShipName(%d, %d): expected '%s', got '%s'",
				tt.factionID, tt.shipID, tt.expected, result)
		}
	}
}

// Test Ship Name Generation with Faction Wrapping
func TestChatWindowGenerateShipNameWrapping(t *testing.T) {
	tempFile := createTestChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	// Test that faction IDs wrap correctly (modulo 4)
	// Faction 4 should wrap to 0 (Green)
	result := cw.GenerateShipName(4, 10)
	if result != "Green10" {
		t.Errorf("Expected faction 4 to wrap to 'Green10', got '%s'", result)
	}

	// Faction 5 should wrap to 1 (Blue)
	result = cw.GenerateShipName(5, 20)
	if result != "Blue20" {
		t.Errorf("Expected faction 5 to wrap to 'Blue20', got '%s'", result)
	}
}

// Test OnKillFighter
func TestChatWindowOnKillFighter(t *testing.T) {
	tempFile := createTestChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	cw.OnKillFighter(0, 42)

	messages := cw.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.SpeakerName != "Green42" {
		t.Errorf("Expected speaker 'Green42', got '%s'", msg.SpeakerName)
	}

	if msg.FactionID != 0 {
		t.Errorf("Expected faction 0, got %d", msg.FactionID)
	}

	// Verify message is one of the templates
	validMessages := []string{"Got him! Bogey down", "Enemy fighter eliminated", "Splash one!"}
	found := false
	for _, valid := range validMessages {
		if msg.Text == valid {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Message text '%s' not in valid templates", msg.Text)
	}
}

// Test OnKillDestroyer
func TestChatWindowOnKillDestroyer(t *testing.T) {
	tempFile := createTestChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	cw.OnKillDestroyer(1, 18)

	messages := cw.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.SpeakerName != "Blue18" {
		t.Errorf("Expected speaker 'Blue18', got '%s'", msg.SpeakerName)
	}

	if msg.FactionID != 1 {
		t.Errorf("Expected faction 1, got %d", msg.FactionID)
	}
}

// Test OnKillTestudon
func TestChatWindowOnKillTestudon(t *testing.T) {
	tempFile := createTestChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	cw.OnKillTestudon(2, 7)

	messages := cw.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.SpeakerName != "Red7" {
		t.Errorf("Expected speaker 'Red7', got '%s'", msg.SpeakerName)
	}
}

// Test OnUnderAttack
func TestChatWindowOnUnderAttack(t *testing.T) {
	tempFile := createTestChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	cw.OnUnderAttack(0, 99)

	messages := cw.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.SpeakerName != "Green99" {
		t.Errorf("Expected speaker 'Green99', got '%s'", msg.SpeakerName)
	}
}

// Test OnFriendlyDestroyerDestroyed
func TestChatWindowOnFriendlyDestroyerDestroyed(t *testing.T) {
	tempFile := createTestChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	cw.OnFriendlyDestroyerDestroyed(3, 50)

	messages := cw.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.SpeakerName != "Yellow50" {
		t.Errorf("Expected speaker 'Yellow50', got '%s'", msg.SpeakerName)
	}
}

// Test OnFriendlyTestudonDestroyed
func TestChatWindowOnFriendlyTestudonDestroyed(t *testing.T) {
	tempFile := createTestChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	cw.OnFriendlyTestudonDestroyed(1, 33)

	messages := cw.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.SpeakerName != "Blue33" {
		t.Errorf("Expected speaker 'Blue33', got '%s'", msg.SpeakerName)
	}
}

// Test Message with Empty Templates
func TestChatWindowEmptyTemplates(t *testing.T) {
	tempFile := createEmptyChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	// Should not add message when template is empty
	cw.OnKillFighter(0, 1)

	messages := cw.GetMessages()
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages with empty template, got %d", len(messages))
	}
}

// Test Message Timestamp
func TestChatWindowMessageTimestamp(t *testing.T) {
	tempFile := createTestChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	beforeAdd := time.Now()
	cw.AddMessage("TestShip", "Test message", 0)
	afterAdd := time.Now()

	messages := cw.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.Timestamp.Before(beforeAdd) || msg.Timestamp.After(afterAdd) {
		t.Error("Message timestamp should be between before and after add time")
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// createTestChatterFile creates a temporary chatter.json file for testing
func createTestChatterFile(t *testing.T) string {
	t.Helper()

	templates := ChatterTemplates{
		KilledFighter:         []string{"Got him! Bogey down", "Enemy fighter eliminated", "Splash one!"},
		KilledDestroyer:       []string{"Destroyer down!", "Enemy capital ship destroyed"},
		KilledTestudon:        []string{"Turtle destroyed!", "Enemy testudon eliminated"},
		UnderAttack:           []string{"I can't shake this guy!", "Need backup!", "Taking fire!"},
		FriendlyDestroyerDown: []string{"We lost a destroyer!", "Friendly capital ship down"},
		FriendlyTestudonDown:  []string{"Our turtle is down!", "Testudon destroyed"},
	}

	data, err := json.MarshalIndent(templates, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal templates: %v", err)
	}

	tempFile := filepath.Join(t.TempDir(), "chatter.json")
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	return tempFile
}

// createEmptyChatterFile creates a chatter file with empty template arrays
func createEmptyChatterFile(t *testing.T) string {
	t.Helper()

	templates := ChatterTemplates{
		KilledFighter:         []string{},
		KilledDestroyer:       []string{},
		KilledTestudon:        []string{},
		UnderAttack:           []string{},
		FriendlyDestroyerDown: []string{},
		FriendlyTestudonDown:  []string{},
	}

	data, err := json.MarshalIndent(templates, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal templates: %v", err)
	}

	tempFile := filepath.Join(t.TempDir(), "chatter_empty.json")
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	return tempFile
}

// Test Clear Method
func TestChatWindowClear(t *testing.T) {
	tempFile := createTestChatterFile(t)
	defer os.Remove(tempFile)

	fsys := os.DirFS(filepath.Dir(tempFile))
	cw, err := NewChatWindow(fsys, filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create ChatWindow: %v", err)
	}

	// Disable throttle for testing
	cw.messageThrottle = 0

	// Add some messages
	cw.AddMessage("Green1", "First message", 0)
	cw.lastMessageTime = time.Time{} // Reset throttle
	cw.AddMessage("Blue2", "Second message", 1)
	cw.lastMessageTime = time.Time{} // Reset throttle
	cw.AddMessage("Red3", "Third message", 2)

	// Verify messages were added
	messages := cw.GetMessages()
	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages before clear, got %d", len(messages))
	}

	// Clear messages
	cw.Clear()

	// Verify messages are cleared
	messages = cw.GetMessages()
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", len(messages))
	}

	// Verify lastMessageTime is reset (should allow immediate message)
	added := cw.AddMessage("Green4", "After clear message", 0)
	if !added {
		t.Error("Message should be added immediately after clear (no throttle)")
	}

	messages = cw.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message after clear and new add, got %d", len(messages))
	}
}
