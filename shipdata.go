package main

import (
	"github.com/nathan/verdant-thane/components"
)

// ShipCharacteristics defines the gameplay stats for a ship class
type ShipCharacteristics struct {
	Class               components.ShipClass
	MaxSpeed            float64 // Pixels per tick
	Acceleration        float64 // Speed increase per tick
	MaxShield           int     // Hit points
	CapacitorChargeRate float64 // Charge per tick (1.0 = fully charged)
	FiringCone          float64 // Radians (for main weapon)
	BaseSpritePath      string  // Path to grayscale base sprite
}

// ShipDatabase holds characteristics for all ship classes
var ShipDatabase = map[components.ShipClass]ShipCharacteristics{
	components.Fighter: {
		Class:               components.Fighter,
		MaxSpeed:            6.0,                             // Standard speed
		Acceleration:        4.0 / 60.0,                      // Pixels per second per tick
		MaxShield:           8,                               // Light armor
		CapacitorChargeRate: 1.0 / ((600.0 / 1000.0) * 60.0), // 600ms charge time
		FiringCone:          30.0 * 0.017453292519943295,     // 30 degrees in radians
		BaseSpritePath:      "assets/fighter.png",
	},
	components.Destroyer: {
		Class:               components.Destroyer,
		MaxSpeed:            5.0,                             // Slower than fighter
		Acceleration:        3.0 / 60.0,                      // Slower acceleration
		MaxShield:           32,                              // Heavy armor (4x fighter)
		CapacitorChargeRate: 1.0 / ((600.0 / 1000.0) * 60.0), // Same as fighter for main gun
		FiringCone:          30.0 * 0.017453292519943295,     // 30 degrees in radians
		BaseSpritePath:      "assets/destroyer_gray.png",
	},
}

// GetShipCharacteristics returns the characteristics for a ship class
// Panics if ship class is not found (should never happen in production)
func GetShipCharacteristics(class components.ShipClass) ShipCharacteristics {
	chars, ok := ShipDatabase[class]
	if !ok {
		// Fallback to fighter if unknown class
		return ShipDatabase[components.Fighter]
	}
	return chars
}
