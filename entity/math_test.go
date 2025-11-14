package entity

import (
	"math"
	"testing"

	"github.com/nathan/verdant-thane/config"
)

func TestNormalizeAngle(t *testing.T) {
	tests := []struct {
		name     string
		angle    float64
		expected float64
	}{
		{"Zero angle", 0, 0},
		{"Positive angle within range", math.Pi / 2, math.Pi / 2},
		{"Negative angle within range", -math.Pi / 2, -math.Pi / 2},
		{"At positive boundary", math.Pi, math.Pi},
		{"At negative boundary", -math.Pi, -math.Pi},
		{"Just over positive boundary", math.Pi + 0.1, -math.Pi + 0.1},
		{"Just under negative boundary", -math.Pi - 0.1, math.Pi - 0.1},
		{"Two full rotations positive", 4 * math.Pi, 0},
		{"Two full rotations negative", -4 * math.Pi, 0},
		{"Large positive angle", 10 * math.Pi, 0},
		{"Large negative angle", -10 * math.Pi, 0},
		{"Large positive non-zero", 10*math.Pi + math.Pi/4, math.Pi / 4},
		{"Large negative non-zero", -10*math.Pi - math.Pi/4, -math.Pi / 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAngle(tt.angle)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("NormalizeAngle(%f) = %f, expected %f", tt.angle, result, tt.expected)
			}
		})
	}
}

func TestWrapPosition(t *testing.T) {
	tests := []struct {
		name                 string
		x, y                 float64
		expectedX, expectedY float64
	}{
		{"Origin", 0, 0, 0, 0},
		{"Center", float64(config.GameWidth) / 2, float64(config.GameHeight) / 2,
			float64(config.GameWidth) / 2, float64(config.GameHeight) / 2},
		{"Just inside bounds", 100, 100, 100, 100},

		// X wrapping
		{"Negative X", -10, 100, float64(config.GameWidth) - 10, 100},
		{"Large negative X", -100, 100, float64(config.GameWidth) - 100, 100},
		{"At X boundary", float64(config.GameWidth), 100, 0, 100},
		{"Just over X boundary", float64(config.GameWidth) + 10, 100, 10, 100},
		{"Large positive X", float64(config.GameWidth) + 100, 100, 100, 100},
		{"Multiple X wraps positive", float64(config.GameWidth)*2 + 50, 100, 50, 100},
		{"Multiple X wraps negative", float64(-config.GameWidth)*2 - 50, 100, float64(config.GameWidth) - 50, 100},

		// Y wrapping
		{"Negative Y", 100, -10, 100, float64(config.GameHeight) - 10},
		{"Large negative Y", 100, -100, 100, float64(config.GameHeight) - 100},
		{"At Y boundary", 100, float64(config.GameHeight), 100, 0},
		{"Just over Y boundary", 100, float64(config.GameHeight) + 10, 100, 10},
		{"Large positive Y", 100, float64(config.GameHeight) + 100, 100, 100},
		{"Multiple Y wraps positive", 100, float64(config.GameHeight)*2 + 50, 100, 50},
		{"Multiple Y wraps negative", 100, float64(-config.GameHeight)*2 - 50, 100, float64(config.GameHeight) - 50},

		// Both axes wrapping
		{"Both negative", -10, -20, float64(config.GameWidth) - 10, float64(config.GameHeight) - 20},
		{"Both over boundary", float64(config.GameWidth) + 10, float64(config.GameHeight) + 20, 10, 20},
		{"X negative Y positive over", -10, float64(config.GameHeight) + 20, float64(config.GameWidth) - 10, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultX, resultY := WrapPosition(tt.x, tt.y)
			if math.Abs(resultX-tt.expectedX) > 1e-9 || math.Abs(resultY-tt.expectedY) > 1e-9 {
				t.Errorf("WrapPosition(%f, %f) = (%f, %f), expected (%f, %f)",
					tt.x, tt.y, resultX, resultY, tt.expectedX, tt.expectedY)
			}
		})
	}
}

