package main

import (
	"testing"

	"github.com/nathan/verdant-thane/entity"
)

// Test EntityManager Creation
func TestNewEntityManager(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	if em == nil {
		t.Fatal("NewEntityManager returned nil")
	}

	if em.nextID != 1 {
		t.Errorf("Expected nextID to start at 1, got %d", em.nextID)
	}

	if em.playerShipID != -1 {
		t.Errorf("Expected playerShipID to start at -1, got %d", em.playerShipID)
	}

	if em.isSpectating {
		t.Error("Should not start in spectate mode")
	}

	score, kills, deaths := em.GetPlayerStats()
	if score != 0 || kills != 0 || deaths != 0 {
		t.Errorf("Expected stats (0, 0, 0), got (%d, %d, %d)", score, kills, deaths)
	}
}

// Test Ship Spawning
func TestEntityManagerSpawnShip(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	// Spawn a fighter
	ship := em.SpawnShip(entity.ClassFighter, 0, 100, 200)

	if ship == nil {
		t.Fatal("SpawnShip should return a ship")
	}

	shipID := ship.GetID()
	if shipID != 1 {
		t.Errorf("Expected first ship ID to be 1, got %d", shipID)
	}

	// Verify ship can be retrieved by ID
	retrievedShip := em.GetShip(shipID)
	if retrievedShip == nil {
		t.Fatal("Spawned ship should be retrievable")
	}

	if ship.GetClass() != entity.ClassFighter {
		t.Errorf("Expected Fighter class, got %v", ship.GetClass())
	}

	if ship.GetFaction() != 0 {
		t.Errorf("Expected faction 0, got %d", ship.GetFaction())
	}

	x, y := ship.GetPosition()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%f, %f)", x, y)
	}

	// Next ID should increment
	if em.nextID != 2 {
		t.Errorf("Expected nextID to be 2, got %d", em.nextID)
	}
}

// Test Multiple Ship Spawning
func TestEntityManagerSpawnMultipleShips(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	ship1 := em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	ship2 := em.SpawnShip(entity.ClassDestroyer, 1, 200, 200)
	ship3 := em.SpawnShip(entity.ClassTestudon, 2, 300, 300)

	id1 := ship1.GetID()
	id2 := ship2.GetID()
	id3 := ship3.GetID()

	if id1 != 1 || id2 != 2 || id3 != 3 {
		t.Errorf("Expected IDs 1, 2, 3, got %d, %d, %d", id1, id2, id3)
	}

	// Verify all ships exist
	if em.GetShip(id1) == nil || em.GetShip(id2) == nil || em.GetShip(id3) == nil {
		t.Error("All spawned ships should exist")
	}

	// Verify ship types
	if ship1.GetClass() != entity.ClassFighter {
		t.Error("Ship 1 should be Fighter")
	}
	if ship2.GetClass() != entity.ClassDestroyer {
		t.Error("Ship 2 should be Destroyer")
	}
	if ship3.GetClass() != entity.ClassTestudon {
		t.Error("Ship 3 should be Testudon")
	}
}

// Test Projectile Spawning
func TestEntityManagerSpawnProjectile(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	cfg := entity.MainGunConfig{
		X:         100,
		Y:         200,
		VelocityX: 5,
		VelocityY: 3,
		OwnerID:   1,
		FactionID: 0,
	}

	em.SpawnProjectile(cfg)

	// Verify projectile was created
	if len(em.projectiles) != 1 {
		t.Fatalf("Expected 1 projectile, got %d", len(em.projectiles))
	}

	// Get the projectile (ID should be 1)
	proj := em.projectiles[1]
	if proj == nil {
		t.Fatal("Projectile with ID 1 should exist")
	}

	if proj.GetOwnerID() != 1 {
		t.Errorf("Expected owner ID 1, got %d", proj.GetOwnerID())
	}

	if proj.GetFaction() != 0 {
		t.Errorf("Expected faction 0, got %d", proj.GetFaction())
	}
}

// Test Missile Spawning
func TestEntityManagerSpawnMissile(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	cfg := entity.MissileConfig{
		X:         100,
		Y:         200,
		VelocityX: 5,
		VelocityY: 3,
		Rotation:  1.5,
		OwnerID:   1,
		FactionID: 0,
		TargetID:  2,
	}

	em.SpawnMissile(cfg)

	// Verify missile was created
	if len(em.projectiles) != 1 {
		t.Fatalf("Expected 1 missile, got %d", len(em.projectiles))
	}

	missile := em.projectiles[1]
	if missile.GetOwnerID() != 1 {
		t.Errorf("Expected owner ID 1, got %d", missile.GetOwnerID())
	}
}

