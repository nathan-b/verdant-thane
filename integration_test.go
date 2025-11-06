package main

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
	"github.com/nathan/verdant-thane/systems"
)

// createTestFactionSprites creates dummy faction sprites for testing
func createTestFactionSprites() *systems.FactionSprites {
	testSprite := ebiten.NewImage(24, 24)
	return &systems.FactionSprites{
		Fighter: &systems.ShipClassSprites{
			Green:  testSprite,
			Blue:   testSprite,
			Red:    testSprite,
			Yellow: testSprite,
		},
		Destroyer: &systems.ShipClassSprites{
			Green:  testSprite,
			Blue:   testSprite,
			Red:    testSprite,
			Yellow: testSprite,
		},
		Testudon: &systems.ShipClassSprites{
			Green:  testSprite,
			Blue:   testSprite,
			Red:    testSprite,
			Yellow: testSprite,
		},
	}
}

// TestFullGameFlowWithCombat tests the complete game flow from initialization through combat
func TestFullGameFlowWithCombat(t *testing.T) {
	// Create a minimal game setup
	world := donburi.NewWorld()
	systems.InitializeFactions(world)

	// Create dummy sprites
	laserSprite := ebiten.NewImage(4, 4)
	missileSprite := ebiten.NewImage(5, 10)
	_ = missileSprite // Not used in this test
	explosionSprite := ebiten.NewImage(400, 70)
	testFactionSprites := createTestFactionSprites()

	// Spawn two ships from different factions positioned for combat
	playerShip, err := systems.SpawnShip(world, systems.ShipConfig{
		Class:              components.Fighter,
		FactionID:          0,
		MaxSpeed:           6.0,
		Acceleration:       4.0 / 60.0,
		MaxHealth:          8,
		CapacitorRate:      1.0 / 36.0,
		FiringCone:         0.523, // ~30 degrees
		FactionSprites:     testFactionSprites,
		IsPlayerControlled: true,
	})
	if err != nil {
		t.Fatalf("Failed to spawn player ship: %v", err)
	}

	enemyShip, err := systems.SpawnShip(world, systems.ShipConfig{
		Class:              components.Fighter,
		FactionID:          1,
		MaxSpeed:           6.0,
		Acceleration:       4.0 / 60.0,
		MaxHealth:          1, // Low health for easy kill
		CapacitorRate:      1.0 / 36.0,
		FiringCone:         0.523,
		FactionSprites:     testFactionSprites,
		IsPlayerControlled: false,
	})
	if err != nil {
		t.Fatalf("Failed to spawn enemy ship: %v", err)
	}

	// Position enemy directly in front of player
	playerEntry := world.Entry(playerShip)
	enemyEntry := world.Entry(enemyShip)

	playerPos := components.Position.Get(playerEntry)
	playerPos.X = 100
	playerPos.Y = 100

	enemyPos := components.Position.Get(enemyEntry)
	enemyPos.X = 120 // 20 pixels away
	enemyPos.Y = 100

	// Point player toward enemy
	// In sprite coordinates: 0 = up, π/2 = right, π = down, 3π/2 = left
	playerRot := components.Rotation.Get(playerEntry)
	playerRot.Angle = math.Pi / 2 // 90 degrees = facing right toward enemy

	// Create player state for score tracking
	playerState := world.Create(components.PlayerState)
	playerStateEntry := world.Entry(playerState)
	components.PlayerState.SetValue(playerStateEntry, components.PlayerStateData{
		ControlledShip: playerShip,
		Score:          0,
		Kills:          0,
	})

	// Charge the weapon fully
	weapon := components.Weapon.Get(playerEntry)
	weapon.Capacitor = 1.0

	// Fire weapon toward enemy
	systems.FireWeapon(world, playerEntry, 110, 100, laserSprite)

	// Verify projectile was created
	projectileQuery := donburi.NewQuery(filter.Contains(components.IsProjectile))
	projectileCount := 0
	for range projectileQuery.Iter(world) {
		projectileCount++
	}
	if projectileCount != 1 {
		t.Errorf("Expected 1 projectile after firing, got %d", projectileCount)
	}

	// Run movement system to move projectile and check collisions
	// At 12 px/tick and 20 pixels distance, should hit in ~2 ticks
	for i := 0; i < 5; i++ {
		systems.UpdateMovement(world)
		systems.UpdateCollisions(world, explosionSprite)
	}

	// Verify enemy was destroyed
	if world.Valid(enemyShip) {
		t.Error("Enemy ship should have been destroyed by projectile")
	}

	// Verify player scored
	state := components.PlayerState.Get(playerStateEntry)
	if state.Score != 10 {
		t.Errorf("Expected player score 10, got %d", state.Score)
	}
	if state.Kills != 1 {
		t.Errorf("Expected player kills 1, got %d", state.Kills)
	}

	// Verify explosion was created
	explosionQuery := donburi.NewQuery(filter.Contains(components.IsExplosion))
	explosionCount := 0
	for range explosionQuery.Iter(world) {
		explosionCount++
	}
	if explosionCount != 1 {
		t.Errorf("Expected 1 explosion after ship destruction, got %d", explosionCount)
	}
}