func TestGetWrappedDistance(t *testing.T) {
	tests := []struct {
		name                   string
		x1, y1, x2, y2         float64
		expectedDx, expectedDy float64
	}{
		{"Same position", 100, 100, 100, 100, 0, 0},
		{"Simple distance", 100, 100, 200, 150, 100, 50},
		{"Negative distance", 200, 150, 100, 100, -100, -50},

		// X wrapping - closer to go across boundary
		{"Wrap across right edge", 100, 100, float64(config.GameWidth) - 100, 100,
			-200, 0}, // Shorter to go left across boundary
		{"Wrap across left edge", float64(config.GameWidth) - 100, 100, 100, 100,
			200, 0}, // Shorter to go right across boundary

		// Y wrapping - closer to go across boundary
		{"Wrap across bottom edge", 100, 100, 100, float64(config.GameHeight) - 100,
			0, -200}, // Shorter to go up across boundary
		{"Wrap across top edge", 100, float64(config.GameHeight) - 100, 100, 100,
			0, 200}, // Shorter to go down across boundary

		// No wrapping - direct path is shorter
		{"No wrap needed X", 100, 100, 1000, 100, 900, 0},
		{"No wrap needed Y", 100, 100, 100, 1000, 0, 900},

		// At exactly half width/height - should not wrap
		{"At half width", 100, 100, 100 + float64(config.GameWidth)/2, 100,
			float64(config.GameWidth) / 2, 0},
		{"At half height", 100, 100, 100, 100 + float64(config.GameHeight)/2,
			0, float64(config.GameHeight) / 2},

		// Both axes wrapping
		{"Both wrap", 100, 100, float64(config.GameWidth) - 100, float64(config.GameHeight) - 100,
			-200, -200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dx, dy := GetWrappedDistance(tt.x1, tt.y1, tt.x2, tt.y2)
			if math.Abs(dx-tt.expectedDx) > 1e-9 || math.Abs(dy-tt.expectedDy) > 1e-9 {
				t.Errorf("GetWrappedDistance(%f, %f, %f, %f) = (%f, %f), expected (%f, %f)",
					tt.x1, tt.y1, tt.x2, tt.y2, dx, dy, tt.expectedDx, tt.expectedDy)
			}
		})
	}
}

func TestDistance(t *testing.T) {
	tests := []struct {
		name           string
		x1, y1, x2, y2 float64
		expected       float64
	}{
		{"Same position", 100, 100, 100, 100, 0},
		{"Simple pythagorean", 0, 0, 3, 4, 5}, // 3-4-5 triangle
		{"Negative coordinates", -3, 0, 0, 4, 5},

		// Wrapping cases
		{"Wrap X boundary", 100, 100, float64(config.GameWidth) - 100, 100, 200},
		{"Wrap Y boundary", 100, 100, 100, float64(config.GameHeight) - 100, 200},
		{"Wrap both", 100, 100, float64(config.GameWidth) - 100, float64(config.GameHeight) - 100,
			math.Sqrt(200*200 + 200*200)},

		// Diagonal across center (no wrapping)
		{"Long diagonal", 0, 0, 1000, 1000, math.Sqrt(1000*1000 + 1000*1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Distance(tt.x1, tt.y1, tt.x2, tt.y2)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Distance(%f, %f, %f, %f) = %f, expected %f",
					tt.x1, tt.y1, tt.x2, tt.y2, result, tt.expected)
			}
		})
	}
}

func TestDistanceSquared(t *testing.T) {
	tests := []struct {
		name           string
		x1, y1, x2, y2 float64
		expected       float64
	}{
		{"Same position", 100, 100, 100, 100, 0},
		{"Simple pythagorean", 0, 0, 3, 4, 25}, // 3^2 + 4^2 = 25
		{"Negative coordinates", -3, 0, 0, 4, 25},

		// Wrapping cases
		{"Wrap X boundary", 100, 100, float64(config.GameWidth) - 100, 100, 200 * 200},
		{"Wrap Y boundary", 100, 100, 100, float64(config.GameHeight) - 100, 200 * 200},
		{"Wrap both", 100, 100, float64(config.GameWidth) - 100, float64(config.GameHeight) - 100,
			200*200 + 200*200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DistanceSquared(tt.x1, tt.y1, tt.x2, tt.y2)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("DistanceSquared(%f, %f, %f, %f) = %f, expected %f",
					tt.x1, tt.y1, tt.x2, tt.y2, result, tt.expected)
			}
		})
	}
}

