package entity

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/nathan/verdant-thane/config"
)

type Weapon struct {
	WeaponCapacitor  float64 // 0.0 to 1.0
	WeaponChargeRate float64
	FiringCone       float64 // Radians
}

// BaseShip contains common data for all ship types
type BaseShip struct {
	// Identity
	ID        int
	FactionID int
	Class     ShipClass

	// Physics
	X, Y                 float64
	VelocityX, VelocityY float64
	Rotation             float64 // Radians

	// Stats
	Health          int
	MaxHealth       int
	Speed           float64 // Current speed
	MaxSpeed        float64
	Accel           float64
	CollisionRadius float64

	// Control
	PlayerControlled bool

	// Weapons
	Weapons []Weapon

	// Afterburner (fighters and destroyers only, not testudons)
	AfterburnerCharge    float64 // 0.0 to 360.0
	AfterburnerActive    bool
	AfterburnerMaxCharge float64
	HasAfterburnerSystem bool

	// AI State
	AITargetID      int // Entity ID of current AI target
	AIRetargetTimer int // Ticks until next target re-evaluation

	// Rendering
	Sprite *ebiten.Image

	// State
	Alive bool
}

// Fighter is a standard fighter ship
type Fighter struct {
	*BaseShip
}

// NewFighter creates a new fighter ship
func NewFighter(id int, factionID int, x, y float64, sprite *ebiten.Image) *Fighter {
	chars := config.GetShipCharacteristics(ClassFighter)

	base := &BaseShip{
		ID:               id,
		FactionID:        factionID,
		Class:            ClassFighter,
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
		Weapons: []Weapon{
			{
				WeaponCapacitor:  1.0, // Start fully charged
				WeaponChargeRate: chars.Weapons[0].CapacitorChargeRate,
				FiringCone:       chars.Weapons[0].FiringCone,
			},
		},
		AfterburnerCharge:    360.0, // Start fully charged
		AfterburnerActive:    false,
		AfterburnerMaxCharge: 360.0,
		HasAfterburnerSystem: true, // Fighters have afterburner
		AITargetID:           -1,   // No target initially
		AIRetargetTimer:      config.AIRetargetInterval,
		Sprite:               sprite,
		Alive:                true,
	}

	return &Fighter{BaseShip: base}
}

// ============================================================================
// Entity Interface Implementation
// ============================================================================

// Update handles all per-frame logic for the fighter
func (f *Fighter) Update(ctx GameContext) error {
	if !f.Alive {
		return nil
	}

	// Update weapons
	f.UpdateWeapons()

	// Update control (AI or player)
	if f.PlayerControlled {
		f.UpdatePlayerInput(ctx)
	} else {
		f.UpdateAI(ctx)
	}

	// Update movement
	f.UpdateMovement()

	return nil
}

// Render draws the ship to the screen (BaseShip method)
func (b *BaseShip) Render(screen *ebiten.Image, cameraX, cameraY float64) {
	if !b.Alive || b.Sprite == nil {
		return
	}

	opts := &ebiten.DrawImageOptions{}

	// Rotation
	opts.GeoM.Translate(-float64(b.Sprite.Bounds().Dx())/2, -float64(b.Sprite.Bounds().Dy())/2)
	opts.GeoM.Rotate(b.Rotation)

	// Position (relative to camera)
	opts.GeoM.Translate(b.X-cameraX, b.Y-cameraY)

	screen.DrawImage(b.Sprite, opts)
}

// GetID returns the entity's unique ID (BaseShip method)
func (b *BaseShip) GetID() int {
	return b.ID
}

// GetPosition returns the ship's position (BaseShip method)
func (b *BaseShip) GetPosition() (float64, float64) {
	return b.X, b.Y
}

// IsAlive returns whether the ship is still alive (BaseShip method)
func (b *BaseShip) IsAlive() bool {
	return b.Alive
}

// ============================================================================
// Ship Interface Implementation
// ============================================================================

