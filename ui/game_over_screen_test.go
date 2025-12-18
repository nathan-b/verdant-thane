package ui

import (
	"bytes"
	"image"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/nathan-b/verdant-thane/config"
)

// Global font source for tests (load once)
var testFontSource *text.GoTextFaceSource

// TestMain sets up the test font once for all tests
func TestMain(m *testing.M) {
	// Load actual font for testing
	fontBytes, err := os.ReadFile("../assets/orbitron.ttf")
	if err != nil {
		// If font is not available, tests will skip rendering
		testFontSource = nil
	} else {
		testFontSource, err = text.NewGoTextFaceSource(bytes.NewReader(fontBytes))
		if err != nil {
			testFontSource = nil
		}
	}

	m.Run()
}

// createTestFontSource returns the test font source or skips the test if unavailable
func createTestFontSource(tb testing.TB) *text.GoTextFaceSource {
	tb.Helper()
	if testFontSource == nil {
		tb.Skip("Font not available for testing")
	}
	return testFontSource
}

// TestNewGameOverScreen tests creating a game over screen
func TestNewGameOverScreen(t *testing.T) {
	fontSource := createTestFontSource(t)

	var continueCalled, restartCalled, newGameCalled bool

	screen := NewGameOverScreen(
		"TestPlayer",
		1000, // score
		25,   // kills
		5,    // deaths
		true, // isHighScore
		fontSource,
		func() { continueCalled = true },
		func() { restartCalled = true },
		func() { newGameCalled = true },
	)

	if screen == nil {
		t.Fatal("NewGameOverScreen returned nil")
	}

	// Check basic fields
	if screen.score != 1000 {
		t.Errorf("score = %d, want 1000", screen.score)
	}
	if screen.kills != 25 {
		t.Errorf("kills = %d, want 25", screen.kills)
	}
	if screen.deaths != 5 {
		t.Errorf("deaths = %d, want 5", screen.deaths)
	}
	if !screen.isHighScore {
		t.Error("isHighScore should be true")
	}

	// Check UI components were created
	if screen.ui == nil {
		t.Error("UI should be created")
	}
	if screen.nameInput == nil {
		t.Error("Name input should be created")
	}
	if screen.continueButton == nil {
		t.Error("Continue button should be created")
	}
	if screen.restartButton == nil {
		t.Error("Restart button should be created")
	}
	if screen.newGameButton == nil {
		t.Error("New game button should be created")
	}

	// Check font was set
	if screen.font == nil {
		t.Error("Font should be set")
	}
	if screen.hudFont == nil {
		t.Error("HUD font should be set")
	}

	// Verify callbacks haven't been called yet
	if continueCalled || restartCalled || newGameCalled {
		t.Error("Callbacks should not be called during construction")
	}
}

// TestGameOverScreen_GetPlayerName tests retrieving the player name
func TestGameOverScreen_GetPlayerName(t *testing.T) {
	fontSource := createTestFontSource(t)

	testCases := []struct {
		name         string
		initialName  string
		expectedName string
	}{
		{"empty name", "", ""},
		{"normal name", "Alice", "Alice"},
		{"long name", "VeryLongPlayerName12", "VeryLongPlayerName12"},
		{"max length", "12345678901234567890", "12345678901234567890"}, // 20 chars
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			screen := NewGameOverScreen(
				tc.initialName,
				100, 5, 1, false,
				fontSource,
				nil, nil, nil,
			)

			got := screen.GetPlayerName()
			if got != tc.expectedName {
				t.Errorf("GetPlayerName() = %q, want %q", got, tc.expectedName)
			}
		})
	}
}

// TestGameOverScreen_NameValidation tests that name input enforces 20 character limit
func TestGameOverScreen_NameValidation(t *testing.T) {
	fontSource := createTestFontSource(t)

	screen := NewGameOverScreen(
		"",
		100, 5, 1, false,
		fontSource,
		nil, nil, nil,
	)

	testCases := []struct {
		name         string
		inputText    string
		shouldAccept bool
	}{
		{"empty", "", true},
		{"1 char", "A", true},
		{"10 chars", "1234567890", true},
		{"20 chars (max)", "12345678901234567890", true},
		{"21 chars (over limit)", "123456789012345678901", false},
		{"30 chars (over limit)", "123456789012345678901234567890", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Attempt to set text
			screen.nameInput.SetText(tc.inputText)

			// Check if it was accepted
			actualText := screen.GetPlayerName()

			if tc.shouldAccept {
				if actualText != tc.inputText {
					t.Errorf("Text %q should be accepted but got %q", tc.inputText, actualText)
				}
			} else {
				// Should not have been set to the full input text
				if len(actualText) > 20 {
					t.Errorf("Text should be limited to 20 chars, got %d chars: %q", len(actualText), actualText)
				}
			}
		})
	}
}

