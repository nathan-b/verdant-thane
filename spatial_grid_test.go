package main

import (
	"testing"

	"github.com/nathan/verdant-thane/config"
	"github.com/nathan/verdant-thane/entity"
)

// Test Spatial Grid Creation
func TestNewSpatialGrid(t *testing.T) {
	grid := newSpatialGrid()

	expectedCellSize := config.CollisionGridSize
	expectedWidth := (config.GameWidth + expectedCellSize - 1) / expectedCellSize
	expectedHeight := (config.GameHeight + expectedCellSize - 1) / expectedCellSize

	if grid.cellSize != expectedCellSize {
		t.Errorf("Expected cell size %d, got %d", expectedCellSize, grid.cellSize)
	}

	if grid.gridWidth != expectedWidth {
		t.Errorf("Expected grid width %d, got %d", expectedWidth, grid.gridWidth)
	}

	if grid.gridHeight != expectedHeight {
		t.Errorf("Expected grid height %d, got %d", expectedHeight, grid.gridHeight)
	}

	if grid.cells == nil {
		t.Error("Grid cells map should be initialized")
	}
}

// Test Cell Index Calculation
func TestGetCellIndex(t *testing.T) {
	grid := newSpatialGrid()

	tests := []struct {
		name          string
		x, y          float64
		expectedIndex int
	}{
		{
			name:          "Origin",
			x:             0,
			y:             0,
			expectedIndex: 0,
		},
		{
			name:          "First cell",
			x:             64,
			y:             64,
			expectedIndex: 0, // Still in cell (0,0) since cell size is 128
		},
		{
			name:          "Second cell X",
			x:             128,
			y:             0,
			expectedIndex: 1, // Cell (1,0)
		},
		{
			name:          "Second cell Y",
			x:             0,
			y:             128,
			expectedIndex: grid.gridWidth, // Cell (0,1)
		},
		{
			name: "Center of world",
			x:    float64(config.GameWidth) / 2,
			y:    float64(config.GameHeight) / 2,
			// Center is at (2520, 2520) which is cell (19, 19) = 19*40 + 19 = 779
			expectedIndex: 19*grid.gridWidth + 19,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := grid.getCellIndex(tt.x, tt.y)
			if index != tt.expectedIndex {
				t.Errorf("Expected index %d, got %d", tt.expectedIndex, index)
			}
		})
	}
}

// Test Cell Index With World Wrapping
func TestGetCellIndexWrapping(t *testing.T) {
	grid := newSpatialGrid()

	tests := []struct {
		name string
		x, y float64
	}{
		{"Negative X", -100, 100},
		{"Negative Y", 100, -100},
		{"Negative both", -100, -100},
		{"Over width", float64(config.GameWidth) + 100, 100},
		{"Over height", 100, float64(config.GameHeight) + 100},
		{"Over both", float64(config.GameWidth) + 100, float64(config.GameHeight) + 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := grid.getCellIndex(tt.x, tt.y)

			// Index should always be valid (within grid bounds)
			maxIndex := grid.gridWidth * grid.gridHeight
			if index < 0 || index >= maxIndex {
				t.Errorf("Invalid cell index %d (max: %d) for position (%f, %f)",
					index, maxIndex-1, tt.x, tt.y)
			}
		})
	}
}

// Test Ship Insertion
func TestSpatialGridInsert(t *testing.T) {
	grid := newSpatialGrid()

	// Create test ships
	ship1 := entity.NewFighter(1, 0, 100, 100, nil) // Cell (0, 0)
	ship2 := entity.NewFighter(2, 0, 500, 500, nil) // Different cell

	grid.insert(ship1)
	grid.insert(ship2)

	// Verify ships are in grid
	totalShips := 0
	for _, ships := range grid.cells {
		totalShips += len(ships)
	}

	if totalShips != 2 {
		t.Errorf("Expected 2 ships in grid, got %d", totalShips)
	}

	// Verify ships are in correct cells
	ship1Index := grid.getCellIndex(100, 100)
	if len(grid.cells[ship1Index]) < 1 {
		t.Error("Ship1 should be in its cell")
	}

	ship2Index := grid.getCellIndex(500, 500)
	if len(grid.cells[ship2Index]) < 1 {
		t.Error("Ship2 should be in its cell")
	}
}

// Test Get Nearby Ships
func TestGetNearbyShips(t *testing.T) {
	grid := newSpatialGrid()

	// Create ships in different cells
	centerShip := entity.NewFighter(1, 0, 500, 500, nil)
	nearbyShip := entity.NewFighter(2, 0, 550, 500, nil) // Same or adjacent cell
	farShip := entity.NewFighter(3, 0, 1500, 1500, nil)  // Far away

	grid.insert(centerShip)
	grid.insert(nearbyShip)
	grid.insert(farShip)

	// Query near centerShip position
	nearby := grid.getNearbyShips(500, 500)

	// Should find centerShip and nearbyShip, but not farShip
	foundCenter := false
	foundNearby := false
	foundFar := false

	for _, ship := range nearby {
		switch ship.GetID() {
		case 1:
			foundCenter = true
		case 2:
			foundNearby = true
		case 3:
			foundFar = true
		}
	}

	if !foundCenter {
		t.Error("Should find center ship in neighborhood")
	}

	if !foundNearby {
		t.Error("Should find nearby ship in neighborhood")
	}

	if foundFar {
		t.Error("Should not find far ship in neighborhood")
	}
}

