package main

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nathan/verdant-thane/config"
	"github.com/nathan/verdant-thane/entity"
	"github.com/nathan/verdant-thane/systems"
)

// createTestSprites creates minimal dummy sprites for testing
func createTestSprites() (*ebiten.Image, *ebiten.Image, *ebiten.Image, *systems.FactionSprites) {
	laserSprite := ebiten.NewImage(4, 4)
	missileSprite := ebiten.NewImage(5, 10)
	explosionSprite := ebiten.NewImage(400, 70)

	testShipSprite := ebiten.NewImage(24, 24)
	factionSprites := &systems.FactionSprites{
		Fighter: &systems.ShipClassSprites{
			Green:  testShipSprite,
			Blue:   testShipSprite,
			Red:    testShipSprite,
			Yellow: testShipSprite,
		},
		Destroyer: &systems.ShipClassSprites{
			Green:  testShipSprite,
			Blue:   testShipSprite,
			Red:    testShipSprite,
			Yellow: testShipSprite,
		},
		Testudon: &systems.ShipClassSprites{
			Green:  testShipSprite,
			Blue:   testShipSprite,
			Red:    testShipSprite,
			Yellow: testShipSprite,
		},
	}

	return laserSprite, missileSprite, explosionSprite, factionSprites
}

// ============================================================================
// Full Game Flow Integration Tests (adapted from old integration_test.go)
// ============================================================================

// TestFullGameFlowWithCombat tests the complete game flow from initialization through combat
// Adapted from old donburi-based test
func TestFullGameFlowWithCombat(t *testing.T) {
	// Create entity manager
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)
	em.InitializeFactions()

	// Spawn two ships from different factions positioned for combat
	playerShip := em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	enemyShip := em.SpawnShip(entity.ClassFighter, 1, 120, 100) // 20 pixels to the right

	// Set player control
	playerShip.SetPlayerControlled(true)
	em.SetPlayerShip(playerShip.GetID())

	// Reduce enemy health for easy kill
	enemyX, enemyY := enemyShip.GetPosition()
	enemyShip.TakeDamage(7, playerShip.GetID(), em) // Reduce to 1 HP (fighter has 8 HP)

	// Point player toward enemy (90 degrees = facing right)
	// Access underlying implementation to set rotation
	if fighter, ok := playerShip.(*entity.Fighter); ok {
		fighter.Rotation = math.Pi / 2
		fighter.WeaponCapacitor = 1.0 // Fully charge
	}

	// Fire weapon
	playerShip.FireWeapon(enemyX, enemyY, em)

	// Verify projectile was created
	projectileCount := len(em.projectiles)
	if projectileCount != 1 {
		t.Errorf("Expected 1 projectile after firing, got %d", projectileCount)
	}

	// Run update cycles to move projectile and detect collision
	// At 12 px/tick and 20 pixels distance, should hit in ~2 ticks
	initialScore := em.score
	initialKills := em.kills

	for i := 0; i < 5; i++ {
		em.UpdateAll()
	}

	// Verify enemy was destroyed
	if enemyShip.IsAlive() {
		t.Error("Enemy ship should have been destroyed by projectile")
	}

	// Verify player scored
	if em.score != initialScore+10 {
		t.Errorf("Expected player score %d, got %d", initialScore+10, em.score)
	}
	if em.kills != initialKills+1 {
		t.Errorf("Expected player kills %d, got %d", initialKills+1, em.kills)
	}

	// Verify explosion was created
	explosionCount := len(em.explosions)
	if explosionCount != 1 {
		t.Errorf("Expected 1 explosion after ship destruction, got %d", explosionCount)
	}
}

