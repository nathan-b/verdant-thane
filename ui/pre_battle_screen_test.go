package ui

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/nathan-b/verdant-thane/config"
	"github.com/nathan-b/verdant-thane/systems"
)

// createMockFontSource creates a minimal font source for testing
func createMockFontSource() *text.GoTextFaceSource {
	// Create a minimal 1x1 font for testing (we won't actually render, just test logic)
	return &text.GoTextFaceSource{}
}

// createMockFactionSprites creates minimal faction sprites for testing
func createMockFactionSprites() *systems.FactionSprites {
	// Create minimal 1x1 images for testing
	mockImage := ebiten.NewImage(1, 1)
	return &systems.FactionSprites{
		Fighter: &systems.ShipClassSprites{
			Green: mockImage,
		},
		Destroyer: &systems.ShipClassSprites{
			Green: mockImage,
		},
		Testudon: &systems.ShipClassSprites{
			Green: mockImage,
		},
	}
}

// TestNewPreBattleScreen_FightersOnly tests creation with only fighters available
func TestNewPreBattleScreen_FightersOnly(t *testing.T) {
	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{
				Fighters:   10,
				Destroyers: 0,
				Testudons:  0,
			},
		},
	}

	fontSource := createMockFontSource()
	factionSprites := createMockFactionSprites()

	screen := NewPreBattleScreen(
		1, fleetConfig, 100, 5, 2, 3, 1,
		fontSource, factionSprites, nil,
	)

	// Should select Fighter as default when only fighters available
	if screen.selectedClass != config.ClassFighter {
		t.Errorf("Expected default class Fighter, got %v", screen.selectedClass)
	}

	// Should have exactly 1 ship button (fighter)
	if len(screen.shipButtons) != 1 {
		t.Errorf("Expected 1 ship button, got %d", len(screen.shipButtons))
	}

	// Button should be for Fighter class
	if screen.shipButtons[0].Class != config.ClassFighter {
		t.Errorf("Expected button for Fighter class, got %v", screen.shipButtons[0].Class)
	}
}

// TestNewPreBattleScreen_DestroyersOnly tests creation with only destroyers available
func TestNewPreBattleScreen_DestroyersOnly(t *testing.T) {
	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{
				Fighters:   0,
				Destroyers: 8,
				Testudons:  0,
			},
		},
	}

	fontSource := createMockFontSource()
	factionSprites := createMockFactionSprites()

	screen := NewPreBattleScreen(
		1, fleetConfig, 50, 2, 1, 1, 0,
		fontSource, factionSprites, nil,
	)

	// Should select Destroyer as default when only destroyers available
	if screen.selectedClass != config.ClassDestroyer {
		t.Errorf("Expected default class Destroyer, got %v", screen.selectedClass)
	}

	// Should have exactly 1 ship button (destroyer)
	if len(screen.shipButtons) != 1 {
		t.Errorf("Expected 1 ship button, got %d", len(screen.shipButtons))
	}

	// Button should be for Destroyer class
	if screen.shipButtons[0].Class != config.ClassDestroyer {
		t.Errorf("Expected button for Destroyer class, got %v", screen.shipButtons[0].Class)
	}
}

// TestNewPreBattleScreen_BothShipTypes tests creation with both fighters and destroyers
func TestNewPreBattleScreen_BothShipTypes(t *testing.T) {
	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{
				Fighters:   10,
				Destroyers: 5,
				Testudons:  2,
			},
		},
	}

	fontSource := createMockFontSource()
	factionSprites := createMockFactionSprites()

	screen := NewPreBattleScreen(
		2, fleetConfig, 250, 15, 7, 10, 3,
		fontSource, factionSprites, nil,
	)

	// Should select Fighter as default when both available
	if screen.selectedClass != config.ClassFighter {
		t.Errorf("Expected default class Fighter, got %v", screen.selectedClass)
	}

	// Should have exactly 2 ship buttons (fighter and destroyer)
	if len(screen.shipButtons) != 2 {
		t.Errorf("Expected 2 ship buttons, got %d", len(screen.shipButtons))
	}

	// First button should be Fighter
	if screen.shipButtons[0].Class != config.ClassFighter {
		t.Errorf("Expected first button for Fighter class, got %v", screen.shipButtons[0].Class)
	}

	// Second button should be Destroyer
	if screen.shipButtons[1].Class != config.ClassDestroyer {
		t.Errorf("Expected second button for Destroyer class, got %v", screen.shipButtons[1].Class)
	}
}