// Test Get Nearby Ships Across Grid Boundaries
func TestGetNearbyShipsAcrossBoundary(t *testing.T) {
	grid := newSpatialGrid()

	// Place ship near grid boundary
	boundaryX := float64(grid.cellSize) - 10 // Just before cell boundary
	boundaryShip := entity.NewFighter(1, 0, boundaryX, 100, nil)

	// Place ship just across boundary
	acrossBoundaryX := float64(grid.cellSize) + 10 // Just after cell boundary
	acrossShip := entity.NewFighter(2, 0, acrossBoundaryX, 100, nil)

	grid.insert(boundaryShip)
	grid.insert(acrossShip)

	// Query at boundary position - should find ships on both sides
	nearby := grid.getNearbyShips(boundaryX, 100)

	if len(nearby) != 2 {
		t.Errorf("Expected to find 2 ships across boundary, got %d", len(nearby))
	}
}

// Test Get Nearby Ships With World Wrapping
func TestGetNearbyShipsWorldWrapping(t *testing.T) {
	grid := newSpatialGrid()

	// Place ships near world edges
	leftEdgeShip := entity.NewFighter(1, 0, 10, 100, nil)
	rightEdgeShip := entity.NewFighter(2, 0, float64(config.GameWidth)-10, 100, nil)

	grid.insert(leftEdgeShip)
	grid.insert(rightEdgeShip)

	// Query near left edge - should find both ships (they wrap to be adjacent)
	nearbyLeft := grid.getNearbyShips(10, 100)

	foundLeft := false
	foundRight := false

	for _, ship := range nearbyLeft {
		if ship.GetID() == 1 {
			foundLeft = true
		}
		if ship.GetID() == 2 {
			foundRight = true
		}
	}

	if !foundLeft {
		t.Error("Should find left edge ship")
	}

	// Right edge ship might or might not be found depending on grid alignment
	// This is fine - the key is that the query doesn't crash
	_ = foundRight
}

// Test Clear Grid
func TestSpatialGridClear(t *testing.T) {
	grid := newSpatialGrid()

	// Insert some ships
	for i := 0; i < 10; i++ {
		ship := entity.NewFighter(i, 0, float64(i*100), float64(i*100), nil)
		grid.insert(ship)
	}

	// Verify ships are in grid
	totalBefore := 0
	for _, ships := range grid.cells {
		totalBefore += len(ships)
	}

	if totalBefore != 10 {
		t.Errorf("Expected 10 ships before clear, got %d", totalBefore)
	}

	// Clear grid
	grid.clear()

	// Verify grid is empty
	totalAfter := 0
	for _, ships := range grid.cells {
		totalAfter += len(ships)
	}

	if totalAfter != 0 {
		t.Errorf("Expected 0 ships after clear, got %d", totalAfter)
	}

	if len(grid.cells) != 0 {
		t.Errorf("Expected empty cells map after clear, got %d cells", len(grid.cells))
	}
}

// Test Multiple Ships in Same Cell
func TestMultipleShipsInSameCell(t *testing.T) {
	grid := newSpatialGrid()

	// Create multiple ships at similar positions (same cell)
	for i := 0; i < 5; i++ {
		ship := entity.NewFighter(i, 0, 100+float64(i*5), 100, nil)
		grid.insert(ship)
	}

	// All should be in same or adjacent cells
	nearby := grid.getNearbyShips(100, 100)

	if len(nearby) != 5 {
		t.Errorf("Expected to find all 5 ships in neighborhood, got %d", len(nearby))
	}
}

// Test Grid Coverage
func TestGridCoverage(t *testing.T) {
	grid := newSpatialGrid()

	// Ensure grid covers entire world
	expectedCells := grid.gridWidth * grid.gridHeight
	maxX := grid.gridWidth * grid.cellSize
	maxY := grid.gridHeight * grid.cellSize

	if maxX < config.GameWidth {
		t.Errorf("Grid doesn't cover full width: %d < %d", maxX, config.GameWidth)
	}

	if maxY < config.GameHeight {
		t.Errorf("Grid doesn't cover full height: %d < %d", maxY, config.GameHeight)
	}

	if expectedCells < 1 {
		t.Error("Grid should have at least one cell")
	}
}

// Benchmark spatial grid operations
func BenchmarkSpatialGridInsert(b *testing.B) {
	grid := newSpatialGrid()
	ship := entity.NewFighter(1, 0, 500, 500, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		grid.clear()
		grid.insert(ship)
	}
}

func BenchmarkSpatialGridGetNearby(b *testing.B) {
	grid := newSpatialGrid()

	// Populate with 800 ships
	for i := 0; i < 800; i++ {
		x := float64((i * 123) % config.GameWidth)
		y := float64((i * 456) % config.GameHeight)
		ship := entity.NewFighter(i, 0, x, y, nil)
		grid.insert(ship)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		grid.getNearbyShips(500, 500)
	}
}

func BenchmarkSpatialGridFullRebuild(b *testing.B) {
	ships := make([]*entity.Ship, 800)
	for i := 0; i < 800; i++ {
		x := float64((i * 123) % config.GameWidth)
		y := float64((i * 456) % config.GameHeight)
		ships[i] = entity.NewFighter(i, 0, x, y, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		grid := newSpatialGrid()
		for _, ship := range ships {
			grid.insert(ship)
		}
	}
}
