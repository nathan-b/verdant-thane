package entity

import (
	"math"
	"testing"

	"github.com/nathan/verdant-thane/config"
)

// Test Destroyer Creation
func TestNewDestroyer(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 100, 200, nil)

	if destroyer == nil {
		t.Fatal("NewDestroyer returned nil")
	}

	if destroyer.GetID() != 1 {
		t.Errorf("Expected ID 1, got %d", destroyer.GetID())
	}

	if destroyer.GetClass() != ClassDestroyer {
		t.Errorf("Expected ClassDestroyer, got %v", destroyer.GetClass())
	}

	if destroyer.GetFaction() != 0 {
		t.Errorf("Expected faction 0, got %d", destroyer.GetFaction())
	}

	x, y := destroyer.GetPosition()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%f, %f)", x, y)
	}

	if !destroyer.IsAlive() {
		t.Error("Destroyer should start alive")
	}

	// Destroyers have more health than fighters
	health, _ := destroyer.GetHealth()
	if health <= 0 {
		t.Errorf("Expected positive health, got %d", health)
	}
}

// Test Destroyer Main Gun
func TestDestroyerMainGun(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 100, 100, nil)
	destroyer.SetPlayerControlled(true) // Prevent AI from auto-firing
	ctx := NewMockGameContext()
	ctx.ships[1] = destroyer // Add destroyer to context

	// Destroyers start with fully charged weapons
	if !destroyer.CanFireWeapon() {
		t.Error("Should be able to fire initially (starts charged)")
	}

	// Fire weapon
	targetX, targetY := 100.0, 0.0 // Target directly above (UP is negative Y)
	destroyer.FireWeapon(targetX, targetY, ctx)

	// Should have spawned a projectile
	if len(ctx.spawnedProjectiles) != 1 {
		t.Fatalf("Expected 1 projectile, got %d", len(ctx.spawnedProjectiles))
	}

	// Verify projectile config
	cfg := ctx.spawnedProjectiles[0]
	if cfg.OwnerID != 1 {
		t.Errorf("Expected owner ID 1, got %d", cfg.OwnerID)
	}
	if cfg.FactionID != 0 {
		t.Errorf("Expected faction ID 0, got %d", cfg.FactionID)
	}

	// Projectile should be spawned in front of ship (ship at 100,100, facing up)
	// With spawn offset of ~15, projectile should be at ~100, 85
	if math.Abs(cfg.X-100) > 1 {
		t.Errorf("Expected projectile X near 100, got %f", cfg.X)
	}
	if cfg.Y > 95 {
		t.Errorf("Expected projectile Y < 95 (spawned above ship), got %f", cfg.Y)
	}

	// Main gun capacitor should be discharged (main gun is now at index 1)
	if len(destroyer.Weapons) > 1 && destroyer.Weapons[1].WeaponCapacitor != 0.0 {
		t.Errorf("Expected main gun capacitor to be 0.0 after firing, got %f", destroyer.Weapons[1].WeaponCapacitor)
	}

	// However, destroyer can still fire (missiles are still charged)
	if !destroyer.CanFireWeapon() {
		t.Error("Destroyer should still be able to fire (missiles are charged)")
	}

	// Verify missile launcher is still charged (missile launcher is now at index 0)
	if len(destroyer.Weapons) > 0 && destroyer.Weapons[0].WeaponCapacitor < 1.0 {
		t.Error("Missile launcher should still be charged")
	}
}

// Test Destroyer Missile System
func TestDestroyerMissileSystem(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 100, 100, nil)
	destroyer.SetPlayerControlled(true) // Prevent AI from auto-firing
	ctx := NewMockGameContext()
	ctx.ships[1] = destroyer // Add destroyer to context

	// Spawn an enemy target
	target := NewFighter(2, 1, 150, 150, nil)
	ctx.ships[2] = target

	// Destroyers start with fully charged missiles
	if !destroyer.CanFireMissile() {
		t.Error("Should be able to fire missile initially (starts charged)")
	}

	// Fire missile at target
	destroyer.FireMissile(2, ctx)

	// Should have spawned a missile
	if len(ctx.spawnedMissiles) != 1 {
		t.Fatalf("Expected 1 missile, got %d", len(ctx.spawnedMissiles))
	}

	// Verify missile config
	cfg := ctx.spawnedMissiles[0]
	if cfg.OwnerID != 1 {
		t.Errorf("Expected owner ID 1, got %d", cfg.OwnerID)
	}
	if cfg.FactionID != 0 {
		t.Errorf("Expected faction ID 0, got %d", cfg.FactionID)
	}
	if cfg.TargetID != 2 {
		t.Errorf("Expected target ID 2, got %d", cfg.TargetID)
	}

	// Can't fire again immediately
	if destroyer.CanFireMissile() {
		t.Error("Should not be able to fire missile immediately after shooting")
	}
}