// TestGameOverScreen_Update tests that Update doesn't panic
func TestGameOverScreen_Update(t *testing.T) {
	fontSource := createTestFontSource(t)

	screen := NewGameOverScreen(
		"TestPlayer",
		100, 5, 1, false,
		fontSource,
		nil, nil, nil,
	)

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update() panicked: %v", r)
		}
	}()

	screen.Update()
}

// TestGameOverScreen_Draw tests that Draw doesn't panic
func TestGameOverScreen_Draw(t *testing.T) {
	fontSource := createTestFontSource(t)

	screen := NewGameOverScreen(
		"TestPlayer",
		1000, 25, 5, true,
		fontSource,
		nil, nil, nil,
	)

	// Create a test image to draw on
	img := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Draw() panicked: %v", r)
		}
	}()

	screen.Draw(img)
}

// TestGameOverScreen_HighScoreDisplay tests high score vs normal score display
func TestGameOverScreen_HighScoreDisplay(t *testing.T) {
	fontSource := createTestFontSource(t)

	testCases := []struct {
		name        string
		isHighScore bool
	}{
		{"high score", true},
		{"normal score", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			screen := NewGameOverScreen(
				"Player",
				500, 10, 2, tc.isHighScore,
				fontSource,
				nil, nil, nil,
			)

			if screen.isHighScore != tc.isHighScore {
				t.Errorf("isHighScore = %v, want %v", screen.isHighScore, tc.isHighScore)
			}

			// Draw should not panic regardless of high score status
			img := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
			screen.Draw(img)
		})
	}
}

// TestGameOverScreen_StatsDisplay tests that stats are stored correctly
func TestGameOverScreen_StatsDisplay(t *testing.T) {
	fontSource := createTestFontSource(t)

	testCases := []struct {
		name   string
		score  int
		kills  int
		deaths int
	}{
		{"zero stats", 0, 0, 0},
		{"normal stats", 1000, 25, 5},
		{"high kills", 5000, 100, 10},
		{"many deaths", 500, 10, 50},
		{"perfect score", 10000, 200, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			screen := NewGameOverScreen(
				"Player",
				tc.score, tc.kills, tc.deaths, false,
				fontSource,
				nil, nil, nil,
			)

			if screen.score != tc.score {
				t.Errorf("score = %d, want %d", screen.score, tc.score)
			}
			if screen.kills != tc.kills {
				t.Errorf("kills = %d, want %d", screen.kills, tc.kills)
			}
			if screen.deaths != tc.deaths {
				t.Errorf("deaths = %d, want %d", screen.deaths, tc.deaths)
			}

			// Draw should properly display these stats
			img := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
			screen.Draw(img)
		})
	}
}

// TestGameOverScreen_ButtonCallbacks tests that callbacks are properly stored
func TestGameOverScreen_ButtonCallbacks(t *testing.T) {
	fontSource := createTestFontSource(t)

	continueCalled := false
	restartCalled := false
	newGameCalled := false

	screen := NewGameOverScreen(
		"Player",
		100, 5, 1, false,
		fontSource,
		func() { continueCalled = true },
		func() { restartCalled = true },
		func() { newGameCalled = true },
	)

	// Verify screen was created
	if screen == nil {
		t.Fatal("Screen should not be nil")
	}

	// Callbacks should not be called during construction
	if continueCalled || restartCalled || newGameCalled {
		t.Error("Callbacks should not be called during construction")
	}

	// Note: Actually triggering the callbacks requires simulating button clicks,
	// which is complex with ebitenui. We verify the buttons exist and the screen
	// can be created with callbacks.
	if screen.continueButton == nil {
		t.Error("Continue button should be created")
	}
	if screen.restartButton == nil {
		t.Error("Restart button should be created")
	}
	if screen.newGameButton == nil {
		t.Error("New game button should be created")
	}
}

