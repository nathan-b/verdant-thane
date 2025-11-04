package systems

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
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
// Updates player score and kill count when player destroys enemy ships
// Creates explosion entities for destroyed ships
func UpdateCollisions(w donburi.World, explosionSprite *ebiten.Image) {
	// Get all projectiles
	projectileQuery := donburi.NewQuery(filter.Contains(
		components.IsProjectile,
		components.Position,
		components.Faction,
		components.Owner,
	))

	// Get all ships
	shipQuery := donburi.NewQuery(filter.Contains(
		components.IsShip,
		components.Position,
		components.Health,
		components.Faction,
	))

	// Track which entities to remove, who killed whom, and where explosions should be
	projectilesToRemove := []donburi.Entity{}
	shipsToRemove := []donburi.Entity{}
	explosionPositions := []struct{ x, y float64 }{}
	kills := []donburi.Entity{} // Track which entity (ship) made each kill

	// Check each projectile against each ship
	for projEntry := range projectileQuery.Iter(w) {
		projPos := components.Position.Get(projEntry)
		projFaction := components.Faction.Get(projEntry)
		projOwner := components.Owner.Get(projEntry)

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
					// Record position for explosion
					explosionPositions = append(explosionPositions, struct{ x, y float64 }{
						x: shipPos.X,
						y: shipPos.Y,
					})
					// Record the kill - track which ship (owner) made the kill
					kills = append(kills, projOwner.OwnerEntity)
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

	// Create explosion entities at destroyed ship positions
	for _, pos := range explosionPositions {
		explosion := w.Create(
			components.IsExplosion,
			components.Position,
			components.Explosion,
			components.Sprite,
		)
		explosionEntry := w.Entry(explosion)
		components.Position.SetValue(explosionEntry, components.PositionData{X: pos.x, Y: pos.y})
		components.Explosion.SetValue(explosionEntry, components.ExplosionData{
			CurrentFrame: 0,
			FrameTimer:   5, // 5 ticks per frame (~12 FPS)
		})
		components.Sprite.SetValue(explosionEntry, components.SpriteData{Image: explosionSprite})
	}

	// Update player score and kills if player got any kills
	// Find player state entity
	playerStateQuery := donburi.NewQuery(filter.Contains(components.PlayerState))
	playerStateEntry, ok := playerStateQuery.First(w)
	if ok {
		state := components.PlayerState.Get(playerStateEntry)
		playerShip := state.ControlledShip

		// Count kills made by the player specifically
		playerKills := 0
		for _, killerEntity := range kills {
			if killerEntity == playerShip {
				playerKills++
			}
		}

		if playerKills > 0 {
			state.Kills += playerKills
			state.Score += playerKills * 10 // 10 points per kill
		}
	}
}
