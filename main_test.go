package main

import (
	"testing"

	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
	"github.com/nathan/verdant-thane/systems"
)

func TestGenerateFleetConfig_Deterministic(t *testing.T) {
	// Test specific configuration
	config := GenerateFleetConfig(3, 10)

	if config.NumFactions != 3 {
		t.Errorf("Expected 3 factions, got %d", config.NumFactions)
	}

	if len(config.Compositions) != 3 {
		t.Errorf("Expected 3 faction entries, got %d", len(config.Compositions))
	}

	for i, comp := range config.Compositions {
		if comp.Fighters != 10 {
			t.Errorf("Expected 10 fighters for faction %d, got %d", i, comp.Fighters)
		}
		if comp.Destroyers != 0 || comp.Testudons != 0 {
			t.Errorf("Expected only fighters for faction %d, got destroyers=%d testudons=%d",
				i, comp.Destroyers, comp.Testudons)
		}
	}
}

func TestGenerateFleetConfig_BoundsChecking(t *testing.T) {
	// Test lower bound (< 2 factions should be clamped to 2)
	config := GenerateFleetConfig(1, 5)
	if config.NumFactions != 2 {
		t.Errorf("Expected 2 factions (clamped from 1), got %d", config.NumFactions)
	}

	// Test upper bound (> 4 factions should be clamped to 4)
	config = GenerateFleetConfig(10, 5)
	if config.NumFactions != 4 {
		t.Errorf("Expected 4 factions (clamped from 10), got %d", config.NumFactions)
	}

	// Test minimum ships (< 1 should be clamped to 1)
	config = GenerateFleetConfig(2, 0)
	for i, comp := range config.Compositions {
		if comp.Fighters != 1 {
			t.Errorf("Expected 1 fighter (clamped from 0) for faction %d, got %d", i, comp.Fighters)
		}
	}
}

func TestGenerateRandomFleetConfig_SeedDeterminism(t *testing.T) {
	const seed = int64(42)

	// Generate two configs with same seed
	config1 := GenerateRandomFleetConfig(seed)
	config2 := GenerateRandomFleetConfig(seed)

	// They should be identical
	if config1.NumFactions != config2.NumFactions {
		t.Errorf("Expected same NumFactions, got %d and %d", config1.NumFactions, config2.NumFactions)
	}

	if len(config1.Compositions) != len(config2.Compositions) {
		t.Fatalf("Expected same number of faction entries, got %d and %d",
			len(config1.Compositions), len(config2.Compositions))
	}

	for i := range config1.Compositions {
		comp1 := config1.Compositions[i]
		comp2 := config2.Compositions[i]
		if comp1.Fighters != comp2.Fighters || comp1.Destroyers != comp2.Destroyers || comp1.Testudons != comp2.Testudons {
			t.Errorf("Expected same composition for faction %d, got %+v and %+v", i, comp1, comp2)
		}
	}
}

func TestGenerateRandomFleetConfig_BoundsChecking(t *testing.T) {
	// Test multiple random configs to ensure they stay within bounds
	for i := 0; i < 20; i++ {
		config := GenerateRandomFleetConfig(int64(i))

		// Check faction count (2-4)
		if config.NumFactions < 2 || config.NumFactions > 4 {
			t.Errorf("Faction count out of bounds: %d (expected 2-4)", config.NumFactions)
		}

		// Check total ships per faction (should be <= 16 since fighter budget is 7-16)
		for factionID, comp := range config.Compositions {
			totalShips := comp.Total()
			if totalShips < 1 || totalShips > 16 {
				t.Errorf("Total ships for faction %d out of bounds: %d (expected 1-16)", factionID, totalShips)
			}
		}
	}
}

func TestSpawnMultipleFactions(t *testing.T) {
	// Create a test fleet config
	config := GenerateFleetConfig(3, 5) // 3 factions, 5 ships each

	// Create world
	world := donburi.NewWorld()

	// Initialize factions
	systems.InitializeFactions(world)

	// Spawn ships manually (simulate what NewGame does, but without loading assets)
	factionSprites := createTestFactionSprites()

	for factionID := 0; factionID < config.NumFactions; factionID++ {
		numShips := config.Compositions[factionID].Total()

		for shipIndex := 0; shipIndex < numShips; shipIndex++ {
			isPlayerControlled := (factionID == 0 && shipIndex == 0)

			_, err := systems.SpawnShip(world, systems.ShipConfig{
				Class:              components.Fighter,
				FactionID:          factionID,
				MaxSpeed:           6.0,
				Acceleration:       0.067,
				MaxHealth:          8,
				CapacitorRate:      0.027,
				FiringCone:         0.524,
				FactionSprites:     factionSprites,
				IsPlayerControlled: isPlayerControlled,
			})

			if err != nil {
				t.Fatalf("Failed to spawn ship for faction %d: %v", factionID, err)
			}
		}
	}

	// Count ships per faction
	shipCounts := make(map[int]int)

	query := donburi.NewQuery(filter.Contains(components.IsShip, components.Faction))
	for entry := range query.Iter(world) {
		faction := components.Faction.Get(entry)
		shipCounts[faction.ID]++
	}

	// Verify counts
	for factionID := 0; factionID < config.NumFactions; factionID++ {
		expected := config.Compositions[factionID].Total()
		actual := shipCounts[factionID]

		if actual != expected {
			t.Errorf("Faction %d: expected %d ships, got %d", factionID, expected, actual)
		}
	}
}

func TestPlayerShipAssignment(t *testing.T) {
	// Create a test fleet config
	config := GenerateFleetConfig(2, 3) // 2 factions, 3 ships each

	// Create world
	world := donburi.NewWorld()

	// Initialize factions
	systems.InitializeFactions(world)

	// Spawn ships
	factionSprites := createTestFactionSprites()
	var playerShip donburi.Entity

	for factionID := 0; factionID < config.NumFactions; factionID++ {
		numShips := config.Compositions[factionID].Total()

		for shipIndex := 0; shipIndex < numShips; shipIndex++ {
			isPlayerControlled := (factionID == 0 && shipIndex == 0)

			ship, err := systems.SpawnShip(world, systems.ShipConfig{
				Class:              components.Fighter,
				FactionID:          factionID,
				MaxSpeed:           6.0,
				Acceleration:       0.067,
				MaxHealth:          8,
				CapacitorRate:      0.027,
				FiringCone:         0.524,
				FactionSprites:     factionSprites,
				IsPlayerControlled: isPlayerControlled,
			})

			if err != nil {
				t.Fatalf("Failed to spawn ship: %v", err)
			}

			if isPlayerControlled {
				playerShip = ship
			}
		}
	}

	// Verify player ship has PlayerControlled tag
	if !world.Valid(playerShip) {
		t.Fatal("Player ship entity is invalid")
	}

	playerEntry := world.Entry(playerShip)
	if !playerEntry.HasComponent(components.PlayerControlled) {
		t.Error("Player ship does not have PlayerControlled tag")
	}

	// Verify player ship is faction 0
	faction := components.Faction.Get(playerEntry)
	if faction.ID != 0 {
		t.Errorf("Player ship should be faction 0, got %d", faction.ID)
	}

	// Count total player-controlled ships (should be exactly 1)
	playerQuery := donburi.NewQuery(filter.Contains(components.PlayerControlled))
	playerCount := 0
	for range playerQuery.Iter(world) {
		playerCount++
	}

	if playerCount != 1 {
		t.Errorf("Expected exactly 1 player-controlled ship, got %d", playerCount)
	}
}
