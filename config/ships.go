package config

// ShipClass defines the type of ship (duplicated to avoid circular import with entity package)
type ShipClass int

const (
	ClassFighter ShipClass = iota
	ClassDestroyer
	ClassTestudon
)

// ShipCharacteristics defines the gameplay stats for a ship class
type ShipCharacteristics struct {
	MaxSpeed            float64 // Pixels per tick
	Acceleration        float64 // Speed increase per tick
	MaxShield           int     // Hit points
	CapacitorChargeRate float64 // Charge per tick (1.0 = fully charged)
	FiringCone          float64 // Radians (for main weapon)
	CollisionRadius     float64 // Collision detection radius in pixels
	BaseSpritePath      string  // Path to grayscale base sprite
}

// ShipDatabase holds characteristics for all ship classes
var ShipDatabase = map[ShipClass]ShipCharacteristics{
	ClassFighter: {
		MaxSpeed:            6.0,                             // Standard speed
		Acceleration:        4.0 / 60.0,                      // Pixels per second per tick
		MaxShield:           8,                               // Light armor
		CapacitorChargeRate: 1.0 / ((600.0 / 1000.0) * 60.0), // 600ms charge time
		FiringCone:          30.0 * 0.017453292519943295,     // 30 degrees in radians
		CollisionRadius:     14.0,                            // 24x24 sprite, slightly larger than half-width for better gameplay
		BaseSpritePath:      "assets/fighter.png",
	},
	ClassDestroyer: {
		MaxSpeed:            5.0,                             // Slower than fighter
		Acceleration:        3.0 / 60.0,                      // Slower acceleration
		MaxShield:           32,                              // Heavy armor (4x fighter)
		CapacitorChargeRate: 1.0 / ((600.0 / 1000.0) * 60.0), // Same as fighter for main gun
		FiringCone:          30.0 * 0.017453292519943295,     // 30 degrees in radians
		CollisionRadius:     30.0,                            // 40x60 sprite, approximate average radius
		BaseSpritePath:      "assets/destroyer.png",
	},
	ClassTestudon: {
		MaxSpeed:            3.0,        // Very slow battleship
		Acceleration:        2.0 / 60.0, // Slow acceleration
		MaxShield:           100,        // Heavy armor (12.5x fighter)
		CapacitorChargeRate: 0.0,        // No projectile weapon
		FiringCone:          0.0,        // No firing cone (360° beam weapon)
		CollisionRadius:     50.0,       // 100x100 sprite, half-width for circular hitbox
		BaseSpritePath:      "assets/testudon.png",
	},
}

// GetShipCharacteristics returns the characteristics for a ship class
// Panics if ship class is not found (should never happen in production)
func GetShipCharacteristics(class ShipClass) ShipCharacteristics {
	chars, ok := ShipDatabase[class]
	if !ok {
		// Fallback to fighter if unknown class
		return ShipDatabase[ClassFighter]
	}
	return chars
}
