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
	// IMPORTANT: Floating point errors can accumulate over time with certain values.
	//
	// Values like 0.1 (1/10) can cause floating point drift:
	//   0.1 is not exactly representable in binary (repeating decimal in binary)
	//   After 10 additions: 0.1*10 = 0.999999... (requires 11 ticks for 1 damage!)
	//   This creates a ~10% DPS reduction from intended design.
	//
	// Safe values are powers of 2: 0.25 (1/4), 0.125 (1/8), 0.0625 (1/16)
	DamagePerTick float64 // For beam weapons: damage dealt per tick (0 for projectile weapons)
}

// WeaponDatabase holds named weapon configurations that can be referenced by ships
var WeaponDatabase = map[string]WeaponCharacteristics{
	"main_gun": {
		CapacitorChargeRate: 6.0 / 60.0,                  // 6 shots per second
		FiringCone:          90.0 * 0.017453292519943295, // 90 degrees in radians
		MaxRange:            0.0,                         // Unlimited range
		SpawnOffset:         20.0,                        // Forward spawn position
		DamagePerTick:       0.0,                         // Projectile weapon, not beam
	},
	"missile": {
		CapacitorChargeRate: 1.0 / ((2000.0 / 1000.0) * 60.0), // 2000ms charge time
		FiringCone:          180.0 * 0.017453292519943295,     // Rear 180 degrees
		MaxRange:            1200.0,                           // Missile lock-on range
		SpawnOffset:         25.0,                             // Rear spawn position (destroyer is larger)
		DamagePerTick:       0.0,                              // Projectile weapon, not beam
	},
	"beam": {
		CapacitorChargeRate: 0.0,    // Beam weapons don't use capacitor charging
		FiringCone:          360.0,  // 360-degree firing arc (testudon can fire in any direction)
		MaxRange:            200.0,  // 200 pixel effective range
		SpawnOffset:         0.0,    // Beam originates from ship center
		DamagePerTick:       0.0625, // 1/16 damage per tick (3.75 damage per second)
	},
}

// ShipCharacteristics defines the gameplay stats for a ship class
type ShipCharacteristics struct {
	MaxSpeed                      float64                 // Pixels per tick
	Acceleration                  float64                 // Speed increase per tick
	MaxShield                     int                     // Hit points
	Weapons                       []WeaponCharacteristics // Ship weapons
	CollisionRadius               float64                 // Collision detection radius in pixels
	BaseSpritePath                string                  // Path to grayscale base sprite
	KillScore                     int                     // Points awarded for destroying this ship
	AfterburnerDrain              float64                 // Afterburner fuel consumed per tick
	AfterburnerRecharge           float64                 // Afterburner recharge rate per tick
	AfterburnerAccelMultiplier    float64                 // Acceleration multiplier when afterburner active
	AfterburnerMaxSpeedMultiplier float64                 // Max speed multiplier when afterburner active
	AIAccurateShotProbability     float64                 // Probability AI fires accurate shot (0.0-1.0)
	AIRandomShotProbability       float64                 // Probability AI fires random shot within cone (0.0-1.0)
}

