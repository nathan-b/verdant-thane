package systems

import "github.com/nathan-b/verdant-thane/config"

// modulo performs proper modulo operation (handles negatives correctly)
func modulo(a, b int) int {
	return ((a % b) + b) % b
}

// hashPosition creates a deterministic hash for a grid position
// Wraps grid coordinates to ensure consistent stars across world boundaries
func hashPosition(gridX, gridY int) int {
	// Calculate number of grid cells in the game world
	gridCountX := config.GameWidth / config.StarGridSize
	gridCountY := config.GameHeight / config.StarGridSize

	// Wrap grid coordinates to ensure tiling
	wrappedX := modulo(gridX, gridCountX)
	wrappedY := modulo(gridY, gridCountY)

	// Simple hash function for deterministic random generation
	h := wrappedX*73856093 ^ wrappedY*19349663
	if h < 0 {
		h = -h
	}
	return h
}

// Star represents a star position in world coordinates
type Star struct {
	X, Y float64
}

// GenerateStarsForGrid generates stars for a specific grid cell
// Returns a slice of star positions in world coordinates
func GenerateStarsForGrid(gridX, gridY int) []Star {
	stars := []Star{}

	// Use hash as seed for this grid cell (hash handles wrapping internally)
	seed := hashPosition(gridX, gridY)

	// Determine number of stars in this grid cell
	area := float64(config.StarGridSize * config.StarGridSize)
	numStars := int(area * config.StarDensity)

	// Calculate grid count (how many grids fit in the world)
	gridCountX := config.GameWidth / config.StarGridSize
	gridCountY := config.GameHeight / config.StarGridSize

	// Wrap grid indices to ensure seamless tiling
	// This handles cases where GameWidth/Height don't evenly divide by StarGridSize
	wrappedGridX := modulo(gridX, gridCountX)
	wrappedGridY := modulo(gridY, gridCountY)

	// Calculate base position using wrapped grid indices
	baseX := float64(wrappedGridX * config.StarGridSize)
	baseY := float64(wrappedGridY * config.StarGridSize)

	// Generate deterministic "random" positions within this grid
	for i := 0; i < numStars; i++ {
		// Simple LCG (Linear Congruential Generator) for deterministic randomness
		seed = (seed*1103515245 + 12345) & 0x7fffffff
		offsetX := float64(seed % config.StarGridSize)

		seed = (seed*1103515245 + 12345) & 0x7fffffff
		offsetY := float64(seed % config.StarGridSize)

		// Star position in world coordinates (base already wrapped, add offset)
		x := baseX + offsetX
		y := baseY + offsetY

		// Wrap again in case offset pushes us past the boundary
		for x >= float64(config.GameWidth) {
			x -= float64(config.GameWidth)
		}
		for y >= float64(config.GameHeight) {
			y -= float64(config.GameHeight)
		}

		stars = append(stars, Star{X: x, Y: y})
	}

	return stars
}