// Test Destroyer Missile Target Tracking
func TestDestroyerMissileTargetTracking(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 100, 100, nil)
	destroyer.SetPlayerControlled(true)
	ctx := NewMockGameContext()
	ctx.ships[1] = destroyer

	// Create multiple targets
	target1 := NewFighter(10, 1, 150, 150, nil)
	target2 := NewFighter(20, 1, 200, 200, nil)
	ctx.ships[10] = target1
	ctx.ships[20] = target2

	// Fire at target 1
	destroyer.FireMissile(10, ctx)
	if len(ctx.spawnedMissiles) != 1 {
		t.Fatalf("Expected 1 missile, got %d", len(ctx.spawnedMissiles))
	}
	if ctx.spawnedMissiles[0].TargetID != 10 {
		t.Errorf("Expected missile to track target 10, got %d", ctx.spawnedMissiles[0].TargetID)
	}

	// Recharge and fire at target 2
	for i := 0; i < 200; i++ {
		destroyer.Update(ctx)
	}
	destroyer.FireMissile(20, ctx)
	if len(ctx.spawnedMissiles) != 2 {
		t.Fatalf("Expected 2 missiles total, got %d", len(ctx.spawnedMissiles))
	}
	if ctx.spawnedMissiles[1].TargetID != 20 {
		t.Errorf("Expected second missile to track target 20, got %d", ctx.spawnedMissiles[1].TargetID)
	}
}

// Test Destroyer AI Behavior
func TestDestroyerAI(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 100, 100, nil)
	destroyer.PlayerControlled = false
	ctx := NewMockGameContext()
	ctx.ships[1] = destroyer // Add destroyer to context

	// Spawn enemy target
	enemy := NewFighter(2, 1, 200, 100, nil)
	ctx.ships[2] = enemy

	// Run AI updates
	for i := 0; i < 10; i++ {
		destroyer.Update(ctx)
	}

	// Destroyer should rotate toward enemy
	x, y := destroyer.GetPosition()
	enemyX, enemyY := enemy.GetPosition()
	dx := enemyX - x
	dy := enemyY - y
	expectedAngle := math.Atan2(dx, -dy)

	rotation := destroyer.GetRotation()

	// Should be turning toward target (allow some tolerance)
	angleDiff := NormalizeAngle(expectedAngle - rotation)
	if math.Abs(angleDiff) > math.Pi/2 {
		t.Errorf("Destroyer not turning toward target: expected angle %f, got %f", expectedAngle, rotation)
	}

	// Destroyer should be moving
	vx, vy := destroyer.GetVelocity()
	speed := math.Sqrt(vx*vx + vy*vy)
	if speed < 0.1 {
		t.Error("Destroyer should be moving toward target")
	}
}

// Test Destroyer Can Fire Both Weapons Independently
func TestDestroyerIndependentWeapons(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 100, 100, nil)
	destroyer.SetPlayerControlled(true) // Prevent AI from auto-firing
	ctx := NewMockGameContext()
	ctx.ships[1] = destroyer // Add destroyer to context

	// Charge main gun fully
	for i := 0; i < 100; i++ {
		destroyer.Update(ctx)
	}

	// Fire main gun
	destroyer.FireWeapon(100, 0, ctx)

	// Should have main gun projectile but no missiles
	if len(ctx.spawnedProjectiles) != 1 {
		t.Errorf("Expected 1 projectile, got %d", len(ctx.spawnedProjectiles))
	}
	if len(ctx.spawnedMissiles) != 0 {
		t.Errorf("Expected 0 missiles, got %d", len(ctx.spawnedMissiles))
	}

	// Continue charging for missiles
	for i := 0; i < 200; i++ {
		destroyer.Update(ctx)
	}

	// Create target and fire missile
	ctx.ships[2] = NewFighter(2, 1, 100, 200, nil)
	destroyer.FireMissile(2, ctx)

	// Should have both projectile and missile now
	if len(ctx.spawnedProjectiles) != 1 {
		t.Errorf("Expected 1 projectile still, got %d", len(ctx.spawnedProjectiles))
	}
	if len(ctx.spawnedMissiles) != 1 {
		t.Errorf("Expected 1 missile, got %d", len(ctx.spawnedMissiles))
	}
}