// ShipDatabase holds characteristics for all ship classes
var ShipDatabase = map[ShipClass]ShipCharacteristics{
	ClassFighter: {
		MaxSpeed:     5.0,        // Standard speed
		Acceleration: 4.0 / 60.0, // Pixels per second per tick
		MaxShield:    8,          // Light armor
		Weapons: []WeaponCharacteristics{
			WeaponDatabase["main_gun"],
		},
		CollisionRadius:               14.0, // 24x24 sprite, slightly larger than half-width for better gameplay
		BaseSpritePath:                "assets/fighter.png",
		KillScore:                     10,   // Points awarded for kill
		AfterburnerDrain:              3.0,  // Fuel consumed per tick when active
		AfterburnerRecharge:           1.0,  // Recharge rate per tick when inactive
		AfterburnerAccelMultiplier:    2.0,  // 2x acceleration when active
		AfterburnerMaxSpeedMultiplier: 1.5,  // 1.5x max speed when active
		AIAccurateShotProbability:     0.40, // 40% chance of accurate shot
		AIRandomShotProbability:       0.20, // 20% chance of random shot (0.40-0.60 range)
	},
	ClassDestroyer: {
		MaxSpeed:     4.0,        // Slower than fighter
		Acceleration: 3.0 / 60.0, // Slower acceleration
		MaxShield:    32,         // Heavy armor (4x fighter)
		Weapons: []WeaponCharacteristics{
			WeaponDatabase["main_gun"],
			WeaponDatabase["missile"],
		},
		CollisionRadius:               30.0, // 40x60 sprite, approximate average radius
		BaseSpritePath:                "assets/destroyer.png",
		KillScore:                     50,
		AfterburnerDrain:              3.0,  // Fuel consumed per tick when active
		AfterburnerRecharge:           1.0,  // Recharge rate per tick when inactive
		AfterburnerAccelMultiplier:    2.0,  // 2x acceleration when active
		AfterburnerMaxSpeedMultiplier: 1.5,  // 1.5x max speed when active
		AIAccurateShotProbability:     0.40, // 40% chance of accurate shot
		AIRandomShotProbability:       0.20, // 20% chance of random shot (0.40-0.60 range)
	},
	ClassTestudon: {
		MaxSpeed:                      3.0,                       // Very slow battleship
		Acceleration:                  2.0 / 60.0,                // Slow acceleration
		MaxShield:                     100,                       // Heavy armor (12.5x fighter)
		Weapons:                       []WeaponCharacteristics{}, // No weapons (beam weapon handled separately for now)
		CollisionRadius:               50.0,                      // 100x100 sprite, half-width for circular hitbox
		BaseSpritePath:                "assets/testudon.png",
		KillScore:                     100,
		AfterburnerDrain:              0.0,
		AfterburnerRecharge:           0.0,
		AfterburnerAccelMultiplier:    1.0,  // Testudon has no afterburner
		AfterburnerMaxSpeedMultiplier: 1.0,  // Testudon has no afterburner
		AIAccurateShotProbability:     0.40, // 40% chance of accurate shot
		AIRandomShotProbability:       0.20, // 20% chance of random shot (0.40-0.60 range)
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

// OverrideFighterFiringCone overrides the fighter's main gun firing cone
// degreesValue should be between 1 and 180
func OverrideFighterFiringCone(degreesValue int) {
	// Convert degrees to radians
	radians := float64(degreesValue) * 0.017453292519943295

	// Update the weapon database
	mainGun := WeaponDatabase["main_gun"]
	mainGun.FiringCone = radians
	WeaponDatabase["main_gun"] = mainGun

	// Update the fighter ship database (which references the weapon database)
	fighter := ShipDatabase[ClassFighter]
	fighter.Weapons[0].FiringCone = radians
	ShipDatabase[ClassFighter] = fighter
}

// OverrideFighterFiringRate overrides the fighter's main gun firing rate
// shotsPerSecond should be between 1 and 10
// This calculates the capacitor charge rate needed to achieve the desired shots per second
func OverrideFighterFiringRate(shotsPerSecond float64) {
	// At 60 ticks per second:
	// - N shots per second means 1/N seconds per shot
	// - That's (1/N) * 60 ticks per shot
	// - Capacitor charge rate = 1.0 / ((1/N) * 60) = N / 60
	chargeRate := shotsPerSecond / 60.0

	// Update the weapon database
	mainGun := WeaponDatabase["main_gun"]
	mainGun.CapacitorChargeRate = chargeRate
	WeaponDatabase["main_gun"] = mainGun

	// Update the fighter ship database (which references the weapon database)
	fighter := ShipDatabase[ClassFighter]
	fighter.Weapons[0].CapacitorChargeRate = chargeRate
	ShipDatabase[ClassFighter] = fighter
}

// OverrideFighterTurnRate overrides the fighter's turn rate
// degreesPerTick should be between 1 and 10
// This converts degrees to radians and updates the RotationSpeed constant
func OverrideFighterTurnRate(degreesPerTick float64) {
	// Convert degrees to radians
	radians := degreesPerTick * 0.017453292519943295
	RotationSpeed = radians
}

// OverrideFighterMaxSpeed overrides the fighter's maximum speed
// pixelsPerTick should be between 4 and 10
func OverrideFighterMaxSpeed(pixelsPerTick float64) {
	// Update the fighter ship database
	fighter := ShipDatabase[ClassFighter]
	fighter.MaxSpeed = pixelsPerTick
	ShipDatabase[ClassFighter] = fighter
}

// OverrideFighterAcceleration overrides the fighter's acceleration
// pixelsPerSecond should be between 2 and 10
// This converts pixels per second to pixels per tick (at 60 TPS)
func OverrideFighterAcceleration(pixelsPerSecond float64) {
	// Convert pixels per second to pixels per tick (60 TPS)
	pixelsPerTick := pixelsPerSecond / 60.0

	// Update the fighter ship database
	fighter := ShipDatabase[ClassFighter]
	fighter.Acceleration = pixelsPerTick
	ShipDatabase[ClassFighter] = fighter
}

// OverrideAIFireProbability overrides the AI firing probability for all ship classes
// fireProbability should be between 0.1 and 1.0
// This maintains the 2:1 ratio between accurate shots and random shots
func OverrideAIFireProbability(fireProbability float64) {
	// Split fire probability: 2/3 accurate, 1/3 random (maintains 2:1 ratio)
	accurateProb := fireProbability * (2.0 / 3.0)
	randomProb := fireProbability * (1.0 / 3.0)

	// Update all ship classes
	for class, chars := range ShipDatabase {
		chars.AIAccurateShotProbability = accurateProb
		chars.AIRandomShotProbability = randomProb
		ShipDatabase[class] = chars
	}
}
