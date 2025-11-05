package systems

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

// Note: Projectile characteristics are now defined in projectiledata.go
// Access them via GetProjectileCharacteristics()

// UpdateWeapons charges weapon capacitors for all ships with weapons
func UpdateWeapons(w donburi.World) {
	// Get charge rates from projectile database
	laserChars := GetProjectileCharacteristics(LaserProjectile)
	missileChars := GetProjectileCharacteristics(MissileProjectile)

	laserChargeRate := 1.0 / (laserChars.ChargeTime * 60.0)
	missileChargeRate := 1.0 / (missileChars.ChargeTime * 60.0)

	// Charge primary weapons
	query := donburi.NewQuery(filter.Contains(components.Weapon))
	for entry := range query.Iter(w) {
		weapon := components.Weapon.Get(entry)
		if weapon.Capacitor < 1.0 {
			weapon.Capacitor += laserChargeRate
			if weapon.Capacitor > 1.0 {
				weapon.Capacitor = 1.0
			}
		}
	}

	// Charge secondary weapons (missiles)
	secondaryQuery := donburi.NewQuery(filter.Contains(components.SecondaryWeapon))
	for entry := range secondaryQuery.Iter(w) {
		secondary := components.SecondaryWeapon.Get(entry)
		if secondary.Capacitor < 1.0 {
			secondary.Capacitor += missileChargeRate
			if secondary.Capacitor > 1.0 {
				secondary.Capacitor = 1.0
			}
		}
	}
}