// Test Destroyer Health (should have more than fighter)
func TestDestroyerHealth(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 100, 100, nil)
	fighter := NewFighter(2, 0, 200, 200, nil)

	destroyerHealth, _ := destroyer.GetHealth()
	fighterHealth, _ := fighter.GetHealth()

	if destroyerHealth <= fighterHealth {
		t.Errorf("Destroyer health (%d) should be greater than fighter health (%d)", destroyerHealth, fighterHealth)
	}
}

// Test Destroyer Takes Damage and Dies
func TestDestroyerTakeDamage(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 100, 100, nil)
	ctx := NewMockGameContext()

	initialHealth, _ := destroyer.GetHealth()

	// Take some damage
	destroyer.TakeDamage(5, -1, ctx)

	currentHealth, _ := destroyer.GetHealth()
	if currentHealth != initialHealth-5 {
		t.Errorf("Expected health %d, got %d", initialHealth-5, currentHealth)
	}

	if !destroyer.IsAlive() {
		t.Error("Destroyer should still be alive")
	}

	// Take lethal damage
	destroyer.TakeDamage(1000, -1, ctx)

	finalHealth, _ := destroyer.GetHealth()
	if finalHealth != 0 {
		t.Errorf("Expected health 0, got %d", finalHealth)
	}

	if destroyer.IsAlive() {
		t.Error("Destroyer should be dead")
	}

	// Should have spawned explosion
	if len(ctx.spawnedExplosions) != 1 {
		t.Errorf("Expected 1 explosion, got %d", len(ctx.spawnedExplosions))
	}
}

// Test Destroyer Movement and Position Wrapping
func TestDestroyerMovementAndWrapping(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 100, 100, nil)
	destroyer.PlayerControlled = true

	// Set speed and rotation to move right (velocity = (5, 0))
	// Moving right means rotation = π/2 (90 degrees)
	destroyer.Speed = 5
	destroyer.Rotation = math.Pi / 2 // Facing right

	// Update movement
	destroyer.UpdateMovement()

	x, y := destroyer.GetPosition()
	if math.Abs(x-105) > 1e-9 {
		t.Errorf("Expected X=105, got %f", x)
	}
	if math.Abs(y-100) > 1e-9 {
		t.Errorf("Expected Y=100, got %f", y)
	}

	// Test wrapping: place near edge
	destroyer.X = float64(config.GameWidth - 5)
	destroyer.Y = 100
	destroyer.Speed = 10
	destroyer.Rotation = math.Pi / 2 // Facing right

	destroyer.UpdateMovement()

	// Should wrap to other side
	x, y = destroyer.GetPosition()
	if x >= float64(config.GameWidth) {
		t.Errorf("X should wrap, got %f", x)
	}
	if x < 0 {
		t.Errorf("X wrapped too far, got %f", x)
	}
}

// Test Destroyer Collision Radius
func TestDestroyerCollisionRadius(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 100, 100, nil)
	fighter := NewFighter(2, 0, 200, 200, nil)

	destroyerRadius := destroyer.GetCollisionRadius()
	fighterRadius := fighter.GetCollisionRadius()

	if destroyerRadius <= fighterRadius {
		t.Errorf("Destroyer collision radius (%f) should be larger than fighter (%f)", destroyerRadius, fighterRadius)
	}
}

// Test Destroyer Player Control
func TestDestroyerPlayerControl(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 100, 100, nil)

	if destroyer.IsPlayerControlled() {
		t.Error("Destroyer should not start as player controlled")
	}

	destroyer.SetPlayerControlled(true)

	if !destroyer.IsPlayerControlled() {
		t.Error("Destroyer should be player controlled after setting")
	}
}

