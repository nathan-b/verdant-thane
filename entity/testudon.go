package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nathan/verdant-thane/config"
)

// Testudon is a slow, heavily-armored ship with a 360° beam weapon
type Testudon struct {
	*BaseShip
	// Beam weapon
	BeamRange             float64
	BeamDamagePerTick     float64
	BeamDamageAccumulator float64
	BeamTargetID          int // Primary movement target
	BeamFiringAtID        int // Currently firing at (for rendering)
	// Under attack tracking (for defensive AI)
	AttackerIDs []int
}

// NewTestudon creates a new testudon ship
func NewTestudon(id int, factionID int, x, y float64, sprite *ebiten.Image) *Testudon {
	chars := config.GetShipCharacteristics(ClassTestudon)
	beamWeapon := config.WeaponDatabase["beam"]

	base := &BaseShip{
		ID:                         id,
		FactionID:                  factionID,
		Class:                      ClassTestudon,
		X:                          x,
		Y:                          y,
		VelocityX:                  0,
		VelocityY:                  0,
		Rotation:                   0,
		Health:                     chars.MaxShield,
		MaxHealth:                  chars.MaxShield,
		Speed:                      0,
		MaxSpeed:                   chars.MaxSpeed,
		Accel:                      chars.Acceleration,
		CollisionRadius:            chars.CollisionRadius,
		PlayerControlled:           false,
		Weapons:                    []Weapon{}, // Testudons don't use projectile weapons (beam weapon only)
		AfterburnerCharge:          0.0,
		AfterburnerActive:          false,
		AfterburnerMaxCharge:       0.0,
		HasAfterburnerSystem:       false, // Testudons do NOT have afterburner
		AfterburnerDrain:           chars.AfterburnerDrain,
		AfterburnerRecharge:        chars.AfterburnerRecharge,
		AfterburnerAccelMultiplier: chars.AfterburnerAccelMultiplier,
		AITargetID:                 -1,
		AIRetargetTimer:            config.AIRetargetInterval,
		AIAccurateShotProbability:  chars.AIAccurateShotProbability,
		AIRandomShotProbability:    chars.AIRandomShotProbability,
		KillScore:                  chars.KillScore,
		Sprite:                     sprite,
		Alive:                      true,
	}

	return &Testudon{
		BaseShip:  base,
		BeamRange: beamWeapon.MaxRange,
		// IMPORTANT: BeamDamagePerTick must be exactly representable in binary floating point
		// to avoid accumulation errors that affect game balance.
		//
		// 0.125 = 1/8, which is exactly representable in binary (2^-3).
		// This means 8 ticks will accumulate to exactly 1.0 damage with no rounding error.
		//
		// DO NOT use values like 0.1 (1/10) which causes floating point drift:
		//   0.1 is not exactly representable in binary (repeating decimal in binary)
		//   After 10 additions: 0.1*10 = 0.999999... (requires 11 ticks for 1 damage!)
		//   This creates a ~10% DPS reduction from intended design.
		//
		// Safe values are powers of 2: 0.5 (1/2), 0.25 (1/4), 0.125 (1/8), 0.0625 (1/16)
		// Current setting from config: 8 ticks per 1 damage = 7.5 DPS at 60 TPS
		BeamDamagePerTick:     beamWeapon.DamagePerTick,
		BeamDamageAccumulator: 0.0,
		BeamTargetID:          -1,
		BeamFiringAtID:        -1,
		AttackerIDs:           []int{},
	}
}

// ============================================================================
// Entity Interface Implementation
// ============================================================================

// Update handles all per-frame logic for the testudon
func (t *Testudon) Update(ctx GameContext) error {
	if !t.Alive {
		return nil
	}

	// Update beam weapon
	t.UpdateBeamWeapon(ctx)

	// Update control (AI only - testudons are never player-controlled)
	if !t.PlayerControlled {
		t.UpdateAI(ctx)
	}

	// Update movement
	t.UpdateMovement()

	return nil
}

// ============================================================================
// Ship Interface Implementation (Beam-specific)
// ============================================================================

