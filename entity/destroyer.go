package entity

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/nathan-b/verdant-thane/config"
)

// NewDestroyer creates a new destroyer ship
func NewDestroyer(id int, factionID int, x, y float64, sprite *ebiten.Image) *Ship {
	ship := newShipBase(id, factionID, ClassDestroyer, x, y, sprite)
	chars := config.GetShipCharacteristics(ClassDestroyer)

	// Configure destroyer-specific weapons (missile launcher + main gun)
	// Weapons stored in priority order (lower priority number = higher priority = earlier in array)
	ship.Weapons = []Weapon{
		{
			WeaponCapacitor:  1.0, // Missile launcher - start fully charged
			WeaponChargeRate: chars.Weapons[1].CapacitorChargeRate,
			FiringCone:       chars.Weapons[1].FiringCone,
			MaxRange:         chars.Weapons[1].MaxRange,
			SpawnOffset:      chars.Weapons[1].SpawnOffset,
			Priority:         1,    // Higher priority than main gun
			RequiresTarget:   true, // Missiles require target lock
			Exclusive:        true, // When missiles fire, don't fire other weapons
			RearFacing:       true, // Fires from rear
			ProjectileType:   config.MissileProjectile,
		},
		{
			WeaponCapacitor:  1.0, // Main gun - start fully charged
			WeaponChargeRate: chars.Weapons[0].CapacitorChargeRate,
			FiringCone:       chars.Weapons[0].FiringCone,
			MaxRange:         chars.Weapons[0].MaxRange,
			SpawnOffset:      chars.Weapons[0].SpawnOffset,
			Priority:         2,
			RequiresTarget:   false,
			Exclusive:        false,
			RearFacing:       false,
			ProjectileType:   config.LaserProjectile,
		},
	}

	// Configure afterburner (destroyers have afterburner)
	ship.AfterburnerCharge = 360.0 // Start fully charged
	ship.AfterburnerActive = false
	ship.AfterburnerMaxCharge = 360.0
	ship.HasAfterburnerSystem = true
	ship.AfterburnerDrain = chars.AfterburnerDrain
	ship.AfterburnerRecharge = chars.AfterburnerRecharge
	ship.AfterburnerAccelMultiplier = chars.AfterburnerAccelMultiplier
	ship.AfterburnerMaxSpeedMultiplier = chars.AfterburnerMaxSpeedMultiplier

	// Beam fields unused by destroyers (zero values from base)
	// AttackerIDs unused by destroyers (nil from base)

	return ship
}
