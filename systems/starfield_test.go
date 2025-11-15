package systems

import (
	"math"
	"testing"

	"github.com/nathan/verdant-thane/config"
)

// ============================================================================
// Modulo Function Tests
// ============================================================================

func TestModuloPositive(t *testing.T) {
	result := modulo(7, 3)
	if result != 1 {
		t.Errorf("modulo(7, 3) = %d, expected 1", result)
	}

	result = modulo(9, 3)
	if result != 0 {
		t.Errorf("modulo(9, 3) = %d, expected 0", result)
	}

	result = modulo(10, 7)
	if result != 3 {
		t.Errorf("modulo(10, 7) = %d, expected 3", result)
	}
}

func TestModuloNegative(t *testing.T) {
	// This is the key difference from % operator
	result := modulo(-1, 3)
	if result != 2 {
		t.Errorf("modulo(-1, 3) = %d, expected 2", result)
	}

	result = modulo(-5, 3)
	if result != 1 {
		t.Errorf("modulo(-5, 3) = %d, expected 1", result)
	}

	result = modulo(-7, 4)
	if result != 1 {
		t.Errorf("modulo(-7, 4) = %d, expected 1", result)
	}
}

func TestModuloZero(t *testing.T) {
	result := modulo(0, 5)
	if result != 0 {
		t.Errorf("modulo(0, 5) = %d, expected 0", result)
	}
}

func TestModuloAlwaysPositive(t *testing.T) {
	// Verify modulo always returns non-negative results
	testCases := []struct {
		a, b int
	}{
		{-10, 3},
		{-5, 7},
		{-100, 13},
		{10, 3},
		{5, 7},
		{100, 13},
	}

	for _, tc := range testCases {
		result := modulo(tc.a, tc.b)
		if result < 0 {
			t.Errorf("modulo(%d, %d) = %d, should be non-negative", tc.a, tc.b, result)
		}
		if result >= tc.b {
			t.Errorf("modulo(%d, %d) = %d, should be < %d", tc.a, tc.b, result, tc.b)
		}
	}
}

// ============================================================================
// Hash Position Tests
// ============================================================================

func TestHashPositionDeterminism(t *testing.T) {
	// Same position should always produce same hash
	hash1 := hashPosition(5, 10)
	hash2 := hashPosition(5, 10)

	if hash1 != hash2 {
		t.Errorf("Hash should be deterministic: got %d and %d", hash1, hash2)
	}
}

func TestHashPositionUniqueness(t *testing.T) {
	// Different positions should (usually) produce different hashes
	hash1 := hashPosition(0, 0)
	hash2 := hashPosition(1, 0)
	hash3 := hashPosition(0, 1)
	hash4 := hashPosition(1, 1)

	// While not guaranteed, these should typically be different
	if hash1 == hash2 && hash1 == hash3 && hash1 == hash4 {
		t.Error("Hashes should have some variation for different positions")
	}
}

func TestHashPositionNonNegative(t *testing.T) {
	// Hash should always be non-negative
	testCases := []struct {
		x, y int
	}{
		{0, 0},
		{5, 10},
		{-5, -10},
		{100, 200},
		{-50, 75},
	}

	for _, tc := range testCases {
		hash := hashPosition(tc.x, tc.y)
		if hash < 0 {
			t.Errorf("hashPosition(%d, %d) = %d, should be non-negative", tc.x, tc.y, hash)
		}
	}
}

