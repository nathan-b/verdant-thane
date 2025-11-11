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
