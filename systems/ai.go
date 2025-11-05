package systems

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

const (
	aiDecisionInterval = 60                    // Ticks between decisions (~1 second at 60 TPS)
	aiRetargetInterval = 60                    // Ticks between target re-evaluation (~1 second at 60 TPS)
	aiRotationSpeed    = 3.0 * math.Pi / 180.0 // 3 degrees per tick rotation toward target
	aiPursuitSpeedMin  = 0.80                  // Minimum speed when pursuing (80% of max)
	aiPursuitSpeedMax  = 1.00                  // Maximum speed when pursuing (100% of max)
	aiPatrolSpeed      = 0.50                  // Speed when no target (50% of max)
)

// normalizeAngle brings an angle into the range [-π, π] (defined in weapons.go but needed here too)
func normalizeAngleAI(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
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
			aiState.RetargetTimer = aiRetargetInterval

			// Find nearest enemy
			newTarget := SelectNearestEnemy(w, entry.Entity())
			aiTarget.TargetEntity = newTarget
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
			if math.Abs(dx) > float64(GameWidth)/2 {
				if dx > 0 {
					dx = dx - float64(GameWidth)
				} else {
					dx = dx + float64(GameWidth)
				}
			}
			if math.Abs(dy) > float64(GameHeight)/2 {
				if dy > 0 {
					dy = dy - float64(GameHeight)
				} else {
					dy = dy + float64(GameHeight)
				}
			}

			targetAngle := math.Atan2(dx, -dy)
			angleDiff := normalizeAngleAI(targetAngle - rotation.Angle)

			// Smoothly rotate toward target
			if math.Abs(angleDiff) > aiRotationSpeed {
				if angleDiff > 0 {
					rotation.Angle += aiRotationSpeed
				} else {
					rotation.Angle -= aiRotationSpeed
				}
			} else {
				// Close enough, snap to target angle
				rotation.Angle = targetAngle
			}

			// Set speed for pursuit (randomized slightly for variety)
			speedVariation := rand.Float64()*(aiPursuitSpeedMax-aiPursuitSpeedMin) + aiPursuitSpeedMin
			shipData.Speed = shipData.MaxSpeed * speedVariation

		} else {
			// PATROL BEHAVIOR: No target, maintain heading and reduce speed
			shipData.Speed = shipData.MaxSpeed * aiPatrolSpeed
		}

		// Update velocity based on rotation and speed
		velocity.X = math.Sin(rotation.Angle) * shipData.Speed
		velocity.Y = -math.Cos(rotation.Angle) * shipData.Speed
	}
}

// UpdateAIFiring handles AI decision making for firing weapons
// AI ships attempt to fire at their current target when in range
func UpdateAIFiring(w donburi.World, playerEntity donburi.Entity, laserSprite *ebiten.Image) {
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
		if math.Abs(dx) > float64(GameWidth)/2 {
			if dx > 0 {
				dx = dx - float64(GameWidth)
			} else {
				dx = dx + float64(GameWidth)
			}
		}
		if math.Abs(dy) > float64(GameHeight)/2 {
			if dy > 0 {
				dy = dy - float64(GameHeight)
			} else {
				dy = dy + float64(GameHeight)
			}
		}

		targetAngle := math.Atan2(dx, -dy)
		angleDiff := normalizeAngleAI(targetAngle - aiRot.Angle)

		// Check if target is within firing cone
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
	}
}