// TakeDamage reduces the ship's health (BaseShip method, can be overridden)
func (b *BaseShip) TakeDamage(amount int, attackerID int, ctx GameContext) {
	if !b.Alive {
		return
	}

	b.Health -= amount
	if b.Health <= 0 {
		b.Health = 0
		b.Alive = false

		// Spawn explosion
		ctx.SpawnExplosion(b.X, b.Y)

		// Notify chat window of ship destruction
		ctx.OnShipDestroyed(b.ID, attackerID)

		// Update score if killed by player
		attacker := ctx.GetShip(attackerID)
		if attacker != nil && attacker.IsPlayerControlled() {
			ctx.AddKill()
			ctx.AddScore(10) // 10 points per kill
		}
	}
}

// GetHealth returns current and max health (BaseShip method)
func (b *BaseShip) GetHealth() (int, int) {
	return b.Health, b.MaxHealth
}

// GetCollisionRadius returns the collision radius (BaseShip method)
func (b *BaseShip) GetCollisionRadius() float64 {
	return b.CollisionRadius
}

// GetFaction returns the faction ID (BaseShip method)
func (b *BaseShip) GetFaction() int {
	return b.FactionID
}

// GetClass returns the ship class (BaseShip method)
func (b *BaseShip) GetClass() ShipClass {
	return b.Class
}

// GetVelocity returns the velocity components (BaseShip method)
func (b *BaseShip) GetVelocity() (float64, float64) {
	return b.VelocityX, b.VelocityY
}

// GetRotation returns the rotation in radians (BaseShip method)
func (b *BaseShip) GetRotation() float64 {
	return b.Rotation
}

// SetPlayerControlled sets whether this ship is player-controlled (BaseShip method)
func (b *BaseShip) SetPlayerControlled(controlled bool) {
	b.PlayerControlled = controlled
}

// IsPlayerControlled returns whether this ship is player-controlled (BaseShip method)
func (b *BaseShip) IsPlayerControlled() bool {
	return b.PlayerControlled
}

// FireWeapon fires the main weapon at a target position (BaseShip method, can be overridden)
func (b *BaseShip) FireWeapon(mouseX, mouseY float64, ctx GameContext) {
	if !b.CanFireWeapon() {
		return
	}

	// Calculate angle to target
	// Sprites face UP (Y-axis), so use atan2(dx, -dy) instead of atan2(dy, dx)
	dx, dy := GetWrappedDistance(b.X, b.Y, mouseX, mouseY)
	angleToTarget := math.Atan2(dx, -dy)

	// Check if target is within firing cone (use primary weapon)
	angleFromForward := NormalizeAngle(angleToTarget - b.Rotation)
	halfCone := b.Weapons[0].FiringCone / 2

	var firingAngle float64
	if math.Abs(angleFromForward) <= halfCone {
		// Target is in cone, fire directly at it
		firingAngle = angleToTarget
	} else {
		// Target is outside cone, fire along nearest cone edge
		if angleFromForward > 0 {
			firingAngle = NormalizeAngle(b.Rotation + halfCone)
		} else {
			firingAngle = NormalizeAngle(b.Rotation - halfCone)
		}
	}

	// Consume capacitor
	b.Weapons[0].WeaponCapacitor = 0.0

	// Calculate projectile velocity
	// Sprites face UP (Y-axis), so use sin/cos adjusted for sprite orientation
	projectileSpeed := 12.0 // 2x max ship speed
	projectileVX := math.Sin(firingAngle) * projectileSpeed
	projectileVY := -math.Cos(firingAngle) * projectileSpeed

	// Add ship velocity to projectile (inheritance)
	projectileVX += b.VelocityX
	projectileVY += b.VelocityY

	// Spawn offset (spawn in front of ship)
	spawnOffset := 20.0
	spawnX := b.X + math.Sin(firingAngle)*spawnOffset
	spawnY := b.Y + -math.Cos(firingAngle)*spawnOffset

	// Spawn projectile via context
	ctx.SpawnProjectile(MainGunConfig{
		X:         spawnX,
		Y:         spawnY,
		VelocityX: projectileVX,
		VelocityY: projectileVY,
		OwnerID:   b.ID,
		FactionID: b.FactionID,
		Sprite:    nil, // Will be set by spawner
	})
}

// FireMissile is not available on base ships (BaseShip method, overridden by Destroyer)
func (b *BaseShip) FireMissile(targetID int, ctx GameContext) {
	// Base ships don't have missiles
}

// CanFireWeapon returns whether the weapon can be fired (BaseShip method)
func (b *BaseShip) CanFireWeapon() bool {
	return b.Alive && len(b.Weapons) > 0 && b.Weapons[0].WeaponCapacitor >= 1.0
}