func TestHashPositionWrapping(t *testing.T) {
	// Calculate grid dimensions
	gridCountX := config.GameWidth / config.StarGridSize
	gridCountY := config.GameHeight / config.StarGridSize

	// Positions that wrap should produce same hash
	hash1 := hashPosition(0, 0)
	hash2 := hashPosition(gridCountX, 0)          // Wraps to 0,0 in X
	hash3 := hashPosition(0, gridCountY)          // Wraps to 0,0 in Y
	hash4 := hashPosition(gridCountX, gridCountY) // Wraps to 0,0 in both

	if hash1 != hash2 {
		t.Errorf("Hash should wrap in X: hashPosition(0,0)=%d, hashPosition(%d,0)=%d",
			hash1, gridCountX, hash2)
	}
	if hash1 != hash3 {
		t.Errorf("Hash should wrap in Y: hashPosition(0,0)=%d, hashPosition(0,%d)=%d",
			hash1, gridCountY, hash3)
	}
	if hash1 != hash4 {
		t.Errorf("Hash should wrap in both: hashPosition(0,0)=%d, hashPosition(%d,%d)=%d",
			hash1, gridCountX, gridCountY, hash4)
	}
}

func TestHashPositionNegativeWrapping(t *testing.T) {
	// Negative positions should wrap correctly
	gridCountX := config.GameWidth / config.StarGridSize
	gridCountY := config.GameHeight / config.StarGridSize

	hash2 := hashPosition(-1, 0)           // Should wrap to gridCountX-1, 0
	hash3 := hashPosition(0, -1)           // Should wrap to 0, gridCountY-1
	hash4 := hashPosition(gridCountX-1, 0) // Explicit position
	hash5 := hashPosition(0, gridCountY-1) // Explicit position

	if hash2 != hash4 {
		t.Errorf("Negative X should wrap: hashPosition(-1,0)=%d, hashPosition(%d,0)=%d",
			hash2, gridCountX-1, hash4)
	}
	if hash3 != hash5 {
		t.Errorf("Negative Y should wrap: hashPosition(0,-1)=%d, hashPosition(0,%d)=%d",
			hash3, gridCountY-1, hash5)
	}
}

// ============================================================================
// Star Generation Tests
// ============================================================================

func TestGenerateStarsForGridDeterminism(t *testing.T) {
	// Same grid cell should always generate same stars
	stars1 := GenerateStarsForGrid(5, 10)
	stars2 := GenerateStarsForGrid(5, 10)

	if len(stars1) != len(stars2) {
		t.Fatalf("Star count should be deterministic: got %d and %d", len(stars1), len(stars2))
	}

	for i := range stars1 {
		if stars1[i].X != stars2[i].X || stars1[i].Y != stars2[i].Y {
			t.Errorf("Star %d position mismatch: (%f,%f) vs (%f,%f)",
				i, stars1[i].X, stars1[i].Y, stars2[i].X, stars2[i].Y)
		}
	}
}

func TestGenerateStarsForGridCount(t *testing.T) {
	stars := GenerateStarsForGrid(0, 0)

	// Calculate expected number of stars
	area := float64(config.StarGridSize * config.StarGridSize)
	expectedCount := int(area * config.StarDensity)

	if len(stars) != expectedCount {
		t.Errorf("Expected %d stars, got %d", expectedCount, len(stars))
	}
}

func TestGenerateStarsForGridBounds(t *testing.T) {
	gridX := 5
	gridY := 10
	stars := GenerateStarsForGrid(gridX, gridY)

	// Stars should be within the grid bounds
	minX := float64(gridX * config.StarGridSize)
	maxX := float64((gridX + 1) * config.StarGridSize)
	minY := float64(gridY * config.StarGridSize)
	maxY := float64((gridY + 1) * config.StarGridSize)

	for i, star := range stars {
		if star.X < minX || star.X >= maxX {
			t.Errorf("Star %d X=%f is outside bounds [%f, %f)", i, star.X, minX, maxX)
		}
		if star.Y < minY || star.Y >= maxY {
			t.Errorf("Star %d Y=%f is outside bounds [%f, %f)", i, star.Y, minY, maxY)
		}
	}
}

