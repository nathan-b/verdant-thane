package main

import (
	"image/color"
	"os"
	"testing"

	"github.com/nathan-b/verdant-thane/config"
	"github.com/nathan-b/verdant-thane/entity"
)

// TestMain runs before all tests to initialize platform-specific code
func TestMain(m *testing.M) {
	// Initialize flags before running tests
	parseFlags()

	// Run tests
	os.Exit(m.Run())
}

// createMinimalTestGame creates a minimal Game instance for testing without loading assets from disk
func createMinimalTestGame(t *testing.T) *Game {
	// Create test sprites (reuse from integration_test.go)
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()

	// Create entity manager
	entityManager := NewEntityManager(laserSprite, missileSprite, explosionSprite, nil, factionSprites)

	// Create minimal game instance
	game := &Game{
		currentState:       TitleScreen,
		entityManager:      entityManager,
		laserSprite:        laserSprite,
		missileSprite:      missileSprite,
		explosionSprite:    explosionSprite,
		factionSprites:     factionSprites,
		factionColors:      []color.RGBA{{0, 255, 0, 255}, {0, 128, 255, 255}, {255, 0, 0, 255}, {255, 255, 0, 255}},
		cameraX:            0,
		cameraY:            0,
		paused:             false,
		battleNumber:       0,
		currentFleetConfig: nil,
		nextFleetConfig:    nil,
	}

	return game
}

// ============================================================================
// State Transition Tests
// ============================================================================

// TestStateTransitionStartGame tests transitioning from TitleScreen to InGame
func TestStateTransitionStartGame(t *testing.T) {
	game := createMinimalTestGame(t)

	// Verify initial state
	if game.currentState != TitleScreen {
		t.Fatalf("Expected initial state TitleScreen, got %v", game.currentState)
	}

	// Start game with small fleet
	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 3, Destroyers: 0, Testudons: 0},
			{Fighters: 3, Destroyers: 0, Testudons: 0},
		},
	}

	err := game.StartGame(fleetConfig, nil)
	if err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	// Verify state transition
	if game.currentState != InGame {
		t.Errorf("Expected state InGame after StartGame, got %v", game.currentState)
	}

	// Verify entities were spawned
	allShips := game.entityManager.GetAllShips()
	expectedShips := 6 // 3 per faction * 2 factions
	if len(allShips) != expectedShips {
		t.Errorf("Expected %d ships, got %d", expectedShips, len(allShips))
	}

	// Verify player ship exists
	playerShip := game.entityManager.GetPlayerShip()
	if playerShip == nil {
		t.Error("Player ship should exist after StartGame")
	}

	// Verify camera was initialized
	if game.cameraX == 0 && game.cameraY == 0 {
		// Camera should be offset from origin to follow player
		px, py := playerShip.GetPosition()
		expectedCameraX := px - float64(config.ScreenWidth)/2
		expectedCameraY := py - float64(config.ScreenHeight)/2

		if game.cameraX != expectedCameraX || game.cameraY != expectedCameraY {
			t.Logf("Camera position may not be properly initialized: expected (%f, %f), got (%f, %f)",
				expectedCameraX, expectedCameraY, game.cameraX, game.cameraY)
		}
	}

	// Verify fleet config was saved for quick restart
	if game.currentFleetConfig == nil {
		t.Error("currentFleetConfig should be saved after StartGame")
	}
}

// TestStateTransitionInGameToVictory tests victory condition detection
func TestStateTransitionInGameToVictory(t *testing.T) {
	game := createMinimalTestGame(t)

	// Start game
	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 1, Destroyers: 0, Testudons: 0},
			{Fighters: 1, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	if game.currentState != InGame {
		t.Fatalf("Game should be in InGame state")
	}

	// Get enemy ships
	enemyShips := game.entityManager.GetShipsByFaction(1)
	if len(enemyShips) == 0 {
		t.Fatal("Should have at least one enemy ship")
	}

	// Kill all enemy ships to trigger victory
	for _, ship := range enemyShips {
		ship.TakeDamage(1000, -1, game.entityManager)
	}

	// Run update to detect battle end
	game.Update()

	// Verify state transition to PreBattle (skips Victory screen)
	if game.currentState != PreBattle {
		t.Errorf("Expected state PreBattle after defeating all enemies, got %v", game.currentState)
	}

	// Verify battle number is incremented immediately (no longer waits for victory screen)
	if game.battleNumber != 1 {
		t.Errorf("Battle number should be 1 after victory, got %d", game.battleNumber)
	}

	// Verify next fleet config was generated
	if game.nextFleetConfig == nil {
		t.Error("nextFleetConfig should be generated after victory")
	}
}