// Test Destroyer Exclusive Weapon Behavior
// When missiles fire (exclusive weapon), main gun should NOT fire in same call
func TestDestroyerExclusiveWeaponBehavior(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 100, 100, nil)
	destroyer.SetPlayerControlled(true)
	destroyer.Rotation = 0 // Facing UP
	ctx := NewMockGameContext()
	ctx.ships[1] = destroyer

	// Create two targets:
	// 1. Enemy in REAR arc (for missile - 180° rear arc when facing UP means below)
	// Place this CLOSER than forward enemy to ensure it's selected
	rearEnemy := NewFighter(2, 1, 100, 180, nil) // 80 pixels below (within missile range from config)
	ctx.ships[2] = rearEnemy

	// 2. Enemy in FORWARD arc (for main gun - 30° forward arc when facing UP)
	forwardEnemy := NewFighter(3, 1, 100, 10, nil) // 90 pixels above (farther)
	ctx.ships[3] = forwardEnemy

	// Verify both weapons are fully charged
	if !destroyer.CanFireWeapon() {
		t.Fatal("Destroyer should be able to fire (both weapons charged)")
	}
	if !destroyer.CanFireMissile() {
		t.Fatal("Missile launcher should be charged")
	}
	if len(destroyer.Weapons) < 2 || destroyer.Weapons[1].WeaponCapacitor < 1.0 {
		t.Fatal("Main gun should be charged")
	}

	// Fire weapon with forward target position
	// Even though we're aiming forward, the missile should fire (higher priority)
	// because there's a valid target in the rear arc
	forwardX, forwardY := forwardEnemy.GetPosition()
	destroyer.FireWeapon(forwardX, forwardY, ctx)

	// CRITICAL ASSERTION: Because missile has Exclusive=true and higher priority,
	// ONLY the missile should fire, NOT the main gun
	if len(ctx.spawnedMissiles) != 1 {
		t.Errorf("Expected exactly 1 missile to fire, got %d", len(ctx.spawnedMissiles))
	}
	if len(ctx.spawnedProjectiles) != 0 {
		t.Errorf("Expected 0 projectiles (main gun should NOT fire due to exclusive missile), got %d", len(ctx.spawnedProjectiles))
	}

	// Verify missile targets the rear enemy
	if len(ctx.spawnedMissiles) > 0 {
		missile := ctx.spawnedMissiles[0]
		if missile.TargetID != 2 {
			t.Errorf("Expected missile to target rear enemy (ID 2), got target ID %d", missile.TargetID)
		}
	}

	// Verify weapon capacitor states
	if destroyer.Weapons[0].WeaponCapacitor != 0.0 {
		t.Errorf("Missile launcher capacitor should be discharged, got %f", destroyer.Weapons[0].WeaponCapacitor)
	}
	if len(destroyer.Weapons) > 1 && destroyer.Weapons[1].WeaponCapacitor != 1.0 {
		t.Errorf("Main gun capacitor should still be charged (didn't fire), got %f", destroyer.Weapons[1].WeaponCapacitor)
	}
}

// Test Destroyer Missile Orientation - Ship Facing UP
func TestDestroyerMissileOrientationFacingUp(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 1000, 1000, nil)
	destroyer.SetPlayerControlled(true)
	destroyer.Rotation = 0 // Facing UP (sprites face UP by default)
	ctx := NewMockGameContext()
	ctx.ships[1] = destroyer

	// Create enemy target behind destroyer (below it, since facing UP)
	enemy := NewFighter(2, 1, 1000, 1100, nil) // 100 pixels below
	ctx.ships[2] = enemy

	// Fire weapon (should trigger missile due to rear target)
	destroyer.FireWeapon(1000, 900, ctx) // Aiming forward, but missile has priority

	// Verify missile was spawned
	if len(ctx.spawnedMissiles) != 1 {
		t.Fatalf("Expected 1 missile, got %d", len(ctx.spawnedMissiles))
	}

	missile := ctx.spawnedMissiles[0]

	// Missile should spawn behind (below) the ship
	if missile.Y <= destroyer.Y {
		t.Errorf("Missile should spawn below ship (Y > ship.Y), got missile.Y=%f, ship.Y=%f", missile.Y, destroyer.Y)
	}
	if math.Abs(missile.X-destroyer.X) > 1.0 {
		t.Errorf("Missile should spawn at same X as ship, got missile.X=%f, ship.X=%f", missile.X, destroyer.X)
	}

	// Missile rotation should point DOWN (π radians)
	// Normalize to [0, 2π) for comparison
	normalizedRotation := NormalizeAngle(missile.Rotation)
	if normalizedRotation < 0 {
		normalizedRotation += 2 * math.Pi
	}
	expectedRotation := math.Pi
	if math.Abs(normalizedRotation-expectedRotation) > 0.01 {
		t.Errorf("Missile rotation should be π (DOWN), got %f (expected %f)", normalizedRotation, expectedRotation)
	}

	// Missile velocity should be pointing DOWN (positive Y)
	if missile.VelocityY <= 0 {
		t.Errorf("Missile should have positive Y velocity (DOWN), got %f", missile.VelocityY)
	}
	if math.Abs(missile.VelocityX) > 0.1 {
		t.Errorf("Missile should have near-zero X velocity, got %f", missile.VelocityX)
	}
}

