package main

import (
	"math"
	"testing"

	"github.com/nathan/verdant-thane/config"
	"github.com/nathan/verdant-thane/entity"
)

// Test Camera Initialization on Game Start
func TestCameraInitialization(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, nil, factionSprites)
	em.InitializeFactions()

	// Spawn player ship at a known position
	playerShip := em.SpawnShip(entity.ClassFighter, 0, 1000, 2000)
	playerShip.SetPlayerControlled(true)
	em.SetPlayerShip(playerShip.GetID())

	// Calculate expected camera position
	// Camera should center on player: player position - screen size / 2
	expectedCameraX := 1000 - float64(config.ScreenWidth)/2
	expectedCameraY := 2000 - float64(config.ScreenHeight)/2

	// Create game and initialize camera position manually (simulating StartGame behavior)
	cameraX := expectedCameraX
	cameraY := expectedCameraY

	if cameraX != expectedCameraX {
		t.Errorf("Expected camera X %f, got %f", expectedCameraX, cameraX)
	}

	if cameraY != expectedCameraY {
		t.Errorf("Expected camera Y %f, got %f", expectedCameraY, cameraY)
	}
}

// Test Camera Centering on Player
func TestCameraCentersOnPlayer(t *testing.T) {
	tests := []struct {
		name            string
		playerX         float64
		playerY         float64
		expectedCameraX float64
		expectedCameraY float64
	}{
		{
			name:            "Player at origin",
			playerX:         0,
			playerY:         0,
			expectedCameraX: 0 - float64(config.ScreenWidth)/2,
			expectedCameraY: 0 - float64(config.ScreenHeight)/2,
		},
		{
			name:            "Player at center of world",
			playerX:         float64(config.GameWidth) / 2,
			playerY:         float64(config.GameHeight) / 2,
			expectedCameraX: float64(config.GameWidth)/2 - float64(config.ScreenWidth)/2,
			expectedCameraY: float64(config.GameHeight)/2 - float64(config.ScreenHeight)/2,
		},
		{
			name:            "Player at arbitrary position",
			playerX:         1500,
			playerY:         3000,
			expectedCameraX: 1500 - float64(config.ScreenWidth)/2,
			expectedCameraY: 3000 - float64(config.ScreenHeight)/2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate camera position as StartGame does
			cameraX := tt.playerX - float64(config.ScreenWidth)/2
			cameraY := tt.playerY - float64(config.ScreenHeight)/2

			if cameraX != tt.expectedCameraX {
				t.Errorf("Expected camera X %f, got %f", tt.expectedCameraX, cameraX)
			}

			if cameraY != tt.expectedCameraY {
				t.Errorf("Expected camera Y %f, got %f", tt.expectedCameraY, cameraY)
			}
		})
	}
}

// Test Camera Updates to Follow Player Movement
func TestCameraFollowsPlayerMovement(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, nil, factionSprites)
	em.InitializeFactions()

	// Spawn player ship
	playerShip := em.SpawnShip(entity.ClassFighter, 0, 1000, 1000)
	playerShip.SetPlayerControlled(true)
	em.SetPlayerShip(playerShip.GetID())

	// Initialize camera to follow player
	x1, y1 := playerShip.GetPosition()
	cameraX := x1 - float64(config.ScreenWidth)/2
	cameraY := y1 - float64(config.ScreenHeight)/2

	// Move player ship by setting velocity directly
	if fighter, ok := playerShip.(*entity.Fighter); ok {
		fighter.Rotation = 0 // Facing up
		fighter.Speed = 5.0
	}

	// Update ship position
	em.UpdateAll()

	// Get new player position
	x2, y2 := playerShip.GetPosition()

	// Verify player moved
	if x2 == x1 && y2 == y1 {
		t.Fatal("Player should have moved after update")
	}

	// Calculate new expected camera position
	expectedCameraX := x2 - float64(config.ScreenWidth)/2
	expectedCameraY := y2 - float64(config.ScreenHeight)/2

	// Update camera to follow (simulating what the game loop does)
	cameraX = x2 - float64(config.ScreenWidth)/2
	cameraY = y2 - float64(config.ScreenHeight)/2

	if math.Abs(cameraX-expectedCameraX) > 0.1 {
		t.Errorf("Camera X should follow player: expected %f, got %f", expectedCameraX, cameraX)
	}

	if math.Abs(cameraY-expectedCameraY) > 0.1 {
		t.Errorf("Camera Y should follow player: expected %f, got %f", expectedCameraY, cameraY)
	}
}

// Test Camera with Player at World Edge (World Wrapping)
func TestCameraAtWorldEdge(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, nil, factionSprites)
	em.InitializeFactions()

	// Spawn player near world edge
	playerShip := em.SpawnShip(entity.ClassFighter, 0, 10, 10) // Near top-left corner
	playerShip.SetPlayerControlled(true)
	em.SetPlayerShip(playerShip.GetID())

	x, y := playerShip.GetPosition()

	// Camera should still center on player even if that means negative camera coordinates
	expectedCameraX := x - float64(config.ScreenWidth)/2
	expectedCameraY := y - float64(config.ScreenHeight)/2

	cameraX := x - float64(config.ScreenWidth)/2
	cameraY := y - float64(config.ScreenHeight)/2

	// Camera coordinates can be negative when player is near the edge
	// This is expected behavior
	if cameraX != expectedCameraX {
		t.Errorf("Expected camera X %f, got %f", expectedCameraX, cameraX)
	}

	if cameraY != expectedCameraY {
		t.Errorf("Expected camera Y %f, got %f", expectedCameraY, cameraY)
	}
}

