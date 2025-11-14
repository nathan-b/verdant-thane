package entity

import (
	"math"

	"github.com/nathan/verdant-thane/config"
)

// NormalizeAngle brings an angle into the range [-π, π]
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
// accounting for toroidal world wrapping
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

// DistanceSquared calculates the squared distance between two points, accounting for world wrapping
func DistanceSquared(x1, y1, x2, y2 float64) float64 {
	dx, dy := GetWrappedDistance(x1, y1, x2, y2)
	return dx*dx + dy*dy
}

// IsInRange checks if two positions are within a given range, accounting for toroidal wrapping
func IsInRange(x1, y1, x2, y2, maxRange float64) bool {
	return DistanceSquared(x1, y1, x2, y2) <= maxRange*maxRange
}

// WrapPosition wraps a position to stay within the game world bounds
func WrapPosition(x, y float64) (float64, float64) {
	for x < 0 {
		x += float64(config.GameWidth)
	}
	for x >= float64(config.GameWidth) {
		x -= float64(config.GameWidth)
	}
	for y < 0 {
		y += float64(config.GameHeight)
	}
	for y >= float64(config.GameHeight) {
		y -= float64(config.GameHeight)
	}
	return x, y
}

// GetWrappedScreenPosition calculates the screen position of an entity,
// accounting for world wrapping. Returns the primary screen position.
// If the entity should be visible at multiple wrapped positions, this returns
// the position closest to the camera center.
func GetWrappedScreenPosition(entityX, entityY, cameraX, cameraY float64) (float64, float64) {
	// Use wrapped distance to get shortest path from camera to entity
	dx, dy := GetWrappedDistance(cameraX, cameraY, entityX, entityY)

	// Screen position is just the wrapped offset from camera
	return dx, dy
}