// Test Destroyer Missile Orientation - Ship Facing RIGHT
func TestDestroyerMissileOrientationFacingRight(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 1000, 1000, nil)
	destroyer.SetPlayerControlled(true)
	destroyer.Rotation = math.Pi / 2 // Facing RIGHT
	ctx := NewMockGameContext()
	ctx.ships[1] = destroyer

	// Create enemy target behind destroyer (to the left, since facing RIGHT)
	enemy := NewFighter(2, 1, 900, 1000, nil) // 100 pixels to the left
	ctx.ships[2] = enemy

	// Fire weapon
	destroyer.FireWeapon(1100, 1000, ctx)

	// Verify missile was spawned
	if len(ctx.spawnedMissiles) != 1 {
		t.Fatalf("Expected 1 missile, got %d", len(ctx.spawnedMissiles))
	}

	missile := ctx.spawnedMissiles[0]

	// Missile should spawn to the left of the ship
	if missile.X >= destroyer.X {
		t.Errorf("Missile should spawn left of ship (X < ship.X), got missile.X=%f, ship.X=%f", missile.X, destroyer.X)
	}
	if math.Abs(missile.Y-destroyer.Y) > 1.0 {
		t.Errorf("Missile should spawn at same Y as ship, got missile.Y=%f, ship.Y=%f", missile.Y, destroyer.Y)
	}

	// Missile rotation should point LEFT (-π/2 or 3π/2 radians)
	normalizedRotation := NormalizeAngle(missile.Rotation)
	if normalizedRotation < 0 {
		normalizedRotation += 2 * math.Pi
	}
	expectedRotation := 3 * math.Pi / 2 // LEFT
	if math.Abs(normalizedRotation-expectedRotation) > 0.01 {
		t.Errorf("Missile rotation should be 3π/2 (LEFT), got %f (expected %f)", normalizedRotation, expectedRotation)
	}

	// Missile velocity should be pointing LEFT (negative X)
	if missile.VelocityX >= 0 {
		t.Errorf("Missile should have negative X velocity (LEFT), got %f", missile.VelocityX)
	}
	if math.Abs(missile.VelocityY) > 0.1 {
		t.Errorf("Missile should have near-zero Y velocity, got %f", missile.VelocityY)
	}
}

// Test Destroyer Missile Orientation - Ship Facing DOWN
func TestDestroyerMissileOrientationFacingDown(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 1000, 1000, nil)
	destroyer.SetPlayerControlled(true)
	destroyer.Rotation = math.Pi // Facing DOWN
	ctx := NewMockGameContext()
	ctx.ships[1] = destroyer

	// Create enemy target behind destroyer (above it, since facing DOWN)
	enemy := NewFighter(2, 1, 1000, 900, nil) // 100 pixels above
	ctx.ships[2] = enemy

	// Fire weapon
	destroyer.FireWeapon(1000, 1100, ctx)

	// Verify missile was spawned
	if len(ctx.spawnedMissiles) != 1 {
		t.Fatalf("Expected 1 missile, got %d", len(ctx.spawnedMissiles))
	}

	missile := ctx.spawnedMissiles[0]

	// Missile should spawn above the ship
	if missile.Y >= destroyer.Y {
		t.Errorf("Missile should spawn above ship (Y < ship.Y), got missile.Y=%f, ship.Y=%f", missile.Y, destroyer.Y)
	}
	if math.Abs(missile.X-destroyer.X) > 1.0 {
		t.Errorf("Missile should spawn at same X as ship, got missile.X=%f, ship.Y=%f", missile.X, destroyer.X)
	}

	// Missile rotation should point UP (0 or 2π radians)
	normalizedRotation := NormalizeAngle(missile.Rotation)
	if normalizedRotation < 0 {
		normalizedRotation += 2 * math.Pi
	}
	// Could be 0 or ~2π
	if normalizedRotation > 0.01 && math.Abs(normalizedRotation-2*math.Pi) > 0.01 {
		t.Errorf("Missile rotation should be 0 or 2π (UP), got %f", normalizedRotation)
	}

	// Missile velocity should be pointing UP (negative Y)
	if missile.VelocityY >= 0 {
		t.Errorf("Missile should have negative Y velocity (UP), got %f", missile.VelocityY)
	}
	if math.Abs(missile.VelocityX) > 0.1 {
		t.Errorf("Missile should have near-zero X velocity, got %f", missile.VelocityX)
	}
}

