package systems

import (
	"math"

	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

// CollisionRadius defines the collision radius for different entity types
const (
	ShipCollisionRadius       = 16.0 // Half the approximate sprite size
	ProjectileCollisionRadius = 2.0  // Small collision for projectiles
)

// distance calculates the distance between two points, accounting for world wrapping
func distance(x1, y1, x2, y2 float64) float64 {
	// Calculate direct distance
	dx := x2 - x1
	dy := y2 - y1

	// Account for world wrapping - find shortest distance considering wrap-around
	if math.Abs(dx) > float64(GameWidth)/2 {
		if dx > 0 {
			dx = dx - float64(GameWidth)
		} else {
			dx = dx + float64(GameWidth)
		}
	}
	if math.Abs(dy) > float64(GameHeight)/2 {
		if dy > 0 {
			dy = dy - float64(GameHeight)
		} else {
			dy = dy + float64(GameHeight)
		}
	}

	return math.Sqrt(dx*dx + dy*dy)
}

// UpdateCollisions checks for collisions between projectiles and ships
// Projectiles damage ships of different factions
func UpdateCollisions(w donburi.World) {
	// Get all projectiles
	projectileQuery := donburi.NewQuery(filter.Contains(
		components.IsProjectile,
		components.Position,
		components.Faction,
	))

	// Get all ships
	shipQuery := donburi.NewQuery(filter.Contains(
		components.IsShip,
		components.Position,
		components.Health,
		components.Faction,
	))

	// Track which entities to remove
	projectilesToRemove := []donburi.Entity{}
	shipsToRemove := []donburi.Entity{}

	// Check each projectile against each ship
	for projEntry := range projectileQuery.Iter(w) {
		projPos := components.Position.Get(projEntry)
		projFaction := components.Faction.Get(projEntry)

		for shipEntry := range shipQuery.Iter(w) {
			shipPos := components.Position.Get(shipEntry)
			shipFaction := components.Faction.Get(shipEntry)
			health := components.Health.Get(shipEntry)

			// Don't check friendly fire
			if projFaction.ID == shipFaction.ID {
				continue
			}

			// Calculate distance with wrapping
			dist := distance(projPos.X, projPos.Y, shipPos.X, shipPos.Y)

			// Check if collision occurred
			collisionDist := ShipCollisionRadius + ProjectileCollisionRadius
			if dist < collisionDist {
				// Hit! Reduce ship health
				health.Current--

				// Mark projectile for removal
				projectilesToRemove = append(projectilesToRemove, projEntry.Entity())

				// Check if ship is destroyed
				if health.Current <= 0 {
					shipsToRemove = append(shipsToRemove, shipEntry.Entity())
				}

				// Break out of ship loop since this projectile hit something
				break
			}
		}
	}

	// Remove projectiles that hit ships
	for _, entity := range projectilesToRemove {
		if w.Valid(entity) {
			w.Remove(entity)
		}
	}

	// Remove destroyed ships
	for _, entity := range shipsToRemove {
		if w.Valid(entity) {
			w.Remove(entity)
		}
	}
}
