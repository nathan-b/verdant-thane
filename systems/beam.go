package systems

import (
	"math"

	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

// UpdateBeamWeapons handles beam weapon targeting and damage application
// This is called every tick for all Testudons with beam weapons
func UpdateBeamWeapons(w donburi.World) {
	beamQuery := donburi.NewQuery(filter.Contains(
		components.BeamWeapon,
		components.Position,
		components.Faction,
		components.Health,
	))

	for entry := range beamQuery.Iter(w) {
		beamWeapon := components.BeamWeapon.Get(entry)
		position := components.Position.Get(entry)
		faction := components.Faction.Get(entry)

		// Check if current target is still valid
		targetValid := false
		if w.Valid(beamWeapon.TargetEntity) {
			targetEntry := w.Entry(beamWeapon.TargetEntity)
			if targetEntry.HasComponent(components.Health) {
				targetHealth := components.Health.Get(targetEntry)
				targetFaction := components.Faction.Get(targetEntry)
				targetPosition := components.Position.Get(targetEntry)

				// Target is valid if:
				// - Still alive (health > 0)
				// - Enemy faction
				// - Within range
				if targetHealth.Current > 0 &&
					targetFaction.ID != faction.ID &&
					IsInRange(position.X, position.Y, targetPosition.X, targetPosition.Y, beamWeapon.Range) {
					targetValid = true

					// Apply damage to valid target
					ApplyBeamDamage(w, entry, targetEntry, beamWeapon)
				}
			}
		}

		// If target is invalid, clear it
		if !targetValid {
			var emptyEntity donburi.Entity
			beamWeapon.TargetEntity = emptyEntity
			beamWeapon.DamageAccumulator = 0.0
		}
	}
}

// ApplyBeamDamage applies damage-over-time to a target entity
func ApplyBeamDamage(w donburi.World, attackerEntry, targetEntry *donburi.Entry, beamWeapon *components.BeamWeaponData) {
	// Accumulate damage
	beamWeapon.DamageAccumulator += beamWeapon.DamagePerTick

	// Apply integer damage when accumulator >= 1.0
	if beamWeapon.DamageAccumulator >= 1.0 {
		damageToApply := int(beamWeapon.DamageAccumulator)
		beamWeapon.DamageAccumulator -= float64(damageToApply)

		// Apply damage to target
		health := components.Health.Get(targetEntry)
		health.Current -= damageToApply
		if health.Current < 0 {
			health.Current = 0
		}

		// Track attacker in target's UnderAttack component (if target has it)
		if targetEntry.HasComponent(components.UnderAttack) {
			TrackAttacker(w, targetEntry, attackerEntry.Entity())
		}
	}
}

// TrackAttacker adds an attacker to a ship's UnderAttack list
func TrackAttacker(w donburi.World, targetEntry *donburi.Entry, attackerEntity donburi.Entity) {
	underAttack := components.UnderAttack.Get(targetEntry)

	// Check if attacker is already in the list
	alreadyTracked := false
	validAttackers := []donburi.Entity{}

	for _, attacker := range underAttack.Attackers {
		// Only keep valid attackers
		if w.Valid(attacker) {
			validAttackers = append(validAttackers, attacker)
			if attacker == attackerEntity {
				alreadyTracked = true
			}
		}
	}

	// Add new attacker if not already tracked
	if !alreadyTracked {
		validAttackers = append(validAttackers, attackerEntity)
	}

	underAttack.Attackers = validAttackers
}

// IsInRange checks if two positions are within a given range, accounting for toroidal wrapping
func IsInRange(x1, y1, x2, y2, maxRange float64) bool {
	dx, dy := GetWrappedDistance(x1, y1, x2, y2)
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance <= maxRange
}

// GetWrappedDistance returns the shortest distance components accounting for wrapping
func GetWrappedDistance(x1, y1, x2, y2 float64) (float64, float64) {
	dx := x2 - x1
	dy := y2 - y1

	// Wrap X coordinate
	if dx > float64(GameWidth)/2 {
		dx -= float64(GameWidth)
	} else if dx < -float64(GameWidth)/2 {
		dx += float64(GameWidth)
	}

	// Wrap Y coordinate
	if dy > float64(GameHeight)/2 {
		dy -= float64(GameHeight)
	} else if dy < -float64(GameHeight)/2 {
		dy += float64(GameHeight)
	}

	return dx, dy
}

// SetBeamTarget assigns a target to a beam weapon
// This is called by AI systems to direct beam weapons at specific targets
func SetBeamTarget(attackerEntry *donburi.Entry, targetEntity donburi.Entity) {
	if !attackerEntry.HasComponent(components.BeamWeapon) {
		return
	}

	beamWeapon := components.BeamWeapon.Get(attackerEntry)
	beamWeapon.TargetEntity = targetEntity
	beamWeapon.DamageAccumulator = 0.0 // Reset accumulator when switching targets
}

// ClearBeamTarget removes the current target from a beam weapon
func ClearBeamTarget(attackerEntry *donburi.Entry) {
	if !attackerEntry.HasComponent(components.BeamWeapon) {
		return
	}

	var emptyEntity donburi.Entity
	beamWeapon := components.BeamWeapon.Get(attackerEntry)
	beamWeapon.TargetEntity = emptyEntity
	beamWeapon.DamageAccumulator = 0.0
}
