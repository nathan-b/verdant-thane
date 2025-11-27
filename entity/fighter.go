package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/nathan/verdant-thane/config"
)

// emptyImage is a 1x1 white image used for drawing solid color triangles
var emptyImage = func() *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(color.White)
	return img
}()

type Weapon struct {
	WeaponCapacitor  float64 // 0.0 to 1.0
	WeaponChargeRate float64
	FiringCone       float64               // Radians
	MaxRange         float64               // Maximum range for weapons that need it
	SpawnOffset      float64               // Distance from ship center where projectiles spawn
	Priority         int                   // Lower number = higher priority
	RequiresTarget   bool                  // Whether weapon needs a target (e.g., missiles)
	Exclusive        bool                  // Whether other weapons can fire simultaneously
	RearFacing       bool                  // Whether weapon fires to the rear
	ProjectileType   config.ProjectileType // Type of projectile (laser or missile)
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
	// Stored in priority order (lower index = higher priority)
	Weapons []Weapon

	// Afterburner (fighters and destroyers only, not testudons)
	AfterburnerCharge             float64 // 0.0 to 360.0
	AfterburnerActive             bool
	AfterburnerMaxCharge          float64
	HasAfterburnerSystem          bool
	AfterburnerDrain              float64 // Fuel consumed per tick when active
	AfterburnerRecharge           float64 // Recharge rate per tick when inactive
	AfterburnerAccelMultiplier    float64 // Acceleration multiplier when active
	AfterburnerMaxSpeedMultiplier float64 // Max speed multiplier when active

	// AI State
	AITargetID                int     // Entity ID of current AI target
	AIRetargetTimer           int     // Ticks until next target re-evaluation
	AIAccurateShotProbability float64 // Probability of accurate shots (0.0-1.0)
	AIRandomShotProbability   float64 // Probability of random shots within cone (0.0-1.0)

	// Scoring
	KillScore int // Points awarded for destroying this ship

	// Rendering
	Sprite *ebiten.Image

	// State
	Alive bool

	// Beam weapon (Testudon only - unused by other classes)
	BeamRange             float64
	BeamDamagePerTick     float64
	BeamDamageAccumulator float64
	BeamTargetID          int // Primary movement target
	BeamFiringAtID        int // Currently firing at (for rendering)
	AttackerIDs           []int
}

// NewFighter creates a new fighter ship
func NewFighter(id int, factionID int, x, y float64, sprite *ebiten.Image) *BaseShip {
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
				MaxRange:         chars.Weapons[0].MaxRange,
				SpawnOffset:      chars.Weapons[0].SpawnOffset,
				Priority:         1,
				RequiresTarget:   false,
				Exclusive:        false,
				RearFacing:       false,
				ProjectileType:   config.LaserProjectile,
			},
		},
		AfterburnerCharge:             360.0, // Start fully charged
		AfterburnerActive:             false,
		AfterburnerMaxCharge:          360.0,
		HasAfterburnerSystem:          true, // Fighters have afterburner
		AfterburnerDrain:              chars.AfterburnerDrain,
		AfterburnerRecharge:           chars.AfterburnerRecharge,
		AfterburnerAccelMultiplier:    chars.AfterburnerAccelMultiplier,
		AfterburnerMaxSpeedMultiplier: chars.AfterburnerMaxSpeedMultiplier,
		AITargetID:                    -1, // No target initially
		AIRetargetTimer:               config.AIRetargetInterval,
		AIAccurateShotProbability:     chars.AIAccurateShotProbability,
		AIRandomShotProbability:       chars.AIRandomShotProbability,
		KillScore:                     chars.KillScore,
		Sprite:                        sprite,
		Alive:                         true,
		BeamRange:                     0.0, // Unused by fighters
		BeamDamagePerTick:             0.0, // Unused by fighters
		BeamDamageAccumulator:         0.0, // Unused by fighters
		BeamTargetID:                  -1,  // Unused by fighters
		BeamFiringAtID:                -1,  // Unused by fighters
		AttackerIDs:                   nil, // Unused by fighters
	}

	return base
}