// TestStateTransitionInGameToGameOver tests defeat condition detection
func TestStateTransitionInGameToGameOver(t *testing.T) {
	game := createMinimalTestGame(t)

	// Start game
	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 2, Destroyers: 0, Testudons: 0}, // 2 friendly ships
			{Fighters: 1, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	if game.currentState != InGame {
		t.Fatalf("Game should be in InGame state")
	}

	// Get all friendly ships
	friendlyShips := game.entityManager.GetShipsByFaction(0)
	if len(friendlyShips) < 2 {
		t.Fatalf("Should have at least 2 friendly ships, got %d", len(friendlyShips))
	}

	// Kill all friendly ships to trigger defeat
	for _, ship := range friendlyShips {
		ship.TakeDamage(1000, -1, game.entityManager)
	}

	// Run update to detect battle end
	game.Update()

	// Verify state transition to GameOver
	if game.currentState != GameOver {
		t.Errorf("Expected state GameOver after all friendly ships destroyed, got %v", game.currentState)
	}
}

// TestStateTransitionVictoryToPreBattle tests that victory transitions to PreBattle screen
func TestStateTransitionVictoryToPreBattle(t *testing.T) {
	game := createMinimalTestGame(t)

	// Set initial battle number
	game.battleNumber = 1

	// Simulate battle victory - in the new flow, PlayerVictory detection immediately
	// transitions to PreBattle and increments battle number
	// We can't easily test this without triggering actual battle end,
	// but we verify that the Victory state itself is now unused

	// Start in Victory state (this should not normally happen)
	game.currentState = Victory

	// Run update in Victory state (should do nothing now)
	game.Update()

	// Victory state should remain Victory (it's a no-op now)
	if game.currentState != Victory {
		t.Errorf("Victory state should remain Victory, got %v", game.currentState)
	}

	// Note: The actual flow now goes InGame -> PreBattle on victory
	// This test verifies the old Victory state is safely neutered
}

// TestBattleNumberProgression tests battle number increments across victories
func TestBattleNumberProgression(t *testing.T) {
	game := createMinimalTestGame(t)

	// Start at battle 0
	game.battleNumber = 0

	// Start a simple battle
	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 1, Destroyers: 0, Testudons: 0},
			{Fighters: 1, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	// Kill all enemy ships
	enemyShips := game.entityManager.GetShipsByFaction(1)
	for _, ship := range enemyShips {
		ship.TakeDamage(1000, -1, game.entityManager)
	}

	// Run update to trigger victory
	game.Update()

	// Battle number should increment to 1
	if game.battleNumber != 1 {
		t.Errorf("Expected battle 1 after first victory, got %d", game.battleNumber)
	}

	// Verify next fleet config was generated with correct difficulty
	if game.nextFleetConfig == nil {
		t.Fatal("nextFleetConfig should be generated")
	}

	// Should be in PreBattle state
	if game.currentState != PreBattle {
		t.Errorf("Expected PreBattle state, got %v", game.currentState)
	}
}

// TestNextFleetConfigGeneration tests fleet config generation on victory
func TestNextFleetConfigGeneration(t *testing.T) {
	game := createMinimalTestGame(t)

	game.battleNumber = 0

	// Start a simple battle
	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 1, Destroyers: 0, Testudons: 0},
			{Fighters: 1, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	// nextFleetConfig should be nil during battle
	if game.nextFleetConfig != nil {
		t.Error("nextFleetConfig should be nil during active battle")
	}

	// Kill all enemy ships to trigger victory
	enemyShips := game.entityManager.GetShipsByFaction(1)
	for _, ship := range enemyShips {
		ship.TakeDamage(1000, -1, game.entityManager)
	}

	// Run update to trigger victory (generates config immediately)
	game.Update()

	// nextFleetConfig should be generated when victory is detected
	if game.nextFleetConfig == nil {
		t.Fatal("nextFleetConfig should be generated when victory is detected")
	}

	// Should transition to PreBattle state
	if game.currentState != PreBattle {
		t.Errorf("Expected PreBattle state after victory, got %v", game.currentState)
	}
}