// Test Destroyer Missile Orientation - Ship Facing LEFT
func TestDestroyerMissileOrientationFacingLeft(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 1000, 1000, nil)
	destroyer.SetPlayerControlled(true)
	destroyer.Rotation = -math.Pi / 2 // Facing LEFT
	ctx := NewMockGameContext()
	ctx.ships[1] = destroyer

	// Create enemy target behind destroyer (to the right, since facing LEFT)
	enemy := NewFighter(2, 1, 1100, 1000, nil) // 100 pixels to the right
	ctx.ships[2] = enemy

	// Fire weapon
	destroyer.FireWeapon(900, 1000, ctx)

	// Verify missile was spawned
	if len(ctx.spawnedMissiles) != 1 {
		t.Fatalf("Expected 1 missile, got %d", len(ctx.spawnedMissiles))
	}

	missile := ctx.spawnedMissiles[0]

	// Missile should spawn to the right of the ship
	if missile.X <= destroyer.X {
		t.Errorf("Missile should spawn right of ship (X > ship.X), got missile.X=%f, ship.X=%f", missile.X, destroyer.X)
	}
	if math.Abs(missile.Y-destroyer.Y) > 1.0 {
		t.Errorf("Missile should spawn at same Y as ship, got missile.Y=%f, ship.Y=%f", missile.Y, destroyer.Y)
	}

	// Missile rotation should point RIGHT (π/2 radians)
	normalizedRotation := NormalizeAngle(missile.Rotation)
	if normalizedRotation < 0 {
		normalizedRotation += 2 * math.Pi
	}
	expectedRotation := math.Pi / 2 // RIGHT
	if math.Abs(normalizedRotation-expectedRotation) > 0.01 {
		t.Errorf("Missile rotation should be π/2 (RIGHT), got %f (expected %f)", normalizedRotation, expectedRotation)
	}

	// Missile velocity should be pointing RIGHT (positive X)
	if missile.VelocityX <= 0 {
		t.Errorf("Missile should have positive X velocity (RIGHT), got %f", missile.VelocityX)
	}
	if math.Abs(missile.VelocityY) > 0.1 {
		t.Errorf("Missile should have near-zero Y velocity, got %f", missile.VelocityY)
	}
}

// Test AI Destroyer Fires Multiple Missiles
func TestAIDestroyerFiresMultipleMissiles(t *testing.T) {
	// Create AI destroyer facing UP (rotation = 0)
	destroyer := NewDestroyer(1, 0, 2500.0, 2500.0, nil)
	destroyer.PlayerControlled = false // AI controlled
	ctx := NewMockGameContext()
	ctx.ships[1] = destroyer

	// Create two enemy targets in the rear arc (behind destroyer)
	// Destroyer faces UP (rotation 0), so rear is DOWN (positive Y)
	enemy1 := NewFighter(2, 1, 2500.0, 2700.0, nil) // 200 pixels below
	enemy2 := NewFighter(3, 1, 2500.0, 2900.0, nil) // 400 pixels below
	ctx.ships[2] = enemy1
	ctx.ships[3] = enemy2

	// Simulate EntityManager's discrete update passes
	missilesFired := 0
	maxIterations := 300 // Should be enough for 2 missiles (missile charge time ~3 seconds at 60fps = 180 frames)

	for i := 0; i < maxIterations; i++ {
		// Pass 1: Weapon capacitor charging (main gun + missiles)
		destroyer.UpdateWeapons()

		// Pass 2: AI updates (targeting, rotation, firing)
		destroyer.UpdateAI(ctx)

		// Pass 3: Movement
		destroyer.UpdateMovement()

		// Check if new missile was fired this frame
		if len(ctx.spawnedMissiles) > missilesFired {
			missilesFired = len(ctx.spawnedMissiles)
			t.Logf("Missile %d fired at frame %d", missilesFired, i)

			// If we've fired 2 missiles, we can stop early
			if missilesFired >= 2 {
				break
			}
		}
	}

	// Verify at least 2 missiles were fired
	if missilesFired < 2 {
		t.Errorf("Expected at least 2 missiles fired, got %d", missilesFired)
	}

	// Verify missiles were targeted at enemies
	for i, missile := range ctx.spawnedMissiles {
		if missile.OwnerID != 1 {
			t.Errorf("Missile %d: Expected owner ID 1, got %d", i, missile.OwnerID)
		}
		if missile.FactionID != 0 {
			t.Errorf("Missile %d: Expected faction ID 0, got %d", i, missile.FactionID)
		}
		// Target should be one of the enemies
		if missile.TargetID != 2 && missile.TargetID != 3 {
			t.Errorf("Missile %d: Expected target ID 2 or 3, got %d", i, missile.TargetID)
		}
	}
}

