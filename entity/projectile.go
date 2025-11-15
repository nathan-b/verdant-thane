package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/nathan/verdant-thane/config"
)

// ============================================================================
// MainGunProjectile
// ============================================================================

// MainGunProjectile is a straight-line laser projectile
type MainGunProjectile struct {
	ID                   int
	X, Y                 float64
	VelocityX, VelocityY float64
	OwnerID              int
	FactionID            int
	Sprite               *ebiten.Image
	Lifetime             int
	MaxLifetime          int
	Alive                bool
	Damage               int
	CollisionRadius      float64
}

// NewMainGunProjectile creates a new projectile for the figter and destroyer main gun
func NewMainGunProjectile(id int, cfg MainGunConfig) *MainGunProjectile {
	chars := config.GetProjectileCharacteristics(config.LaserProjectile)

	return &MainGunProjectile{
		ID:              id,
		X:               cfg.X,
		Y:               cfg.Y,
		VelocityX:       cfg.VelocityX,
		VelocityY:       cfg.VelocityY,
		OwnerID:         cfg.OwnerID,
		FactionID:       cfg.FactionID,
		Sprite:          cfg.Sprite,
		Lifetime:        chars.Lifetime,
		MaxLifetime:     chars.Lifetime,
		Alive:           true,
		Damage:          chars.Damage,
		CollisionRadius: chars.CollisionRadius,
	}
}

// ============================================================================
// Entity Interface Implementation
// ============================================================================

// Update handles per-frame logic for the laser
func (l *MainGunProjectile) Update(ctx GameContext) error {
	if !l.Alive {
		return nil
	}

	// Update lifetime
	l.Lifetime--
	if l.Lifetime <= 0 {
		l.Alive = false
		return nil
	}

	// Update position
	l.X += l.VelocityX
	l.Y += l.VelocityY

	// Wrap position
	l.X, l.Y = WrapPosition(l.X, l.Y)

	return nil
}

// Render draws the laser to the screen
func (l *MainGunProjectile) Render(screen *ebiten.Image, cameraX, cameraY float64) {
	if !l.Alive || l.Sprite == nil {
		return
	}

	// Get wrapped screen position (handles world wrapping)
	screenX, screenY := GetWrappedScreenPosition(l.X, l.Y, cameraX, cameraY)

	opts := &ebiten.DrawImageOptions{}

	// Center sprite
	opts.GeoM.Translate(-float64(l.Sprite.Bounds().Dx())/2, -float64(l.Sprite.Bounds().Dy())/2)

	// Position (relative to camera, accounting for world wrapping)
	opts.GeoM.Translate(screenX, screenY)

	screen.DrawImage(l.Sprite, opts)
}

// GetID returns the projectile's ID
func (l *MainGunProjectile) GetID() int {
	return l.ID
}

// GetPosition returns the projectile's position
func (l *MainGunProjectile) GetPosition() (float64, float64) {
	return l.X, l.Y
}

// IsAlive returns whether the projectile is still active
func (l *MainGunProjectile) IsAlive() bool {
	return l.Alive
}

// ============================================================================
// Projectile Interface Implementation
// ============================================================================

// GetOwnerID returns the ID of the ship that fired this projectile
func (l *MainGunProjectile) GetOwnerID() int {
	return l.OwnerID
}

// GetDamage returns the damage this projectile deals
func (l *MainGunProjectile) GetDamage() int {
	return l.Damage
}

// CheckCollision checks if this projectile collides with a ship
func (l *MainGunProjectile) CheckCollision(ship Ship) bool {
	if !l.Alive || !ship.IsAlive() {
		return false
	}

	// Don't collide with same faction
	if ship.GetFaction() == l.FactionID {
		return false
	}

	// Check distance
	shipX, shipY := ship.GetPosition()
	dist := Distance(l.X, l.Y, shipX, shipY)

	// Collision if within ship's collision radius + projectile radius
	return dist <= ship.GetCollisionRadius()+l.CollisionRadius
}

// GetFaction returns the faction ID (inherited from owner)
func (l *MainGunProjectile) GetFaction() int {
	return l.FactionID
}

// ============================================================================
// MissileProjectile
// ============================================================================

// MissileProjectile is a homing missile that tracks a target
type MissileProjectile struct {
	ID                   int
	X, Y                 float64
	VelocityX, VelocityY float64
	Rotation             float64 // Current rotation angle
	OwnerID              int
	FactionID            int
	TargetID             int // ID of ship being tracked
	Sprite               *ebiten.Image
	Lifetime             int
	MaxLifetime          int
	Acceleration         float64
	Alive                bool
	Damage               int
	CollisionRadius      float64
}

// NewMissileProjectile creates a new missile projectile
func NewMissileProjectile(id int, cfg MissileConfig) *MissileProjectile {
	chars := config.GetProjectileCharacteristics(config.MissileProjectile)

	return &MissileProjectile{
		ID:              id,
		X:               cfg.X,
		Y:               cfg.Y,
		VelocityX:       cfg.VelocityX,
		VelocityY:       cfg.VelocityY,
		Rotation:        cfg.Rotation,
		OwnerID:         cfg.OwnerID,
		FactionID:       cfg.FactionID,
		TargetID:        cfg.TargetID,
		Sprite:          cfg.Sprite,
		Lifetime:        chars.Lifetime,
		MaxLifetime:     chars.Lifetime,
		Acceleration:    chars.Acceleration,
		Alive:           true,
		Damage:          chars.Damage,
		CollisionRadius: chars.CollisionRadius,
	}
}

