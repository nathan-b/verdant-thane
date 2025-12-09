package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

// ChatMessage represents a single chat message
type ChatMessage struct {
	SpeakerName string
	Text        string
	FactionID   int
	Timestamp   time.Time
}

// ChatterTemplates holds the loaded message templates from chatter.json
type ChatterTemplates struct {
	KilledFighter         []string `json:"killed_fighter"`
	KilledDestroyer       []string `json:"killed_destroyer"`
	KilledTestudon        []string `json:"killed_testudon"`
	UnderAttack           []string `json:"under_attack"`
	FriendlyDestroyerDown []string `json:"friendly_destroyer_down"`
	FriendlyTestudonDown  []string `json:"friendly_testudon_down"`
}

// ChatWindow manages the in-game chat system
type ChatWindow struct {
	messages          []ChatMessage
	maxMessages       int
	lastMessageTime   time.Time
	messageThrottle   time.Duration
	templates         *ChatterTemplates
	factionColorNames []string // e.g., ["Green", "Blue", "Red", "Yellow"]
}

// NewChatWindow creates a new chat window
func NewChatWindow(chatterFilePath string) (*ChatWindow, error) {
	// Load chatter templates
	templates, err := loadChatterTemplates(chatterFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chatter templates: %w", err)
	}

	return &ChatWindow{
		messages:          make([]ChatMessage, 0),
		maxMessages:       10, // Keep last 10 messages
		lastMessageTime:   time.Time{},
		messageThrottle:   1 * time.Second, // Max 1 message per second
		templates:         templates,
		factionColorNames: []string{"Green", "Blue", "Red", "Yellow"},
	}, nil
}

// loadChatterTemplates loads message templates from JSON file
func loadChatterTemplates(filePath string) (*ChatterTemplates, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var templates ChatterTemplates
	if err := json.Unmarshal(data, &templates); err != nil {
		return nil, err
	}

	return &templates, nil
}

// AddMessage attempts to add a message to the chat window
// Returns true if message was added, false if throttled
func (cw *ChatWindow) AddMessage(speakerName, text string, factionID int) bool {
	now := time.Now()

	// Throttle: no more than 1 message per second
	if !cw.lastMessageTime.IsZero() && now.Sub(cw.lastMessageTime) < cw.messageThrottle {
		return false // Message throttled
	}

	// Add message
	msg := ChatMessage{
		SpeakerName: speakerName,
		Text:        text,
		FactionID:   factionID,
		Timestamp:   now,
	}

	cw.messages = append(cw.messages, msg)

	// Keep only last N messages
	if len(cw.messages) > cw.maxMessages {
		cw.messages = cw.messages[1:]
	}

	cw.lastMessageTime = now
	return true
}

// GetMessages returns the current list of messages
func (cw *ChatWindow) GetMessages() []ChatMessage {
	return cw.messages
}

// GenerateShipName generates a ship name like "Green42" or "Blue18"
func (cw *ChatWindow) GenerateShipName(factionID, shipID int) string {
	colorName := cw.factionColorNames[factionID%len(cw.factionColorNames)]
	return fmt.Sprintf("%s%d", colorName, shipID)
}

// OnKillFighter generates a random kill message for a fighter
func (cw *ChatWindow) OnKillFighter(killerFactionID, killerShipID int) {
	if len(cw.templates.KilledFighter) == 0 {
		return
	}

	speakerName := cw.GenerateShipName(killerFactionID, killerShipID)
	text := cw.templates.KilledFighter[rand.Intn(len(cw.templates.KilledFighter))]
	cw.AddMessage(speakerName, text, killerFactionID)
}

// OnKillDestroyer generates a random kill message for a destroyer
func (cw *ChatWindow) OnKillDestroyer(killerFactionID, killerShipID int) {
	if len(cw.templates.KilledDestroyer) == 0 {
		return
	}

	speakerName := cw.GenerateShipName(killerFactionID, killerShipID)
	text := cw.templates.KilledDestroyer[rand.Intn(len(cw.templates.KilledDestroyer))]
	cw.AddMessage(speakerName, text, killerFactionID)
}

// OnKillTestudon generates a random kill message for a testudon
func (cw *ChatWindow) OnKillTestudon(killerFactionID, killerShipID int) {
	if len(cw.templates.KilledTestudon) == 0 {
		return
	}

	speakerName := cw.GenerateShipName(killerFactionID, killerShipID)
	text := cw.templates.KilledTestudon[rand.Intn(len(cw.templates.KilledTestudon))]
	cw.AddMessage(speakerName, text, killerFactionID)
}

// OnUnderAttack generates a random under attack message
func (cw *ChatWindow) OnUnderAttack(shipFactionID, shipID int) {
	if len(cw.templates.UnderAttack) == 0 {
		return
	}

	speakerName := cw.GenerateShipName(shipFactionID, shipID)
	text := cw.templates.UnderAttack[rand.Intn(len(cw.templates.UnderAttack))]
	cw.AddMessage(speakerName, text, shipFactionID)
}

// OnFriendlyDestroyerDestroyed generates a message when friendly destroyer is destroyed
func (cw *ChatWindow) OnFriendlyDestroyerDestroyed(observerFactionID, observerShipID int) {
	if len(cw.templates.FriendlyDestroyerDown) == 0 {
		return
	}

	speakerName := cw.GenerateShipName(observerFactionID, observerShipID)
	text := cw.templates.FriendlyDestroyerDown[rand.Intn(len(cw.templates.FriendlyDestroyerDown))]
	cw.AddMessage(speakerName, text, observerFactionID)
}

// OnFriendlyTestudonDestroyed generates a message when friendly testudon is destroyed
func (cw *ChatWindow) OnFriendlyTestudonDestroyed(observerFactionID, observerShipID int) {
	if len(cw.templates.FriendlyTestudonDown) == 0 {
		return
	}

	speakerName := cw.GenerateShipName(observerFactionID, observerShipID)
	text := cw.templates.FriendlyTestudonDown[rand.Intn(len(cw.templates.FriendlyTestudonDown))]
	cw.AddMessage(speakerName, text, observerFactionID)
}

// Clear removes all messages from the chat window
func (cw *ChatWindow) Clear() {
	cw.messages = make([]ChatMessage, 0)
	cw.lastMessageTime = time.Time{}
}