func TestIsInRange(t *testing.T) {
	tests := []struct {
		name           string
		x1, y1, x2, y2 float64
		maxRange       float64
		expected       bool
	}{
		{"Same position in range", 100, 100, 100, 100, 10, true},
		{"Just inside range", 0, 0, 3, 4, 5, true}, // distance = 5
		{"At exact range", 0, 0, 3, 4, 5, true},
		{"Just outside range", 0, 0, 3, 4, 4.9, false},
		{"Far outside range", 0, 0, 1000, 1000, 100, false},

		// Wrapping cases
		{"In range with X wrap", 100, 100, float64(config.GameWidth) - 100, 100, 250, true},
		{"Out of range with X wrap", 100, 100, float64(config.GameWidth) - 100, 100, 150, false},
		{"In range with Y wrap", 100, 100, 100, float64(config.GameHeight) - 100, 250, true},
		{"Out of range with Y wrap", 100, 100, 100, float64(config.GameHeight) - 100, 150, false},

		// Edge case: exactly at half-world distance
		{"At half world X", 0, 0, float64(config.GameWidth) / 2, 0,
			float64(config.GameWidth)/2 + 10, true},
		{"At half world Y", 0, 0, 0, float64(config.GameHeight) / 2,
			float64(config.GameHeight)/2 + 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInRange(tt.x1, tt.y1, tt.x2, tt.y2, tt.maxRange)
			if result != tt.expected {
				actualDist := Distance(tt.x1, tt.y1, tt.x2, tt.y2)
				t.Errorf("IsInRange(%f, %f, %f, %f, %f) = %v, expected %v (actual distance: %f)",
					tt.x1, tt.y1, tt.x2, tt.y2, tt.maxRange, result, tt.expected, actualDist)
			}
		})
	}
}

func TestGetWrappedScreenPosition(t *testing.T) {
	tests := []struct {
		name                     string
		entityX, entityY         float64
		cameraX, cameraY         float64
		expectedScreenX, screenY float64
	}{
		// No wrapping cases
		{
			name:            "Entity and camera at same position",
			entityX:         1000,
			entityY:         1000,
			cameraX:         1000,
			cameraY:         1000,
			expectedScreenX: 0,
			screenY:         0,
		},
		{
			name:            "Entity directly in front of camera",
			entityX:         1000,
			entityY:         1000,
			cameraX:         500,
			cameraY:         500,
			expectedScreenX: 500,
			screenY:         500,
		},
		{
			name:            "Entity behind camera (negative offset)",
			entityX:         500,
			entityY:         500,
			cameraX:         1000,
			cameraY:         1000,
			expectedScreenX: -500,
			screenY:         -500,
		},
		{
			name:            "Entity to the right of camera",
			entityX:         2500,
			entityY:         1000,
			cameraX:         2000,
			cameraY:         1000,
			expectedScreenX: 500,
			screenY:         0,
		},
		{
			name:            "Entity below camera",
			entityX:         1000,
			entityY:         2500,
			cameraX:         1000,
			cameraY:         2000,
			expectedScreenX: 0,
			screenY:         500,
		},

		// Horizontal wrapping cases
		{
			name:    "Entity near left edge, camera near right edge (wrap right to left)",
			entityX: 100,
			entityY: 2500,
			cameraX: float64(config.GameWidth) - 100, // Near right edge
			cameraY: 2500,
			// Wrapped distance: 100 - (5040-100) = 100 - 4940 = -4840
			// But wrapping: dx > GameWidth/2, so dx -= GameWidth = -4840 - (-5040) = 200
			// Wait, let me recalculate. dx = 100 - 4940 = -4840
			// Since dx < -GameWidth/2 (-2520), we add GameWidth: -4840 + 5040 = 200
			expectedScreenX: 200,
			screenY:         0,
		},
		{
			name:    "Entity near right edge, camera near left edge (wrap left to right)",
			entityX: float64(config.GameWidth) - 100, // Near right edge
			entityY: 2500,
			cameraX: 100,
			cameraY: 2500,
			// dx = (5040-100) - 100 = 4840
			// Since dx > GameWidth/2 (2520), we subtract GameWidth: 4840 - 5040 = -200
			expectedScreenX: -200,
			screenY:         0,
		},

		// Vertical wrapping cases
		{
			name:    "Entity near top edge, camera near bottom edge (wrap bottom to top)",
			entityX: 2500,
			entityY: 100,
			cameraX: 2500,
			cameraY: float64(config.GameHeight) - 100, // Near bottom edge
			// dy = 100 - (5040-100) = 100 - 4940 = -4840
			// Since dy < -GameHeight/2 (-2520), we add GameHeight: -4840 + 5040 = 200
			expectedScreenX: 0,
			screenY:         200,
		},
		{
			name:    "Entity near bottom edge, camera near top edge (wrap top to bottom)",
			entityX: 2500,
			entityY: float64(config.GameHeight) - 100, // Near bottom edge
			cameraX: 2500,
			cameraY: 100,
			// dy = (5040-100) - 100 = 4840
			// Since dy > GameHeight/2 (2520), we subtract GameHeight: 4840 - 5040 = -200
			expectedScreenX: 0,
			screenY:         -200,
		},

		// Corner wrapping cases (both axes)
		{
			name:    "Entity at top-left corner, camera at bottom-right corner",
			entityX: 100,
			entityY: 100,
			cameraX: float64(config.GameWidth) - 100,
			cameraY: float64(config.GameHeight) - 100,
			// dx: 100 - 4940 = -4840, wrapped = 200
			// dy: 100 - 4940 = -4840, wrapped = 200
			expectedScreenX: 200,
			screenY:         200,
		},
		{
			name:    "Entity at bottom-right corner, camera at top-left corner",
			entityX: float64(config.GameWidth) - 100,
			entityY: float64(config.GameHeight) - 100,
			cameraX: 100,
			cameraY: 100,
			// dx: 4940 - 100 = 4840, wrapped = -200
			// dy: 4940 - 100 = 4840, wrapped = -200
			expectedScreenX: -200,
			screenY:         -200,
		},
		{
			name:    "Entity at top-right corner, camera at bottom-left corner",
			entityX: float64(config.GameWidth) - 100,
			entityY: 100,
			cameraX: 100,
			cameraY: float64(config.GameHeight) - 100,
			// dx: 4940 - 100 = 4840, wrapped = -200
			// dy: 100 - 4940 = -4840, wrapped = 200
			expectedScreenX: -200,
			screenY:         200,
		},

		// Edge cases at world boundaries
		{
			name:            "Entity at world origin, camera at world origin",
			entityX:         0,
			entityY:         0,
			cameraX:         0,
			cameraY:         0,
			expectedScreenX: 0,
			screenY:         0,
		},
		{
			name:    "Entity at world max, camera at world origin",
			entityX: float64(config.GameWidth) - 1,
			entityY: float64(config.GameHeight) - 1,
			cameraX: 0,
			cameraY: 0,
			// dx: (5040-1) - 0 = 5039, wrapped = -1 (wraps around)
			// dy: (5040-1) - 0 = 5039, wrapped = -1
			expectedScreenX: -1,
			screenY:         -1,
		},

		// Cases at exactly half-world distance (boundary between wrapping/not wrapping)
		{
			name:    "At exactly half world width apart",
			entityX: 0,
			entityY: 1000,
			cameraX: float64(config.GameWidth) / 2,
			cameraY: 1000,
			// dx: 0 - 2520 = -2520 (exactly at boundary, should not wrap)
			expectedScreenX: -float64(config.GameWidth) / 2,
			screenY:         0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screenX, screenY := GetWrappedScreenPosition(tt.entityX, tt.entityY, tt.cameraX, tt.cameraY)
			if math.Abs(screenX-tt.expectedScreenX) > 1e-9 || math.Abs(screenY-tt.screenY) > 1e-9 {
				t.Errorf("GetWrappedScreenPosition(entity: %f,%f, camera: %f,%f) = (%f, %f), expected (%f, %f)",
					tt.entityX, tt.entityY, tt.cameraX, tt.cameraY,
					screenX, screenY, tt.expectedScreenX, tt.screenY)
			}
		})
	}
}