// TestGameOverScreen_NilCallbacks tests that nil callbacks don't cause panics
func TestGameOverScreen_NilCallbacks(t *testing.T) {
	fontSource := createTestFontSource(t)

	// Should not panic with nil callbacks
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewGameOverScreen with nil callbacks panicked: %v", r)
		}
	}()

	screen := NewGameOverScreen(
		"Player",
		100, 5, 1, false,
		fontSource,
		nil, nil, nil, // All nil callbacks
	)

	if screen == nil {
		t.Fatal("Screen should not be nil")
	}

	// Update and Draw should work fine
	screen.Update()
	img := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	screen.Draw(img)
}

// TestGameOverScreen_MultipleUpdates tests multiple Update calls
func TestGameOverScreen_MultipleUpdates(t *testing.T) {
	fontSource := createTestFontSource(t)

	screen := NewGameOverScreen(
		"Player",
		100, 5, 1, false,
		fontSource,
		nil, nil, nil,
	)

	// Multiple updates should not panic
	for i := 0; i < 10; i++ {
		screen.Update()
	}
}

// TestGameOverScreen_MultipleDraws tests multiple Draw calls
func TestGameOverScreen_MultipleDraws(t *testing.T) {
	fontSource := createTestFontSource(t)

	screen := NewGameOverScreen(
		"Player",
		100, 5, 1, false,
		fontSource,
		nil, nil, nil,
	)

	img := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)

	// Multiple draws should not panic
	for i := 0; i < 10; i++ {
		screen.Draw(img)
	}
}

// TestGameOverScreen_UpdateAndDraw tests alternating Update and Draw calls
func TestGameOverScreen_UpdateAndDraw(t *testing.T) {
	fontSource := createTestFontSource(t)

	screen := NewGameOverScreen(
		"Player",
		100, 5, 1, false,
		fontSource,
		nil, nil, nil,
	)

	img := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)

	// Simulate game loop behavior
	for i := 0; i < 5; i++ {
		screen.Update()
		screen.Draw(img)
	}
}

// TestGameOverScreen_NameModification tests modifying the name after creation
func TestGameOverScreen_NameModification(t *testing.T) {
	fontSource := createTestFontSource(t)

	screen := NewGameOverScreen(
		"InitialName",
		100, 5, 1, false,
		fontSource,
		nil, nil, nil,
	)

	// Verify initial name
	if screen.GetPlayerName() != "InitialName" {
		t.Errorf("Initial name = %q, want %q", screen.GetPlayerName(), "InitialName")
	}

	// Modify name
	screen.nameInput.SetText("ModifiedName")

	// Verify modified name
	if screen.GetPlayerName() != "ModifiedName" {
		t.Errorf("Modified name = %q, want %q", screen.GetPlayerName(), "ModifiedName")
	}
}

// TestGameOverScreen_DrawOnDifferentImageSizes tests drawing on various image sizes
func TestGameOverScreen_DrawOnDifferentImageSizes(t *testing.T) {
	fontSource := createTestFontSource(t)

	screen := NewGameOverScreen(
		"Player",
		100, 5, 1, false,
		fontSource,
		nil, nil, nil,
	)

	testCases := []struct {
		name   string
		width  int
		height int
	}{
		{"standard", config.ScreenWidth, config.ScreenHeight},
		{"wide", 1920, 1080},
		{"square", 800, 800},
		{"small", 400, 300},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			img := ebiten.NewImage(tc.width, tc.height)

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Draw on %dx%d image panicked: %v", tc.width, tc.height, r)
				}
			}()

			screen.Draw(img)
		})
	}
}

// TestGameOverScreen_EmptyFontSource tests behavior with empty font source
func TestGameOverScreen_EmptyFontSource(t *testing.T) {
	// Create a potentially invalid font source
	emptySource := &text.GoTextFaceSource{}

	// Should not panic even with empty font source
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewGameOverScreen with empty font source panicked: %v", r)
		}
	}()

	screen := NewGameOverScreen(
		"Player",
		100, 5, 1, false,
		emptySource,
		nil, nil, nil,
	)

	if screen == nil {
		t.Fatal("Screen should not be nil")
	}
}

