package systems

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
	"github.com/nathan/verdant-thane/config"
)

// Spatial grid for collision detection
type spatialGrid struct {
	cellSize   int
	gridWidth  int
	gridHeight int
	cells      map[int][]*donburi.Entry // Grid cell index -> list of ship entries
}

// newSpatialGrid creates a new spatial grid for the game world
func newSpatialGrid() *spatialGrid {
	return &spatialGrid{
		cellSize:   config.CollisionGridSize,
		gridWidth:  (config.GameWidth + config.CollisionGridSize - 1) / config.CollisionGridSize,
		gridHeight: (config.GameHeight + config.CollisionGridSize - 1) / config.CollisionGridSize,
		cells:      make(map[int][]*donburi.Entry),
	}
}

// getCellIndex returns the grid cell index for a position
func (g *spatialGrid) getCellIndex(x, y float64) int {
	// Wrap coordinates to world bounds
	wx := int(x) % config.GameWidth
	wy := int(y) % config.GameHeight
	if wx < 0 {
		wx += config.GameWidth
	}
	if wy < 0 {
		wy += config.GameHeight
	}

	cellX := wx / g.cellSize
	cellY := wy / g.cellSize

	return cellY*g.gridWidth + cellX
}

// insert adds a ship entry to the appropriate grid cell
func (g *spatialGrid) insert(entry *donburi.Entry, x, y float64) {
	idx := g.getCellIndex(x, y)
	g.cells[idx] = append(g.cells[idx], entry)
}

// getNearbyShips returns all ships in the same cell and adjacent cells
func (g *spatialGrid) getNearbyShips(x, y float64) []*donburi.Entry {
	// Wrap coordinates
	wx := int(x) % config.GameWidth
	wy := int(y) % config.GameHeight
	if wx < 0 {
		wx += config.GameWidth
	}
	if wy < 0 {
		wy += config.GameHeight
	}

	cellX := wx / g.cellSize
	cellY := wy / g.cellSize

	nearby := make([]*donburi.Entry, 0, 32) // Pre-allocate reasonable capacity

	// Check 3x3 grid of cells around the projectile
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			// Handle wrapping at grid boundaries
			checkX := (cellX + dx + g.gridWidth) % g.gridWidth
			checkY := (cellY + dy + g.gridHeight) % g.gridHeight

			idx := checkY*g.gridWidth + checkX
			if ships, ok := g.cells[idx]; ok {
				nearby = append(nearby, ships...)
			}
		}
	}

	return nearby
}

// clear resets the grid for the next frame
func (g *spatialGrid) clear() {
	// Reuse the map to avoid allocations
	for k := range g.cells {
		delete(g.cells, k)
	}
}

// distance calculates the distance between two points, accounting for world wrapping
func distance(x1, y1, x2, y2 float64) float64 {
	// Calculate direct distance
	dx := x2 - x1
	dy := y2 - y1

	// Account for world wrapping - find shortest distance considering wrap-around
	if math.Abs(dx) > float64(config.GameWidth)/2 {
		if dx > 0 {
			dx = dx - float64(config.GameWidth)
		} else {
			dx = dx + float64(config.GameWidth)
		}
	}
	if math.Abs(dy) > float64(config.GameHeight)/2 {
		if dy > 0 {
			dy = dy - float64(config.GameHeight)
		} else {
			dy = dy + float64(config.GameHeight)
		}
	}

	return math.Sqrt(dx*dx + dy*dy)
}

// distanceSquared calculates the squared distance between two points, accounting for world wrapping
// Faster than distance() since it avoids the sqrt() operation
func distanceSquared(x1, y1, x2, y2 float64) float64 {
	// Calculate direct distance
	dx := x2 - x1
	dy := y2 - y1

	// Account for world wrapping - find shortest distance considering wrap-around
	if math.Abs(dx) > float64(config.GameWidth)/2 {
		if dx > 0 {
			dx = dx - float64(config.GameWidth)
		} else {
			dx = dx + float64(config.GameWidth)
		}
	}
	if math.Abs(dy) > float64(config.GameHeight)/2 {
		if dy > 0 {
			dy = dy - float64(config.GameHeight)
		} else {
			dy = dy + float64(config.GameHeight)
		}
	}

	return dx*dx + dy*dy
}

// UpdateCollisions checks for collisions between projectiles and ships
// Projectiles damage ships of different factions
// Updates player score and kill count when player destroys enemy ships
// Creates explosion entities for destroyed ships
// OPTIMIZED: Uses spatial grid partitioning to reduce collision checks from O(n*m) to O(n*k)
func UpdateCollisions(w donburi.World, explosionSprite *ebiten.Image) {
	// Pre-calculate collision distance squared (avoids sqrt in inner loop)
	collisionDistSq := (config.ShipCollisionRadius + config.ProjectileCollisionRadius) * (config.ShipCollisionRadius + config.ProjectileCollisionRadius)

	// Build spatial grid of ships
	grid := newSpatialGrid()
	shipQuery := donburi.NewQuery(filter.Contains(
		components.IsShip,
		components.Position,
		components.Health,
		components.Faction,
	))

	// Populate grid with all ships
	for shipEntry := range shipQuery.Iter(w) {
		shipPos := components.Position.Get(shipEntry)
		grid.insert(shipEntry, shipPos.X, shipPos.Y)
	}

	// Get all projectiles
	projectileQuery := donburi.NewQuery(filter.Contains(
		components.IsProjectile,
		components.Position,
		components.Faction,
		components.Owner,
	))

	// Track which entities to remove, who killed whom, and where explosions should be
	projectilesToRemove := make([]donburi.Entity, 0, 64)
	shipsToRemove := make([]donburi.Entity, 0, 16)
	explosionPositions := make([]struct{ x, y float64 }, 0, 16)
	kills := make([]donburi.Entity, 0, 16) // Track which entity (ship) made each kill

	// Check each projectile against only nearby ships using spatial grid
	for projEntry := range projectileQuery.Iter(w) {
		projPos := components.Position.Get(projEntry)
		projFaction := components.Faction.Get(projEntry)
		projOwner := components.Owner.Get(projEntry)

		// Get only ships in nearby grid cells (9 cells max instead of all ships)
		nearbyShips := grid.getNearbyShips(projPos.X, projPos.Y)

		for _, shipEntry := range nearbyShips {
			shipPos := components.Position.Get(shipEntry)
			shipFaction := components.Faction.Get(shipEntry)

			// Don't check friendly fire
			if projFaction.ID == shipFaction.ID {
				continue
			}

			// Calculate squared distance (avoids expensive sqrt)
			distSq := distanceSquared(projPos.X, projPos.Y, shipPos.X, shipPos.Y)

			// Check if collision occurred
			if distSq < collisionDistSq {
				health := components.Health.Get(shipEntry)

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
