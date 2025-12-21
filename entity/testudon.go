package entity

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/nathan-b/verdant-thane/config"
)

// NewTestudon creates a new testudon ship
func NewTestudon(id int, factionID int, x, y float64, sprite *ebiten.Image) *Ship {
	ship := newShipBase(id, factionID, ClassTestudon, x, y, sprite)
	chars := config.GetShipCharacteristics(ClassTestudon)
	beamWeapon := config.WeaponDatabase["beam"]

	// Configure testudon-specific: no projectile weapons (beam weapon only)
	ship.Weapons = []Weapon{}

	// No afterburner for testudons
	ship.AfterburnerCharge = 0.0
	ship.AfterburnerActive = false
	ship.AfterburnerMaxCharge = 0.0
	ship.HasAfterburnerSystem = false
	ship.AfterburnerDrain = chars.AfterburnerDrain
	ship.AfterburnerRecharge = chars.AfterburnerRecharge
	ship.AfterburnerAccelMultiplier = chars.AfterburnerAccelMultiplier
	ship.AfterburnerMaxSpeedMultiplier = chars.AfterburnerMaxSpeedMultiplier

	// Configure beam weapon
	ship.BeamRange = beamWeapon.MaxRange
	ship.BeamDamagePerTick = beamWeapon.DamagePerTick
	ship.BeamDamageAccumulator = 0.0

	// Testudons use AttackerIDs for AI behavior
	ship.AttackerIDs = []int{}

	return ship
}
