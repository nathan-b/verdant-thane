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

	// Can't fire again immediately
	if destroyer.CanFireWeapon() {
		t.Error("Should not be able to fire immediately after shooting")
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
		destroyer.UpdateMissileWeapon()

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