// ============================================================================
// Entity Interface Implementation
// ============================================================================

// Update handles per-frame logic for the missile (including tracking)
func (m *MissileProjectile) Update(ctx GameContext) error {
	if !m.Alive {
		return nil
	}

	// Update lifetime
	m.Lifetime--
	if m.Lifetime <= 0 {
		m.Alive = false
		return nil
	}

	// Update tracking
	m.UpdateTracking(ctx)

	// Update position
	m.X += m.VelocityX
	m.Y += m.VelocityY

	// Wrap position
	m.X, m.Y = WrapPosition(m.X, m.Y)

	return nil
}

// Render draws the missile to the screen
func (m *MissileProjectile) Render(screen *ebiten.Image, cameraX, cameraY float64) {
	if !m.Alive || m.Sprite == nil {
		return
	}

	// Get wrapped screen position (handles world wrapping)
	screenX, screenY := GetWrappedScreenPosition(m.X, m.Y, cameraX, cameraY)

	opts := &ebiten.DrawImageOptions{}

	// Rotation
	opts.GeoM.Translate(-float64(m.Sprite.Bounds().Dx())/2, -float64(m.Sprite.Bounds().Dy())/2)
	opts.GeoM.Rotate(m.Rotation)

	// Position (relative to camera, accounting for world wrapping)
	opts.GeoM.Translate(screenX, screenY)

	screen.DrawImage(m.Sprite, opts)
}

// GetID returns the missile's ID
func (m *MissileProjectile) GetID() int {
	return m.ID
}

// GetPosition returns the missile's position
func (m *MissileProjectile) GetPosition() (float64, float64) {
	return m.X, m.Y
}

// IsAlive returns whether the missile is still active
func (m *MissileProjectile) IsAlive() bool {
	return m.Alive
}

// ============================================================================
// Projectile Interface Implementation
// ============================================================================

// GetOwnerID returns the ID of the ship that fired this missile
func (m *MissileProjectile) GetOwnerID() int {
	return m.OwnerID
}

// GetDamage returns the damage this missile deals
func (m *MissileProjectile) GetDamage() int {
	return m.Damage
}

// CheckCollision checks if this missile collides with a ship
func (m *MissileProjectile) CheckCollision(ship Ship) bool {
	if !m.Alive || !ship.IsAlive() {
		return false
	}

	// Don't collide with same faction
	if ship.GetFaction() == m.FactionID {
		return false
	}

	// Check distance
	shipX, shipY := ship.GetPosition()
	dist := Distance(m.X, m.Y, shipX, shipY)

	// Collision if within ship's collision radius + projectile radius
	return dist <= ship.GetCollisionRadius()+m.CollisionRadius
}

// GetFaction returns the faction ID (inherited from owner)
func (m *MissileProjectile) GetFaction() int {
	return m.FactionID
}

// ============================================================================
// Missile-specific Methods
// ============================================================================

// UpdateTracking handles homing behavior toward target
func (m *MissileProjectile) UpdateTracking(ctx GameContext) {
	// Check if target is still valid
	target := ctx.GetShip(m.TargetID)
	if target == nil || !target.IsAlive() {
		// Target lost, continue on current trajectory
		return
	}

	// Calculate direction to target
	targetX, targetY := target.GetPosition()
	dx, dy := GetWrappedDistance(m.X, m.Y, targetX, targetY)
	// Sprites face UP (Y-axis), so use atan2(dx, -dy)
	angleToTarget := math.Atan2(dx, -dy)

	// Current velocity angle
	currentSpeed := math.Sqrt(m.VelocityX*m.VelocityX + m.VelocityY*m.VelocityY)
	if currentSpeed < 0.01 {
		// Edge case: very slow, just point at target
		m.Rotation = angleToTarget
	} else {
		// Sprites face UP (Y-axis), so use atan2(vx, -vy) for velocity angle
		currentAngle := math.Atan2(m.VelocityX, -m.VelocityY)

		// Calculate angle difference
		angleDiff := NormalizeAngle(angleToTarget - currentAngle)

		// Get turn rate from config
		chars := config.GetProjectileCharacteristics(config.MissileProjectile)
		maxTurn := chars.TurnRate

		// Apply turn limit
		if math.Abs(angleDiff) <= maxTurn {
			currentAngle = angleToTarget
		} else if angleDiff > 0 {
			currentAngle += maxTurn
		} else {
			currentAngle -= maxTurn
		}
		currentAngle = NormalizeAngle(currentAngle)

		// Update rotation for rendering
		m.Rotation = currentAngle

		// Accelerate in current direction
		// Sprites face UP (Y-axis), so use sin/cos adjusted for sprite orientation
		currentSpeed += m.Acceleration
		m.VelocityX = math.Sin(currentAngle) * currentSpeed
		m.VelocityY = -math.Cos(currentAngle) * currentSpeed
	}
}