// TestMultiFactionBattle tests a battle scenario with 3 factions
func TestMultiFactionBattle(t *testing.T) {
	world := donburi.NewWorld()
	systems.InitializeFactions(world)

	laserSprite := ebiten.NewImage(4, 4)
	missileSprite := ebiten.NewImage(5, 10)
	factionSprites := createTestFactionSprites()

	// Spawn 3 ships from different factions
	faction0Ship, _ := systems.SpawnShip(world, systems.ShipConfig{
		Class:              components.Fighter,
		FactionID:          0,
		MaxSpeed:           6.0,
		Acceleration:       4.0 / 60.0,
		MaxHealth:          8,
		CapacitorRate:      1.0 / 36.0,
		FiringCone:         0.523,
		FactionSprites:     factionSprites,
		IsPlayerControlled: true,
	})

	faction1Ship, _ := systems.SpawnShip(world, systems.ShipConfig{
		Class:              components.Fighter,
		FactionID:          1,
		MaxSpeed:           6.0,
		Acceleration:       4.0 / 60.0,
		MaxHealth:          8,
		CapacitorRate:      1.0 / 36.0,
		FiringCone:         0.523,
		FactionSprites:     factionSprites,
		IsPlayerControlled: false,
	})

	faction2Ship, _ := systems.SpawnShip(world, systems.ShipConfig{
		Class:              components.Fighter,
		FactionID:          2,
		MaxSpeed:           6.0,
		Acceleration:       4.0 / 60.0,
		MaxHealth:          8,
		CapacitorRate:      1.0 / 36.0,
		FiringCone:         0.523,
		FactionSprites:     factionSprites,
		IsPlayerControlled: false,
	})

	// Verify all three factions are present
	factionCount := make(map[int]int)
	shipQuery := donburi.NewQuery(filter.Contains(components.IsShip, components.Faction))
	for shipEntry := range shipQuery.Iter(world) {
		faction := components.Faction.Get(shipEntry)
		factionCount[faction.ID]++
	}

	if len(factionCount) != 3 {
		t.Errorf("Expected 3 factions, got %d", len(factionCount))
	}

	// Run AI targeting for faction 1 and 2 ships
	systems.UpdateAIMovement(world)

	// Verify AI ships have selected targets
	faction1Entry := world.Entry(faction1Ship)
	if faction1Entry.HasComponent(components.AITarget) {
		target := components.AITarget.Get(faction1Entry)
		// Target should be valid and from a different faction
		if world.Valid(target.TargetEntity) {
			targetEntry := world.Entry(target.TargetEntity)
			targetFaction := components.Faction.Get(targetEntry)
			if targetFaction.ID == 1 {
				t.Error("Faction 1 ship should not target its own faction")
			}
		}
	}

	faction2Entry := world.Entry(faction2Ship)
	if faction2Entry.HasComponent(components.AITarget) {
		target := components.AITarget.Get(faction2Entry)
		if world.Valid(target.TargetEntity) {
			targetEntry := world.Entry(target.TargetEntity)
			targetFaction := components.Faction.Get(targetEntry)
			if targetFaction.ID == 2 {
				t.Error("Faction 2 ship should not target its own faction")
			}
		}
	}

	// Run AI firing system
	systems.UpdateAIFiring(world, faction0Ship, laserSprite, missileSprite)

	// AI ships may have fired, check if projectiles exist
	projectileQuery := donburi.NewQuery(filter.Contains(components.IsProjectile))
	projectileCount := 0
	for range projectileQuery.Iter(world) {
		projectileCount++
	}

	// We don't know if AI fired (depends on targeting and weapon charge), but system should not crash
	t.Logf("AI created %d projectiles", projectileCount)
}