// Test Explosion Spawning
func TestEntityManagerSpawnExplosion(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	em.SpawnExplosion(100, 200)

	if len(em.explosions) != 1 {
		t.Fatalf("Expected 1 explosion, got %d", len(em.explosions))
	}

	explosion := em.explosions[1]
	if explosion == nil {
		t.Fatal("Explosion should exist")
	}

	x, y := explosion.GetPosition()
	if x != 100 || y != 200 {
		t.Errorf("Expected explosion at (100, 200), got (%f, %f)", x, y)
	}
}

// Test Player Ship Assignment
func TestEntityManagerPlayerShip(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	ship := em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	shipID := ship.GetID()
	em.SetPlayerShip(shipID)

	if em.playerShipID != shipID {
		t.Errorf("Expected player ship ID %d, got %d", shipID, em.playerShipID)
	}

	playerShip := em.GetPlayerShip()
	if playerShip == nil {
		t.Fatal("Player ship should exist")
	}

	if !playerShip.IsPlayerControlled() {
		t.Error("Player ship should be marked as player-controlled")
	}

	if playerShip.GetID() != shipID {
		t.Errorf("Player ship ID mismatch: expected %d, got %d", shipID, playerShip.GetID())
	}
}

// Test Player Stats Tracking
func TestEntityManagerPlayerStats(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	// Add some kills and score
	em.AddKill()
	em.AddKill()
	em.AddScore(25)

	score, kills, deaths := em.GetPlayerStats()
	if kills != 2 {
		t.Errorf("Expected 2 kills, got %d", kills)
	}
	if score != 25 {
		t.Errorf("Expected score 25, got %d", score)
	}
	if deaths != 0 {
		t.Errorf("Expected 0 deaths, got %d", deaths)
	}
}

// Test GetAllShips
func TestEntityManagerGetAllShips(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	em.SpawnShip(entity.ClassFighter, 1, 200, 200)
	em.SpawnShip(entity.ClassDestroyer, 0, 300, 300)

	ships := em.GetAllShips()
	if len(ships) != 3 {
		t.Errorf("Expected 3 ships, got %d", len(ships))
	}
}

// Test GetShipsByFaction
func TestEntityManagerGetShipsByFaction(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	em.SpawnShip(entity.ClassFighter, 0, 150, 150)
	em.SpawnShip(entity.ClassFighter, 1, 200, 200)
	em.SpawnShip(entity.ClassDestroyer, 1, 250, 250)

	faction0Ships := em.GetShipsByFaction(0)
	faction1Ships := em.GetShipsByFaction(1)

	if len(faction0Ships) != 2 {
		t.Errorf("Expected 2 ships in faction 0, got %d", len(faction0Ships))
	}

	if len(faction1Ships) != 2 {
		t.Errorf("Expected 2 ships in faction 1, got %d", len(faction1Ships))
	}

	// Verify faction IDs
	for _, ship := range faction0Ships {
		if ship.GetFaction() != 0 {
			t.Error("Faction 0 ships should have faction ID 0")
		}
	}

	for _, ship := range faction1Ships {
		if ship.GetFaction() != 1 {
			t.Error("Faction 1 ships should have faction ID 1")
		}
	}
}

// Test FindNearestEnemy
func TestEntityManagerFindNearestEnemy(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	// Spawn test ships
	ship1 := em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	em.SpawnShip(entity.ClassFighter, 1, 150, 100) // Close enemy (distance 50)
	em.SpawnShip(entity.ClassFighter, 1, 500, 500) // Far enemy
	em.SpawnShip(entity.ClassFighter, 0, 110, 100) // Friendly

	nearest, distance := em.FindNearestEnemy(ship1)

	if nearest == nil {
		t.Fatal("Should find nearest enemy")
	}

	// Should find the close enemy at (150, 100)
	x, y := nearest.GetPosition()
	if x != 150 || y != 100 {
		t.Errorf("Expected nearest enemy at (150, 100), got (%f, %f)", x, y)
	}

	if distance != 50 {
		t.Errorf("Expected distance 50, got %f", distance)
	}

	// Should not find friendly
	if nearest.GetFaction() == ship1.GetFaction() {
		t.Error("Nearest enemy should not be from same faction")
	}
}