// TestNewPreBattleScreen_DataStorage tests that screen stores passed data correctly
func TestNewPreBattleScreen_DataStorage(t *testing.T) {
	fleetConfig := config.FleetConfig{
		NumFactions: 3,
		Compositions: []config.FactionComposition{
			{
				Fighters:   15,
				Destroyers: 10,
				Testudons:  3,
			},
		},
	}

	fontSource := createMockFontSource()
	factionSprites := createMockFactionSprites()

	battleNumber := 5
	score := 1234
	totalKills := 50
	totalDeaths := 12
	battleKills := 15
	battleDeaths := 3

	screen := NewPreBattleScreen(
		battleNumber, fleetConfig, score, totalKills, totalDeaths, battleKills, battleDeaths,
		fontSource, factionSprites, nil,
	)

	// Verify all data is stored correctly
	if screen.battleNumber != battleNumber {
		t.Errorf("Expected battleNumber %d, got %d", battleNumber, screen.battleNumber)
	}

	if screen.score != score {
		t.Errorf("Expected score %d, got %d", score, screen.score)
	}

	if screen.totalKills != totalKills {
		t.Errorf("Expected totalKills %d, got %d", totalKills, screen.totalKills)
	}

	if screen.totalDeaths != totalDeaths {
		t.Errorf("Expected totalDeaths %d, got %d", totalDeaths, screen.totalDeaths)
	}

	if screen.battleKills != battleKills {
		t.Errorf("Expected battleKills %d, got %d", battleKills, screen.battleKills)
	}

	if screen.battleDeaths != battleDeaths {
		t.Errorf("Expected battleDeaths %d, got %d", battleDeaths, screen.battleDeaths)
	}

	// Verify fleet config is stored
	if screen.fleetConfig.NumFactions != 3 {
		t.Errorf("Expected NumFactions 3, got %d", screen.fleetConfig.NumFactions)
	}

	if screen.playerFaction.Fighters != 15 {
		t.Errorf("Expected playerFaction Fighters 15, got %d", screen.playerFaction.Fighters)
	}

	if screen.playerFaction.Destroyers != 10 {
		t.Errorf("Expected playerFaction Destroyers 10, got %d", screen.playerFaction.Destroyers)
	}

	if screen.playerFaction.Testudons != 3 {
		t.Errorf("Expected playerFaction Testudons 3, got %d", screen.playerFaction.Testudons)
	}
}

// TestPreBattleScreen_ButtonLayout tests that buttons are positioned correctly
func TestPreBattleScreen_ButtonLayout(t *testing.T) {
	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{
				Fighters:   10,
				Destroyers: 5,
				Testudons:  0,
			},
		},
	}

	fontSource := createMockFontSource()
	factionSprites := createMockFactionSprites()

	screen := NewPreBattleScreen(
		1, fleetConfig, 0, 0, 0, 0, 0,
		fontSource, factionSprites, nil,
	)

	// Verify buttons have correct dimensions
	for i, btn := range screen.shipButtons {
		if btn.Width != 140.0 {
			t.Errorf("Button %d: expected width 140.0, got %f", i, btn.Width)
		}
		if btn.Height != 120.0 {
			t.Errorf("Button %d: expected height 120.0, got %f", i, btn.Height)
		}
		if !btn.Available {
			t.Errorf("Button %d: expected Available=true, got false", i)
		}
	}

	// Verify buttons are horizontally centered
	// With 2 buttons: total width = 2*140 + 30 = 310
	// Start X should be (ScreenWidth/2) - 155
	centerX := float64(config.ScreenWidth) / 2
	totalWidth := 2*140.0 + 30.0
	expectedStartX := centerX - totalWidth/2

	if screen.shipButtons[0].X != expectedStartX {
		t.Errorf("First button X: expected %f, got %f", expectedStartX, screen.shipButtons[0].X)
	}

	// Second button should be offset by button width + spacing
	expectedSecondX := expectedStartX + 140.0 + 30.0
	if screen.shipButtons[1].X != expectedSecondX {
		t.Errorf("Second button X: expected %f, got %f", expectedSecondX, screen.shipButtons[1].X)
	}

	// Start button should be centered
	expectedStartButtonX := centerX - 100.0 // 200px wide button
	if screen.startButton.X != expectedStartButtonX {
		t.Errorf("Start button X: expected %f, got %f", expectedStartButtonX, screen.startButton.X)
	}

	// Start button dimensions
	if screen.startButton.Width != 200 {
		t.Errorf("Start button width: expected 200, got %f", screen.startButton.Width)
	}
	if screen.startButton.Height != 50 {
		t.Errorf("Start button height: expected 50, got %f", screen.startButton.Height)
	}
	if screen.startButton.Label != "START BATTLE" {
		t.Errorf("Start button label: expected 'START BATTLE', got '%s'", screen.startButton.Label)
	}
}

// TestPreBattleScreen_OnStartCallback tests that the callback is invoked with correct class
func TestPreBattleScreen_OnStartCallback(t *testing.T) {
	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{
				Fighters:   10,
				Destroyers: 5,
				Testudons:  0,
			},
		},
	}

	fontSource := createMockFontSource()
	factionSprites := createMockFactionSprites()

	// Track callback invocations
	callbackInvoked := false
	var receivedClass config.ShipClass

	onStart := func(selectedClass config.ShipClass) {
		callbackInvoked = true
		receivedClass = selectedClass
	}

	screen := NewPreBattleScreen(
		1, fleetConfig, 0, 0, 0, 0, 0,
		fontSource, factionSprites, onStart,
	)

	// Manually invoke the callback to verify it works
	if screen.onStart != nil {
		screen.onStart(config.ClassDestroyer)
	}

	if !callbackInvoked {
		t.Error("Expected callback to be invoked")
	}

	if receivedClass != config.ClassDestroyer {
		t.Errorf("Expected callback to receive ClassDestroyer, got %v", receivedClass)
	}
}

// TestPreBattleScreen_SingleButtonLayout tests layout with only one ship type
func TestPreBattleScreen_SingleButtonLayout(t *testing.T) {
	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{
				Fighters:   10,
				Destroyers: 0,
				Testudons:  0,
			},
		},
	}

	fontSource := createMockFontSource()
	factionSprites := createMockFactionSprites()

	screen := NewPreBattleScreen(
		1, fleetConfig, 0, 0, 0, 0, 0,
		fontSource, factionSprites, nil,
	)

	// With 1 button: total width = 140
	// Start X should be (ScreenWidth/2) - 70
	centerX := float64(config.ScreenWidth) / 2
	expectedX := centerX - 70.0

	if screen.shipButtons[0].X != expectedX {
		t.Errorf("Single button X: expected %f, got %f", expectedX, screen.shipButtons[0].X)
	}
}