// TestAICombatBehavior tests AI targeting and firing behavior
func TestAICombatBehavior(t *testing.T) {
	world := donburi.NewWorld()
	systems.InitializeFactions(world)

	laserSprite := ebiten.NewImage(4, 4)
	missileSprite := ebiten.NewImage(5, 10)
	factionSprites := createTestFactionSprites()

	// Spawn AI ship with fully charged weapon
	aiShip, _ := systems.SpawnShip(world, systems.ShipConfig{
		Class:              components.Fighter,
		FactionID:          0,
		MaxSpeed:           6.0,
		Acceleration:       4.0 / 60.0,
		MaxHealth:          8,
		CapacitorRate:      1.0 / 36.0,
		FiringCone:         0.523,
		FactionSprites:     factionSprites,
		IsPlayerControlled: false,
	})

	// Spawn enemy ship directly in front
	enemyShip, _ := systems.SpawnShip(world, systems.ShipConfig{
		Class:              components.Fighter,
		FactionID:          1,
		MaxSpeed:           6.0,
		Acceleration:       4.0 / 60.0,
		MaxHealth:          8,
		CapacitorRate:      1.0 / 36.0,
		FiringCone:         0.523,
		FactionSprites:     factionSprites,
		IsPlayerControlled: false,
	})

	// Position ships
	aiEntry := world.Entry(aiShip)
	enemyEntry := world.Entry(enemyShip)

	aiPos := components.Position.Get(aiEntry)
	aiPos.X = 100
	aiPos.Y = 100

	enemyPos := components.Position.Get(enemyEntry)
	enemyPos.X = 150
	enemyPos.Y = 100

	// Point AI ship toward enemy
	aiRot := components.Rotation.Get(aiEntry)
	aiRot.Angle = 0

	// Fully charge AI weapon
	aiWeapon := components.Weapon.Get(aiEntry)
	aiWeapon.Capacitor = 1.0

	// Initialize AI target
	aiState := components.AIState.Get(aiEntry)
	aiState.RetargetTimer = 0

	// Run AI systems
	systems.UpdateAIMovement(world) // This should select enemy as target
	systems.UpdateAIFiring(world, aiShip, laserSprite, missileSprite)

	// Verify AI selected the enemy as target
	aiTarget := components.AITarget.Get(aiEntry)
	if !world.Valid(aiTarget.TargetEntity) {
		t.Error("AI should have selected a valid target")
	}
	if aiTarget.TargetEntity != enemyShip {
		t.Error("AI should have selected the enemy ship as target")
	}

	// AI may or may not have fired (50% chance based on AI firing logic)
	// But the system should not crash
	projectileQuery := donburi.NewQuery(filter.Contains(components.IsProjectile))
	projectileCount := 0
	for range projectileQuery.Iter(world) {
		projectileCount++
	}
	t.Logf("AI fired %d projectiles", projectileCount)
}

