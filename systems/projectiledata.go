package systems

// ProjectileType identifies different projectile types
type ProjectileType int

const (
	LaserProjectile ProjectileType = iota
	MissileProjectile
)

// ProjectileCharacteristics defines the gameplay stats for a projectile type
type ProjectileCharacteristics struct {
	Type         ProjectileType
	Speed        float64 // Initial speed in pixels per tick
	Acceleration float64 // Acceleration per tick (0 for non-tracking projectiles)
	TurnRate     float64 // Maximum turn rate in radians per tick (0 = no limit)
	Lifetime     int     // Ticks before expiration
	ChargeTime   float64 // Seconds to fully charge weapon
	SpritePath   string  // Path to sprite asset
}

// ProjectileDatabase holds characteristics for all projectile types
var ProjectileDatabase = map[ProjectileType]ProjectileCharacteristics{
	LaserProjectile: {
		Type:         LaserProjectile,
		Speed:        12.0, // 2x max fighter speed
		Acceleration: 0.0,  // No acceleration (constant velocity)
		TurnRate:     0.0,  // No turning (straight line)
		Lifetime:     180,  // 3 seconds at 60 TPS
		ChargeTime:   0.6,  // 600ms charge time
		SpritePath:   "assets/laser.png",
	},
	MissileProjectile: {
		Type:         MissileProjectile,
		Speed:        4.0,               // Slow initial speed
		Acceleration: 0.15,              // Gradual acceleration toward target
		TurnRate:     3.0 * 0.017453292, // ~3 degrees per tick turn radius
		Lifetime:     150,               // 2.5 seconds at 60 TPS
		ChargeTime:   1.2,               // 1200ms charge time (half fire rate)
		SpritePath:   "assets/missile.png",
	},
}

// GetProjectileCharacteristics returns the characteristics for a projectile type
func GetProjectileCharacteristics(projType ProjectileType) ProjectileCharacteristics {
	chars, ok := ProjectileDatabase[projType]
	if !ok {
		// Fallback to laser if unknown type
		return ProjectileDatabase[LaserProjectile]
	}
	return chars
}
