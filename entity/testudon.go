package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nathan/verdant-thane/config"
)

// NewTestudon creates a new testudon ship
func NewTestudon(id int, factionID int, x, y float64, sprite *ebiten.Image) *Ship {
	chars := config.GetShipCharacteristics(ClassTestudon)
	beamWeapon := config.WeaponDatabase["beam"]

	base := &Ship{
		ID:                            id,
		FactionID:                     factionID,
		Class:                         ClassTestudon,
		X:                             x,
		Y:                             y,
		VelocityX:                     0,
		VelocityY:                     0,
		Rotation:                      0,
		Health:                        chars.MaxShield,
		MaxHealth:                     chars.MaxShield,
		Speed:                         0,
		MaxSpeed:                      chars.MaxSpeed,
		Accel:                         chars.Acceleration,
		CollisionRadius:               chars.CollisionRadius,
		PlayerControlled:              false,
		Weapons:                       []Weapon{}, // Testudons don't use projectile weapons (beam weapon only)
		AfterburnerCharge:             0.0,
		AfterburnerActive:             false,
		AfterburnerMaxCharge:          0.0,
		HasAfterburnerSystem:          false, // Testudons do NOT have afterburner
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
		BeamRange:                     beamWeapon.MaxRange,
		// IMPORTANT: BeamDamagePerTick must be exactly representable in binary floating point
		// to avoid accumulation errors that affect game balance.
		//
		// 0.125 = 1/8, which is exactly representable in binary (2^-3).
		// This means 8 ticks will accumulate to exactly 1.0 damage with no rounding error.
		//
		// DO NOT use values like 0.1 (1/10) which causes floating point drift:
		//   0.1 is not exactly representable in binary (repeating decimal in binary)
		//   After 10 additions: 0.1*10 = 0.999999... (requires 11 ticks for 1 damage!)
		//   This creates a ~10% DPS reduction from intended design.
		//
		// Safe values are powers of 2: 0.5 (1/2), 0.25 (1/4), 0.125 (1/8), 0.0625 (1/16)
		// Current setting from config: 8 ticks per 1 damage = 7.5 DPS at 60 TPS
		BeamDamagePerTick:     beamWeapon.DamagePerTick,
		BeamDamageAccumulator: 0.0,
		BeamTargetID:          -1,
		BeamFiringAtID:        -1,
		AttackerIDs:           []int{},
	}

	return base
}