// TestFleetSpawningIntegration tests fleet config integration with game systems
func TestFleetSpawningIntegration(t *testing.T) {
	world := donburi.NewWorld()
	systems.InitializeFactions(world)

	factionSprites := createTestFactionSprites()

	// Create a fleet configuration
	fleetConfig := GenerateFleetConfig(3, 5) // 3 factions, 5 ships each

	// Spawn all ships
	var playerShip donburi.Entity
	for factionID := 0; factionID < fleetConfig.NumFactions; factionID++ {
		for shipIndex := 0; shipIndex < fleetConfig.ShipsPerFaction[factionID]; shipIndex++ {
			isPlayer := (factionID == 0 && shipIndex == 0)

			ship, err := systems.SpawnShip(world, systems.ShipConfig{
				Class:              components.Fighter,
				FactionID:          factionID,
				MaxSpeed:           6.0,
				Acceleration:       4.0 / 60.0,
				MaxHealth:          8,
				CapacitorRate:      1.0 / 36.0,
				FiringCone:         0.523,
				FactionSprites:     factionSprites,
				IsPlayerControlled: isPlayer,
			})
			if err != nil {
				t.Fatalf("Failed to spawn ship: %v", err)
			}

			if isPlayer {
				playerShip = ship
			}
		}
	}

	// Verify correct number of ships spawned
	shipQuery := donburi.NewQuery(filter.Contains(components.IsShip))
	shipCount := 0
	for range shipQuery.Iter(world) {
		shipCount++
	}

	expectedShips := fleetConfig.NumFactions * fleetConfig.ShipsPerFaction[0]
	if shipCount != expectedShips {
		t.Errorf("Expected %d ships, got %d", expectedShips, shipCount)
	}

	// Verify player ship exists and is controlled
	if !world.Valid(playerShip) {
		t.Fatal("Player ship is not valid")
	}

	playerEntry := world.Entry(playerShip)
	if !playerEntry.HasComponent(components.PlayerControlled) {
		t.Error("Player ship should have PlayerControlled component")
	}

	// Verify all factions are represented
	factionCount := make(map[int]int)
	for shipEntry := range shipQuery.Iter(world) {
		faction := components.Faction.Get(shipEntry)
		factionCount[faction.ID]++
	}

	if len(factionCount) != fleetConfig.NumFactions {
		t.Errorf("Expected %d factions, got %d", fleetConfig.NumFactions, len(factionCount))
	}

	for factionID := 0; factionID < fleetConfig.NumFactions; factionID++ {
		if factionCount[factionID] != fleetConfig.ShipsPerFaction[factionID] {
			t.Errorf("Faction %d: expected %d ships, got %d",
				factionID, fleetConfig.ShipsPerFaction[factionID], factionCount[factionID])
		}
	}
}

// TestProjectileLifecycleIntegration tests the full lifecycle of projectiles
func TestProjectileLifecycleIntegration(t *testing.T) {
	world := donburi.NewWorld()
	systems.InitializeFactions(world)

	laserSprite := ebiten.NewImage(4, 4)
	missileSprite := ebiten.NewImage(5, 10)
	_ = missileSprite // Not used in this test
	explosionSprite := ebiten.NewImage(400, 70)

	// Spawn ship and fire multiple projectiles
	ship, _ := systems.SpawnShip(world, systems.ShipConfig{
		Class:              components.Fighter,
		FactionID:          0,
		MaxSpeed:           6.0,
		Acceleration:       4.0 / 60.0,
		MaxHealth:          8,
		CapacitorRate:      1.0 / 36.0,
		FiringCone:         0.523,
		FactionSprites:     createTestFactionSprites(),
		IsPlayerControlled: true,
	})

	shipEntry := world.Entry(ship)
	weapon := components.Weapon.Get(shipEntry)

	// Fire 3 projectiles
	for i := 0; i < 3; i++ {
		weapon.Capacitor = 1.0
		systems.FireWeapon(world, shipEntry, 200, 100, laserSprite)
	}

	// Verify 3 projectiles exist
	projectileQuery := donburi.NewQuery(filter.Contains(components.IsProjectile))
	projectileCount := 0
	for range projectileQuery.Iter(world) {
		projectileCount++
	}
	if projectileCount != 3 {
		t.Errorf("Expected 3 projectiles, got %d", projectileCount)
	}

	// Run movement and lifetime systems
	for tick := 0; tick < 500; tick++ {
		systems.UpdateMovement(world)
		systems.UpdateProjectileLifetime(world)
		systems.UpdateCollisions(world, explosionSprite)
	}

	// All projectiles should have expired (lifetime is ~300 ticks)
	projectileCount = 0
	for range projectileQuery.Iter(world) {
		projectileCount++
	}
	if projectileCount != 0 {
		t.Errorf("Expected 0 projectiles after expiration, got %d", projectileCount)
	}
}