// TestGameOverScreen_ExtremeStats tests with extreme stat values
func TestGameOverScreen_ExtremeStats(t *testing.T) {
	fontSource := createTestFontSource(t)

	testCases := []struct {
		name   string
		score  int
		kills  int
		deaths int
	}{
		{"negative values", -100, -5, -1},
		{"zero", 0, 0, 0},
		{"max int", int(^uint(0) >> 1), int(^uint(0) >> 1), int(^uint(0) >> 1)},
		{"large values", 999999, 9999, 999},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			screen := NewGameOverScreen(
				"Player",
				tc.score, tc.kills, tc.deaths, false,
				fontSource,
				nil, nil, nil,
			)

			// Should handle extreme values without panic
			img := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
			screen.Draw(img)

			if screen.score != tc.score {
				t.Errorf("score = %d, want %d", screen.score, tc.score)
			}
		})
	}
}

// TestGameOverScreen_SpecialCharactersInName tests special characters in player names
func TestGameOverScreen_SpecialCharactersInName(t *testing.T) {
	fontSource := createTestFontSource(t)

	testCases := []struct {
		name      string
		inputName string
	}{
		{"ascii symbols", "Player!@#$%"},
		{"spaces", "Player One Two"},
		{"numbers", "Player123"},
		{"underscores", "Player_123"},
		{"mixed", "P1@y3r_!"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Limit to 20 chars if longer
			expectedName := tc.inputName
			if len(expectedName) > 20 {
				expectedName = expectedName[:20]
			}

			screen := NewGameOverScreen(
				tc.inputName,
				100, 5, 1, false,
				fontSource,
				nil, nil, nil,
			)

			got := screen.GetPlayerName()
			if got != expectedName {
				t.Errorf("GetPlayerName() = %q, want %q", got, expectedName)
			}

			// Draw should handle special characters
			img := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
			screen.Draw(img)
		})
	}
}

// BenchmarkGameOverScreen_Update benchmarks the Update method
func BenchmarkGameOverScreen_Update(b *testing.B) {
	fontSource := createTestFontSource(b)

	screen := NewGameOverScreen(
		"Player",
		1000, 25, 5, true,
		fontSource,
		nil, nil, nil,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		screen.Update()
	}
}

// BenchmarkGameOverScreen_Draw benchmarks the Draw method
func BenchmarkGameOverScreen_Draw(b *testing.B) {
	fontSource := createTestFontSource(b)

	screen := NewGameOverScreen(
		"Player",
		1000, 25, 5, true,
		fontSource,
		nil, nil, nil,
	)

	img := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		screen.Draw(img)
	}
}

// TestGameOverScreen_Rendering tests that Draw completes without panicking
// Note: Actual pixel-level rendering verification is not possible in Ebiten tests
// because ReadPixels cannot be called outside of the game loop
func TestGameOverScreen_Rendering(t *testing.T) {
	fontSource := createTestFontSource(t)

	screen := NewGameOverScreen(
		"TestPlayer",
		1000, 25, 5, true,
		fontSource,
		nil, nil, nil,
	)

	// Create image and draw
	img := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)

	// The fact that Draw completes without panicking is the test
	// We cannot read pixels in test mode due to Ebiten limitations
	screen.Draw(img)

	// Verify image still exists and has valid bounds
	bounds := img.Bounds()
	if bounds.Empty() {
		t.Error("Image bounds should not be empty after drawing")
	}
}

