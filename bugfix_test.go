package main

import (
	"testing"

	"github.com/nathan/verdant-thane/config"
)

// TestRetryBattleResetsKills verifies that retrying a battle resets kills to battle start
func TestRetryBattleResetsKills(t *testing.T) {
	// Create entity manager
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, nil, factionSprites)

	// Set initial stats (as if player completed previous battle with some kills)
	em.SetPlayerStats(100, 10, 2) // score=100, kills=10, deaths=2

	// Simulate starting a new battle (this saves the battle start stats)
	battleStartScore, battleStartKills, _ := em.GetPlayerStats()

	// Player gets some kills during this battle
	em.AddKill()
	em.AddKill()
	em.AddKill()

	// Verify kills increased
	_, kills, _ := em.GetPlayerStats()
	if kills != 13 {
		t.Errorf("Expected 13 kills after adding 3, got %d", kills)
	}

	// Player dies and retries the battle
	// This should restore score and kills to battle start, but keep current deaths
	_, _, currentDeaths := em.GetPlayerStats()
	em.SetPlayerStats(battleStartScore, battleStartKills, currentDeaths)

	// Verify kills were reset to battle start, but deaths stayed the same
	score, kills, deaths := em.GetPlayerStats()
	if score != 100 {
		t.Errorf("Expected score=100, got %d", score)
	}
	if kills != 10 {
		t.Errorf("Expected kills to reset to 10, got %d", kills)
	}
	if deaths != 2 {
		t.Errorf("Expected deaths to stay at 2, got %d", deaths)
	}
}

// TestDeathCounterIncrements verifies death counter increments when player dies
func TestDeathCounterIncrements(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, nil, factionSprites)
	em.InitializeFactions()

	// Spawn player and friendly ship
	playerShip := em.SpawnShip(config.ClassFighter, 0, 100, 100)
	playerShip.SetPlayerControlled(true)
	em.SetPlayerShip(playerShip.GetID())

	friendlyShip := em.SpawnShip(config.ClassFighter, 0, 200, 200)

	// Initial deaths should be 0
	_, _, deaths := em.GetPlayerStats()
	if deaths != 0 {
		t.Errorf("Expected initial deaths=0, got %d", deaths)
	}

	// Kill player ship
	playerShip.TakeDamage(100, -1, em)

	// Process the death
	em.UpdateAll()

	// Verify death was counted
	_, _, deaths = em.GetPlayerStats()
	if deaths != 1 {
		t.Errorf("Expected deaths=1 after player death, got %d", deaths)
	}

	// Verify player is spectating the friendly ship
	if !em.IsSpectating() {
		t.Error("Player should be in spectate mode")
	}

	spectatedShip := em.GetSpectatedShip()
	if spectatedShip == nil || spectatedShip.GetID() != friendlyShip.GetID() {
		t.Error("Player should be spectating the friendly ship")
	}
}

// TestSpectateTargetSwitchDoesNotIncrementDeaths verifies that when spectated ship dies,
// finding a new spectate target doesn't increment death counter
func TestSpectateTargetSwitchDoesNotIncrementDeaths(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, nil, factionSprites)
	em.InitializeFactions()

	// Spawn player and two friendly ships
	playerShip := em.SpawnShip(config.ClassFighter, 0, 100, 100)
	playerShip.SetPlayerControlled(true)
	em.SetPlayerShip(playerShip.GetID())

	_ = em.SpawnShip(config.ClassFighter, 0, 200, 200)
	_ = em.SpawnShip(config.ClassFighter, 0, 300, 300)

	// Kill player ship
	playerShip.TakeDamage(100, -1, em)
	em.UpdateAll()

	// Verify death counter is 1
	_, _, deaths := em.GetPlayerStats()
	if deaths != 1 {
		t.Errorf("Expected deaths=1 after player death, got %d", deaths)
	}

	// Player is now spectating friendlyShip1 (or friendlyShip2, doesn't matter)
	spectatedID := em.GetSpectatedShip().GetID()

	// Kill the spectated ship
	spectatedShip := em.ships[spectatedID]
	spectatedShip.TakeDamage(100, -1, em)

	// Update spectate mode to find new target
	em.UpdateSpectateMode()

	// Verify death counter is STILL 1 (not incremented to 2)
	_, _, deaths = em.GetPlayerStats()
	if deaths != 1 {
		t.Errorf("Expected deaths to stay at 1 when spectated ship dies, got %d", deaths)
	}

	// Verify player is now spectating the other friendly ship
	newSpectatedShip := em.GetSpectatedShip()
	if newSpectatedShip == nil {
		t.Error("Player should be spectating another friendly ship")
	}
	if newSpectatedShip.GetID() == spectatedID {
		t.Error("Player should have switched to a different spectate target")
	}
}