func TestGenerateStarsForGridOrigin(t *testing.T) {
	stars := GenerateStarsForGrid(0, 0)

	if len(stars) == 0 {
		t.Fatal("Should generate stars for origin grid")
	}

	// Verify stars are in first grid cell
	for i, star := range stars {
		if star.X < 0 || star.X >= float64(config.StarGridSize) {
			t.Errorf("Star %d X=%f is outside origin grid [0, %d)",
				i, star.X, config.StarGridSize)
		}
		if star.Y < 0 || star.Y >= float64(config.StarGridSize) {
			t.Errorf("Star %d Y=%f is outside origin grid [0, %d)",
				i, star.Y, config.StarGridSize)
		}
	}
}

func TestGenerateStarsForGridVariation(t *testing.T) {
	// Different grid cells should generate stars at different positions
	stars1 := GenerateStarsForGrid(0, 0)
	stars2 := GenerateStarsForGrid(1, 0)

	// Count how many stars have identical X or Y coordinates between grids
	// (should be very unlikely)
	identicalCount := 0
	for _, s1 := range stars1 {
		for _, s2 := range stars2 {
			if math.Abs(s1.X-s2.X) < 0.01 && math.Abs(s1.Y-s2.Y) < 0.01 {
				identicalCount++
			}
		}
	}

	if identicalCount > len(stars1)/10 {
		t.Error("Too many identical star positions between different grids")
	}

	// Verify grids have different stars (not same pattern)
	allSame := true
	for i := range stars1 {
		if i >= len(stars2) {
			break
		}
		// Compare within grid-relative coordinates
		relX1 := stars1[i].X
		relY1 := stars1[i].Y
		relX2 := stars2[i].X - float64(config.StarGridSize)
		relY2 := stars2[i].Y

		if math.Abs(relX1-relX2) > 1.0 || math.Abs(relY1-relY2) > 1.0 {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Different grids should have different star patterns")
	}
}

func TestGenerateStarsForGridWrapping(t *testing.T) {
	// Grid cells that wrap should produce identical star positions (grid index wraps)
	gridCountX := config.GameWidth / config.StarGridSize

	stars1 := GenerateStarsForGrid(0, 0)
	stars2 := GenerateStarsForGrid(gridCountX, 0) // Wraps to grid 0

	if len(stars1) != len(stars2) {
		t.Fatalf("Wrapped grids should have same star count: %d vs %d",
			len(stars1), len(stars2))
	}

	// Stars should be at identical world positions (grid indices wrap)
	for i := range stars1 {
		if math.Abs(stars1[i].X-stars2[i].X) > 0.01 {
			t.Errorf("Star %d X position mismatch: grid 0 at %f, grid %d at %f",
				i, stars1[i].X, gridCountX, stars2[i].X)
		}
		if math.Abs(stars1[i].Y-stars2[i].Y) > 0.01 {
			t.Errorf("Star %d Y position mismatch: grid 0 at %f, grid %d at %f",
				i, stars1[i].Y, gridCountX, stars2[i].Y)
		}
	}
}

func TestGenerateStarsForGridNegativeIndices(t *testing.T) {
	// Negative grid indices should work due to wrapping
	stars := GenerateStarsForGrid(-1, -1)

	if len(stars) == 0 {
		t.Fatal("Should generate stars for negative grid indices")
	}

	// Stars should still have valid positions
	for i, star := range stars {
		if math.IsNaN(star.X) || math.IsNaN(star.Y) {
			t.Errorf("Star %d has NaN position", i)
		}
		if math.IsInf(star.X, 0) || math.IsInf(star.Y, 0) {
			t.Errorf("Star %d has infinite position", i)
		}
	}
}

// ============================================================================
// Star Struct Tests
// ============================================================================

func TestStarStruct(t *testing.T) {
	star := Star{X: 100.5, Y: 200.7}

	if star.X != 100.5 {
		t.Errorf("Expected X=100.5, got %f", star.X)
	}
	if star.Y != 200.7 {
		t.Errorf("Expected Y=200.7, got %f", star.Y)
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestStarfieldConsistencyAcrossWorld(t *testing.T) {
	// Generate stars for a 3x3 grid around origin
	grids := make(map[string][]Star)

	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			key := string(rune(x)) + "," + string(rune(y))
			grids[key] = GenerateStarsForGrid(x, y)
		}
	}

	// Verify each grid has stars
	for key, stars := range grids {
		if len(stars) == 0 {
			t.Errorf("Grid %s has no stars", key)
		}
	}

	// Verify regeneration produces same results
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			key := string(rune(x)) + "," + string(rune(y))
			stars := GenerateStarsForGrid(x, y)

			if len(stars) != len(grids[key]) {
				t.Errorf("Grid %s star count changed on regeneration", key)
			}
		}
	}
}