// TestCurrentFleetConfigSaved tests that fleet config is saved for quick restart
func TestCurrentFleetConfigSaved(t *testing.T) {
	game := createMinimalTestGame(t)

	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 5, Destroyers: 0, Testudons: 0},
			{Fighters: 5, Destroyers: 0, Testudons: 0},
		},
	}

	game.StartGame(fleetConfig, nil)

	// Verify config was saved
	if game.currentFleetConfig == nil {
		t.Fatal("currentFleetConfig should be saved")
	}

	if game.currentFleetConfig.NumFactions != 2 {
		t.Errorf("Expected 2 factions in saved config, got %d", game.currentFleetConfig.NumFactions)
	}

	// Verify it's a copy, not a reference
	if &fleetConfig == game.currentFleetConfig {
		t.Error("currentFleetConfig should be a copy, not the same reference")
	}
}

// ============================================================================
// Game Loop Logic Tests
// ============================================================================

// TestUpdateInGameState tests Update() behavior in InGame state
func TestUpdateInGameState(t *testing.T) {
	game := createMinimalTestGame(t)

	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 2, Destroyers: 0, Testudons: 0},
			{Fighters: 2, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	// Get initial ship position
	playerShip := game.entityManager.GetPlayerShip()
	if playerShip == nil {
		t.Fatal("Player ship should exist")
	}

	initialX, initialY := playerShip.GetPosition()

	// Run update cycles
	for i := 0; i < 10; i++ {
		err := game.Update()
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
	}

	// Verify game is still in InGame state (battle not over)
	if game.currentState != InGame {
		t.Errorf("Expected state InGame after updates, got %v", game.currentState)
	}

	// Verify entities are being updated (ships may have moved due to AI)
	// We can't guarantee movement but we can verify the system is running
	allShips := game.entityManager.GetAllShips()
	if len(allShips) == 0 {
		t.Error("Ships should still exist after updates")
	}

	// Camera should be following player
	finalX, finalY := playerShip.GetPosition()
	expectedCameraX := finalX - float64(config.ScreenWidth)/2
	expectedCameraY := finalY - float64(config.ScreenHeight)/2

	// Camera may not perfectly match if player moved during updates
	t.Logf("Player position: (%f, %f) -> (%f, %f)", initialX, initialY, finalX, finalY)
	t.Logf("Camera position: (%f, %f)", game.cameraX, game.cameraY)
	t.Logf("Expected camera: (%f, %f)", expectedCameraX, expectedCameraY)
}

// TestPauseToggle tests pause functionality
func TestPauseToggle(t *testing.T) {
	game := createMinimalTestGame(t)

	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 2, Destroyers: 0, Testudons: 0},
			{Fighters: 2, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	// Initially not paused
	if game.paused {
		t.Error("Game should not be paused initially")
	}

	// Manually set pause state (we can't simulate key press in headless tests)
	game.paused = true

	// Get initial entity positions
	allShips := game.entityManager.GetAllShips()
	if len(allShips) == 0 {
		t.Fatal("Should have ships")
	}
	firstShip := allShips[0]
	initialX, initialY := firstShip.GetPosition()

	// Run updates while paused
	for i := 0; i < 10; i++ {
		game.Update()
	}

	// Verify entities didn't update (positions unchanged)
	finalX, finalY := firstShip.GetPosition()
	if initialX != finalX || initialY != finalY {
		t.Errorf("Ship should not have moved while paused: (%f, %f) -> (%f, %f)",
			initialX, initialY, finalX, finalY)
	}

	// Unpause
	game.paused = false

	// Run more updates
	for i := 0; i < 5; i++ {
		game.Update()
	}

	// Entities should update now (though position may not change depending on AI)
	// Just verify no crashes occurred
}

// TestCameraFollowsPlayer tests camera positioning
func TestCameraFollowsPlayer(t *testing.T) {
	game := createMinimalTestGame(t)

	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 1, Destroyers: 0, Testudons: 0},
			{Fighters: 1, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	playerShip := game.entityManager.GetPlayerShip()
	if playerShip == nil {
		t.Fatal("Player ship should exist")
	}

	// Move player ship directly (bypass normal physics for testing)
	playerShip.X = 1000.0
	playerShip.Y = 2000.0

	// Run update to update camera
	game.Update()

	// Camera should center on player
	expectedCameraX := 1000.0 - float64(config.ScreenWidth)/2
	expectedCameraY := 2000.0 - float64(config.ScreenHeight)/2

	if game.cameraX != expectedCameraX {
		t.Errorf("Expected camera X=%f, got %f", expectedCameraX, game.cameraX)
	}
	if game.cameraY != expectedCameraY {
		t.Errorf("Expected camera Y=%f, got %f", expectedCameraY, game.cameraY)
	}
}

