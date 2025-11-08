package systems

import (
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

// UpdateCollisions checks for collisions between projectiles and ships
// Projectiles damage ships of different factions
// Updates player score and kill count when player destroys enemy ships
// Creates explosion entities for destroyed ships
// OPTIMIZED: Uses spatial grid partitioning to reduce collision checks from O(n*m) to O(n*k)
func UpdateCollisions(w donburi.World, explosionSprite *ebiten.Image) {
	// Build spatial grid of ships
	grid := newSpatialGrid()
	shipQuery := donburi.NewQuery(filter.Contains(
		components.IsShip,
		components.Position,
		components.Health,
		components.Faction,
		components.CollisionRadius,
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

	// Track which entities to remove and ship destruction info
	projectilesToRemove := make([]donburi.Entity, 0, 64)
	shipsToDestroy := make([]*donburi.Entry, 0, 16)
	killerEntities := make([]donburi.Entity, 0, 16)   // Track which entity (ship) made each kill
	destroyedShipSet := make(map[donburi.Entity]bool) // Prevent double-processing ships hit by multiple projectiles

	// Check each projectile against only nearby ships using spatial grid
	for projEntry := range projectileQuery.Iter(w) {
		projPos := components.Position.Get(projEntry)
		projFaction := components.Faction.Get(projEntry)
		projOwner := components.Owner.Get(projEntry)

		// Get only ships in nearby grid cells (9 cells max instead of all ships)
		nearbyShips := grid.getNearbyShips(projPos.X, projPos.Y)

		for _, shipEntry := range nearbyShips {
			shipEntity := shipEntry.Entity()

			// Skip ships that are already marked for destruction
			if destroyedShipSet[shipEntity] {
				continue
			}

			shipPos := components.Position.Get(shipEntry)
			shipFaction := components.Faction.Get(shipEntry)
			shipCollisionRadius := components.CollisionRadius.Get(shipEntry)

			// Don't check friendly fire
			if projFaction.ID == shipFaction.ID {
				continue
			}

			// Calculate collision distance squared using ship's actual collision radius
			// (avoids expensive sqrt in inner loop)
			collisionDist := shipCollisionRadius.Radius + config.ProjectileCollisionRadius
			collisionDistSq := collisionDist * collisionDist

			// Calculate squared distance (avoids expensive sqrt)
			distSq := DistanceSquared(projPos.X, projPos.Y, shipPos.X, shipPos.Y)

			// Check if collision occurred
			if distSq < collisionDistSq {
				health := components.Health.Get(shipEntry)

				// Hit! Reduce ship health
				health.Current--

				// Mark projectile for removal
				projectilesToRemove = append(projectilesToRemove, projEntry.Entity())

				// Check if ship is destroyed
				if health.Current <= 0 {
					// Mark ship for destruction (we already checked destroyedShipSet at loop start)
					destroyedShipSet[shipEntity] = true
					shipsToDestroy = append(shipsToDestroy, shipEntry)
					// Record the kill - track which ship (owner) made the kill
					killerEntities = append(killerEntities, projOwner.OwnerEntity)
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

	// Destroy all ships marked for destruction
	// DestroyShip handles: explosion creation, player respawn, score updates, ship removal
	for i, shipEntry := range shipsToDestroy {
		killerEntity := killerEntities[i]
		DestroyShip(w, shipEntry, killerEntity, explosionSprite)
	}
}