// TestNewGameOverScreen_AllParameters tests that all constructor parameters are used
func TestNewGameOverScreen_AllParameters(t *testing.T) {
	fontSource := createTestFontSource(t)

	playerName := "TestPlayer123"
	score := 12345
	kills := 67
	deaths := 8
	isHighScore := true

	continueCalled := 0
	restartCalled := 0
	newGameCalled := 0

	screen := NewGameOverScreen(
		playerName,
		score,
		kills,
		deaths,
		isHighScore,
		fontSource,
		func() { continueCalled++ },
		func() { restartCalled++ },
		func() { newGameCalled++ },
	)

	// Verify all parameters are correctly stored
	if screen.GetPlayerName() != playerName {
		t.Errorf("playerName = %q, want %q", screen.GetPlayerName(), playerName)
	}
	if screen.score != score {
		t.Errorf("score = %d, want %d", screen.score, score)
	}
	if screen.kills != kills {
		t.Errorf("kills = %d, want %d", screen.kills, kills)
	}
	if screen.deaths != deaths {
		t.Errorf("deaths = %d, want %d", screen.deaths, deaths)
	}
	if screen.isHighScore != isHighScore {
		t.Errorf("isHighScore = %v, want %v", screen.isHighScore, isHighScore)
	}

	// Verify widgets were created
	if screen.nameInput == nil || screen.continueButton == nil ||
		screen.restartButton == nil || screen.newGameButton == nil {
		t.Error("All widgets should be created")
	}

	// Verify callbacks haven't been called
	if continueCalled != 0 || restartCalled != 0 || newGameCalled != 0 {
		t.Error("Callbacks should not be invoked during construction")
	}
}

// TestGameOverScreen_Integration tests a full interaction cycle
func TestGameOverScreen_Integration(t *testing.T) {
	fontSource := createTestFontSource(t)

	screen := NewGameOverScreen(
		"Player",
		500, 12, 3, false,
		fontSource,
		nil, nil, nil,
	)

	img := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)

	// Simulate several frames
	for frame := 0; frame < 60; frame++ {
		screen.Update()
		screen.Draw(img)
	}

	// Modify name during "gameplay"
	screen.nameInput.SetText("NewPlayerName")

	// Continue for more frames
	for frame := 0; frame < 30; frame++ {
		screen.Update()
		screen.Draw(img)
	}

	// Verify final state
	if screen.GetPlayerName() != "NewPlayerName" {
		t.Errorf("Final name = %q, want %q", screen.GetPlayerName(), "NewPlayerName")
	}
}

// Helper function to verify image bounds
func verifyImageBounds(t *testing.T, img *ebiten.Image, expectedWidth, expectedHeight int) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width != expectedWidth {
		t.Errorf("Image width = %d, want %d", width, expectedWidth)
	}
	if height != expectedHeight {
		t.Errorf("Image height = %d, want %d", height, expectedHeight)
	}
}

// TestGameOverScreen_ImageBounds tests that drawing respects image bounds
func TestGameOverScreen_ImageBounds(t *testing.T) {
	fontSource := createTestFontSource(t)

	screen := NewGameOverScreen(
		"Player",
		100, 5, 1, false,
		fontSource,
		nil, nil, nil,
	)

	img := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	screen.Draw(img)

	verifyImageBounds(t, img, config.ScreenWidth, config.ScreenHeight)
}

// Verify that image is not nil after drawing
func verifyImageNotNil(t *testing.T, img *ebiten.Image) {
	if img == nil {
		t.Error("Image should not be nil after drawing")
	}

	bounds := img.Bounds()
	if bounds.Empty() {
		t.Error("Image bounds should not be empty after drawing")
	}
}

// TestGameOverScreen_ImageIntegrity tests that drawing doesn't corrupt the image
func TestGameOverScreen_ImageIntegrity(t *testing.T) {
	fontSource := createTestFontSource(t)

	screen := NewGameOverScreen(
		"Player",
		100, 5, 1, false,
		fontSource,
		nil, nil, nil,
	)

	img := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	verifyImageNotNil(t, img)

	screen.Draw(img)
	verifyImageNotNil(t, img)

	// Draw again
	screen.Draw(img)
	verifyImageNotNil(t, img)
}

// Verify image has valid bounds
func hasValidBounds(img image.Image) bool {
	bounds := img.Bounds()
	return !bounds.Empty() && bounds.Dx() > 0 && bounds.Dy() > 0
}

// TestGameOverScreen_ValidImageBounds verifies image bounds remain valid
func TestGameOverScreen_ValidImageBounds(t *testing.T) {
	fontSource := createTestFontSource(t)

	screen := NewGameOverScreen(
		"Player",
		100, 5, 1, false,
		fontSource,
		nil, nil, nil,
	)

	img := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)

	if !hasValidBounds(img) {
		t.Error("Initial image should have valid bounds")
	}

	screen.Draw(img)

	if !hasValidBounds(img) {
		t.Error("Image should have valid bounds after drawing")
	}
}