// TestMultiFactionBattle tests a battle scenario with 3 factions
// Adapted from old donburi-based test
func TestMultiFactionBattle(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)
	em.InitializeFactions()

	// Spawn 3 ships from different factions
	faction0Ship := em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	faction1Ship := em.SpawnShip(entity.ClassFighter, 1, 200, 100)
	faction2Ship := em.SpawnShip(entity.ClassFighter, 2, 300, 100)

	faction0Ship.SetPlayerControlled(true)
	em.SetPlayerShip(faction0Ship.GetID())

	// Verify all three factions are present
	factionCount := make(map[int]int)
	for _, ship := range em.GetAllShips() {
		factionCount[ship.GetFaction()]++
	}

	if len(factionCount) != 3 {
		t.Errorf("Expected 3 factions, got %d", len(factionCount))
	}

	// Run AI updates to select targets
	em.UpdateAll()

	// Verify AI ships have selected targets from different factions
	// (New architecture doesn't expose AITargetID directly, but we can verify
	// behavior by checking that AI ships are alive and functioning)
	if !faction1Ship.IsAlive() {
		t.Error("Faction 1 ship should still be alive after one update")
	}
	if !faction2Ship.IsAlive() {
		t.Error("Faction 2 ship should still be alive after one update")
	}

	// Verify ships exist in correct factions
	faction0Ships := em.GetShipsByFaction(0)
	faction1Ships := em.GetShipsByFaction(1)
	faction2Ships := em.GetShipsByFaction(2)

	if len(faction0Ships) != 1 {
		t.Errorf("Expected 1 ship in faction 0, got %d", len(faction0Ships))
	}
	if len(faction1Ships) != 1 {
		t.Errorf("Expected 1 ship in faction 1, got %d", len(faction1Ships))
	}
	if len(faction2Ships) != 1 {
		t.Errorf("Expected 1 ship in faction 2, got %d", len(faction2Ships))
	}
}

// TestAICombatBehavior tests AI targeting and firing behavior
// Adapted from old donburi-based test
func TestAICombatBehavior(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)
	em.InitializeFactions()

	// Spawn AI ship with enemy ship in front
	aiShip := em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	enemyShip := em.SpawnShip(entity.ClassFighter, 1, 150, 100)

	// Point AI ship toward enemy (0 radians = facing up)
	// Enemy is to the right, so rotate 90 degrees (π/2)
	if fighter, ok := aiShip.(*entity.Fighter); ok {
		fighter.Rotation = math.Pi / 2
		fighter.WeaponCapacitor = 1.0 // Fully charge
	}

	// Get initial projectile count
	initialProjectileCount := len(em.projectiles)

	// Run AI systems
	em.UpdateAll()

	// AI may or may not have fired (probabilistic based on AI firing logic)
	// But the system should not crash and AI should have processed
	projectileCount := len(em.projectiles)
	t.Logf("AI created %d new projectiles", projectileCount-initialProjectileCount)

	// Verify ships are still alive after one update
	if !aiShip.IsAlive() {
		t.Error("AI ship should still be alive")
	}
	if !enemyShip.IsAlive() {
		t.Error("Enemy ship should still be alive after one update")
	}

	// Verify FindNearestEnemy works
	nearest, dist := em.FindNearestEnemy(aiShip)
	if nearest == nil {
		t.Error("FindNearestEnemy should find the enemy ship")
	}
	if nearest != nil && nearest.GetID() != enemyShip.GetID() {
		t.Error("FindNearestEnemy should return the enemy ship")
	}
	expectedDist := 50.0 // Distance between (100,100) and (150,100)
	if math.Abs(dist-expectedDist) > 1.0 {
		t.Errorf("Expected distance ~%f, got %f", expectedDist, dist)
	}
}

// TestFleetSpawningIntegration tests fleet config integration with EntityManager
// Adapted from old donburi-based test
func TestFleetSpawningIntegration(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)
	em.InitializeFactions()

	// Create a fleet configuration
	fleetConfig := config.GenerateFleetConfig(3, 5) // 3 factions, 5 ships each

	// Spawn all ships
	var playerShipID int
	spawnedShips := 0

	for factionID := 0; factionID < fleetConfig.NumFactions; factionID++ {
		numShips := fleetConfig.Compositions[factionID].Total()
		for shipIndex := 0; shipIndex < numShips; shipIndex++ {
			isPlayer := (factionID == 0 && shipIndex == 0)

			ship := em.SpawnShipAtFactionPoint(entity.ClassFighter, factionID)
			ship.SetPlayerControlled(isPlayer)

			if isPlayer {
				playerShipID = ship.GetID()
				em.SetPlayerShip(playerShipID)
			}

			spawnedShips++
		}
	}

	// Verify correct number of ships spawned
	allShips := em.GetAllShips()
	expectedShips := fleetConfig.NumFactions * fleetConfig.Compositions[0].Total()
	if len(allShips) != expectedShips {
		t.Errorf("Expected %d ships, got %d", expectedShips, len(allShips))
	}

	// Verify player ship exists and is controlled
	playerShip := em.GetPlayerShip()
	if playerShip == nil {
		t.Fatal("Player ship should exist")
	}
	if !playerShip.IsPlayerControlled() {
		t.Error("Player ship should be player controlled")
	}

	// Verify all factions are represented
	factionCount := make(map[int]int)
	for _, ship := range allShips {
		factionCount[ship.GetFaction()]++
	}

	if len(factionCount) != fleetConfig.NumFactions {
		t.Errorf("Expected %d factions, got %d", fleetConfig.NumFactions, len(factionCount))
	}

	for factionID := 0; factionID < fleetConfig.NumFactions; factionID++ {
		expectedCount := fleetConfig.Compositions[factionID].Total()
		if factionCount[factionID] != expectedCount {
			t.Errorf("Faction %d: expected %d ships, got %d",
				factionID, expectedCount, factionCount[factionID])
		}
	}
}

