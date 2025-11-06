package systems

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
	"github.com/nathan/verdant-thane/config"
)

// SelectTestudonTarget finds the best target for a Testudon using priority-based targeting
// Priority: 1. Attackers (defensive), 2. Enemy Testudons, 3. Enemy Destroyers, 4. Enemy Fighters
// Returns the selected enemy entity, or an invalid entity if no enemies found
func SelectTestudonTarget(w donburi.World, aiEntity donburi.Entity) donburi.Entity {
	if !w.Valid(aiEntity) {
		var emptyEntity donburi.Entity
		return emptyEntity
	}

	aiEntry := w.Entry(aiEntity)
	aiPos := components.Position.Get(aiEntry)
	aiFaction := components.Faction.Get(aiEntry)

	// Check if under attack - prioritize attackers
	if aiEntry.HasComponent(components.UnderAttack) {
		underAttack := components.UnderAttack.Get(aiEntry)
		if len(underAttack.Attackers) > 0 {
			// Find nearest attacker
			minDistance := math.MaxFloat64
			var nearestAttacker donburi.Entity

			for _, attackerEntity := range underAttack.Attackers {
				if !w.Valid(attackerEntity) {
					continue
				}

				attackerEntry := w.Entry(attackerEntity)
				if !attackerEntry.HasComponent(components.Position) {
					continue
				}

				attackerPos := components.Position.Get(attackerEntry)
				dist := distance(aiPos.X, aiPos.Y, attackerPos.X, attackerPos.Y)

				if dist < minDistance {
					minDistance = dist
					nearestAttacker = attackerEntity
				}
			}

			if w.Valid(nearestAttacker) {
				return nearestAttacker
			}
		}
	}

	// Priority-based targeting: Testudons > Destroyers > Fighters
	// We'll search for each class in priority order, selecting the nearest of the highest priority class found

	var bestTarget donburi.Entity
	var bestDistance float64 = math.MaxFloat64
	bestPriority := -1

	enemyQuery := donburi.NewQuery(filter.Contains(
		components.IsShip,
		components.Position,
		components.Faction,
		components.Ship,
	))

	for enemyEntry := range enemyQuery.Iter(w) {
		enemyFaction := components.Faction.Get(enemyEntry)

		// Skip friendlies
		if enemyFaction.ID == aiFaction.ID {
			continue
		}

		enemyShip := components.Ship.Get(enemyEntry)
		enemyPos := components.Position.Get(enemyEntry)

		// Determine priority based on ship class
		var priority int
		switch enemyShip.Class {
		case components.Testudon:
			priority = 3 // Highest priority
		case components.Destroyer:
			priority = 2
		case components.Fighter:
			priority = 1
		default:
			priority = 0
		}

		dist := distance(aiPos.X, aiPos.Y, enemyPos.X, enemyPos.Y)

		// Update best target if this is higher priority, or same priority but closer
		if priority > bestPriority || (priority == bestPriority && dist < bestDistance) {
			bestPriority = priority
			bestDistance = dist
			bestTarget = enemyEntry.Entity()
		}
	}

	return bestTarget
}

// SelectNearestEnemy finds the nearest enemy ship for an AI entity
// Returns the nearest enemy entity, or an invalid entity if no enemies found
func SelectNearestEnemy(w donburi.World, aiEntity donburi.Entity) donburi.Entity {
	if !w.Valid(aiEntity) {
		var emptyEntity donburi.Entity
		return emptyEntity
	}

	aiEntry := w.Entry(aiEntity)
	aiPos := components.Position.Get(aiEntry)
	aiFaction := components.Faction.Get(aiEntry)

	// Query all enemy ships
	enemyQuery := donburi.NewQuery(filter.Contains(
		components.IsShip,
		components.Position,
		components.Faction,
	))

	var nearestEnemy donburi.Entity
	minDistance := math.MaxFloat64

	for enemyEntry := range enemyQuery.Iter(w) {
		enemyFaction := components.Faction.Get(enemyEntry)

		// Skip friendlies
		if enemyFaction.ID == aiFaction.ID {
			continue
		}

		enemyPos := components.Position.Get(enemyEntry)

		// Calculate distance using world-wrapping distance function from collision.go
		dist := distance(aiPos.X, aiPos.Y, enemyPos.X, enemyPos.Y)

		if dist < minDistance {
			minDistance = dist
			nearestEnemy = enemyEntry.Entity()
		}
	}

	return nearestEnemy
}