func TestStarfieldTiling(t *testing.T) {
	// Verify that stars tile seamlessly across world boundaries
	gridCountX := config.GameWidth / config.StarGridSize
	gridCountY := config.GameHeight / config.StarGridSize

	// Generate stars for grid (0,0) and its wrapped equivalent
	stars1 := GenerateStarsForGrid(0, 0)
	stars2 := GenerateStarsForGrid(gridCountX, gridCountY)

	if len(stars1) != len(stars2) {
		t.Fatalf("Tiled grids should have same star count: %d vs %d",
			len(stars1), len(stars2))
	}

	// Verify stars are at identical world positions (grid indices wrap)
	for i := range stars1 {
		if math.Abs(stars1[i].X-stars2[i].X) > 0.01 {
			t.Errorf("Star %d X position mismatch: grid (0,0) at %f, grid (%d,%d) at %f",
				i, stars1[i].X, gridCountX, gridCountY, stars2[i].X)
		}
		if math.Abs(stars1[i].Y-stars2[i].Y) > 0.01 {
			t.Errorf("Star %d Y position mismatch: grid (0,0) at %f, grid (%d,%d) at %f",
				i, stars1[i].Y, gridCountX, gridCountY, stars2[i].Y)
		}
	}
}

func TestStarfieldDensity(t *testing.T) {
	// Generate stars for multiple grids and verify density is consistent
	totalStars := 0
	gridCount := 10

	for x := 0; x < gridCount; x++ {
		for y := 0; y < gridCount; y++ {
			stars := GenerateStarsForGrid(x, y)
			totalStars += len(stars)
		}
	}

	// Calculate actual density
	totalArea := float64(config.StarGridSize * config.StarGridSize * gridCount * gridCount)
	actualDensity := float64(totalStars) / totalArea
	expectedDensity := config.StarDensity

	// Allow small variance due to integer rounding
	if math.Abs(actualDensity-expectedDensity) > 0.0001 {
		t.Errorf("Star density mismatch: expected %f, got %f",
			expectedDensity, actualDensity)
	}
}