// CanFireMissile returns whether missiles can be fired (BaseShip method, overridden by Destroyer)
func (b *BaseShip) CanFireMissile() bool {
	return false // Base ships (fighters, testudons) don't have missiles
}

// GetAfterburnerCharge returns current afterburner charge (BaseShip method)
func (b *BaseShip) GetAfterburnerCharge() float64 {
	return b.AfterburnerCharge
}

// IsAfterburnerActive returns whether afterburner is currently active (BaseShip method)
func (b *BaseShip) IsAfterburnerActive() bool {
	return b.AfterburnerActive
}

// HasAfterburner returns whether this ship has an afterburner system (BaseShip method)
func (b *BaseShip) HasAfterburner() bool {
	return b.HasAfterburnerSystem
}

// ============================================================================
// Update Methods
// ============================================================================

// UpdateWeapons charges the weapon capacitor and afterburner (BaseShip method)
func (b *BaseShip) UpdateWeapons() {
	// Charge all weapon capacitors
	allWeaponsCharged := true
	for i := range b.Weapons {
		if b.Weapons[i].WeaponCapacitor < 1.0 {
			b.Weapons[i].WeaponCapacitor += b.Weapons[i].WeaponChargeRate
			if b.Weapons[i].WeaponCapacitor > 1.0 {
				b.Weapons[i].WeaponCapacitor = 1.0
			} else {
				allWeaponsCharged = false
			}
		}
	}

	// Only charge afterburner when all weapon capacitors are full and afterburner is not active
	if allWeaponsCharged && b.HasAfterburnerSystem && !b.AfterburnerActive {
		if b.AfterburnerCharge < b.AfterburnerMaxCharge {
			b.AfterburnerCharge += 1.0 // Charge at 1 per tick
			if b.AfterburnerCharge > b.AfterburnerMaxCharge {
				b.AfterburnerCharge = b.AfterburnerMaxCharge
			}
		}
	}
}

// UpdateMovement applies velocity and handles world wrapping (BaseShip method)
func (b *BaseShip) UpdateMovement() {
	// Calculate velocity from Speed and Rotation
	// Sprites face UP (along Y-axis), so rotation 0 = facing up
	// Use sin for X and -cos for Y to account for sprite orientation
	b.VelocityX = math.Sin(b.Rotation) * b.Speed
	b.VelocityY = -math.Cos(b.Rotation) * b.Speed

	// Apply velocity
	b.X += b.VelocityX
	b.Y += b.VelocityY

	// Wrap position
	b.X, b.Y = WrapPosition(b.X, b.Y)
}

// UpdatePlayerInput handles WASD controls for player-controlled ships (BaseShip method)
func (b *BaseShip) UpdatePlayerInput(ctx GameContext) {
	// Rotation
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		b.Rotation -= config.RotationSpeed
		b.Rotation = NormalizeAngle(b.Rotation)
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		b.Rotation += config.RotationSpeed
		b.Rotation = NormalizeAngle(b.Rotation)
	}

	// Afterburner activation (space bar)
	if b.HasAfterburnerSystem {
		if ebiten.IsKeyPressed(ebiten.KeySpace) && b.AfterburnerCharge > 0 {
			b.AfterburnerActive = true
			// Consume fuel at 3 per tick
			b.AfterburnerCharge -= 3.0
			if b.AfterburnerCharge < 0 {
				b.AfterburnerCharge = 0
			}

			// Spawn afterburner particles (orange exhaust from rear of ship)
			// Rear direction is opposite of rotation (rotation + π)
			rearAngle := b.Rotation + math.Pi
			particleOffset := 15.0
			particleX := b.X + math.Sin(rearAngle)*particleOffset
			particleY := b.Y + -math.Cos(rearAngle)*particleOffset

			// Spawn 1-2 particles per tick when afterburner is active
			ctx.SpawnParticle(particleX, particleY, b.VelocityX, b.VelocityY)
		} else {
			b.AfterburnerActive = false
		}
	}

	// Determine effective acceleration (double if afterburner active)
	effectiveAccel := b.Accel
	if b.AfterburnerActive {
		effectiveAccel *= 2.0
	}

	// Acceleration/Deceleration (modify Speed scalar, not velocity)
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		// Accelerate
		b.Speed += effectiveAccel
		if b.Speed > b.MaxSpeed {
			b.Speed = b.MaxSpeed
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		// Decelerate
		decel := effectiveAccel * 0.5
		b.Speed -= decel
		if b.Speed < 0 {
			b.Speed = 0
		}
	}

	// Weapon firing (handled externally since it needs mouse position)
	// The game will call FireWeapon() when mouse button is pressed
}