// UpdateAIMovement handles AI decision making for movement with intelligent targeting
// AI ships pursue the nearest enemy and re-evaluate targets periodically
func UpdateAIMovement(w donburi.World) {
	query := donburi.NewQuery(filter.Contains(
		components.AIControlled,
		components.AIState,
		components.AITarget,
		components.Position,
		components.Ship,
		components.Rotation,
		components.Velocity,
	))

	// Process all AI ships every frame
	for entry := range query.Iter(w) {
		aiState := components.AIState.Get(entry)
		aiTarget := components.AITarget.Get(entry)
		aiPos := components.Position.Get(entry)
		shipData := components.Ship.Get(entry)
		rotation := components.Rotation.Get(entry)
		velocity := components.Velocity.Get(entry)

		// Countdown retarget timer
		aiState.RetargetTimer--

		// Re-evaluate target when timer expires or current target is invalid
		if aiState.RetargetTimer <= 0 || !w.Valid(aiTarget.TargetEntity) {
			aiState.RetargetTimer = config.AIRetargetInterval

			// Use appropriate targeting logic based on ship class
			var newTarget donburi.Entity
			if shipData.Class == components.Testudon {
				// Testudons use priority-based targeting
				newTarget = SelectTestudonTarget(w, entry.Entity())
			} else {
				// Fighters and Destroyers use nearest-enemy targeting
				newTarget = SelectNearestEnemy(w, entry.Entity())
			}
			aiTarget.TargetEntity = newTarget

			// Update beam weapon target if this is a Testudon
			if shipData.Class == components.Testudon && entry.HasComponent(components.BeamWeapon) {
				if w.Valid(newTarget) {
					SetBeamTarget(entry, newTarget)
				} else {
					ClearBeamTarget(entry)
				}
			}
		}

		// Behavior based on whether we have a valid target
		hasValidTarget := w.Valid(aiTarget.TargetEntity)

		if hasValidTarget {
			// PURSUIT BEHAVIOR: Rotate toward target and fly at high speed
			targetEntry := w.Entry(aiTarget.TargetEntity)
			targetPos := components.Position.Get(targetEntry)

			// Calculate angle to target (with world wrapping)
			dx := targetPos.X - aiPos.X
			dy := targetPos.Y - aiPos.Y

			// Handle world wrapping for shortest path
			if math.Abs(dx) > float64(config.GameWidth)/2 {
				if dx > 0 {
					dx = dx - float64(config.GameWidth)
				} else {
					dx = dx + float64(config.GameWidth)
				}
			}
			if math.Abs(dy) > float64(config.GameHeight)/2 {
				if dy > 0 {
					dy = dy - float64(config.GameHeight)
				} else {
					dy = dy + float64(config.GameHeight)
				}
			}

			targetAngle := math.Atan2(dx, -dy)
			angleDiff := NormalizeAngle(targetAngle - rotation.Angle)

			// Smoothly rotate toward target
			if math.Abs(angleDiff) > config.AIRotationSpeed {
				if angleDiff > 0 {
					rotation.Angle += config.AIRotationSpeed
				} else {
					rotation.Angle -= config.AIRotationSpeed
				}
			} else {
				// Close enough, snap to target angle
				rotation.Angle = targetAngle
			}

			// Set speed for pursuit (randomized slightly for variety)
			speedVariation := rand.Float64()*(config.AIPursuitSpeedMax-config.AIPursuitSpeedMin) + config.AIPursuitSpeedMin
			shipData.Speed = shipData.MaxSpeed * speedVariation

		} else {
			// PATROL BEHAVIOR: No target, maintain heading and reduce speed
			shipData.Speed = shipData.MaxSpeed * config.AIPatrolSpeed
		}

		// Update velocity based on rotation and speed
		velocity.X = math.Sin(rotation.Angle) * shipData.Speed
		velocity.Y = -math.Cos(rotation.Angle) * shipData.Speed
	}
}