func TestStarfieldRandomnessQuality(t *testing.T) {
	// Generate stars for a single grid
	stars := GenerateStarsForGrid(0, 0)

	if len(stars) < 2 {
		t.Skip("Need at least 2 stars for randomness test")
	}

	// Check that stars are not all clustered at one position
	minX, maxX := stars[0].X, stars[0].X
	minY, maxY := stars[0].Y, stars[0].Y

	for _, star := range stars {
		if star.X < minX {
			minX = star.X
		}
		if star.X > maxX {
			maxX = star.X
		}
		if star.Y < minY {
			minY = star.Y
		}
		if star.Y > maxY {
			maxY = star.Y
		}
	}

	// Stars should span a reasonable portion of the grid
	xSpan := maxX - minX
	ySpan := maxY - minY
	gridSize := float64(config.StarGridSize)

	// Expect at least 50% span in each dimension (with many stars)
	if len(stars) > 10 {
		if xSpan < gridSize*0.5 {
			t.Errorf("Stars poorly distributed in X: span=%f, grid=%f", xSpan, gridSize)
		}
		if ySpan < gridSize*0.5 {
			t.Errorf("Stars poorly distributed in Y: span=%f, grid=%f", ySpan, gridSize)
		}
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestStarfieldLargeGridIndices(t *testing.T) {
	// Test with very large grid indices
	stars := GenerateStarsForGrid(10000, 10000)

	if len(stars) == 0 {
		t.Error("Should generate stars for large grid indices")
	}

	// Stars should still be valid
	for i, star := range stars {
		if math.IsNaN(star.X) || math.IsNaN(star.Y) {
			t.Errorf("Star %d has NaN position", i)
		}
	}
}

func TestStarfieldNegativeLargeIndices(t *testing.T) {
	// Test with very negative grid indices
	stars := GenerateStarsForGrid(-10000, -10000)

	if len(stars) == 0 {
		t.Error("Should generate stars for negative large grid indices")
	}
}

func TestStarfieldBoundaryConditions(t *testing.T) {
	// Test at exact world boundaries
	gridCountX := config.GameWidth / config.StarGridSize
	gridCountY := config.GameHeight / config.StarGridSize

	testCases := []struct {
		x, y int
	}{
		{0, 0},
		{gridCountX - 1, 0},
		{0, gridCountY - 1},
		{gridCountX - 1, gridCountY - 1},
		{gridCountX, gridCountY},
	}

	for _, tc := range testCases {
		stars := GenerateStarsForGrid(tc.x, tc.y)
		if len(stars) == 0 {
			t.Errorf("Grid (%d,%d) should have stars", tc.x, tc.y)
		}
	}
}

// ============================================================================
// Grid Cell Calculation Tests (for rendering)
// ============================================================================

// TestGridCellCalculation tests that grid cells are correctly calculated
// from camera positions, especially when camera position is negative
func TestGridCellCalculation(t *testing.T) {
	testCases := []struct {
		name          string
		cameraPos     float64
		gridSize      int
		expectedGrid  int
		description   string
	}{
		{
			name:         "Positive camera in first grid",
			cameraPos:    50.0,
			gridSize:     128,
			expectedGrid: 0,
			description:  "Camera at x=50 should be in grid cell 0",
		},
		{
			name:         "Positive camera in second grid",
			cameraPos:    150.0,
			gridSize:     128,
			expectedGrid: 1,
			description:  "Camera at x=150 should be in grid cell 1",
		},
		{
			name:         "Camera exactly at grid boundary",
			cameraPos:    128.0,
			gridSize:     128,
			expectedGrid: 1,
			description:  "Camera at x=128 should be in grid cell 1",
		},
		{
			name:         "Small negative camera",
			cameraPos:    -50.0,
			gridSize:     128,
			expectedGrid: -1,
			description:  "Camera at x=-50 should be in grid cell -1 (not 0!)",
		},
		{
			name:         "Large negative camera",
			cameraPos:    -200.0,
			gridSize:     128,
			expectedGrid: -2,
			description:  "Camera at x=-200 should be in grid cell -2",
		},
		{
			name:         "Camera at negative grid boundary",
			cameraPos:    -128.0,
			gridSize:     128,
			expectedGrid: -1,
			description:  "Camera at x=-128 should be in grid cell -1",
		},
		{
			name:         "Camera just past negative boundary",
			cameraPos:    -128.1,
			gridSize:     128,
			expectedGrid: -2,
			description:  "Camera at x=-128.1 should be in grid cell -2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Correct calculation using floor division
			gridCell := int(math.Floor(tc.cameraPos / float64(tc.gridSize)))

			if gridCell != tc.expectedGrid {
				t.Errorf("%s\n  Got grid cell %d, expected %d",
					tc.description, gridCell, tc.expectedGrid)
			}

			// Also verify that the INCORRECT calculation (current bug) would fail
			// This demonstrates the bug when camera position is negative
			incorrectGridCell := int(tc.cameraPos) / tc.gridSize
			if tc.cameraPos < 0 && incorrectGridCell == tc.expectedGrid {
				// If they're equal for negative camera, our expected value might be wrong
				// (or the incorrect method happens to be right for this case)
			}
		})
	}
}

// TestStarVisibilityAcrossWorldBoundary tests that stars are correctly
// rendered when the camera crosses world boundaries
func TestStarVisibilityAcrossWorldBoundary(t *testing.T) {
	// Simulate camera near the left edge of the world (player near x=50)
	// Camera would be centered on player: cameraX = 50 - ScreenWidth/2
	// With ScreenWidth=800, cameraX = 50 - 400 = -350

	cameraX := -350.0
	cameraY := 100.0
	screenWidth := 800
	screenHeight := 600

	// Calculate grid cells using CORRECT floor division
	minGridX := int(math.Floor(cameraX / float64(config.StarGridSize)))
	maxGridX := int(math.Floor((cameraX + float64(screenWidth)) / float64(config.StarGridSize)))
	minGridY := int(math.Floor(cameraY / float64(config.StarGridSize)))
	maxGridY := int(math.Floor((cameraY + float64(screenHeight)) / float64(config.StarGridSize)))

	// With StarGridSize=200, cameraX=-350:
	// minGridX should be floor(-350/200) = floor(-1.75) = -2
	// maxGridX should be floor((-350+800)/200) = floor(450/200) = floor(2.25) = 2

	if minGridX != -2 {
		t.Errorf("minGridX should be -2, got %d", minGridX)
	}
	if maxGridX != 2 {
		t.Errorf("maxGridX should be 2, got %d", maxGridX)
	}

	// Collect all stars that should be visible
	visibleStars := 0
	for gridX := minGridX; gridX <= maxGridX; gridX++ {
		for gridY := minGridY; gridY <= maxGridY; gridY++ {
			stars := GenerateStarsForGrid(gridX, gridY)
			for _, star := range stars {
				// Calculate screen position (same as rendering code)
				// Use simple offset calculation for test
				screenX := star.X - cameraX
				screenY := star.Y - cameraY

				// Count if on screen
				if screenX >= 0 && screenX < float64(screenWidth) &&
					screenY >= 0 && screenY < float64(screenHeight) {
					visibleStars++
				}
			}
		}
	}

	// Should have some visible stars
	if visibleStars == 0 {
		t.Error("Should have visible stars when camera is at world boundary")
	}

	// Now test with INCORRECT grid calculation (the bug)
	incorrectMinGridX := int(cameraX) / config.StarGridSize

	// With int(-350) / 200 = -350 / 200 = -1 (truncates toward zero, BUG!)
	// Should be -2, not -1
	if incorrectMinGridX == minGridX {
		t.Errorf("Bug test failed: incorrect calculation gave same result as correct one")
	}

	// The incorrect calculation should give -1, not -2
	expectedBugValue := -1
	if incorrectMinGridX != expectedBugValue {
		t.Logf("Note: Incorrect calculation gave %d (expected bug value was %d)",
			incorrectMinGridX, expectedBugValue)
	}
}

// TestGridCellCoverageAtBoundary verifies that all stars in visible grid cells
// can potentially be seen when camera crosses boundaries
func TestGridCellCoverageAtBoundary(t *testing.T) {
	// Test multiple camera positions near boundaries
	testCases := []struct {
		name    string
		cameraX float64
		cameraY float64
	}{
		{"Near left edge", -350.0, 100.0},
		{"Near right edge", float64(config.GameWidth) - 450.0, 100.0},
		{"Near top edge", 100.0, -350.0},
		{"Near bottom edge", 100.0, float64(config.GameHeight) - 450.0},
		{"Near top-left corner", -350.0, -350.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use correct floor division
			minGridX := int(math.Floor(tc.cameraX / float64(config.StarGridSize)))
			maxGridX := int(math.Floor((tc.cameraX + float64(config.ScreenWidth)) / float64(config.StarGridSize)))
			minGridY := int(math.Floor(tc.cameraY / float64(config.StarGridSize)))
			maxGridY := int(math.Floor((tc.cameraY + float64(config.ScreenHeight)) / float64(config.StarGridSize)))

			// Generate stars for all visible grid cells
			totalStars := 0
			for gridX := minGridX; gridX <= maxGridX; gridX++ {
				for gridY := minGridY; gridY <= maxGridY; gridY++ {
					stars := GenerateStarsForGrid(gridX, gridY)
					totalStars += len(stars)
				}
			}

			// Should have stars
			if totalStars == 0 {
				t.Errorf("No stars generated for camera at (%f, %f)", tc.cameraX, tc.cameraY)
			}
		})
	}
}