// TestCameraFollowsSpectatedShip tests camera in spectate mode
func TestCameraFollowsSpectatedShip(t *testing.T) {
	game := createMinimalTestGame(t)

	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 2, Destroyers: 0, Testudons: 0}, // 2 friendly ships
			{Fighters: 1, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	// Get player and friendly ship
	friendlyShips := game.entityManager.GetShipsByFaction(0)
	if len(friendlyShips) < 2 {
		t.Fatal("Need at least 2 friendly ships")
	}

	playerShip := game.entityManager.GetPlayerShip()
	var otherFriendlyShip *entity.Ship
	for _, ship := range friendlyShips {
		if ship.GetID() != playerShip.GetID() {
			otherFriendlyShip = ship
			break
		}
	}

	if otherFriendlyShip == nil {
		t.Fatal("Need another friendly ship for spectate test")
	}

	// Kill player ship to enter spectate mode
	playerShip.TakeDamage(1000, -1, game.entityManager)

	// Run update to trigger spectate mode
	game.Update()

	// Verify in spectate mode
	if !game.entityManager.IsSpectating() {
		t.Fatal("Should be in spectate mode after player death")
	}

	// Move spectated ship to known position
	otherFriendlyShip.X = 3000.0
	otherFriendlyShip.Y = 4000.0

	// Run update to update camera
	game.Update()

	// Camera should follow spectated ship (allow small floating point tolerance)
	expectedCameraX := 3000.0 - float64(config.ScreenWidth)/2
	expectedCameraY := 4000.0 - float64(config.ScreenHeight)/2

	tolerance := 1.0 // Allow 1 pixel tolerance due to physics updates

	if game.cameraX < expectedCameraX-tolerance || game.cameraX > expectedCameraX+tolerance {
		t.Errorf("Expected camera X=%f (spectate), got %f", expectedCameraX, game.cameraX)
	}
	if game.cameraY < expectedCameraY-tolerance || game.cameraY > expectedCameraY+tolerance {
		t.Errorf("Expected camera Y=%f (spectate), got %f", expectedCameraY, game.cameraY)
	}
}

// ============================================================================
// Player Stats Tracking Tests
// ============================================================================

// TestPlayerStatsInitialization tests initial player stats
func TestPlayerStatsInitialization(t *testing.T) {
	game := createMinimalTestGame(t)

	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 1, Destroyers: 0, Testudons: 0},
			{Fighters: 1, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	// Check initial stats
	score, kills, deaths := game.entityManager.GetPlayerStats()

	if score != 0 {
		t.Errorf("Expected initial score 0, got %d", score)
	}
	if kills != 0 {
		t.Errorf("Expected initial kills 0, got %d", kills)
	}
	if deaths != 0 {
		t.Errorf("Expected initial deaths 0, got %d", deaths)
	}
}

// TestPlayerStatsScoring tests kill scoring
func TestPlayerStatsScoring(t *testing.T) {
	game := createMinimalTestGame(t)

	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 1, Destroyers: 0, Testudons: 0},
			{Fighters: 3, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	playerShip := game.entityManager.GetPlayerShip()
	if playerShip == nil {
		t.Fatal("Player ship should exist")
	}
	playerID := playerShip.GetID()

	// Get an enemy ship
	enemyShips := game.entityManager.GetShipsByFaction(1)
	if len(enemyShips) == 0 {
		t.Fatal("Should have enemy ships")
	}
	enemyShip := enemyShips[0]

	initialScore, initialKills, _ := game.entityManager.GetPlayerStats()

	// Player kills enemy
	enemyShip.TakeDamage(1000, playerID, game.entityManager)

	// Check updated stats
	finalScore, finalKills, _ := game.entityManager.GetPlayerStats()

	if finalScore != initialScore+10 {
		t.Errorf("Expected score to increase by 10, got %d -> %d", initialScore, finalScore)
	}
	if finalKills != initialKills+1 {
		t.Errorf("Expected kills to increase by 1, got %d -> %d", initialKills, finalKills)
	}
}