// UpdateAIFiring handles AI decision making for firing weapons
// AI ships attempt to fire at their current target when in range
// If the ship has a secondary weapon (missiles), it will fire at targets in the rear arc
func UpdateAIFiring(w donburi.World, playerEntity donburi.Entity, laserSprite, missileSprite *ebiten.Image) {
	// Query all AI ships with targets
	aiQuery := donburi.NewQuery(filter.Contains(
		components.AIControlled,
		components.AITarget,
		components.Position,
		components.Rotation,
		components.Weapon,
	))

	for aiEntry := range aiQuery.Iter(w) {
		aiTarget := components.AITarget.Get(aiEntry)

		// Skip if no valid target
		if !w.Valid(aiTarget.TargetEntity) {
			continue
		}

		targetEntry := w.Entry(aiTarget.TargetEntity)
		targetPos := components.Position.Get(targetEntry)

		aiPos := components.Position.Get(aiEntry)
		aiRot := components.Rotation.Get(aiEntry)
		aiWeapon := components.Weapon.Get(aiEntry)

		// Calculate angle to target (accounting for world wrapping)
		dx := targetPos.X - aiPos.X
		dy := targetPos.Y - aiPos.Y

		// Handle world wrapping for shortest distance
		if math.Abs(dx) > float64(config.GameWidth)/2 {
			if dx > 0 {
				dx = dx - float64(config.GameWidth)
			} else {
				dx = dx + float64(config.GameWidth)
			}
		}
		if math.Abs(dy) > float64(config.GameHeight)/2 {
			if dy > 0 {
				dy = dy - float64(config.GameHeight)
			} else {
				dy = dy + float64(config.GameHeight)
			}
		}

		targetAngle := math.Atan2(dx, -dy)
		angleDiff := NormalizeAngle(targetAngle - aiRot.Angle)

		// Check if target is within firing cone (front)
		halfCone := aiWeapon.FiringCone / 2
		if math.Abs(angleDiff) <= halfCone {
			// Target is in firing cone!
			// Decide what to do based on probabilities
			fireRoll := rand.Float64()

			if fireRoll < 0.50 {
				// 50% chance: Fire toward target
				FireWeapon(w, aiEntry, targetPos.X, targetPos.Y, laserSprite)
			} else if fireRoll < 0.75 {
				// 25% chance: Fire randomly within cone
				randomAngle := aiRot.Angle + (rand.Float64()*2-1)*halfCone
				// Calculate random target point in that direction
				randomTargetX := aiPos.X + math.Sin(randomAngle)*1000
				randomTargetY := aiPos.Y - math.Cos(randomAngle)*1000
				FireWeapon(w, aiEntry, randomTargetX, randomTargetY, laserSprite)
			}
			// else: 25% chance: Don't fire
		}

		// Check if ship has secondary weapon (missiles) and target is in rear arc
		if aiEntry.HasComponent(components.SecondaryWeapon) {
			// Check if target is in rear 180° arc
			// Rear arc is centered at 180° from facing direction
			rearAngleDiff := NormalizeAngle(angleDiff)

			// Target is in rear arc if angle difference is > 90° (π/2)
			if math.Abs(rearAngleDiff) > math.Pi/2 {
				// Target is behind us! Fire missile with same probability as main gun
				fireRoll := rand.Float64()

				if fireRoll < 0.50 {
					// 50% chance: Fire missile at target
					FireMissile(w, aiEntry, aiTarget.TargetEntity, missileSprite)
				}
				// else: Don't fire missile
			}
		}
	}
}