// TestExplosionLifecycleIntegration tests explosion creation and animation
func TestExplosionLifecycleIntegration(t *testing.T) {
	world := donburi.NewWorld()
	systems.InitializeFactions(world)

	laserSprite := ebiten.NewImage(4, 4)
	missileSprite := ebiten.NewImage(5, 10)
	_ = missileSprite // Not used in this test
	explosionSprite := ebiten.NewImage(400, 70)

	// Spawn ships for combat
	playerShip, _ := systems.SpawnShip(world, systems.ShipConfig{
		Class:              components.Fighter,
		FactionID:          0,
		MaxSpeed:           6.0,
		Acceleration:       4.0 / 60.0,
		MaxHealth:          8,
		CapacitorRate:      1.0 / 36.0,
		FiringCone:         0.523,
		FactionSprites:     createTestFactionSprites(),
		IsPlayerControlled: true,
	})

	enemyShip, _ := systems.SpawnShip(world, systems.ShipConfig{
		Class:              components.Fighter,
		FactionID:          1,
		MaxSpeed:           6.0,
		Acceleration:       4.0 / 60.0,
		MaxHealth:          1, // Low health
		CapacitorRate:      1.0 / 36.0,
		FiringCone:         0.523,
		FactionSprites:     createTestFactionSprites(),
		IsPlayerControlled: false,
	})

	// Position for collision
	playerEntry := world.Entry(playerShip)
	enemyEntry := world.Entry(enemyShip)

	playerPos := components.Position.Get(playerEntry)
	playerPos.X = 100
	playerPos.Y = 100

	enemyPos := components.Position.Get(enemyEntry)
	enemyPos.X = 120 // 20 pixels away
	enemyPos.Y = 100

	// Point player toward enemy
	// In sprite coordinates: π/2 = 90 degrees = facing right
	playerRot := components.Rotation.Get(playerEntry)
	playerRot.Angle = math.Pi / 2 // Facing right (toward enemy)

	// Fire and hit enemy
	weapon := components.Weapon.Get(playerEntry)
	weapon.Capacitor = 1.0
	systems.FireWeapon(world, playerEntry, 120, 100, laserSprite)

	// Run systems to cause collision
	// At 12 px/tick and 20 pixels distance, should hit in ~2 ticks
	for i := 0; i < 5; i++ {
		systems.UpdateMovement(world)
		systems.UpdateCollisions(world, explosionSprite)
	}

	// Verify explosion was created
	explosionQuery := donburi.NewQuery(filter.Contains(components.IsExplosion))
	explosionCount := 0
	var explosionEntity donburi.Entity
	var initialFrame int
	for entry := range explosionQuery.Iter(world) {
		explosionCount++
		explosionEntity = entry.Entity()
		initialFrame = components.Explosion.Get(entry).CurrentFrame
	}

	if explosionCount != 1 {
		t.Errorf("Expected 1 explosion, got %d", explosionCount)
		// Don't proceed if no explosion was created
		return
	}

	// Run explosion animation (already got initial frame above)

	// Run for one full animation cycle (4 frames * 5 ticks per frame = 20 ticks)
	for i := 0; i < 25; i++ {
		systems.UpdateExplosions(world)
	}

	// Explosion should be removed after animation completes
	if world.Valid(explosionEntity) {
		t.Error("Explosion should be removed after animation completes")
	}

	// Verify initial frame was 0
	if initialFrame != 0 {
		t.Errorf("Expected initial explosion frame 0, got %d", initialFrame)
	}
}