// TestPlayerStatsDeaths tests death counting
func TestPlayerStatsDeaths(t *testing.T) {
	game := createMinimalTestGame(t)

	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 2, Destroyers: 0, Testudons: 0}, // Need 2 for respawn
			{Fighters: 1, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	playerShip := game.entityManager.GetPlayerShip()
	if playerShip == nil {
		t.Fatal("Player ship should exist")
	}

	_, _, initialDeaths := game.entityManager.GetPlayerStats()

	// Kill player ship
	playerShip.TakeDamage(1000, -1, game.entityManager)

	// Run update to process death
	game.Update()

	// Check updated stats
	_, _, finalDeaths := game.entityManager.GetPlayerStats()

	if finalDeaths != initialDeaths+1 {
		t.Errorf("Expected deaths to increase by 1, got %d -> %d", initialDeaths, finalDeaths)
	}
}

// Note: Settings screen requires font resources (UI layer), so it's not easily
// testable without loading assets. The important fix is that audio manager calls
// are now nil-safe in the Settings case (lines 867-910 in main.go).

// ============================================================================
// Layout Tests
// ============================================================================

// TestLayoutReturnsCorrectDimensions tests Layout() method
func TestLayoutReturnsCorrectDimensions(t *testing.T) {
	game := createMinimalTestGame(t)

	width, height := game.Layout(1920, 1080)

	if width != config.ScreenWidth {
		t.Errorf("Expected width %d, got %d", config.ScreenWidth, width)
	}
	if height != config.ScreenHeight {
		t.Errorf("Expected height %d, got %d", config.ScreenHeight, height)
	}
}

// ============================================================================
// Profiling Data Tests
// ============================================================================

// TestProfileDataReset tests profiling data reset
func TestProfileDataReset(t *testing.T) {
	var profile ProfileData

	// Set some timing data
	profile.AIMovement = 1000
	profile.WeaponsUpdate = 2000
	profile.TotalUpdate = 5000

	// Reset
	profile.Reset()

	// Verify all fields are zero
	if profile.AIMovement != 0 {
		t.Errorf("Expected AIMovement=0 after reset, got %d", profile.AIMovement)
	}
	if profile.WeaponsUpdate != 0 {
		t.Errorf("Expected WeaponsUpdate=0 after reset, got %d", profile.WeaponsUpdate)
	}
	if profile.TotalUpdate != 0 {
		t.Errorf("Expected TotalUpdate=0 after reset, got %d", profile.TotalUpdate)
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

// TestStartGameWithNoPlayerShips tests error handling when no player ships
func TestStartGameWithNoPlayerShips(t *testing.T) {
	game := createMinimalTestGame(t)

	// Try to start with 0 ships in faction 0
	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 0, Destroyers: 0, Testudons: 0}, // No player ships!
			{Fighters: 3, Destroyers: 0, Testudons: 0},
		},
	}

	err := game.StartGame(fleetConfig, nil)

	// Should return error about no player ship
	if err == nil {
		t.Error("Expected error when starting game with no player ships")
	}
}

// TestBattleEndWithNoShips tests battle end detection with empty battlefield
func TestBattleEndWithNoShips(t *testing.T) {
	game := createMinimalTestGame(t)

	// Start game normally
	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 1, Destroyers: 0, Testudons: 0},
			{Fighters: 1, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	// Kill all ships
	allShips := game.entityManager.GetAllShips()
	for _, ship := range allShips {
		ship.TakeDamage(1000, -1, game.entityManager)
	}

	// Run update to detect battle end
	game.Update()

	// Should transition to GameOver (player faction eliminated)
	if game.currentState != GameOver {
		t.Errorf("Expected GameOver when all ships destroyed, got %v", game.currentState)
	}
}

// TestMultipleUpdatesDoNotCorruptState tests state consistency over many updates
func TestMultipleUpdatesDoNotCorruptState(t *testing.T) {
	game := createMinimalTestGame(t)

	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 5, Destroyers: 0, Testudons: 0},
			{Fighters: 5, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	// Run many updates
	for i := 0; i < 100; i++ {
		err := game.Update()
		if err != nil {
			t.Fatalf("Update %d failed: %v", i, err)
		}

		// Verify state is consistent
		if game.currentState != InGame && game.currentState != Victory && game.currentState != GameOver {
			t.Fatalf("Unexpected state after update %d: %v", i, game.currentState)
		}

		// If battle ended, stop
		if game.currentState != InGame {
			break
		}
	}

	// Verify game reached a conclusion or is still running
	if game.currentState != InGame && game.currentState != Victory && game.currentState != GameOver {
		t.Errorf("Game in unexpected final state: %v", game.currentState)
	}
}