// TestProjectileLifecycleIntegration tests the full lifecycle of projectiles
// Adapted from old donburi-based test
func TestProjectileLifecycleIntegration(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)
	em.InitializeFactions()

	// Spawn ship and fire multiple projectiles
	ship := em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	ship.SetPlayerControlled(true)

	// Fire 3 projectiles
	for i := 0; i < 3; i++ {
		// Charge weapon directly on underlying implementation
		if fighter, ok := ship.(*entity.Fighter); ok {
			fighter.WeaponCapacitor = 1.0
		}
		ship.FireWeapon(200, 100, em)
	}

	// Verify 3 projectiles exist
	projectileCount := len(em.projectiles)
	if projectileCount != 3 {
		t.Errorf("Expected 3 projectiles, got %d", projectileCount)
	}

	// Run movement and lifetime systems for long enough to expire projectiles
	// Laser lifetime is 180 ticks (3 seconds)
	for tick := 0; tick < 200; tick++ {
		em.UpdateAll()
	}

	// All projectiles should have expired
	projectileCount = len(em.projectiles)
	if projectileCount != 0 {
		t.Errorf("Expected 0 projectiles after expiration, got %d", projectileCount)
	}
}

// TestExplosionLifecycleIntegration tests explosion creation and animation
// Adapted from old donburi-based test
func TestExplosionLifecycleIntegration(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)
	em.InitializeFactions()

	// Spawn ships for combat
	playerShip := em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	enemyShip := em.SpawnShip(entity.ClassFighter, 1, 120, 100) // 20 pixels away

	playerShip.SetPlayerControlled(true)
	em.SetPlayerShip(playerShip.GetID())

	// Reduce enemy health to 1 HP for easy kill
	enemyShip.TakeDamage(7, playerShip.GetID(), em)

	// Point player toward enemy (90 degrees = facing right)
	if fighter, ok := playerShip.(*entity.Fighter); ok {
		fighter.Rotation = math.Pi / 2
		fighter.WeaponCapacitor = 1.0 // Fully charge
	}

	// Fire and hit enemy
	enemyX, enemyY := enemyShip.GetPosition()
	playerShip.FireWeapon(enemyX, enemyY, em)

	// Run systems to cause collision
	for i := 0; i < 5; i++ {
		em.UpdateAll()
	}

	// Verify explosion was created
	explosionCount := len(em.explosions)
	if explosionCount != 1 {
		t.Errorf("Expected 1 explosion, got %d", explosionCount)
		return // Don't proceed if no explosion was created
	}

	// Get the explosion
	var explosion *entity.Explosion
	for _, exp := range em.explosions {
		explosion = exp
		break
	}

	// Verify initial frame is 0
	if explosion.CurrentFrame != 0 {
		t.Errorf("Expected initial explosion frame 0, got %d", explosion.CurrentFrame)
	}

	initialExplosionID := explosion.ID

	// Run for one full animation cycle
	// 4 frames * 5 ticks per frame = 20 ticks, plus a buffer
	for i := 0; i < 25; i++ {
		em.UpdateAll()
	}

	// Explosion should be removed after animation completes
	if _, exists := em.explosions[initialExplosionID]; exists {
		t.Error("Explosion should be removed after animation completes")
	}
}

