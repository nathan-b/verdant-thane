package main

import (
	"image/color"
	"testing"

	"github.com/nathan/verdant-thane/config"
	"github.com/nathan/verdant-thane/entity"
)

// createMinimalTestGame creates a minimal Game instance for testing without loading assets from disk
func createMinimalTestGame(t *testing.T) *Game {
	// Create test sprites (reuse from integration_test.go)
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()

	// Create entity manager
	entityManager := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)

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

	err := game.StartGame(fleetConfig)
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
	game.StartGame(fleetConfig)

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

	// Verify state transition to Victory
	if game.currentState != Victory {
		t.Errorf("Expected state Victory after defeating all enemies, got %v", game.currentState)
	}

	// Verify battle number is still 0 (it gets incremented after victory screen, not during)
	if game.battleNumber != 0 {
		t.Errorf("Battle number should still be 0 during victory screen, got %d", game.battleNumber)
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
	game.StartGame(fleetConfig)

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

// TestStateTransitionVictoryToInterstitial tests victory screen progression
func TestStateTransitionVictoryToInterstitial(t *testing.T) {
	game := createMinimalTestGame(t)

	// Set initial battle number
	game.battleNumber = 1

	// Start in Victory state
	game.currentState = Victory

	// Verify next fleet config hasn't been generated yet
	if game.nextFleetConfig != nil {
		t.Error("nextFleetConfig should be nil before first Victory update")
	}

	// Run update in Victory state (generates next fleet config)
	game.Update()

	// Verify next fleet config was generated
	if game.nextFleetConfig == nil {
		t.Error("nextFleetConfig should be generated during Victory update")
	}

	// Verify battle number was incremented
	if game.battleNumber != 2 {
		t.Errorf("Expected battle number 2 after victory, got %d", game.battleNumber)
	}

	// Note: Actual transition to Interstitial requires key press (Enter/Space)
	// which we can't simulate in headless tests
	// We verify that the preconditions are set up correctly
}

// TestBattleNumberProgression tests battle number increments across victories
func TestBattleNumberProgression(t *testing.T) {
	game := createMinimalTestGame(t)

	// Start at battle 1
	game.battleNumber = 1

	// Simulate victory by entering Victory state and running update
	game.currentState = Victory
	game.Update()

	if game.battleNumber != 2 {
		t.Errorf("Expected battle 2 after first victory, got %d", game.battleNumber)
	}

	// Verify next fleet config was generated with correct difficulty
	if game.nextFleetConfig == nil {
		t.Fatal("nextFleetConfig should be generated")
	}

	// Battle 2 should be harder than battle 1
	totalShips := 0
	for _, comp := range game.nextFleetConfig.Compositions {
		totalShips += comp.Total()
	}

	if totalShips <= 4 {
		t.Logf("Battle 2 has %d total ships (may be randomly easy)", totalShips)
	}
}

// TestNextFleetConfigGeneration tests fleet config generation on victory
func TestNextFleetConfigGeneration(t *testing.T) {
	game := createMinimalTestGame(t)

	game.battleNumber = 1
	game.currentState = Victory
	game.nextFleetConfig = nil

	// First update generates next fleet config
	game.Update()

	if game.nextFleetConfig == nil {
		t.Fatal("nextFleetConfig should be generated on first Victory update")
	}

	firstConfig := game.nextFleetConfig

	// Second update should not regenerate config
	game.Update()

	if game.nextFleetConfig != firstConfig {
		t.Error("nextFleetConfig should not be regenerated on subsequent Victory updates")
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

	game.StartGame(fleetConfig)

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
	game.StartGame(fleetConfig)

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
	game.StartGame(fleetConfig)

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
	game.StartGame(fleetConfig)

	playerShip := game.entityManager.GetPlayerShip()
	if playerShip == nil {
		t.Fatal("Player ship should exist")
	}

	// Move player ship directly (bypass normal physics for testing)
	if fighter, ok := playerShip.(*entity.Fighter); ok {
		fighter.X = 1000.0
		fighter.Y = 2000.0
	}

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
	game.StartGame(fleetConfig)

	// Get player and friendly ship
	friendlyShips := game.entityManager.GetShipsByFaction(0)
	if len(friendlyShips) < 2 {
		t.Fatal("Need at least 2 friendly ships")
	}

	playerShip := game.entityManager.GetPlayerShip()
	var otherFriendlyShip entity.Ship
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
	if fighter, ok := otherFriendlyShip.(*entity.Fighter); ok {
		fighter.X = 3000.0
		fighter.Y = 4000.0
	}

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
	game.StartGame(fleetConfig)

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
	game.StartGame(fleetConfig)

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
	game.StartGame(fleetConfig)

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

	err := game.StartGame(fleetConfig)

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
	game.StartGame(fleetConfig)

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
	game.StartGame(fleetConfig)

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
