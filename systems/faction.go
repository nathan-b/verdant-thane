package systems

import (
	"fmt"
	"math/rand"

	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

// InitializeFactions creates four faction entities with randomly assigned spawn points
// Spawn points are at the four cardinal directions from the center
func InitializeFactions(w donburi.World) {
	// Calculate spawn point positions
	// Each spawn is halfway between center and edge in each cardinal direction
	spawnPoints := []struct{ x, y float64 }{
		{float64(GameWidth) / 2, float64(GameHeight) / 4},     // North
		{float64(GameWidth) / 2, 3 * float64(GameHeight) / 4}, // South
		{3 * float64(GameWidth) / 4, float64(GameHeight) / 2}, // East
		{float64(GameWidth) / 4, float64(GameHeight) / 2},     // West
	}

	// Shuffle spawn points to randomly assign them to factions
	rand.Shuffle(len(spawnPoints), func(i, j int) {
		spawnPoints[i], spawnPoints[j] = spawnPoints[j], spawnPoints[i]
	})

	// Create faction entities with assigned spawn points
	// Faction IDs: 0 = Green (player), 1 = Blue, 2 = Red, 3 = Yellow
	for factionID := 0; factionID < 4; factionID++ {
		factionEntity := w.Create(components.FactionInfo)
		factionEntry := w.Entry(factionEntity)
		components.FactionInfo.SetValue(factionEntry, components.FactionInfoData{
			FactionID: factionID,
			SpawnX:    spawnPoints[factionID].x,
			SpawnY:    spawnPoints[factionID].y,
		})
	}
}

// GetFactionSpawnPoint returns the spawn coordinates for a given faction ID
// Returns (0, 0) if faction not found
func GetFactionSpawnPoint(w donburi.World, factionID int) (float64, float64) {
	query := donburi.NewQuery(filter.Contains(components.FactionInfo))

	for entry := range query.Iter(w) {
		info := components.FactionInfo.Get(entry)
		if info.FactionID == factionID {
			return info.SpawnX, info.SpawnY
		}
	}

	return 0, 0 // Fallback if faction not found
}

// ShipConfig holds configuration parameters for spawning a ship
type ShipConfig struct {
	Class              components.ShipClass
	FactionID          int
	MaxSpeed           float64
	Acceleration       float64
	MaxHealth          int
	CapacitorRate      float64
	FiringCone         float64
	FactionSprites     *FactionSprites
	IsPlayerControlled bool
}

// SpawnShip creates a new ship entity at the faction's spawn point
// Returns the created entity
func SpawnShip(w donburi.World, config ShipConfig) (donburi.Entity, error) {
	// Get spawn position for this faction
	spawnX, spawnY := GetFactionSpawnPoint(w, config.FactionID)
	if spawnX == 0 && spawnY == 0 {
		var emptyEntity donburi.Entity
		return emptyEntity, fmt.Errorf("faction %d not found", config.FactionID)
	}

	// Create ship entity with all necessary components
	ship := w.Create(
		components.IsShip,
		components.Position,
		components.Velocity,
		components.Rotation,
		components.Ship,
		components.Faction,
		components.Health,
		components.Weapon,
		components.Sprite,
	)

	// Add control tags and AI state
	entry := w.Entry(ship)
	if config.IsPlayerControlled {
		entry.AddComponent(components.PlayerControlled)
	} else {
		entry.AddComponent(components.AIControlled)
		entry.AddComponent(components.AIState)
		// Initialize AI with first decision in ~1 second
		components.AIState.SetValue(entry, components.AIStateData{
			DecisionTimer: 60, // 60 ticks ≈ 1 second
		})
	}

	// Set component values
	components.Position.SetValue(entry, components.PositionData{X: spawnX, Y: spawnY})
	components.Velocity.SetValue(entry, components.VelocityData{X: 0, Y: 0})
	components.Rotation.SetValue(entry, components.RotationData{Angle: 0})
	components.Ship.SetValue(entry, components.ShipData{
		Class:    config.Class,
		Speed:    0,
		MaxSpeed: config.MaxSpeed,
		Accel:    config.Acceleration,
	})
	components.Faction.SetValue(entry, components.FactionData{ID: config.FactionID})
	components.Health.SetValue(entry, components.HealthData{Current: config.MaxHealth, Max: config.MaxHealth})
	components.Weapon.SetValue(entry, components.WeaponData{
		Capacitor:  1.0, // Start fully charged
		ChargeRate: config.CapacitorRate,
		FiringCone: config.FiringCone,
	})

	// Get the correct sprite for this faction
	sprite := config.FactionSprites.GetSpriteForFaction(config.FactionID)
	components.Sprite.SetValue(entry, components.SpriteData{Image: sprite})

	return ship, nil
}

