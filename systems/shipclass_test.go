package systems

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/nathan/verdant-thane/components"
)

// TestFighterComponents verifies Fighters have correct components
func TestFighterComponents(t *testing.T) {
	world := donburi.NewWorld()
	InitializeFactions(world)

	// Create test sprites
	factionSprites := createTestFactionSprites()

	// Spawn a Fighter
	fighterEntity, err := SpawnShip(world, ShipConfig{
		Class:              components.Fighter,
		FactionID:          0,
		MaxSpeed:           6.0,
		Acceleration:       0.067,
		MaxHealth:          8,
		CapacitorRate:      0.028,
		FiringCone:         0.524,
		FactionSprites:     factionSprites,
		IsPlayerControlled: false,
	})

	if err != nil {
		t.Fatalf("Failed to spawn Fighter: %v", err)
	}

	entry := world.Entry(fighterEntity)

	// Fighters should have primary weapon
	if !entry.HasComponent(components.Weapon) {
		t.Error("Fighter should have Weapon component")
	}

	// Fighters should NOT have secondary weapon (missiles)
	if entry.HasComponent(components.SecondaryWeapon) {
		t.Error("Fighter should NOT have SecondaryWeapon component")
	}

	// Fighters should NOT have beam weapon
	if entry.HasComponent(components.BeamWeapon) {
		t.Error("Fighter should NOT have BeamWeapon component")
	}

	// Fighters should NOT have UnderAttack component by default
	if entry.HasComponent(components.UnderAttack) {
		t.Error("Fighter should NOT have UnderAttack component")
	}
}

// TestDestroyerComponents verifies Destroyers have correct components
func TestDestroyerComponents(t *testing.T) {
	world := donburi.NewWorld()
	InitializeFactions(world)

	factionSprites := createTestFactionSprites()

	// Spawn a Destroyer
	destroyerEntity, err := SpawnShip(world, ShipConfig{
		Class:              components.Destroyer,
		FactionID:          0,
		MaxSpeed:           5.0,
		Acceleration:       0.05,
		MaxHealth:          32,
		CapacitorRate:      0.028,
		FiringCone:         0.524,
		FactionSprites:     factionSprites,
		IsPlayerControlled: false,
	})

	if err != nil {
		t.Fatalf("Failed to spawn Destroyer: %v", err)
	}

	entry := world.Entry(destroyerEntity)

	// Destroyers should have primary weapon
	if !entry.HasComponent(components.Weapon) {
		t.Error("Destroyer should have Weapon component")
	}

	// Destroyers should have secondary weapon (missiles)
	if !entry.HasComponent(components.SecondaryWeapon) {
		t.Error("Destroyer should have SecondaryWeapon component")
	}

	// Destroyers should NOT have beam weapon
	if entry.HasComponent(components.BeamWeapon) {
		t.Error("Destroyer should NOT have BeamWeapon component")
	}

	// Destroyers should NOT have UnderAttack component by default
	if entry.HasComponent(components.UnderAttack) {
		t.Error("Destroyer should NOT have UnderAttack component")
	}
}

// TestTestudonComponents verifies Testudons have correct components
func TestTestudonComponents(t *testing.T) {
	world := donburi.NewWorld()
	InitializeFactions(world)

	factionSprites := createTestFactionSprites()

	// Spawn a Testudon
	testudonEntity, err := SpawnShip(world, ShipConfig{
		Class:              components.Testudon,
		FactionID:          0,
		MaxSpeed:           3.0,
		Acceleration:       0.033,
		MaxHealth:          100,
		CapacitorRate:      0.0,   // No projectile weapon
		FiringCone:         0.0,   // No firing cone
		FactionSprites:     factionSprites,
		IsPlayerControlled: false,
	})

	if err != nil {
		t.Fatalf("Failed to spawn Testudon: %v", err)
	}

	entry := world.Entry(testudonEntity)

	// Testudons should NOT have primary weapon
	if entry.HasComponent(components.Weapon) {
		t.Error("Testudon should NOT have Weapon component")
	}

	// Testudons should NOT have secondary weapon
	if entry.HasComponent(components.SecondaryWeapon) {
		t.Error("Testudon should NOT have SecondaryWeapon component")
	}

	// Testudons SHOULD have beam weapon
	if !entry.HasComponent(components.BeamWeapon) {
		t.Error("Testudon should have BeamWeapon component")
	}

	// Testudons SHOULD have UnderAttack component for defensive AI
	if !entry.HasComponent(components.UnderAttack) {
		t.Error("Testudon should have UnderAttack component")
	}

	// Verify beam weapon configuration
	beamWeapon := components.BeamWeapon.Get(entry)
	if beamWeapon.Range != 200.0 {
		t.Errorf("Expected beam range 200, got %f", beamWeapon.Range)
	}
	if beamWeapon.DamagePerTick != 0.1 {
		t.Errorf("Expected damage per tick 0.1, got %f", beamWeapon.DamagePerTick)
	}
}