// Test that GetWrappedScreenPosition is consistent with GetWrappedDistance
func TestGetWrappedScreenPositionConsistency(t *testing.T) {
	testPositions := []struct {
		entityX, entityY float64
		cameraX, cameraY float64
	}{
		{100, 100, 200, 200},
		{4900, 100, 100, 200},
		{100, 4900, 200, 100},
		{4900, 4900, 100, 100},
		{2520, 2520, 100, 100},
		{0, 0, 5039, 5039},
	}

	for _, pos := range testPositions {
		t.Run("", func(t *testing.T) {
			// GetWrappedScreenPosition should return the same values as GetWrappedDistance
			screenX, screenY := GetWrappedScreenPosition(pos.entityX, pos.entityY, pos.cameraX, pos.cameraY)
			dx, dy := GetWrappedDistance(pos.cameraX, pos.cameraY, pos.entityX, pos.entityY)

			if math.Abs(screenX-dx) > 1e-9 || math.Abs(screenY-dy) > 1e-9 {
				t.Errorf("GetWrappedScreenPosition and GetWrappedDistance inconsistent:\n"+
					"  entity: (%f, %f), camera: (%f, %f)\n"+
					"  GetWrappedScreenPosition: (%f, %f)\n"+
					"  GetWrappedDistance: (%f, %f)",
					pos.entityX, pos.entityY, pos.cameraX, pos.cameraY,
					screenX, screenY, dx, dy)
			}
		})
	}
}

// Benchmark tests
func BenchmarkDistance(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Distance(100, 100, 200, 200)
	}
}

func BenchmarkDistanceSquared(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DistanceSquared(100, 100, 200, 200)
	}
}

func BenchmarkWrapPosition(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WrapPosition(float64(config.GameWidth)+100, float64(config.GameHeight)+100)
	}
}

func BenchmarkGetWrappedScreenPosition(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetWrappedScreenPosition(100, 100, 4900, 4900)
	}
}