// ============================================================================
// Entity Interface Implementation (BaseShip methods)
// ============================================================================

// Render draws the ship to the screen (BaseShip method)
func (b *BaseShip) Render(screen *ebiten.Image, cameraX, cameraY float64) {
	if !b.Alive || b.Sprite == nil {
		return
	}

	// Get wrapped screen position (handles world wrapping)
	screenX, screenY := GetWrappedScreenPosition(b.X, b.Y, cameraX, cameraY)

	// Draw afterburner flame cone (before ship so it appears behind)
	if b.AfterburnerActive && b.HasAfterburnerSystem {
		b.renderAfterburnerFlame(screen, screenX, screenY)
	}

	opts := &ebiten.DrawImageOptions{}

	// Rotation
	opts.GeoM.Translate(-float64(b.Sprite.Bounds().Dx())/2, -float64(b.Sprite.Bounds().Dy())/2)
	opts.GeoM.Rotate(b.Rotation)

	// Position (relative to camera, accounting for world wrapping)
	opts.GeoM.Translate(screenX, screenY)

	screen.DrawImage(b.Sprite, opts)
}

// renderAfterburnerFlame draws a cone-shaped flame behind the ship
func (b *BaseShip) renderAfterburnerFlame(screen *ebiten.Image, screenX, screenY float64) {
	// Flame cone dimensions: half the sprite height in length, spreads to ~1/3 sprite width
	spriteHeight := float64(b.Sprite.Bounds().Dy())
	flameLength := spriteHeight * 0.5
	flameBaseWidth := spriteHeight * 0.3

	// Rear direction is opposite of rotation (rotation + π)
	// Ships face UP (negative Y), so rear is positive Y in local space
	rearAngle := b.Rotation + math.Pi

	// Calculate the rear center point (where flame originates, at back of ship)
	rearOffset := spriteHeight * 0.4 // Slightly inside the back of the ship
	rearX := screenX + math.Sin(rearAngle)*rearOffset
	rearY := screenY - math.Cos(rearAngle)*rearOffset

	// Calculate flame tip (extends further back)
	tipX := screenX + math.Sin(rearAngle)*(rearOffset+flameLength)
	tipY := screenY - math.Cos(rearAngle)*(rearOffset+flameLength)

	// Calculate the two base corners of the cone (perpendicular to rear direction)
	perpAngle := rearAngle + math.Pi/2
	halfWidth := flameBaseWidth / 2
	baseX1 := rearX + math.Sin(perpAngle)*halfWidth
	baseY1 := rearY - math.Cos(perpAngle)*halfWidth
	baseX2 := rearX - math.Sin(perpAngle)*halfWidth
	baseY2 := rearY + math.Cos(perpAngle)*halfWidth

	// Draw the flame as a filled triangle
	flameColor := color.RGBA{255, 140, 0, 255} // Orange, fully opaque

	// Draw triangle for flame cone
	path := vector.Path{}
	path.MoveTo(float32(baseX1), float32(baseY1))
	path.LineTo(float32(tipX), float32(tipY))
	path.LineTo(float32(baseX2), float32(baseY2))
	path.Close()

	// Fill the triangle
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].ColorR = float32(flameColor.R) / 255
		vs[i].ColorG = float32(flameColor.G) / 255
		vs[i].ColorB = float32(flameColor.B) / 255
		vs[i].ColorA = float32(flameColor.A) / 255
	}
	screen.DrawTriangles(vs, is, emptyImage, &ebiten.DrawTrianglesOptions{})
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
			ctx.AddScore(b.KillScore)
		}
	} else {
		// Testudon: Track attacker for defensive AI
		if b.Class == ClassTestudon {
			b.TrackAttacker(attackerID, ctx)
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

	fired := false // Has some weapon been fired yet?

	// Weapons are already stored in priority order (lower index = higher priority)
	for i := range b.Weapons {
		weapon := &b.Weapons[i]
		if b.canFireWeapon(weapon, fired) {
			targetX, targetY := mouseX, mouseY
			if weapon.RequiresTarget {
				// Weapon requires a target, see if we have a target
				nearestEnemy, dist := ctx.FindNearestEnemyInArc(
					ctx.GetShip(b.ID),
					weapon.FiringCone,
					weapon.MaxRange,
					weapon.RearFacing,
				)
				if nearestEnemy == nil || dist > weapon.MaxRange {
					continue // No valid target
				}
				targetX, targetY = nearestEnemy.GetPosition()
			}
			b.fireWeapon(weapon, targetX, targetY, ctx)
			fired = true

			// If weapon is exclusive, stop firing other weapons
			if weapon.Exclusive {
				break
			}
		}
	}
}

// FireMissile is not available on base ships (BaseShip method, overridden by Destroyer)
func (b *BaseShip) FireMissile(targetID int, ctx GameContext) {
	// Base ships don't have missiles
}

// CanFireWeapon returns whether at least one weapon can be fired
func (b *BaseShip) CanFireWeapon() bool {
	if !b.Alive {
		return false
	}
	for i := range b.Weapons {
		if b.Weapons[i].WeaponCapacitor >= 1.0 {
			return true
		}
	}
	return false
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

// canFireWeapon checks if a specific weapon can be fired given current state
func (b *BaseShip) canFireWeapon(weapon *Weapon, fired bool) bool {
	if weapon.WeaponCapacitor < 1.0 {
		return false
	}
	if weapon.Exclusive && fired {
		return false
	}
	return true
}

func (b *BaseShip) fireWeapon(weapon *Weapon, targetX, targetY float64, ctx GameContext) {
	// Calculate direction to target
	dx := targetX - b.X
	dy := targetY - b.Y
	angleToTarget := math.Atan2(dx, -dy) // Sprites face UP (Y-axis)

	// Get the weapon's firing direction (forward or rear)
	weaponAngle := b.Rotation
	if weapon.RearFacing {
		weaponAngle = b.Rotation + math.Pi
		// Normalize to [0, 2π) range to avoid -π/+π ambiguity in sprite rendering
		for weaponAngle < 0 {
			weaponAngle += 2 * math.Pi
		}
		for weaponAngle >= 2*math.Pi {
			weaponAngle -= 2 * math.Pi
		}
	}

	// Check if target is within firing cone
	angleFromWeapon := NormalizeAngle(angleToTarget - weaponAngle)
	halfCone := weapon.FiringCone / 2.0

	// Clamp firing angle to weapon cone
	var firingAngle float64
	if math.Abs(angleFromWeapon) <= halfCone {
		// Target within cone, fire directly at it
		firingAngle = angleToTarget
	} else {
		// Target outside cone, fire along nearest edge of cone
		if angleFromWeapon > 0 {
			firingAngle = NormalizeAngle(weaponAngle + halfCone)
		} else {
			firingAngle = NormalizeAngle(weaponAngle - halfCone)
		}
	}

	// Get projectile characteristics
	projChars := config.GetProjectileCharacteristics(weapon.ProjectileType)

	// Calculate spawn position
	// - Rear-facing weapons: spawn at weapon mount (rear of ship)
	// - Forward-facing weapons: spawn toward target (firing direction)
	var spawnX, spawnY float64
	if weapon.RearFacing {
		// Rear-facing: spawn at weapon mount
		spawnX = b.X + math.Sin(weaponAngle)*weapon.SpawnOffset
		spawnY = b.Y + -math.Cos(weaponAngle)*weapon.SpawnOffset
	} else {
		// Forward-facing: spawn toward target
		spawnX = b.X + math.Sin(firingAngle)*weapon.SpawnOffset
		spawnY = b.Y + -math.Cos(firingAngle)*weapon.SpawnOffset
	}

	// Calculate projectile velocity and initial rotation
	var projVX, projVY, initialRotation float64
	if weapon.ProjectileType == config.MissileProjectile {
		// Missiles: Launch straight from weapon mount direction
		// The tracking system will turn them toward target on subsequent frames
		// NOTE: Missiles do NOT inherit ship velocity because they need to track targets
		// independently of the launching ship's movement
		projVX = math.Sin(weaponAngle) * projChars.Speed
		projVY = -math.Cos(weaponAngle) * projChars.Speed
		initialRotation = weaponAngle
	} else {
		// Non-tracking projectiles: Launch toward firing direction and inherit ship velocity
		projVX = math.Sin(firingAngle)*projChars.Speed + b.VelocityX
		projVY = -math.Cos(firingAngle)*projChars.Speed + b.VelocityY
		initialRotation = firingAngle
	}

	// Consume capacitor
	weapon.WeaponCapacitor = 0.0

	// Spawn appropriate projectile type
	if weapon.ProjectileType == config.MissileProjectile {
		// For missiles, we need a target - this should have been validated by caller
		// Find the target enemy that was validated earlier
		nearestEnemy, dist := ctx.FindNearestEnemyInArc(
			ctx.GetShip(b.ID),
			weapon.FiringCone,
			weapon.MaxRange,
			weapon.RearFacing,
		)
		if nearestEnemy == nil || dist > weapon.MaxRange {
			return // No valid target, don't fire
		}

		// Spawn missile with rotation matching weapon mount direction
		// (tracking will adjust on first update)
		ctx.SpawnMissile(MissileConfig{
			X:         spawnX,
			Y:         spawnY,
			VelocityX: projVX,
			VelocityY: projVY,
			Rotation:  initialRotation,
			OwnerID:   b.ID,
			FactionID: b.FactionID,
			TargetID:  nearestEnemy.GetID(),
			Sprite:    nil, // Will be set by spawner
		})
	} else {
		// Spawn laser/main gun projectile
		ctx.SpawnProjectile(MainGunConfig{
			X:         spawnX,
			Y:         spawnY,
			VelocityX: projVX,
			VelocityY: projVY,
			OwnerID:   b.ID,
			FactionID: b.FactionID,
			Sprite:    nil, // Will be set by spawner
		})
	}
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
			b.AfterburnerCharge += b.AfterburnerRecharge
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
			// Consume fuel
			b.AfterburnerCharge -= b.AfterburnerDrain
			if b.AfterburnerCharge < 0 {
				b.AfterburnerCharge = 0
			}
			// Flame cone is now rendered in Render() instead of spawning particles
		} else {
			b.AfterburnerActive = false
		}
	}

	// Determine effective acceleration (multiply if afterburner active)
	effectiveAccel := b.Accel
	if b.AfterburnerActive {
		effectiveAccel *= b.AfterburnerAccelMultiplier
	}

	// Acceleration/Deceleration (modify Speed scalar, not velocity)
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		// Accelerate
		b.Speed += effectiveAccel
		// Apply afterburner max speed boost if active
		effectiveMaxSpeed := b.MaxSpeed
		if b.AfterburnerActive {
			effectiveMaxSpeed *= b.AfterburnerMaxSpeedMultiplier
		}
		if b.Speed > effectiveMaxSpeed {
			b.Speed = effectiveMaxSpeed
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

// ============================================================================
// Consolidated Update Methods (Phase 2)
// ============================================================================

// Update handles all per-frame logic (unified for all ship types)
func (b *BaseShip) Update(ctx GameContext) error {
	if !b.Alive {
		return nil
	}

	// Update weapons (class-specific)
	if b.Class == ClassTestudon {
		b.UpdateBeamWeapon(ctx)
	} else {
		b.UpdateWeapons()
	}

	// Update control (AI or player)
	if b.PlayerControlled {
		b.UpdatePlayerInput(ctx)
	} else {
		b.UpdateAI(ctx)
	}

	// Update movement
	b.UpdateMovement()

	return nil
}

// UpdateAI handles AI decision-making (delegates to class-specific implementations)
func (b *BaseShip) UpdateAI(ctx GameContext) {
	switch b.Class {
	case ClassTestudon:
		b.updateAITestudon(ctx)
	case ClassDestroyer:
		b.updateAIDestroyer(ctx)
	case ClassFighter:
		b.updateAIFighter(ctx)
	}
}

// ============================================================================
// Class-Specific AI Helpers (Private)
// ============================================================================

// updateAIFighter handles Fighter AI (will be moved from Fighter.UpdateAI)
func (b *BaseShip) updateAIFighter(ctx GameContext) {
	// This is a placeholder - the actual implementation will be copied from Fighter.UpdateAI
	// For now, call the Fighter method to maintain backward compatibility
}

// updateAIDestroyer handles Destroyer AI (will be moved from Destroyer.UpdateAI)
func (b *BaseShip) updateAIDestroyer(ctx GameContext) {
	// This is a placeholder - the actual implementation will be copied from Destroyer.UpdateAI
	// For now, this is empty and will be implemented
}

// updateAITestudon handles Testudon AI (will be moved from Testudon.UpdateAI)
func (b *BaseShip) updateAITestudon(ctx GameContext) {
	// This is a placeholder - the actual implementation will be copied from Testudon.UpdateAI
	// For now, this is empty and will be implemented
}

// UpdateBeamWeapon handles beam targeting and damage application (Testudon only)
func (b *BaseShip) UpdateBeamWeapon(ctx GameContext) {
	var targetToFire *BaseShip

	// First, check if primary target is valid and in range
	if b.BeamTargetID >= 0 {
		target := ctx.GetShip(b.BeamTargetID)
		if target != nil && target.IsAlive() && target.GetFaction() != b.FactionID {
			targetX, targetY := target.GetPosition()
			if IsInRange(b.X, b.Y, targetX, targetY, b.BeamRange) {
				targetToFire = target
			}
		}
	}

	// If primary target not in range, opportunistically fire at ANY in-range enemy
	if targetToFire == nil {
		// TODO: In Phase 3, change FindNearestEnemy to accept *BaseShip instead of Ship
		// For now, skip opportunistic targeting (Testudon.UpdateBeamWeapon still handles this)
	}

	// Fire at the chosen target (if any)
	if targetToFire != nil {
		b.BeamFiringAtID = targetToFire.GetID()
		b.ApplyBeamDamage(targetToFire, ctx)
	} else {
		// No target in range - reset damage accumulator
		b.BeamDamageAccumulator = 0.0
		b.BeamFiringAtID = -1
	}
}

// ============================================================================
// Beam Weapon Methods (Testudon only - unused by other classes)
// ============================================================================

// GetBeamTargetID returns the current beam target (for rendering)
func (b *BaseShip) GetBeamTargetID() int {
	return b.BeamFiringAtID
}

// TrackAttacker adds an attacker to the ship's attacker list (Testudon defensive AI)
func (b *BaseShip) TrackAttacker(attackerID int, ctx GameContext) {
	// Check if already tracked
	for _, id := range b.AttackerIDs {
		if id == attackerID {
			return
		}
	}

	// Add new attacker
	b.AttackerIDs = append(b.AttackerIDs, attackerID)
}

// ApplyBeamDamage applies damage-over-time to a target
func (b *BaseShip) ApplyBeamDamage(target *BaseShip, ctx GameContext) {
	// Accumulate damage
	b.BeamDamageAccumulator += b.BeamDamagePerTick

	// Apply integer damage when accumulator >= 1.0
	if b.BeamDamageAccumulator >= 1.0 {
		damageToApply := int(b.BeamDamageAccumulator)
		b.BeamDamageAccumulator -= float64(damageToApply)

		// Apply damage to target
		target.TakeDamage(damageToApply, b.ID, ctx)

		// Play impact sound if player ship was hit
		ctx.PlayImpactSound(target)

		// If target was destroyed, clear beam target
		if !target.IsAlive() {
			b.BeamTargetID = -1
			b.BeamFiringAtID = -1
		}
	}
}