// TakeDamage reduces the testudon's health and tracks attackers
func (t *Testudon) TakeDamage(amount int, attackerID int, ctx GameContext) {
	if !t.Alive {
		return
	}

	t.Health -= amount
	if t.Health <= 0 {
		t.Health = 0
		t.Alive = false

		// Spawn explosion
		ctx.SpawnExplosion(t.X, t.Y)

		// Notify chat window of ship destruction
		ctx.OnShipDestroyed(t.ID, attackerID)

		// Update score if killed by player
		attacker := ctx.GetShip(attackerID)
		if attacker != nil && attacker.IsPlayerControlled() {
			ctx.AddKill()
			ctx.AddScore(t.KillScore)
		}
	} else {
		// Track attacker for defensive AI
		t.TrackAttacker(attackerID, ctx)
	}
}

// FireWeapon is not used by testudons (they have beam weapons)
func (t *Testudon) FireWeapon(targetX, targetY float64, ctx GameContext) {
	// Testudons use beam weapons, not projectiles
}

// FireMissile is not available on testudons
func (t *Testudon) FireMissile(targetID int, ctx GameContext) {
	// Testudons don't have missiles
}

// CanFireWeapon returns false (testudons use beams)
func (t *Testudon) CanFireWeapon() bool {
	return false
}

// CanFireMissile returns false
func (t *Testudon) CanFireMissile() bool {
	return false
}

// GetBeamTargetID returns the current beam target (for rendering)
func (t *Testudon) GetBeamTargetID() int {
	return t.BeamFiringAtID
}

// ============================================================================
// AI and Targeting Methods
// ============================================================================

// UpdateAI handles AI decision-making for testudons (priority-based targeting)
func (t *Testudon) UpdateAI(ctx GameContext) {
	// Retarget timer
	t.AIRetargetTimer--
	if t.AIRetargetTimer <= 0 || t.BeamTargetID < 0 {
		t.SelectTarget(ctx)
		t.AIRetargetTimer = config.AIRetargetInterval
	}

	// Validate current target
	target := ctx.GetShip(t.BeamTargetID)
	if target == nil || !target.IsAlive() {
		t.SelectTarget(ctx)
		t.AIRetargetTimer = config.AIRetargetInterval
		target = ctx.GetShip(t.BeamTargetID)
	}

	// Movement AI (move toward target)
	if target != nil {
		targetX, targetY := target.GetPosition()
		dx, dy := GetWrappedDistance(t.X, t.Y, targetX, targetY)
		// Sprites face UP (Y-axis), so use atan2(dx, -dy)
		angleToTarget := math.Atan2(dx, -dy)

		// Rotate toward target
		angleDiff := NormalizeAngle(angleToTarget - t.Rotation)
		if math.Abs(angleDiff) > config.AIRotationSpeed {
			if angleDiff > 0 {
				t.Rotation += config.AIRotationSpeed
			} else {
				t.Rotation -= config.AIRotationSpeed
			}
			t.Rotation = NormalizeAngle(t.Rotation)
		} else {
			t.Rotation = angleToTarget
		}

		// Move toward target at max speed
		targetSpeed := t.MaxSpeed

		// Adjust speed toward target speed
		if t.Speed < targetSpeed {
			t.Speed += t.Accel
			if t.Speed > targetSpeed {
				t.Speed = targetSpeed
			}
		} else if t.Speed > targetSpeed {
			t.Speed -= t.Accel
			if t.Speed < targetSpeed {
				t.Speed = targetSpeed
			}
		}
	} else {
		// Patrol behavior
		targetSpeed := t.MaxSpeed * config.AIPatrolSpeed

		// Adjust speed toward target speed
		if t.Speed < targetSpeed {
			t.Speed += t.Accel
			if t.Speed > targetSpeed {
				t.Speed = targetSpeed
			}
		} else if t.Speed > targetSpeed {
			t.Speed -= t.Accel
			if t.Speed < targetSpeed {
				t.Speed = targetSpeed
			}
		}
	}
}