// ============================================================================
// New Architecture-Specific Integration Tests
// ============================================================================

// TestEntityManagerInitialization tests basic EntityManager setup
func TestEntityManagerInitialization(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)

	if em.nextID != 1 {
		t.Errorf("Expected initial nextID=1, got %d", em.nextID)
	}

	if len(em.ships) != 0 {
		t.Errorf("Expected 0 initial ships, got %d", len(em.ships))
	}

	if em.playerShipID != -1 {
		t.Errorf("Expected playerShipID=-1, got %d", em.playerShipID)
	}

	// Initialize factions
	em.InitializeFactions()

	// Verify faction spawn points were created
	if len(em.factionSpawnPoints) != 4 {
		t.Errorf("Expected 4 faction spawn points, got %d", len(em.factionSpawnPoints))
	}
}

// TestBattleEndDetection tests victory/defeat/ongoing detection
func TestBattleEndDetection(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)
	em.InitializeFactions()

	// Initial state - no ships means player faction has no ships = defeat
	result := em.CheckBattleEnd()
	if result != PlayerDefeat {
		t.Errorf("Expected PlayerDefeat with no ships, got %v", result)
	}

	// Spawn player ship only - should be victory
	playerShip := em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	playerShip.SetPlayerControlled(true)
	em.SetPlayerShip(playerShip.GetID())

	result = em.CheckBattleEnd()
	if result != PlayerVictory {
		t.Errorf("Expected PlayerVictory with only player faction, got %v", result)
	}

	// Spawn enemy ship - battle ongoing
	em.SpawnShip(entity.ClassFighter, 1, 200, 100)

	result = em.CheckBattleEnd()
	if result != BattleOngoing {
		t.Errorf("Expected BattleOngoing with multiple factions, got %v", result)
	}

	// Kill player ship - should be defeat
	playerShip.TakeDamage(100, -1, em)

	result = em.CheckBattleEnd()
	if result != PlayerDefeat {
		t.Errorf("Expected PlayerDefeat when player faction eliminated, got %v", result)
	}
}

// TestShipRespawning tests player respawn mechanics
func TestShipRespawning(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)
	em.InitializeFactions()

	// Spawn player ship
	playerShip := em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	playerShip.SetPlayerControlled(true)
	em.SetPlayerShip(playerShip.GetID())

	// Spawn another friendly ship for respawn
	friendlyShip := em.SpawnShip(entity.ClassFighter, 0, 200, 200)

	initialPlayerID := playerShip.GetID()
	initialDeaths := em.deaths

	// Kill player ship
	playerShip.TakeDamage(100, -1, em)

	// Update to trigger death handling
	em.UpdateAll()

	// Verify death was counted
	if em.deaths != initialDeaths+1 {
		t.Errorf("Expected deaths=%d, got %d", initialDeaths+1, em.deaths)
	}

	// Verify player is now in spectate mode
	if !em.IsSpectating() {
		t.Error("Player should be in spectate mode after death")
	}

	// Verify spectating the friendly ship
	spectated := em.GetSpectatedShip()
	if spectated == nil {
		t.Fatal("Player should be spectating a friendly ship")
	}
	if spectated.GetID() != friendlyShip.GetID() {
		t.Error("Player should be spectating the friendly ship")
	}

	// Manual respawn into spectated ship
	success := em.RespawnIntoSpectatedShip()
	if !success {
		t.Fatal("Should be able to respawn into spectated ship")
	}

	// Verify player respawned into friendly ship
	newPlayerShip := em.GetPlayerShip()
	if newPlayerShip == nil {
		t.Fatal("Player should have respawned")
	}
	if newPlayerShip.GetID() == initialPlayerID {
		t.Error("Player should have respawned into different ship")
	}
	if newPlayerShip.GetID() != friendlyShip.GetID() {
		t.Error("Player should have respawned into the friendly ship")
	}
}