// UpdateAI handles AI decision-making and movement
func (f *Fighter) UpdateAI(ctx GameContext) {
	// Retarget timer
	f.AIRetargetTimer--
	if f.AIRetargetTimer <= 0 || f.AITargetID < 0 {
		f.SelectTarget(ctx)
		f.AIRetargetTimer = config.AIRetargetInterval
	}

	// Validate current target
	target := ctx.GetShip(f.AITargetID)
	if target == nil || !target.IsAlive() {
		f.SelectTarget(ctx)
		f.AIRetargetTimer = config.AIRetargetInterval
		target = ctx.GetShip(f.AITargetID)
	}

	// AI behavior
	if target != nil {
		// Pursuit: Rotate toward target and fly at 80-100% speed
		targetX, targetY := target.GetPosition()
		dx, dy := GetWrappedDistance(f.X, f.Y, targetX, targetY)
		// Sprites face UP (Y-axis), so use atan2(dx, -dy)
		angleToTarget := math.Atan2(dx, -dy)

		// Rotate toward target
		angleDiff := NormalizeAngle(angleToTarget - f.Rotation)
		if math.Abs(angleDiff) > config.AIRotationSpeed {
			if angleDiff > 0 {
				f.Rotation += config.AIRotationSpeed
			} else {
				f.Rotation -= config.AIRotationSpeed
			}
			f.Rotation = NormalizeAngle(f.Rotation)
		} else {
			f.Rotation = angleToTarget
		}

		// Accelerate toward target at random speed (80-100%)
		targetSpeed := f.MaxSpeed * (config.AIPursuitSpeedMin + rand.Float64()*(config.AIPursuitSpeedMax-config.AIPursuitSpeedMin))

		// Adjust speed toward target speed
		if f.Speed < targetSpeed {
			f.Speed += f.Accel
			if f.Speed > targetSpeed {
				f.Speed = targetSpeed
			}
		} else if f.Speed > targetSpeed {
			f.Speed -= f.Accel
			if f.Speed < targetSpeed {
				f.Speed = targetSpeed
			}
		}

		// Fire weapon if target in arc
		if f.CanFireWeapon() {
			angleFromForward := math.Abs(NormalizeAngle(angleToTarget - f.Rotation))
			if angleFromForward <= f.Weapons[0].FiringCone/2 {
				// 50% accurate, 25% random, 25% no fire
				roll := rand.Float64()
				if roll < 0.50 {
					// Accurate shot
					f.FireWeapon(targetX, targetY, ctx)
				} else if roll < 0.75 {
					// Random shot within cone
					randomAngle := f.Rotation + (rand.Float64()-0.5)*f.Weapons[0].FiringCone
					randomTargetX := f.X + math.Cos(randomAngle)*1000
					randomTargetY := f.Y + math.Sin(randomAngle)*1000
					f.FireWeapon(randomTargetX, randomTargetY, ctx)
				}
				// Else: don't fire
			}
		}
	} else {
		// Patrol: Maintain heading at 50% speed
		targetSpeed := f.MaxSpeed * config.AIPatrolSpeed

		// Adjust speed toward target speed
		if f.Speed < targetSpeed {
			f.Speed += f.Accel
			if f.Speed > targetSpeed {
				f.Speed = targetSpeed
			}
		} else if f.Speed > targetSpeed {
			f.Speed -= f.Accel
			if f.Speed < targetSpeed {
				f.Speed = targetSpeed
			}
		}
	}
}

// SelectTarget finds the nearest enemy ship
func (f *Fighter) SelectTarget(ctx GameContext) {
	nearestShip, _ := ctx.FindNearestEnemy(f)
	if nearestShip != nil {
		f.AITargetID = nearestShip.GetID()
	} else {
		f.AITargetID = -1
	}
}