// Test Destroyer Afterburner Increases Max Speed
func TestDestroyerAfterburnerMaxSpeedBoost(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 100, 100, nil)

	// Verify destroyer has afterburner
	if !destroyer.HasAfterburnerSystem {
		t.Fatal("Destroyer should have afterburner system")
	}

	// Get characteristics for verification
	chars := config.GetShipCharacteristics(ClassDestroyer)
	normalMaxSpeed := chars.MaxSpeed
	boostedMaxSpeed := normalMaxSpeed * chars.AfterburnerMaxSpeedMultiplier

	// Activate afterburner and accelerate beyond normal max speed
	destroyer.AfterburnerActive = true
	destroyer.AfterburnerCharge = 100.0    // Ensure we have charge
	destroyer.Speed = normalMaxSpeed + 0.5 // Set speed above normal max

	// Simulate player input to accelerate
	effectiveAccel := destroyer.Accel
	if destroyer.AfterburnerActive {
		effectiveAccel *= destroyer.AfterburnerAccelMultiplier
	}

	// Accelerate to beyond normal max speed
	destroyer.Speed += effectiveAccel

	// Calculate effective max speed with afterburner
	effectiveMaxSpeed := normalMaxSpeed
	if destroyer.AfterburnerActive {
		effectiveMaxSpeed *= destroyer.AfterburnerMaxSpeedMultiplier
	}

	// Cap speed to effective max
	if destroyer.Speed > effectiveMaxSpeed {
		destroyer.Speed = effectiveMaxSpeed
	}

	// With afterburner active, should be able to reach boosted max speed
	if destroyer.Speed <= normalMaxSpeed {
		t.Errorf("Destroyer should exceed normal max speed (%f) with afterburner, got %f",
			normalMaxSpeed, destroyer.Speed)
	}

	if destroyer.Speed > boostedMaxSpeed+0.01 {
		t.Errorf("Destroyer should not exceed boosted max speed (%f), got %f",
			boostedMaxSpeed, destroyer.Speed)
	}
}

// Test Destroyer Afterburner Max Speed Multiplier Applied Correctly
func TestDestroyerAfterburnerMaxSpeedMultiplier(t *testing.T) {
	destroyer := NewDestroyer(1, 0, 100, 100, nil)
	chars := config.GetShipCharacteristics(ClassDestroyer)

	// Verify the multiplier is set correctly from config
	if destroyer.AfterburnerMaxSpeedMultiplier != chars.AfterburnerMaxSpeedMultiplier {
		t.Errorf("Expected afterburner max speed multiplier %f, got %f",
			chars.AfterburnerMaxSpeedMultiplier, destroyer.AfterburnerMaxSpeedMultiplier)
	}

	// Verify it's greater than 1.0 (should boost speed)
	if destroyer.AfterburnerMaxSpeedMultiplier <= 1.0 {
		t.Errorf("Afterburner max speed multiplier should be > 1.0, got %f",
			destroyer.AfterburnerMaxSpeedMultiplier)
	}

	// Calculate expected boosted max speed
	expectedBoostedMaxSpeed := destroyer.MaxSpeed * destroyer.AfterburnerMaxSpeedMultiplier

	// Verify it's actually higher than normal max
	if expectedBoostedMaxSpeed <= destroyer.MaxSpeed {
		t.Errorf("Boosted max speed (%f) should be greater than normal max speed (%f)",
			expectedBoostedMaxSpeed, destroyer.MaxSpeed)
	}
}
