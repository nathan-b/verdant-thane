package systems

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

// UpdateBeamWeapons handles beam weapon targeting and damage application
// This is called every tick for all Testudons with beam weapons
func UpdateBeamWeapons(w donburi.World, explosionSprite *ebiten.Image) {
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

		var targetToFire *donburi.Entry

		// First, check if current AI target is valid and in range
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
					targetToFire = targetEntry
				}
			}
		}

		// If primary target is not in range, opportunistically fire at ANY in-range enemy
		// This makes Testudons more aggressive - they fire while moving toward distant targets
		if targetToFire == nil {
			targetToFire = FindNearestInRangeEnemy(w, entry, beamWeapon.Range)
		}

		// Fire at the chosen target (if any)
		if targetToFire != nil {
			// Update FiringAtEntity for beam rendering
			beamWeapon.FiringAtEntity = targetToFire.Entity()
			ApplyBeamDamage(w, entry, targetToFire, beamWeapon, explosionSprite)
		} else {
			// No target in range - reset damage accumulator and clear firing target
			beamWeapon.DamageAccumulator = 0.0
			var emptyEntity donburi.Entity
			beamWeapon.FiringAtEntity = emptyEntity
		}
	}
}

// ApplyBeamDamage applies damage-over-time to a target entity
func ApplyBeamDamage(w donburi.World, attackerEntry, targetEntry *donburi.Entry, beamWeapon *components.BeamWeaponData, explosionSprite *ebiten.Image) {
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

		// Check if target was destroyed
		if health.Current <= 0 {
			// Destroy the ship (creates explosion, handles player respawn, updates score)
			DestroyShip(w, targetEntry, attackerEntry.Entity(), explosionSprite)
			// Clear the beam target since it's now destroyed
			ClearBeamTarget(attackerEntry)
			return
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
	beamWeapon.FiringAtEntity = emptyEntity
	beamWeapon.DamageAccumulator = 0.0
}

// FindNearestInRangeEnemy finds the nearest enemy within range for opportunistic beam firing
// Returns nil if no enemies are in range
func FindNearestInRangeEnemy(w donburi.World, attackerEntry *donburi.Entry, maxRange float64) *donburi.Entry {
	attackerPos := components.Position.Get(attackerEntry)
	attackerFaction := components.Faction.Get(attackerEntry)

	// Query all enemy ships
	enemyQuery := donburi.NewQuery(filter.Contains(
		components.IsShip,
		components.Position,
		components.Faction,
		components.Health,
	))

	var closestEnemy *donburi.Entry
	closestDistance := math.MaxFloat64 // Start with infinity

	for enemyEntry := range enemyQuery.Iter(w) {
		enemyFaction := components.Faction.Get(enemyEntry)
		enemyHealth := components.Health.Get(enemyEntry)

		// Skip friendlies and dead ships
		if enemyFaction.ID == attackerFaction.ID || enemyHealth.Current <= 0 {
			continue
		}

		enemyPos := components.Position.Get(enemyEntry)

		// Calculate distance with wrapping
		dx, dy := GetWrappedDistance(attackerPos.X, attackerPos.Y, enemyPos.X, enemyPos.Y)
		distance := math.Sqrt(dx*dx + dy*dy)

		// CRITICAL: Only consider enemies within max range AND closer than current closest
		if distance <= maxRange && distance < closestDistance {
			closestDistance = distance
			closestEnemy = enemyEntry
		}
	}

	return closestEnemy
}
