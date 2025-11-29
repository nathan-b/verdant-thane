package main

import (
	"encoding/json"
	"image/color"
	"math"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/nathan-b/verdant-thane/config"
)

// KillstreakText holds the mapping of kill counts to display messages
type KillstreakText struct {
	messages map[int]string // Maps kill count to message text
}

// LoadKillstreakText loads the killtext.json file and returns the mapping
func LoadKillstreakText(filepath string) (*KillstreakText, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	kt := &KillstreakText{
		messages: make(map[int]string),
	}

	// Try to unmarshal directly into map[int]string
	// Go's JSON library will convert string keys to int if possible
	if err := json.Unmarshal(data, &kt.messages); err != nil {
		return nil, err
	}

	return kt, nil
}

// GetMessage returns the message for a specific kill count, or empty string if none
func (kt *KillstreakText) GetMessage(killCount int) string {
	return kt.messages[killCount]
}

// KillstreakDisplay manages the display of killstreak messages
type KillstreakDisplay struct {
	// Current display state
	currentMessage string    // Currently displayed message
	displayStart   time.Time // When the current message started displaying
	isActive       bool      // Whether a message is currently being displayed

	// Display configuration
	displayDuration time.Duration // How long to display each message (3 seconds)

	// Killstreak text mapping
	killtext *KillstreakText

	// Font for rendering
	fontSource *text.GoTextFaceSource
}

// NewKillstreakDisplay creates a new killstreak display
func NewKillstreakDisplay(killtext *KillstreakText, fontSource *text.GoTextFaceSource) *KillstreakDisplay {
	return &KillstreakDisplay{
		currentMessage:  "",
		isActive:        false,
		displayDuration: 3 * time.Second,
		killtext:        killtext,
		fontSource:      fontSource,
	}
}

// OnKill is called when the player gets a kill (per-round kill count)
func (kd *KillstreakDisplay) OnKill(roundKillCount int) {
	// Check if there's a message for this kill count
	message := kd.killtext.GetMessage(roundKillCount)
	if message != "" {
		// Display the message (overwrites any existing message)
		kd.currentMessage = message
		kd.displayStart = time.Now()
		kd.isActive = true
	}
}

// Update updates the killstreak display state
func (kd *KillstreakDisplay) Update() {
	if !kd.isActive {
		return
	}

	// Check if display duration has elapsed
	if time.Since(kd.displayStart) >= kd.displayDuration {
		kd.isActive = false
		kd.currentMessage = ""
	}
}

// Render draws the killstreak message to the screen
func (kd *KillstreakDisplay) Render(screen *ebiten.Image) {
	if !kd.isActive || kd.currentMessage == "" {
		return
	}

	// Create large font for killstreak text
	const fontSize = 48
	font := &text.GoTextFace{
		Source: kd.fontSource,
		Size:   fontSize,
	}

	// Calculate position (center horizontal, 25% vertical)
	messageWidth, _ := text.Measure(kd.currentMessage, font, 0)
	centerX := (float64(config.ScreenWidth) - messageWidth) / 2
	centerY := float64(config.ScreenHeight) * 0.25

	// Calculate color (cycle through spectrum based on time)
	// Use time since display start to create smooth color cycling
	elapsedSec := time.Since(kd.displayStart).Seconds()
	// Complete 3 full color cycles over 3 seconds (1 cycle per second)
	hue := math.Mod(elapsedSec*360.0, 360.0) // 0-360 degrees

	// Convert HSV to RGB (S=1.0, V=1.0 for full saturation and brightness)
	r, g, b := hsvToRGB(hue, 1.0, 1.0)
	textColor := color.RGBA{R: r, G: g, B: b, A: 255}

	// Draw the text
	textOpts := &text.DrawOptions{}
	textOpts.GeoM.Translate(centerX, centerY)
	textOpts.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, kd.currentMessage, font, textOpts)
}

// hsvToRGB converts HSV color space to RGB
// h: 0-360, s: 0-1, v: 0-1
// Returns RGB values 0-255
func hsvToRGB(h, s, v float64) (uint8, uint8, uint8) {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := v - c

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return uint8((r + m) * 255), uint8((g + m) * 255), uint8((b + m) * 255)
}
