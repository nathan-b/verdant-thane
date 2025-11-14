package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/nathan/verdant-thane/config"
)

// Destroyer is a heavy fighter with rear-facing missiles
type Destroyer struct {
	*BaseShip
	// Weapon 0: Main gun (forward)
	// Weapon 1: Missile launcher (rear 180° arc)
}

// NewDestroyer creates a new destroyer ship
func NewDestroyer(id int, factionID int, x, y float64, sprite *ebiten.Image) *Destroyer {
	chars := config.GetShipCharacteristics(ClassDestroyer)

	base := &BaseShip{
		ID:               id,
		FactionID:        factionID,
		Class:            ClassDestroyer,
		X:                x,
		Y:                y,
		VelocityX:        0,
		VelocityY:        0,
		Rotation:         0,
		Health:           chars.MaxShield,
		MaxHealth:        chars.MaxShield,
		Speed:            0,
		MaxSpeed:         chars.MaxSpeed,
		Accel:            chars.Acceleration,
		CollisionRadius:  chars.CollisionRadius,
		PlayerControlled: false,
		// Weapons stored in priority order (lower priority number = higher priority = earlier in array)
		Weapons: []Weapon{
			{
				WeaponCapacitor:  1.0, // Missile launcher - start fully charged
				WeaponChargeRate: chars.Weapons[1].CapacitorChargeRate,
				FiringCone:       chars.Weapons[1].FiringCone,
				MaxRange:         chars.Weapons[1].MaxRange, // From config
				Priority:         1,                         // Higher priority than main gun
				RequiresTarget:   true,                      // Missiles require target lock
				Exclusive:        true,                      // When missiles fire, don't fire other weapons
				RearFacing:       true,                      // Fires from rear
				ProjectileType:   config.MissileProjectile,
			},
			{
				WeaponCapacitor:  1.0, // Main gun - start fully charged
				WeaponChargeRate: chars.Weapons[0].CapacitorChargeRate,
				FiringCone:       chars.Weapons[0].FiringCone,
				MaxRange:         chars.Weapons[0].MaxRange, // From config
				Priority:         2,
				RequiresTarget:   false,
				Exclusive:        false,
				RearFacing:       false,
				ProjectileType:   config.LaserProjectile,
			},
		},
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
		BaseShip: base,
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

	// Update all weapons (main gun and missiles)
	d.UpdateWeapons()

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
	if d.CanFireMissile() && len(d.Weapons) > 0 {
		// Missile launcher is now at index 0 (highest priority)
		rearTarget, _ := ctx.FindNearestEnemyInArc(d, d.Weapons[0].FiringCone, d.Weapons[0].MaxRange, true)
		if rearTarget != nil {
			d.FireMissile(rearTarget.GetID(), ctx)
		}
	}
}

// ============================================================================
// Ship Interface Implementation (Missile-specific)
// ============================================================================

func (d *Destroyer) FireWeapon(mouseX, mouseY float64, ctx GameContext) {
	// Use the unified weapon system which already handles priority and target requirements
	d.BaseShip.FireWeapon(mouseX, mouseY, ctx)
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

	// Consume capacitor (missile launcher is at index 0)
	d.Weapons[0].WeaponCapacitor = 0.0

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
	// Missile launcher is at index 0 (highest priority)
	return d.Alive && len(d.Weapons) > 0 && d.Weapons[0].WeaponCapacitor >= 1.0
}
