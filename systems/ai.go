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
	aiDecisionInterval = 60 // Ticks between decisions (~1 second at 60 TPS)
	aiTurnAngle        = 30.0 * math.Pi / 180.0 // 30 degrees in radians
	aiSpeedAdjustment  = 0.10 // 10% speed adjustment
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

// UpdateAIMovement handles AI decision making for movement
// AI ships make decisions approximately once per second
func UpdateAIMovement(w donburi.World) {
	query := donburi.NewQuery(filter.Contains(
		components.AIControlled,
		components.AIState,
		components.Ship,
		components.Rotation,
		components.Velocity,
	))

	for entry := range query.Iter(w) {
		aiState := components.AIState.Get(entry)
		shipData := components.Ship.Get(entry)
		rotation := components.Rotation.Get(entry)
		velocity := components.Velocity.Get(entry)

		// Countdown decision timer
		aiState.DecisionTimer--

		// Make decision when timer reaches zero
		if aiState.DecisionTimer <= 0 {
			// Reset timer for next decision
			aiState.DecisionTimer = aiDecisionInterval

			// Randomly decide on speed adjustment
			speedRoll := rand.Float64()
			if speedRoll < 0.33 {
				// Increase speed by 10%
				shipData.Speed *= (1.0 + aiSpeedAdjustment)
				if shipData.Speed > shipData.MaxSpeed {
					shipData.Speed = shipData.MaxSpeed
				}
			} else if speedRoll < 0.66 {
				// Decrease speed by 10%
				shipData.Speed *= (1.0 - aiSpeedAdjustment)
				if shipData.Speed < 0 {
					shipData.Speed = 0
				}
			}
			// else: no speed change (0.66-1.0)

			// Randomly decide on rotation
			turnRoll := rand.Float64()
			if turnRoll < 0.33 {
				// Turn clockwise
				rotation.Angle += aiTurnAngle
			} else if turnRoll < 0.66 {
				// Turn counter-clockwise
				rotation.Angle -= aiTurnAngle
			}
			// else: no turn (0.66-1.0)

			// Update velocity based on new rotation and speed
			velocity.X = math.Sin(rotation.Angle) * shipData.Speed
			velocity.Y = -math.Cos(rotation.Angle) * shipData.Speed
		}
	}
}

// UpdateAIFiring handles AI decision making for firing weapons
// AI ships attempt to fire at player ships when in range
func UpdateAIFiring(w donburi.World, playerEntity donburi.Entity, laserSprite *ebiten.Image) {
	// Check if player is valid
	if !w.Valid(playerEntity) {
		return
	}

	playerEntry := w.Entry(playerEntity)
	playerPos := components.Position.Get(playerEntry)
	playerFaction := components.Faction.Get(playerEntry)

	// Query all AI ships
	aiQuery := donburi.NewQuery(filter.Contains(
		components.AIControlled,
		components.Position,
		components.Rotation,
		components.Weapon,
		components.Faction,
	))

	for aiEntry := range aiQuery.Iter(w) {
		aiPos := components.Position.Get(aiEntry)
		aiRot := components.Rotation.Get(aiEntry)
		aiWeapon := components.Weapon.Get(aiEntry)
		aiFaction := components.Faction.Get(aiEntry)

		// Don't shoot at own faction
		if aiFaction.ID == playerFaction.ID {
			continue
		}

		// Calculate angle to player (accounting for world wrapping)
		dx := playerPos.X - aiPos.X
		dy := playerPos.Y - aiPos.Y

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

		// Check if player is within firing cone
		halfCone := aiWeapon.FiringCone / 2
		if math.Abs(angleDiff) <= halfCone {
			// Player is in firing cone!
			// Decide what to do based on probabilities
			fireRoll := rand.Float64()

			if fireRoll < 0.50 {
				// 50% chance: Fire toward player
				FireWeapon(w, aiEntry, playerPos.X, playerPos.Y, laserSprite)
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

