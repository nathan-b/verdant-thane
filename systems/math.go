package systems

import (
	"math"

	"github.com/nathan/verdant-thane/config"
)

// NormalizeAngle brings an angle into the range [-π, π]
// This is used for angle comparisons and calculations throughout the game
func NormalizeAngle(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

// GetWrappedDistance returns the shortest distance components between two points,
// accounting for toroidal world wrapping. Returns (dx, dy) where each component
// represents the shortest distance considering wrap-around.
func GetWrappedDistance(x1, y1, x2, y2 float64) (float64, float64) {
	dx := x2 - x1
	dy := y2 - y1

	// Wrap X coordinate
	if dx > float64(config.GameWidth)/2 {
		dx -= float64(config.GameWidth)
	} else if dx < -float64(config.GameWidth)/2 {
		dx += float64(config.GameWidth)
	}

	// Wrap Y coordinate
	if dy > float64(config.GameHeight)/2 {
		dy -= float64(config.GameHeight)
	} else if dy < -float64(config.GameHeight)/2 {
		dy += float64(config.GameHeight)
	}

	return dx, dy
}

// Distance calculates the actual distance between two points, accounting for world wrapping
func Distance(x1, y1, x2, y2 float64) float64 {
	dx, dy := GetWrappedDistance(x1, y1, x2, y2)
	return math.Sqrt(dx*dx + dy*dy)
}

// DistanceSquared calculates the squared distance between two points, accounting for world wrapping.
// This is more efficient than Distance() when you only need to compare distances.
func DistanceSquared(x1, y1, x2, y2 float64) float64 {
	dx, dy := GetWrappedDistance(x1, y1, x2, y2)
	return dx*dx + dy*dy
}

// IsInRange checks if two positions are within a given range, accounting for toroidal wrapping
func IsInRange(x1, y1, x2, y2, maxRange float64) bool {
	dx, dy := GetWrappedDistance(x1, y1, x2, y2)
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance <= maxRange
}
