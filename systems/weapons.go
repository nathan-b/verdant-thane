package systems

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

const (
	projectileSpeed    = 12.0            // 2x max ship speed
	projectileLifetime = 180             // ticks (3 seconds at 60 TPS)
	capacitorChargeTime = 600.0 / 1000.0 // 600ms in seconds
	capacitorChargeRate = 1.0 / (capacitorChargeTime * 60.0) // charge per tick
)

// UpdateWeapons charges weapon capacitors for all ships with weapons
func UpdateWeapons(w donburi.World) {
	query := donburi.NewQuery(filter.Contains(components.Weapon))

	for entry := range query.Iter(w) {
		weapon := components.Weapon.Get(entry)
		if weapon.Capacitor < 1.0 {
			weapon.Capacitor += capacitorChargeRate
			if weapon.Capacitor > 1.0 {
				weapon.Capacitor = 1.0
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
	vx := math.Sin(firingAngle) * projectileSpeed
	vy := -math.Cos(firingAngle) * projectileSpeed

	// Create projectile entity
	projectile := w.Create(
		components.IsProjectile,
		components.Position,
		components.Velocity,
		components.Faction,
		components.Projectile,
		components.Sprite,
	)

	entry := w.Entry(projectile)
	components.Position.SetValue(entry, components.PositionData{X: pos.X, Y: pos.Y})
	components.Velocity.SetValue(entry, components.VelocityData{X: vx, Y: vy})
	components.Faction.SetValue(entry, components.FactionData{ID: faction.ID})
	components.Projectile.SetValue(entry, components.ProjectileData{
		Lifetime:    projectileLifetime,
		MaxLifetime: projectileLifetime,
	})
	components.Sprite.SetValue(entry, components.SpriteData{Image: laserSprite})

	// Drain capacitor
	weapon.Capacitor = 0.0

	return true
}
