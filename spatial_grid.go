package main

import (
	"github.com/nathan-b/verdant-thane/config"
	"github.com/nathan-b/verdant-thane/entity"
)

// spatialGrid implements spatial partitioning for efficient collision detection
// Divides the game world into cells and allows quick lookup of ships in a neighborhood
type spatialGrid struct {
	cellSize   int                    // Size of each grid cell in pixels (typically 128)
	gridWidth  int                    // Number of cells horizontally
	gridHeight int                    // Number of cells vertically
	cells      map[int][]*entity.Ship // Cell index → ships in that cell
}

// newSpatialGrid creates a new spatial grid for the game world
func newSpatialGrid() *spatialGrid {
	cellSize := config.CollisionGridSize // 128 pixels
	gridWidth := (config.GameWidth + cellSize - 1) / cellSize
	gridHeight := (config.GameHeight + cellSize - 1) / cellSize

	return &spatialGrid{
		cellSize:   cellSize,
		gridWidth:  gridWidth,
		gridHeight: gridHeight,
		cells:      make(map[int][]*entity.Ship, gridWidth*gridHeight),
	}
}

// getCellIndex converts a world position to a grid cell index
// Handles world wrapping for toroidal game board
func (g *spatialGrid) getCellIndex(x, y float64) int {
	// Handle world wrapping
	wx := int(x) % config.GameWidth
	wy := int(y) % config.GameHeight

	if wx < 0 {
		wx += config.GameWidth
	}
	if wy < 0 {
		wy += config.GameHeight
	}

	// Calculate cell coordinates
	cellX := wx / g.cellSize
	cellY := wy / g.cellSize

	// Ensure within bounds (should always be true with wrapping, but defensive)
	if cellX >= g.gridWidth {
		cellX = g.gridWidth - 1
	}
	if cellY >= g.gridHeight {
		cellY = g.gridHeight - 1
	}

	return cellY*g.gridWidth + cellX
}

// insert adds a ship to the appropriate grid cell
func (g *spatialGrid) insert(ship *entity.Ship) {
	x, y := ship.GetPosition()
	cellIndex := g.getCellIndex(x, y)

	// Append to cell's ship list
	g.cells[cellIndex] = append(g.cells[cellIndex], ship)
}

// findNearestEnemy searches outward from a position to find the nearest enemy ship
// Uses expanding ring search for O(k) performance instead of O(n)
// Returns nil if no enemy found
func (g *spatialGrid) findNearestEnemy(ship *entity.Ship) (*entity.Ship, float64) {
	shipX, shipY := ship.GetPosition()
	shipFaction := ship.GetFaction()
	shipID := ship.GetID()

	var nearestShip *entity.Ship
	minDistance := 999999.0

	// Get starting cell coordinates
	wx := int(shipX) % config.GameWidth
	wy := int(shipY) % config.GameHeight
	if wx < 0 {
		wx += config.GameWidth
	}
	if wy < 0 {
		wy += config.GameHeight
	}
	centerCellX := wx / g.cellSize
	centerCellY := wy / g.cellSize

	// Search outward in expanding rings until we've checked all cells within minDistance
	// Start with radius 0 (own cell), then 1 (3x3), then 2 (5x5), etc.
	maxRadius := max(g.gridWidth, g.gridHeight) // Maximum possible radius

	for radius := 0; radius <= maxRadius; radius++ {
		// Early termination: if we've found a ship and all cells at this radius
		// are farther than minDistance, we can stop
		radiusMinDist := float64(radius * g.cellSize)
		if nearestShip != nil && radiusMinDist > minDistance {
			break
		}

		// Check all cells in the ring at this radius
		for dx := -radius; dx <= radius; dx++ {
			for dy := -radius; dy <= radius; dy++ {
				// Only check cells on the perimeter of this radius
				// (Interior cells were checked in previous iterations)
				if radius > 0 && abs(dx) != radius && abs(dy) != radius {
					continue
				}

				// Calculate wrapped cell coordinates
				cellX := (centerCellX + dx + g.gridWidth*1000) % g.gridWidth
				cellY := (centerCellY + dy + g.gridHeight*1000) % g.gridHeight
				cellIndex := cellY*g.gridWidth + cellX

				// Check all ships in this cell
				for _, otherShip := range g.cells[cellIndex] {
					// Skip same faction
					if otherShip.GetFaction() == shipFaction {
						continue
					}
					// Skip self
					if otherShip.GetID() == shipID {
						continue
					}
					// Skip dead ships
					if !otherShip.IsAlive() {
						continue
					}

					// Calculate distance
					otherX, otherY := otherShip.GetPosition()
					dist := entity.Distance(shipX, shipY, otherX, otherY)

					if dist < minDistance {
						minDistance = dist
						nearestShip = otherShip
					}
				}
			}
		}
	}

	return nearestShip, minDistance
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// getNearbyShips returns all ships in the 3×3 cell neighborhood around a position
// This includes the cell at (x,y) plus all 8 adjacent cells
func (g *spatialGrid) getNearbyShips(x, y float64) []*entity.Ship {
	// Get the center cell
	wx := int(x) % config.GameWidth
	wy := int(y) % config.GameHeight

	if wx < 0 {
		wx += config.GameWidth
	}
	if wy < 0 {
		wy += config.GameHeight
	}

	centerCellX := wx / g.cellSize
	centerCellY := wy / g.cellSize

	// Ensure within bounds
	if centerCellX >= g.gridWidth {
		centerCellX = g.gridWidth - 1
	}
	if centerCellY >= g.gridHeight {
		centerCellY = g.gridHeight - 1
	}

	// Collect ships from 3×3 neighborhood
	nearby := make([]*entity.Ship, 0, 20) // Pre-allocate for typical case

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			// Calculate neighbor cell coordinates with wrapping
			neighborX := (centerCellX + dx + g.gridWidth) % g.gridWidth
			neighborY := (centerCellY + dy + g.gridHeight) % g.gridHeight

			cellIndex := neighborY*g.gridWidth + neighborX

			// Append all ships from this cell
			if ships, exists := g.cells[cellIndex]; exists {
				nearby = append(nearby, ships...)
			}
		}
	}

	return nearby
}

// clear resets the grid (call at start of each frame)
func (g *spatialGrid) clear() {
	// Clear all cells
	// Reusing the map is more efficient than recreating it
	for k := range g.cells {
		delete(g.cells, k)
	}
}