// normalizeAngle brings an angle into the range [-π, π]
func normalizeAngle(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

// IsTargetInFiringArc checks if the target coordinates are within the ship's firing arc
func IsTargetInFiringArc(shipEntry *donburi.Entry, targetX, targetY float64) bool {
	if !shipEntry.HasComponent(components.Position) || !shipEntry.HasComponent(components.Rotation) || !shipEntry.HasComponent(components.Weapon) {
		return false
	}

	pos := components.Position.Get(shipEntry)
	rot := components.Rotation.Get(shipEntry)
	weapon := components.Weapon.Get(shipEntry)

	// Calculate angle to target
	dx := targetX - pos.X
	dy := targetY - pos.Y
	targetAngle := math.Atan2(dx, -dy)

	// Calculate angle difference from ship's facing direction
	angleDiff := normalizeAngle(targetAngle - rot.Angle)

	// Check if within firing cone
	halfCone := weapon.FiringCone / 2
	return math.Abs(angleDiff) <= halfCone
}

// FireWeapon attempts to fire a weapon from the given ship toward the target coordinates
// Returns true if the weapon was fired, false otherwise
func FireWeapon(w donburi.World, shipEntry *donburi.Entry, targetX, targetY float64, laserSprite *ebiten.Image) bool {
	// Check if ship has a weapon
	if !shipEntry.HasComponent(components.Weapon) {
		return false
	}

	weapon := components.Weapon.Get(shipEntry)
	if weapon.Capacitor < 1.0 {
		return false
	}

	pos := components.Position.Get(shipEntry)
	rot := components.Rotation.Get(shipEntry)
	faction := components.Faction.Get(shipEntry)

	// Get laser projectile characteristics
	laserChars := GetProjectileCharacteristics(LaserProjectile)

	// Calculate angle to target
	dx := targetX - pos.X
	dy := targetY - pos.Y
	// Adjust for sprite orientation (sprite faces up at angle 0)
	targetAngle := math.Atan2(dx, -dy)

	// Calculate angle difference from ship's facing direction
	angleDiff := normalizeAngle(targetAngle - rot.Angle)

	// Constrain to firing cone
	firingAngle := rot.Angle
	halfCone := weapon.FiringCone / 2
	if angleDiff > halfCone {
		firingAngle += halfCone
	} else if angleDiff < -halfCone {
		firingAngle -= halfCone
	} else {
		firingAngle = targetAngle
	}

	// Create projectile velocity
	vx := math.Sin(firingAngle) * laserChars.Speed
	vy := -math.Cos(firingAngle) * laserChars.Speed

	// Create projectile entity
	projectile := w.Create(
		components.IsProjectile,
		components.Position,
		components.Velocity,
		components.Faction,
		components.Projectile,
		components.Sprite,
		components.Owner,
	)

	entry := w.Entry(projectile)
	components.Position.SetValue(entry, components.PositionData{X: pos.X, Y: pos.Y})
	components.Velocity.SetValue(entry, components.VelocityData{X: vx, Y: vy})
	components.Faction.SetValue(entry, components.FactionData{ID: faction.ID})
	components.Projectile.SetValue(entry, components.ProjectileData{
		Lifetime:    laserChars.Lifetime,
		MaxLifetime: laserChars.Lifetime,
	})
	components.Owner.SetValue(entry, components.OwnerData{OwnerEntity: shipEntry.Entity()})
	components.Sprite.SetValue(entry, components.SpriteData{Image: laserSprite})

	// Drain capacitor
	weapon.Capacitor = 0.0

	return true
}

// FireMissile attempts to fire a homing missile from the given ship toward the target entity
// Returns true if the missile was fired, false otherwise
func FireMissile(w donburi.World, shipEntry *donburi.Entry, targetEntity donburi.Entity, missileSprite *ebiten.Image) bool {
	// Check if ship has a secondary weapon
	if !shipEntry.HasComponent(components.SecondaryWeapon) {
		return false
	}

	secondary := components.SecondaryWeapon.Get(shipEntry)
	if secondary.Capacitor < 1.0 {
		return false
	}

	// Check if target is valid
	if !w.Valid(targetEntity) {
		return false
	}

	pos := components.Position.Get(shipEntry)
	rot := components.Rotation.Get(shipEntry)
	faction := components.Faction.Get(shipEntry)

	// Get missile projectile characteristics
	missileChars := GetProjectileCharacteristics(MissileProjectile)

	// Calculate initial missile velocity based on ship's facing direction
	// Missile launches from rear of ship (180° from facing)
	launchAngle := rot.Angle + math.Pi
	vx := math.Sin(launchAngle) * missileChars.Speed
	vy := -math.Cos(launchAngle) * missileChars.Speed

	// Create missile entity
	missile := w.Create(
		components.IsProjectile,
		components.IsMissile,
		components.Position,
		components.Velocity,
		components.Rotation,
		components.Faction,
		components.Projectile,
		components.Missile,
		components.Sprite,
		components.Owner,
	)

	entry := w.Entry(missile)
	components.Position.SetValue(entry, components.PositionData{X: pos.X, Y: pos.Y})
	components.Velocity.SetValue(entry, components.VelocityData{X: vx, Y: vy})
	components.Rotation.SetValue(entry, components.RotationData{Angle: launchAngle})
	components.Faction.SetValue(entry, components.FactionData{ID: faction.ID})
	components.Projectile.SetValue(entry, components.ProjectileData{
		Lifetime:    missileChars.Lifetime,
		MaxLifetime: missileChars.Lifetime,
	})
	components.Missile.SetValue(entry, components.MissileData{
		TargetEntity: targetEntity,
		Acceleration: missileChars.Acceleration,
	})
	components.Owner.SetValue(entry, components.OwnerData{OwnerEntity: shipEntry.Entity()})
	components.Sprite.SetValue(entry, components.SpriteData{Image: missileSprite})

	// Drain capacitor
	secondary.Capacitor = 0.0

	return true
}

// FindNearestEnemyInRearArc finds the nearest enemy ship in the rear 180° arc of the given ship
// Returns the enemy entity and true if found, or an invalid entity and false if no enemy found
func FindNearestEnemyInRearArc(w donburi.World, shipEntry *donburi.Entry) (donburi.Entity, bool) {
	if !shipEntry.HasComponent(components.Position) || !shipEntry.HasComponent(components.Rotation) || !shipEntry.HasComponent(components.Faction) {
		var emptyEntity donburi.Entity
		return emptyEntity, false
	}

	shipPos := components.Position.Get(shipEntry)
	shipRot := components.Rotation.Get(shipEntry)
	shipFaction := components.Faction.Get(shipEntry)

	// Query all enemy ships
	enemyQuery := donburi.NewQuery(filter.Contains(
		components.IsShip,
		components.Position,
		components.Faction,
	))

	var nearestEnemy donburi.Entity
	nearestDistance := math.MaxFloat64
	found := false

	for enemyEntry := range enemyQuery.Iter(w) {
		// Skip same faction
		enemyFaction := components.Faction.Get(enemyEntry)
		if enemyFaction.ID == shipFaction.ID {
			continue
		}

		enemyPos := components.Position.Get(enemyEntry)

		// Calculate direction to enemy (accounting for world wrapping)
		dx := enemyPos.X - shipPos.X
		dy := enemyPos.Y - shipPos.Y

		// Handle world wrapping - choose shortest path
		if dx > GameWidth/2 {
			dx -= GameWidth
		} else if dx < -GameWidth/2 {
			dx += GameWidth
		}
		if dy > GameHeight/2 {
			dy -= GameHeight
		} else if dy < -GameHeight/2 {
			dy += GameHeight
		}

		// Calculate angle to enemy
		enemyAngle := math.Atan2(dx, -dy)
		angleDiff := normalizeAngle(enemyAngle - shipRot.Angle)

		// Check if enemy is in rear arc (> 90° from facing direction)
		if math.Abs(angleDiff) > math.Pi/2 {
			// Enemy is in rear arc
			distance := math.Sqrt(dx*dx + dy*dy)
			if distance < nearestDistance {
				nearestDistance = distance
				nearestEnemy = enemyEntry.Entity()
				found = true
			}
		}
	}

	return nearestEnemy, found
}

// UpdateMissileTracking updates all missiles to accelerate toward their targets
func UpdateMissileTracking(w donburi.World) {
	// Get missile characteristics for turn rate
	missileChars := GetProjectileCharacteristics(MissileProjectile)

	query := donburi.NewQuery(filter.Contains(components.IsMissile, components.Missile, components.Velocity, components.Position, components.Rotation))

	for entry := range query.Iter(w) {
		missile := components.Missile.Get(entry)

		// Check if target is still valid
		if !w.Valid(missile.TargetEntity) {
			// Target destroyed - missile continues on current trajectory
			continue
		}

		// Get target position
		targetEntry := w.Entry(missile.TargetEntity)
		if !targetEntry.HasComponent(components.Position) {
			continue
		}

		targetPos := components.Position.Get(targetEntry)
		missilePos := components.Position.Get(entry)
		velocity := components.Velocity.Get(entry)
		rotation := components.Rotation.Get(entry)

		// Calculate direction to target (accounting for world wrapping)
		dx := targetPos.X - missilePos.X
		dy := targetPos.Y - missilePos.Y

		// Handle world wrapping - choose shortest path
		if dx > GameWidth/2 {
			dx -= GameWidth
		} else if dx < -GameWidth/2 {
			dx += GameWidth
		}
		if dy > GameHeight/2 {
			dy -= GameHeight
		} else if dy < -GameHeight/2 {
			dy += GameHeight
		}

		// Calculate current velocity direction and desired direction
		currentSpeed := math.Sqrt(velocity.X*velocity.X + velocity.Y*velocity.Y)
		if currentSpeed > 0 {
			// Current direction
			currentAngle := math.Atan2(velocity.X, -velocity.Y)

			// Desired direction (toward target)
			desiredAngle := math.Atan2(dx, -dy)

			// Calculate angle difference
			angleDiff := normalizeAngle(desiredAngle - currentAngle)

			// Constrain turn rate
			actualTurn := angleDiff
			if missileChars.TurnRate > 0 {
				if angleDiff > missileChars.TurnRate {
					actualTurn = missileChars.TurnRate
				} else if angleDiff < -missileChars.TurnRate {
					actualTurn = -missileChars.TurnRate
				}
			}

			// Rotate the velocity vector by the turn rate
			// This ensures we actually turn at the specified rate
			newAngle := currentAngle + actualTurn
			velocity.X = math.Sin(newAngle) * currentSpeed
			velocity.Y = -math.Cos(newAngle) * currentSpeed

			// Apply acceleration in the new direction (to increase speed)
			accelX := math.Sin(newAngle) * missile.Acceleration
			accelY := -math.Cos(newAngle) * missile.Acceleration
			velocity.X += accelX
			velocity.Y += accelY

			// Update rotation to face velocity direction
			rotation.Angle = newAngle
		}
	}
}
