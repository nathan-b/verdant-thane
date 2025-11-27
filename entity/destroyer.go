package entity

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/nathan-b/verdant-thane/config"
)

// NewDestroyer creates a new destroyer ship
func NewDestroyer(id int, factionID int, x, y float64, sprite *ebiten.Image) *Ship {
	chars := config.GetShipCharacteristics(ClassDestroyer)

	base := &Ship{
		ID:               id,
		FactionID:        factionID,
		Class:            ClassDestroyer,
		X:                x,
		Y:                y,
		VelocityX:        0,
		VelocityY:        0,
		Rotation:         0,
		Health:           chars.MaxShield,
		MaxHealth:        chars.MaxShield,
		Speed:            0,
		MaxSpeed:         chars.MaxSpeed,
		Accel:            chars.Acceleration,
		CollisionRadius:  chars.CollisionRadius,
		PlayerControlled: false,
		// Weapons stored in priority order (lower priority number = higher priority = earlier in array)
		Weapons: []Weapon{
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
		},
		AfterburnerCharge:             360.0, // Start fully charged
		AfterburnerActive:             false,
		AfterburnerMaxCharge:          360.0,
		HasAfterburnerSystem:          true, // Destroyers have afterburner
		AfterburnerDrain:              chars.AfterburnerDrain,
		AfterburnerRecharge:           chars.AfterburnerRecharge,
		AfterburnerAccelMultiplier:    chars.AfterburnerAccelMultiplier,
		AfterburnerMaxSpeedMultiplier: chars.AfterburnerMaxSpeedMultiplier,
		AITargetID:                    -1,
		AIRetargetTimer:               config.AIRetargetInterval,
		AIAccurateShotProbability:     chars.AIAccurateShotProbability,
		AIRandomShotProbability:       chars.AIRandomShotProbability,
		KillScore:                     chars.KillScore,
		Sprite:                        sprite,
		Alive:                         true,
		BeamRange:                     0.0, // Unused by destroyers
		BeamDamagePerTick:             0.0, // Unused by destroyers
		BeamDamageAccumulator:         0.0, // Unused by destroyers
		BeamTargetID:                  -1,  // Unused by destroyers
		BeamFiringAtID:                -1,  // Unused by destroyers
		AttackerIDs:                   nil, // Unused by destroyers
	}

	return base
}