// Test Clear Preserves Player Stats
func TestEntityManagerClearPreservesStats(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	// Set up some state
	em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	em.SpawnProjectile(entity.MainGunConfig{X: 50, Y: 50, OwnerID: 1, FactionID: 0})
	em.AddKill()
	em.AddKill()
	em.AddScore(50)

	// Clear
	em.Clear()

	// Entities should be cleared
	if len(em.ships) != 0 {
		t.Errorf("Expected 0 ships after clear, got %d", len(em.ships))
	}
	if len(em.projectiles) != 0 {
		t.Errorf("Expected 0 projectiles after clear, got %d", len(em.projectiles))
	}

	// Player stats should be preserved
	score, kills, _ := em.GetPlayerStats()
	if kills != 2 {
		t.Errorf("Expected kills to be preserved (2), got %d", kills)
	}
	if score != 50 {
		t.Errorf("Expected score to be preserved (50), got %d", score)
	}

	// Player ship ID should be reset
	if em.playerShipID != -1 {
		t.Errorf("Expected player ship ID to be reset to -1, got %d", em.playerShipID)
	}

	// Next ID should reset
	if em.nextID != 1 {
		t.Errorf("Expected nextID to reset to 1, got %d", em.nextID)
	}
}

// Test Battle End Detection - Player Victory
func TestEntityManagerBattleEndPlayerVictory(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	// Player faction (0) has ships, no other factions do
	em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	em.SpawnShip(entity.ClassFighter, 0, 200, 200)

	result := em.CheckBattleEnd()
	if result != PlayerVictory {
		t.Errorf("Expected PlayerVictory, got %v", result)
	}
}

// Test Battle End Detection - Player Defeat
func TestEntityManagerBattleEndPlayerDefeat(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	// Enemy faction has ships, player faction (0) does not
	em.SpawnShip(entity.ClassFighter, 1, 100, 100)
	em.SpawnShip(entity.ClassFighter, 1, 200, 200)

	result := em.CheckBattleEnd()
	if result != PlayerDefeat {
		t.Errorf("Expected PlayerDefeat, got %v", result)
	}
}

// Test Battle End Detection - Ongoing
func TestEntityManagerBattleEndOngoing(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	// Multiple factions with ships
	em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	em.SpawnShip(entity.ClassFighter, 1, 200, 200)

	result := em.CheckBattleEnd()
	if result != BattleOngoing {
		t.Errorf("Expected BattleOngoing, got %v", result)
	}
}

// Test UpdateAll Removes Dead Entities
func TestEntityManagerUpdateRemovesDeadEntities(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	// Spawn entities
	ship := em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	shipID := ship.GetID()
	em.SpawnProjectile(entity.MainGunConfig{X: 50, Y: 50, OwnerID: 1, FactionID: 0})
	em.SpawnExplosion(75, 75)

	// Kill the entities
	ship.TakeDamage(1000, -1, em) // Kill with massive damage

	for id := range em.projectiles {
		delete(em.projectiles, id)
	}

	for _, explosion := range em.explosions {
		// Set explosion to last frame to make it die
		explosion.CurrentFrame = 100
		explosion.Alive = false
	}

	// Update to clean up
	em.UpdateAll()

	// Dead ship should be removed
	if em.GetShip(shipID) != nil {
		t.Error("Dead ship should be removed")
	}

	// Dead explosions should be removed (after animation completes)
	if len(em.explosions) != 0 {
		t.Errorf("Expected dead explosions to be removed, got %d", len(em.explosions))
	}
}

// Test ID Generation Doesn't Reuse IDs
func TestEntityManagerIDGeneration(t *testing.T) {
	em := NewEntityManager(nil, nil, nil, nil)

	ids := make(map[int]bool)

	// Spawn many entities
	for i := 0; i < 100; i++ {
		ship := em.SpawnShip(entity.ClassFighter, 0, float64(i*10), 100)
		id := ship.GetID()
		if ids[id] {
			t.Fatalf("ID %d was reused!", id)
		}
		ids[id] = true
	}

	// All IDs should be unique
	if len(ids) != 100 {
		t.Errorf("Expected 100 unique IDs, got %d", len(ids))
	}
}