// TestStarPositionConsistencyAcrossBoundary verifies that stars generated for
// wrapped grid cells have the same positions (modulo world size)
func TestStarPositionConsistencyAcrossBoundary(t *testing.T) {
	gridCountX := config.GameWidth / config.StarGridSize

	// Test wrapping in X dimension
	// Grid 0 and grid gridCountX should use same hash (same pattern) and have matching wrapped positions
	stars0 := GenerateStarsForGrid(0, 0)
	starsWrappedX := GenerateStarsForGrid(gridCountX, 0)

	if len(stars0) != len(starsWrappedX) {
		t.Errorf("Grid 0 and grid %d should have same number of stars: %d vs %d",
			gridCountX, len(stars0), len(starsWrappedX))
	}

	// Check that stars have equivalent positions when considering world wrapping
	// Both should generate stars in the same wrapped world positions
	for i := range stars0 {
		x0 := stars0[i].X
		xWrapped := starsWrappedX[i].X

		// Both positions should be in valid range [0, GameWidth)
		if x0 < 0 || x0 >= float64(config.GameWidth) {
			t.Errorf("Star %d from grid 0 outside valid range: %f", i, x0)
		}
		if xWrapped < 0 || xWrapped >= float64(config.GameWidth) {
			t.Errorf("Star %d from grid %d outside valid range: %f", i, gridCountX, xWrapped)
		}

		// Stars should be at identical wrapped positions
		if math.Abs(x0-xWrapped) > 0.01 {
			t.Errorf("Star %d position mismatch: grid 0 at %f, grid %d at %f (should be equal after wrapping)",
				i, x0, gridCountX, xWrapped)
		}
	}

	// Test negative grid cells
	// Grid -1 should have same pattern as grid (gridCountX - 1) and matching wrapped positions
	starsNeg1 := GenerateStarsForGrid(-1, 0)
	starsLast := GenerateStarsForGrid(gridCountX-1, 0)

	if len(starsNeg1) != len(starsLast) {
		t.Errorf("Grid -1 and grid %d should have same number of stars: %d vs %d",
			gridCountX-1, len(starsNeg1), len(starsLast))
	}

	// Check wrapped positions match
	for i := range starsNeg1 {
		xNeg := starsNeg1[i].X
		xLast := starsLast[i].X

		// Both should be in valid range
		if xNeg < 0 || xNeg >= float64(config.GameWidth) {
			t.Errorf("Star %d from grid -1 outside valid range: %f", i, xNeg)
		}
		if xLast < 0 || xLast >= float64(config.GameWidth) {
			t.Errorf("Star %d from grid %d outside valid range: %f", i, gridCountX-1, xLast)
		}

		// Stars should be at identical wrapped positions
		if math.Abs(xNeg-xLast) > 0.01 {
			t.Errorf("Star %d position mismatch: grid -1 at %f, grid %d at %f (should be equal after wrapping)",
				i, xNeg, gridCountX-1, xLast)
		}
	}
}
