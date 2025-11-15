package config

// ShipClass defines the type of ship (duplicated to avoid circular import with entity package)
type ShipClass int

const (
	ClassFighter ShipClass = iota
	ClassDestroyer
	ClassTestudon
)

type WeaponCharacteristics struct {
	CapacitorChargeRate float64 // Charge per tick (1.0 = fully charged)
	FiringCone          float64 // Radians
	MaxRange            float64 // Maximum effective range in pixels (0 = unlimited)
	SpawnOffset         float64 // Distance from ship center where projectiles spawn
	DamagePerTick       float64 // For beam weapons: damage dealt per tick (0 for projectile weapons)
}

// WeaponDatabase holds named weapon configurations that can be referenced by ships
var WeaponDatabase = map[string]WeaponCharacteristics{
	"main_gun": {
		CapacitorChargeRate: 1.0 / ((600.0 / 1000.0) * 60.0), // 600ms charge time
		FiringCone:          30.0 * 0.017453292519943295,     // 30 degrees in radians
		MaxRange:            0.0,                             // Unlimited range
		SpawnOffset:         20.0,                            // Forward spawn position
		DamagePerTick:       0.0,                             // Projectile weapon, not beam
	},
	"missile": {
		CapacitorChargeRate: 1.0 / ((2000.0 / 1000.0) * 60.0), // 2000ms charge time
		FiringCone:          180.0 * 0.017453292519943295,     // Rear 180 degrees
		MaxRange:            1200.0,                           // Missile lock-on range
		SpawnOffset:         25.0,                             // Rear spawn position (destroyer is larger)
		DamagePerTick:       0.0,                              // Projectile weapon, not beam
	},
	"beam": {
		CapacitorChargeRate: 0.0,   // Beam weapons don't use capacitor charging
		FiringCone:          360.0, // 360-degree firing arc (testudon can fire in any direction)
		MaxRange:            200.0, // 200 pixel effective range
		SpawnOffset:         0.0,   // Beam originates from ship center
		DamagePerTick:       0.125, // 0.125 damage per tick (8 ticks = 1 HP)
	},
}

// ShipCharacteristics defines the gameplay stats for a ship class
type ShipCharacteristics struct {
	MaxSpeed                   float64                 // Pixels per tick
	Acceleration               float64                 // Speed increase per tick
	MaxShield                  int                     // Hit points
	Weapons                    []WeaponCharacteristics // Ship weapons
	CollisionRadius            float64                 // Collision detection radius in pixels
	BaseSpritePath             string                  // Path to grayscale base sprite
	KillScore                  int                     // Points awarded for destroying this ship
	AfterburnerDrain           float64                 // Afterburner fuel consumed per tick
	AfterburnerRecharge        float64                 // Afterburner recharge rate per tick
	AfterburnerAccelMultiplier float64                 // Acceleration multiplier when afterburner active
	AIAccurateShotProbability  float64                 // Probability AI fires accurate shot (0.0-1.0)
	AIRandomShotProbability    float64                 // Probability AI fires random shot within cone (0.0-1.0)
}

// ShipDatabase holds characteristics for all ship classes
var ShipDatabase = map[ShipClass]ShipCharacteristics{
	ClassFighter: {
		MaxSpeed:     6.0,        // Standard speed
		Acceleration: 4.0 / 60.0, // Pixels per second per tick
		MaxShield:    8,          // Light armor
		Weapons: []WeaponCharacteristics{
			WeaponDatabase["main_gun"],
		},
		CollisionRadius:            14.0, // 24x24 sprite, slightly larger than half-width for better gameplay
		BaseSpritePath:             "assets/fighter.png",
		KillScore:                  10,   // Points awarded for kill
		AfterburnerDrain:           3.0,  // Fuel consumed per tick when active
		AfterburnerRecharge:        1.0,  // Recharge rate per tick when inactive
		AfterburnerAccelMultiplier: 2.0,  // 2x acceleration when active
		AIAccurateShotProbability:  0.50, // 50% chance of accurate shot
		AIRandomShotProbability:    0.25, // 25% chance of random shot (0.50-0.75 range)
	},
	ClassDestroyer: {
		MaxSpeed:     5.0,        // Slower than fighter
		Acceleration: 3.0 / 60.0, // Slower acceleration
		MaxShield:    32,         // Heavy armor (4x fighter)
		Weapons: []WeaponCharacteristics{
			WeaponDatabase["main_gun"],
			WeaponDatabase["missile"],
		},
		CollisionRadius:            30.0, // 40x60 sprite, approximate average radius
		BaseSpritePath:             "assets/destroyer.png",
		KillScore:                  50,
		AfterburnerDrain:           3.0,  // Fuel consumed per tick when active
		AfterburnerRecharge:        1.0,  // Recharge rate per tick when inactive
		AfterburnerAccelMultiplier: 2.0,  // 2x acceleration when active
		AIAccurateShotProbability:  0.50, // 50% chance of accurate shot
		AIRandomShotProbability:    0.25, // 25% chance of random shot (0.50-0.75 range)
	},
	ClassTestudon: {
		MaxSpeed:                   3.0,                       // Very slow battleship
		Acceleration:               2.0 / 60.0,                // Slow acceleration
		MaxShield:                  100,                       // Heavy armor (12.5x fighter)
		Weapons:                    []WeaponCharacteristics{}, // No weapons (beam weapon handled separately for now)
		CollisionRadius:            50.0,                      // 100x100 sprite, half-width for circular hitbox
		BaseSpritePath:             "assets/testudon.png",
		KillScore:                  100,
		AfterburnerDrain:           0.0,
		AfterburnerRecharge:        0.0,
		AfterburnerAccelMultiplier: 1.0,  // Testudon has no afterburner
		AIAccurateShotProbability:  0.50, // 50% chance of accurate shot
		AIRandomShotProbability:    0.25, // 25% chance of random shot (0.50-0.75 range)
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