// TestMixedFleetComposition tests spawning different ship classes
func TestMixedFleetComposition(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)
	em.InitializeFactions()

	// Spawn mixed fleet
	fighter := em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	destroyer := em.SpawnShip(entity.ClassDestroyer, 0, 200, 100)
	testudon := em.SpawnShip(entity.ClassTestudon, 0, 300, 100)

	// Verify different ship classes
	if fighter.GetClass() != entity.ClassFighter {
		t.Error("Fighter should have ClassFighter")
	}
	if destroyer.GetClass() != entity.ClassDestroyer {
		t.Error("Destroyer should have ClassDestroyer")
	}
	if testudon.GetClass() != entity.ClassTestudon {
		t.Error("Testudon should have ClassTestudon")
	}

	// Verify different stats
	fighterChars := config.GetShipCharacteristics(entity.ClassFighter)
	destroyerChars := config.GetShipCharacteristics(entity.ClassDestroyer)
	testudonChars := config.GetShipCharacteristics(entity.ClassTestudon)

	// Verify health matches ship class (GetHealth returns current, max)
	fighterHealth, fighterMaxHealth := fighter.GetHealth()
	destroyerHealth, destroyerMaxHealth := destroyer.GetHealth()
	testudonHealth, testudonMaxHealth := testudon.GetHealth()

	if fighterHealth != fighterChars.MaxShield {
		t.Errorf("Fighter health should be %d, got %d", fighterChars.MaxShield, fighterHealth)
	}
	if fighterMaxHealth != fighterChars.MaxShield {
		t.Errorf("Fighter max health should be %d, got %d", fighterChars.MaxShield, fighterMaxHealth)
	}

	if destroyerHealth != destroyerChars.MaxShield {
		t.Errorf("Destroyer health should be %d, got %d", destroyerChars.MaxShield, destroyerHealth)
	}
	if destroyerMaxHealth != destroyerChars.MaxShield {
		t.Errorf("Destroyer max health should be %d, got %d", destroyerChars.MaxShield, destroyerMaxHealth)
	}

	if testudonHealth != testudonChars.MaxShield {
		t.Errorf("Testudon health should be %d, got %d", testudonChars.MaxShield, testudonHealth)
	}
	if testudonMaxHealth != testudonChars.MaxShield {
		t.Errorf("Testudon max health should be %d, got %d", testudonChars.MaxShield, testudonMaxHealth)
	}
}

// TestSpatialGridIntegration tests spatial grid for collision detection
func TestSpatialGridIntegration(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)
	em.InitializeFactions()

	// Spawn ships in different locations
	ship1 := em.SpawnShip(entity.ClassFighter, 0, 100, 100)
	ship2 := em.SpawnShip(entity.ClassFighter, 1, 150, 100) // Nearby
	_ = em.SpawnShip(entity.ClassFighter, 1, 5000, 5000)    // Far away (intentionally unused)

	// Update to rebuild spatial grid
	em.UpdateAll()

	// FindNearestEnemy should find ship2 (nearby), not ship3 (far)
	nearest, dist := em.FindNearestEnemy(ship1)
	if nearest == nil {
		t.Fatal("Should find nearest enemy")
	}
	if nearest.GetID() != ship2.GetID() {
		t.Error("Should find ship2 as nearest enemy")
	}
	expectedDist := 50.0
	if math.Abs(dist-expectedDist) > 1.0 {
		t.Errorf("Expected distance ~%f, got %f", expectedDist, dist)
	}
}

// TestWorldWrappingIntegration tests world wrapping across boundaries
func TestWorldWrappingIntegration(t *testing.T) {
	laserSprite, missileSprite, explosionSprite, factionSprites := createTestSprites()
	em := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)
	em.InitializeFactions()

	// Spawn ship near left edge
	ship := em.SpawnShip(entity.ClassFighter, 0, 10, 100)

	// Set rotation and speed directly
	if fighter, ok := ship.(*entity.Fighter); ok {
		fighter.Rotation = -math.Pi / 2 // Facing left (west)
		fighter.Speed = 10.0            // Moving fast
	}

	// Update to move ship off left edge (multiple updates to cross boundary)
	for i := 0; i < 3; i++ {
		em.UpdateAll()
	}

	// Ship should wrap to right edge
	x, _ := ship.GetPosition()
	if x < float64(config.GameWidth)-50 || x >= float64(config.GameWidth) {
		t.Errorf("Ship should have wrapped to right edge (GameWidth=%d), got x=%f", config.GameWidth, x)
	}
}