// SelectTarget uses priority-based targeting
// Priority: 1. Attackers (defensive), 2. Enemy Testudons, 3. Enemy Destroyers, 4. Enemy Fighters
func (t *Testudon) SelectTarget(ctx GameContext) {
	// First, check for attackers (highest priority)
	if len(t.AttackerIDs) > 0 {
		// Clean up invalid attackers and find nearest
		validAttackers := []int{}
		var nearestAttackerID int = -1
		nearestDistance := math.MaxFloat64

		for _, attackerID := range t.AttackerIDs {
			attacker := ctx.GetShip(attackerID)
			if attacker != nil && attacker.IsAlive() {
				validAttackers = append(validAttackers, attackerID)
				attackerX, attackerY := attacker.GetPosition()
				dist := Distance(t.X, t.Y, attackerX, attackerY)
				if dist < nearestDistance {
					nearestDistance = dist
					nearestAttackerID = attackerID
				}
			}
		}
		t.AttackerIDs = validAttackers

		if nearestAttackerID >= 0 {
			t.BeamTargetID = nearestAttackerID
			return
		}
	}

	// Priority-based targeting: Testudons > Destroyers > Fighters
	allShips := ctx.GetAllShips()

	var bestTargetID int = -1
	var bestDistance float64 = math.MaxFloat64
	bestPriority := -1

	for _, ship := range allShips {
		// Skip friendlies
		if ship.GetFaction() == t.FactionID {
			continue
		}

		// Skip self (shouldn't happen, but safety check)
		if ship.GetID() == t.ID {
			continue
		}

		// Skip dead ships
		if !ship.IsAlive() {
			continue
		}

		// Determine priority based on ship class
		var priority int
		switch ship.GetClass() {
		case ClassTestudon:
			priority = 3 // Highest priority
		case ClassDestroyer:
			priority = 2
		case ClassFighter:
			priority = 1
		default:
			priority = 0
		}

		shipX, shipY := ship.GetPosition()
		dist := Distance(t.X, t.Y, shipX, shipY)

		// Select ship if higher priority, or same priority but closer
		if priority > bestPriority || (priority == bestPriority && dist < bestDistance) {
			bestPriority = priority
			bestDistance = dist
			bestTargetID = ship.GetID()
		}
	}

	t.BeamTargetID = bestTargetID
}

// TrackAttacker adds an attacker to the testudon's attacker list
func (t *Testudon) TrackAttacker(attackerID int, ctx GameContext) {
	// Check if already tracked
	for _, id := range t.AttackerIDs {
		if id == attackerID {
			return
		}
	}

	// Add new attacker
	t.AttackerIDs = append(t.AttackerIDs, attackerID)
}

// ============================================================================
// Beam Weapon Methods
// ============================================================================

// UpdateBeamWeapon handles beam targeting and damage application
func (t *Testudon) UpdateBeamWeapon(ctx GameContext) {
	var targetToFire Ship

	// First, check if primary target is valid and in range
	if t.BeamTargetID >= 0 {
		target := ctx.GetShip(t.BeamTargetID)
		if target != nil && target.IsAlive() && target.GetFaction() != t.FactionID {
			targetX, targetY := target.GetPosition()
			if IsInRange(t.X, t.Y, targetX, targetY, t.BeamRange) {
				targetToFire = target
			}
		}
	}

	// If primary target not in range, opportunistically fire at ANY in-range enemy
	if targetToFire == nil {
		nearestInRange, dist := ctx.FindNearestEnemy(t)
		if nearestInRange != nil && dist <= t.BeamRange {
			targetToFire = nearestInRange
		}
	}

	// Fire at the chosen target (if any)
	if targetToFire != nil {
		t.BeamFiringAtID = targetToFire.GetID()
		t.ApplyBeamDamage(targetToFire, ctx)
	} else {
		// No target in range - reset damage accumulator
		t.BeamDamageAccumulator = 0.0
		t.BeamFiringAtID = -1
	}
}

// ApplyBeamDamage applies damage-over-time to a target
func (t *Testudon) ApplyBeamDamage(target Ship, ctx GameContext) {
	// Accumulate damage
	t.BeamDamageAccumulator += t.BeamDamagePerTick

	// Apply integer damage when accumulator >= 1.0
	if t.BeamDamageAccumulator >= 1.0 {
		damageToApply := int(t.BeamDamageAccumulator)
		t.BeamDamageAccumulator -= float64(damageToApply)

		// Apply damage to target
		target.TakeDamage(damageToApply, t.ID, ctx)

		// Play impact sound if player ship was hit
		ctx.PlayImpactSound(target)

		// If target was destroyed, clear beam target
		if !target.IsAlive() {
			t.BeamTargetID = -1
			t.BeamFiringAtID = -1
		}
	}
}