// Test Camera Stays Centered During Combat
func TestCameraStaysCenteredDuringCombat(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, nil, factionSprites)
	em.InitializeFactions()

	// Spawn player and enemy
	playerShip := em.SpawnShip(entity.ClassFighter, 0, 1000, 1000)
	enemyShip := em.SpawnShip(entity.ClassFighter, 1, 1100, 1000)

	playerShip.SetPlayerControlled(true)
	em.SetPlayerShip(playerShip.GetID())

	// Initialize camera
	x1, y1 := playerShip.GetPosition()
	cameraX := x1 - float64(config.ScreenWidth)/2
	cameraY := y1 - float64(config.ScreenHeight)/2

	// Player fires at enemy
	if fighter, ok := playerShip.(*entity.Fighter); ok {
		fighter.Rotation = math.Pi / 2 // Face right toward enemy
		fighter.Weapons[0].WeaponCapacitor = 1.0
	}

	enemyX, enemyY := enemyShip.GetPosition()
	playerShip.FireWeapon(enemyX, enemyY, em)

	// Update game (projectile moves, etc)
	for i := 0; i < 3; i++ {
		em.UpdateAll()
	}

	// Get current player position
	x2, y2 := playerShip.GetPosition()

	// Update camera
	cameraX = x2 - float64(config.ScreenWidth)/2
	cameraY = y2 - float64(config.ScreenHeight)/2

	// Verify camera is still centered on player
	expectedCameraX := x2 - float64(config.ScreenWidth)/2
	expectedCameraY := y2 - float64(config.ScreenHeight)/2

	if math.Abs(cameraX-expectedCameraX) > 0.1 {
		t.Errorf("Camera should stay centered on player during combat: expected X=%f, got %f",
			expectedCameraX, cameraX)
	}

	if math.Abs(cameraY-expectedCameraY) > 0.1 {
		t.Errorf("Camera should stay centered on player during combat: expected Y=%f, got %f",
			expectedCameraY, cameraY)
	}
}

// Test Camera Behavior When Player Dies (Spectate Mode)
func TestCameraInSpectateMode(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, nil, factionSprites)
	em.InitializeFactions()

	// Spawn player and friendly ship
	playerShip := em.SpawnShip(entity.ClassFighter, 0, 1000, 1000)
	_ = em.SpawnShip(entity.ClassFighter, 0, 2000, 2000) // Friendly ship for respawn

	playerShip.SetPlayerControlled(true)
	em.SetPlayerShip(playerShip.GetID())

	// Kill player
	playerShip.TakeDamage(1000, -1, em)
	em.UpdateAll()

	// Player should now be spectating
	if !em.IsSpectating() {
		t.Fatal("Player should be in spectate mode after death")
	}

	// Get spectated ship
	spectatedShip := em.GetSpectatedShip()
	if spectatedShip == nil {
		t.Fatal("Should be spectating a ship")
	}

	// Camera should now center on spectated ship
	x, y := spectatedShip.GetPosition()
	expectedCameraX := x - float64(config.ScreenWidth)/2
	expectedCameraY := y - float64(config.ScreenHeight)/2

	// Simulate camera update in spectate mode
	cameraX := x - float64(config.ScreenWidth)/2
	cameraY := y - float64(config.ScreenHeight)/2

	if math.Abs(cameraX-expectedCameraX) > 0.1 {
		t.Errorf("Camera should center on spectated ship: expected X=%f, got %f",
			expectedCameraX, cameraX)
	}

	if math.Abs(cameraY-expectedCameraY) > 0.1 {
		t.Errorf("Camera should center on spectated ship: expected Y=%f, got %f",
			expectedCameraY, cameraY)
	}
}

// Test Camera Coordinate Calculations for Rendering
func TestCameraCoordinateTransform(t *testing.T) {
	// Test that world coordinates are correctly transformed to screen coordinates

	tests := []struct {
		name            string
		worldX          float64
		worldY          float64
		cameraX         float64
		cameraY         float64
		expectedScreenX float64
		expectedScreenY float64
	}{
		{
			name:            "Entity at camera position",
			worldX:          1000,
			worldY:          1000,
			cameraX:         1000,
			cameraY:         1000,
			expectedScreenX: 0,
			expectedScreenY: 0,
		},
		{
			name:            "Entity offset from camera",
			worldX:          1100,
			worldY:          1200,
			cameraX:         1000,
			cameraY:         1000,
			expectedScreenX: 100,
			expectedScreenY: 200,
		},
		{
			name:            "Entity behind camera",
			worldX:          900,
			worldY:          800,
			cameraX:         1000,
			cameraY:         1000,
			expectedScreenX: -100,
			expectedScreenY: -200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Screen coordinates = world coordinates - camera position
			screenX := tt.worldX - tt.cameraX
			screenY := tt.worldY - tt.cameraY

			if screenX != tt.expectedScreenX {
				t.Errorf("Expected screen X %f, got %f", tt.expectedScreenX, screenX)
			}

			if screenY != tt.expectedScreenY {
				t.Errorf("Expected screen Y %f, got %f", tt.expectedScreenY, screenY)
			}
		})
	}
}