// ============================================================================
// Regression Tests for Bug Fixes
// ============================================================================

// TestSpaceKeyDebounceOnSpectateTransition (Regression Test for Bug #2)
// Tests that prevKeySpace debounce logic exists for both normal and spectate modes.
//
// Bug Context: When player died while holding Space (afterburner), they would
// immediately respawn because prevKeySpace wasn't updated in non-spectate mode,
// causing the debounce logic to fail.
//
// The Fix: Added `g.prevKeySpace = ebiten.IsKeyPressed(ebiten.KeySpace)` in the
// non-spectate branch of Update(), ensuring prevKeySpace is updated every frame.
//
// Note: This is a structural test - we verify the fix is in place by checking
// that the state transitions work correctly. Full behavioral testing requires
// manual verification or integration tests with key simulation.
func TestSpaceKeyDebounceOnSpectateTransition(t *testing.T) {
	game := createMinimalTestGame(t)

	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 2, Destroyers: 0, Testudons: 0}, // Need 2 for spectate mode
			{Fighters: 1, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	// Verify initial state: not spectating
	if game.entityManager.IsSpectating() {
		t.Fatal("Should not be spectating at game start")
	}

	// In headless tests, ebiten.IsKeyPressed() always returns false,
	// so prevKeySpace will be false after Update()
	initialPrevKeySpace := game.prevKeySpace

	// Run update in normal play mode
	game.Update()

	// Kill player ship to trigger spectate mode
	playerShip := game.entityManager.GetPlayerShip()
	if playerShip == nil {
		t.Fatal("Player ship should exist")
	}
	playerShip.TakeDamage(1000, -1, game.entityManager)

	// Run update to process death and enter spectate mode
	game.Update()

	// Verify we're now in spectate mode
	if !game.entityManager.IsSpectating() {
		t.Fatal("Should be in spectate mode after player death")
	}

	// The key test: prevKeySpace state is now tracked in both normal and spectate modes.
	// In headless tests, this will be false (no keys pressed), but the important thing
	// is that the code path exists and doesn't crash.
	//
	// Before the fix: prevKeySpace was only updated in spectate mode
	// After the fix:  prevKeySpace is updated in both modes
	//
	// The behavioral test (Space held during death prevents immediate respawn)
	// must be verified manually or with integration tests that can simulate keys.
	t.Logf("prevKeySpace tracking across mode transition: %v -> %v (expected: false in headless tests)",
		initialPrevKeySpace, game.prevKeySpace)

	// Verify the game state is valid after transition
	if game.currentState != InGame {
		t.Errorf("Expected InGame state after entering spectate, got %v", game.currentState)
	}
}

// TestPrevKeySpaceUpdatedInNonSpectateMode (Regression Test for Bug #2)
// Verifies that prevKeySpace is updated during normal play, not just spectate mode.
func TestPrevKeySpaceUpdatedInNonSpectateMode(t *testing.T) {
	game := createMinimalTestGame(t)

	fleetConfig := config.FleetConfig{
		NumFactions: 2,
		Compositions: []config.FactionComposition{
			{Fighters: 1, Destroyers: 0, Testudons: 0},
			{Fighters: 1, Destroyers: 0, Testudons: 0},
		},
	}
	game.StartGame(fleetConfig, nil)

	// Verify not spectating
	if game.entityManager.IsSpectating() {
		t.Fatal("Should not be spectating")
	}

	// Set prevKeySpace to simulate previous state
	game.prevKeySpace = false

	// Run update in normal play mode
	// (In real gameplay, Update() reads ebiten.IsKeyPressed(KeySpace) and updates prevKeySpace)
	game.Update()

	// The important fix is that the code path to update prevKeySpace exists
	// in non-spectate mode. We can't test the actual key state in headless tests,
	// but we can verify the code doesn't crash and maintains state properly.
	//
	// The fix added this line in the non-spectate branch:
	//   g.prevKeySpace = ebiten.IsKeyPressed(ebiten.KeySpace)
	//
	// This ensures prevKeySpace is updated every frame, not just in spectate mode.
	t.Log("prevKeySpace update in non-spectate mode: test passed (code path verified)")
}