// TestShipClassCharacteristics verifies each ship class has correct stats
func TestShipClassCharacteristics(t *testing.T) {
	tests := []struct {
		class       components.ShipClass
		name        string
		maxSpeed    float64
		maxShield   int
		hasWeapon   bool
		hasMissiles bool
		hasBeam     bool
	}{
		{
			class:       components.Fighter,
			name:        "Fighter",
			maxSpeed:    6.0,
			maxShield:   8,
			hasWeapon:   true,
			hasMissiles: false,
			hasBeam:     false,
		},
		{
			class:       components.Destroyer,
			name:        "Destroyer",
			maxSpeed:    5.0,
			maxShield:   32,
			hasWeapon:   true,
			hasMissiles: true,
			hasBeam:     false,
		},
		{
			class:       components.Testudon,
			name:        "Testudon",
			maxSpeed:    3.0,
			maxShield:   100,
			hasWeapon:   false,
			hasMissiles: false,
			hasBeam:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := donburi.NewWorld()
			InitializeFactions(world)
			factionSprites := createTestFactionSprites()

			// Spawn ship of this class
			entity, err := SpawnShip(world, ShipConfig{
				Class:              tt.class,
				FactionID:          0,
				MaxSpeed:           tt.maxSpeed,
				Acceleration:       0.05,
				MaxHealth:          tt.maxShield,
				CapacitorRate:      0.028,
				FiringCone:         0.524,
				FactionSprites:     factionSprites,
				IsPlayerControlled: false,
			})

			if err != nil {
				t.Fatalf("Failed to spawn %s: %v", tt.name, err)
			}

			entry := world.Entry(entity)

			// Check components
			if entry.HasComponent(components.Weapon) != tt.hasWeapon {
				t.Errorf("%s: Weapon component mismatch (expected %v)", tt.name, tt.hasWeapon)
			}
			if entry.HasComponent(components.SecondaryWeapon) != tt.hasMissiles {
				t.Errorf("%s: SecondaryWeapon component mismatch (expected %v)", tt.name, tt.hasMissiles)
			}
			if entry.HasComponent(components.BeamWeapon) != tt.hasBeam {
				t.Errorf("%s: BeamWeapon component mismatch (expected %v)", tt.name, tt.hasBeam)
			}

			// Check stats
			ship := components.Ship.Get(entry)
			if ship.MaxSpeed != tt.maxSpeed {
				t.Errorf("%s: MaxSpeed mismatch (expected %.1f, got %.1f)", tt.name, tt.maxSpeed, ship.MaxSpeed)
			}

			health := components.Health.Get(entry)
			if health.Max != tt.maxShield {
				t.Errorf("%s: MaxShield mismatch (expected %d, got %d)", tt.name, tt.maxShield, health.Max)
			}
		})
	}
}

// Helper function to create test sprites (minimal for testing)
func createTestFactionSprites() *FactionSprites {
	// Create minimal test sprites
	testImage := ebiten.NewImage(32, 32)

	testClassSprites := &ShipClassSprites{
		baseSprite: testImage,
		Green:      testImage,
		Blue:       testImage,
		Red:        testImage,
		Yellow:     testImage,
	}

	return &FactionSprites{
		Fighter:   testClassSprites,
		Destroyer: testClassSprites,
		Testudon:  testClassSprites,
	}
}
