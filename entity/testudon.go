package entity

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/nathan-b/verdant-thane/config"
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
		BeamDamagePerTick:             beamWeapon.DamagePerTick,
		BeamDamageAccumulator:         0.0,
		BeamTargetID:                  -1,
		BeamFiringAtID:                -1,
		AttackerIDs:                   []int{},
	}

	return base
}
