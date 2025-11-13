package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/nathan/verdant-thane/config"
)

// Destroyer is a heavy fighter with rear-facing missiles
type Destroyer struct {
	*BaseShip
	// Secondary weapon (missiles)
	MissileCapacitor  float64
	MissileChargeRate float64
	MissileFiringArc  float64 // Rear 180° arc
}

// NewDestroyer creates a new destroyer ship
func NewDestroyer(id int, factionID int, x, y float64, sprite *ebiten.Image) *Destroyer {
	chars := config.GetShipCharacteristics(ClassDestroyer)

	// Get missile characteristics for charge rate
	missileChars := config.GetProjectileCharacteristics(config.MissileProjectile)
	missileChargeRate := 1.0 / (missileChars.ChargeTime * 60.0)

	base := &BaseShip{
		ID:                   id,
		FactionID:            factionID,
		Class:                ClassDestroyer,
		X:                    x,
		Y:                    y,
		VelocityX:            0,
		VelocityY:            0,
		Rotation:             0,
		Health:               chars.MaxShield,
		MaxHealth:            chars.MaxShield,
		Speed:                0,
		MaxSpeed:             chars.MaxSpeed,
		Accel:                chars.Acceleration,
		CollisionRadius:      chars.CollisionRadius,
		PlayerControlled:     false,
		WeaponCapacitor:      1.0,
		WeaponChargeRate:     chars.CapacitorChargeRate,
		FiringCone:           chars.FiringCone,
		AfterburnerCharge:    360.0, // Start fully charged
		AfterburnerActive:    false,
		AfterburnerMaxCharge: 360.0,
		HasAfterburnerSystem: true, // Destroyers have afterburner
		AITargetID:           -1,
		AIRetargetTimer:      config.AIRetargetInterval,
		Sprite:               sprite,
		Alive:                true,
	}

	return &Destroyer{
		BaseShip:          base,
		MissileCapacitor:  1.0, // Start fully charged
		MissileChargeRate: missileChargeRate,
		MissileFiringArc:  math.Pi, // 180° rear arc
	}
}

// ============================================================================
// Override Update Methods
// ============================================================================

// Update handles all per-frame logic for the destroyer
func (d *Destroyer) Update(ctx GameContext) error {
	if !d.Alive {
		return nil
	}

	// Update both weapons
	d.UpdateWeapons()
	d.UpdateMissileWeapon()

	// Update control (AI or player)
	if d.PlayerControlled {
		d.UpdatePlayerInput(ctx)
	} else {
		d.UpdateAI(ctx)
	}

	// Update movement
	d.UpdateMovement()

	return nil
}

// UpdateAI handles AI decision-making for destroyers (includes missile firing)
func (d *Destroyer) UpdateAI(ctx GameContext) {
	// Use base fighter AI for movement and main gun
	// (We need to convert *Destroyer to *Fighter temporarily)
	fighter := &Fighter{BaseShip: d.BaseShip}
	fighter.UpdateAI(ctx)

	// Additional destroyer behavior: Fire missiles at rear targets
	if d.CanFireMissile() {
		rearTarget, _ := ctx.FindNearestEnemyInArc(d, d.MissileFiringArc, 2000.0, true) // 2000px range, rear-facing
		if rearTarget != nil {
			d.FireMissile(rearTarget.GetID(), ctx)
		}
	}
}

// ============================================================================
// Ship Interface Implementation (Missile-specific)
// ============================================================================

func (d *Destroyer) FireWeapon(mouseX, mouseY float64, ctx GameContext) {
	playerShip := ctx.GetShip(d.ID)
	// Find nearest enemy in rear arc for missile
	nearestEnemy, dist := ctx.FindNearestEnemyInArc(
		playerShip,
		math.Pi, // 180 degree arc
		1000.0,  // Max range (TODO: make this part of the projectile stats)
		true,    // Rear-facing
	)
	if d.CanFireMissile() && nearestEnemy != nil && dist < 1000.0 { // TODO: definitely don't hardcode this twice
		d.FireMissile(nearestEnemy.GetID(), ctx)
	} else {
		// No valid target for a missile, just fire the main gun
		d.BaseShip.FireWeapon(mouseX, mouseY, ctx)
	}
}

// FireMissile fires a homing missile at a target (this is a destroyer-specific method)
// It's the responsibility of the caller to validate that the target is in a position
// to be fired upon (e.g. in rear arc, within range)
func (d *Destroyer) FireMissile(targetID int, ctx GameContext) {
	if !d.CanFireMissile() {
		return
	}

	// Check if target exists
	target := ctx.GetShip(targetID)
	if target == nil || !target.IsAlive() {
		return
	}

	// Consume capacitor
	d.MissileCapacitor = 0.0

	// Get missile characteristics
	missileChars := config.GetProjectileCharacteristics(config.MissileProjectile)

	// Launch from rear of ship
	// Sprites face UP (Y-axis), so rotation + π points to rear
	launchAngle := NormalizeAngle(d.Rotation + math.Pi)
	spawnOffset := 25.0
	// Use sin/cos adjusted for sprite orientation
	spawnX := d.X + math.Sin(launchAngle)*spawnOffset
	spawnY := d.Y + -math.Cos(launchAngle)*spawnOffset

	// Initial velocity (launches from rear)
	// Use sin/cos adjusted for sprite orientation
	missileVX := math.Sin(launchAngle) * missileChars.Speed
	missileVY := -math.Cos(launchAngle) * missileChars.Speed

	// Spawn missile via context
	ctx.SpawnMissile(MissileConfig{
		X:         spawnX,
		Y:         spawnY,
		VelocityX: missileVX,
		VelocityY: missileVY,
		Rotation:  launchAngle,
		OwnerID:   d.ID,
		FactionID: d.FactionID,
		TargetID:  targetID,
		Sprite:    nil, // Will be set by spawner
	})
}

// CanFireMissile returns whether missiles can be fired
func (d *Destroyer) CanFireMissile() bool {
	return d.Alive && d.MissileCapacitor >= 1.0
}

// ============================================================================
// Weapon Update Methods
// ============================================================================

// UpdateMissileWeapon charges the missile capacitor
func (d *Destroyer) UpdateMissileWeapon() {
	if d.MissileCapacitor < 1.0 {
		d.MissileCapacitor += d.MissileChargeRate
		if d.MissileCapacitor > 1.0 {
			d.MissileCapacitor = 1.0
		}
	}
	// Note: Afterburner charging is handled in UpdateWeapons (BaseShip method)
	// which charges afterburner when the main weapon capacitor is full
}
