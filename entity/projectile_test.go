package entity

import (
	"math"
	"testing"

	"github.com/nathan/verdant-thane/config"
)

// Test MainGunProjectile Creation
func TestNewMainGunProjectile(t *testing.T) {
	cfg := MainGunConfig{
		X:         100,
		Y:         200,
		VelocityX: 5,
		VelocityY: 3,
		OwnerID:   1,
		FactionID: 0,
		Sprite:    nil,
	}

	proj := NewMainGunProjectile(42, cfg)

	// Verify identity
	if proj.GetID() != 42 {
		t.Errorf("Expected ID 42, got %d", proj.GetID())
	}
	if proj.GetFaction() != 0 {
		t.Errorf("Expected faction 0, got %d", proj.GetFaction())
	}
	if proj.GetOwnerID() != 1 {
		t.Errorf("Expected owner ID 1, got %d", proj.GetOwnerID())
	}

	// Verify position
	x, y := proj.GetPosition()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%f, %f)", x, y)
	}

	// Verify velocity
	if proj.VelocityX != 5 || proj.VelocityY != 3 {
		t.Errorf("Expected velocity (5, 3), got (%f, %f)", proj.VelocityX, proj.VelocityY)
	}

	// Verify starts alive
	if !proj.IsAlive() {
		t.Error("Projectile should start alive")
	}

	// Verify lifetime from config
	chars := config.GetProjectileCharacteristics(config.LaserProjectile)
	if proj.Lifetime != chars.Lifetime {
		t.Errorf("Expected lifetime %d, got %d", chars.Lifetime, proj.Lifetime)
	}

	// Verify damage
	if proj.GetDamage() != 1 {
		t.Errorf("Expected damage 1, got %d", proj.GetDamage())
	}
}

// Test MainGunProjectile Movement
func TestMainGunProjectileMovement(t *testing.T) {
	ctx := NewMockGameContext()
	cfg := MainGunConfig{
		X:         100,
		Y:         100,
		VelocityX: 10,
		VelocityY: 5,
		OwnerID:   1,
		FactionID: 0,
	}

	proj := NewMainGunProjectile(1, cfg)
	proj.Update(ctx)

	// Should have moved
	x, y := proj.GetPosition()
	if x != 110 || y != 105 {
		t.Errorf("Expected position (110, 105) after movement, got (%f, %f)", x, y)
	}

	// Lifetime should decrease
	chars := config.GetProjectileCharacteristics(config.LaserProjectile)
	if proj.Lifetime != chars.Lifetime-1 {
		t.Errorf("Expected lifetime %d, got %d", chars.Lifetime-1, proj.Lifetime)
	}
}

// Test MainGunProjectile Lifetime Expiration
func TestMainGunProjectileLifetimeExpiration(t *testing.T) {
	ctx := NewMockGameContext()
	cfg := MainGunConfig{
		X:         100,
		Y:         100,
		VelocityX: 1,
		VelocityY: 0,
		OwnerID:   1,
		FactionID: 0,
	}

	proj := NewMainGunProjectile(1, cfg)
	proj.Lifetime = 2 // Set to almost expired

	// Update once - should still be alive
	proj.Update(ctx)
	if !proj.IsAlive() {
		t.Error("Projectile should still be alive after first update")
	}
	if proj.Lifetime != 1 {
		t.Errorf("Expected lifetime 1, got %d", proj.Lifetime)
	}

	// Update again - should expire
	proj.Update(ctx)
	if proj.IsAlive() {
		t.Error("Projectile should be dead after lifetime expires")
	}
	if proj.Lifetime != 0 {
		t.Errorf("Expected lifetime 0, got %d", proj.Lifetime)
	}

	// Further updates should do nothing
	initialX, initialY := proj.GetPosition()
	proj.Update(ctx)
	x, y := proj.GetPosition()
	if x != initialX || y != initialY {
		t.Error("Dead projectile should not move")
	}
}

// Test MainGunProjectile World Wrapping
func TestMainGunProjectileWrapping(t *testing.T) {
	ctx := NewMockGameContext()

	tests := []struct {
		name                 string
		startX, startY       float64
		velocityX, velocityY float64
		expectedX, expectedY float64
	}{
		{"Wrap right edge", float64(config.GameWidth) - 5, 100, 10, 0, 5, 100},
		{"Wrap left edge", 5, 100, -10, 0, float64(config.GameWidth) - 5, 100},
		{"Wrap bottom edge", 100, float64(config.GameHeight) - 5, 0, 10, 100, 5},
		{"Wrap top edge", 100, 5, 0, -10, 100, float64(config.GameHeight) - 5},
		{"No wrapping", 100, 100, 5, 5, 105, 105},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := MainGunConfig{
				X:         tt.startX,
				Y:         tt.startY,
				VelocityX: tt.velocityX,
				VelocityY: tt.velocityY,
				OwnerID:   1,
				FactionID: 0,
			}

			proj := NewMainGunProjectile(1, cfg)
			proj.Update(ctx)

			x, y := proj.GetPosition()
			if math.Abs(x-tt.expectedX) > 1e-9 || math.Abs(y-tt.expectedY) > 1e-9 {
				t.Errorf("Expected position (%f, %f), got (%f, %f)", tt.expectedX, tt.expectedY, x, y)
			}
		})
	}
}

// Test MainGunProjectile Collision Detection
func TestMainGunProjectileCollisionDetection(t *testing.T) {
	// Create enemy ship
	enemy := NewFighter(2, 1, 200, 200, nil) // Faction 1, at (200, 200)

	// Create friendly ship
	friendly := NewFighter(3, 0, 250, 200, nil) // Faction 0, at (250, 200)

	tests := []struct {
		name            string
		projX, projY    float64
		projFaction     int
		target          *Ship
		expectedCollide bool
		reason          string
	}{
		{"Direct hit on enemy", 200, 200, 0, enemy, true, "should collide with enemy at same position"},
		{"Near hit on enemy", 205, 200, 0, enemy, true, "should collide within collision radius"},
		{"Far from enemy", 300, 200, 0, enemy, false, "should not collide when far away"},
		{"Friendly fire", 250, 200, 0, friendly, false, "should not collide with friendly ship"},
		{"Enemy projectile on friendly", 250, 200, 1, friendly, true, "enemy projectile should hit friendly ship"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := MainGunConfig{
				X:         tt.projX,
				Y:         tt.projY,
				VelocityX: 1,
				VelocityY: 0,
				OwnerID:   1,
				FactionID: tt.projFaction,
			}

			proj := NewMainGunProjectile(1, cfg)
			collided := proj.CheckCollision(tt.target)

			if collided != tt.expectedCollide {
				t.Errorf("%s: expected collision=%v, got %v", tt.reason, tt.expectedCollide, collided)
			}
		})
	}
}

// Test Projectile Does Not Collide With Dead Ship
func TestProjectileNoCollisionWithDeadShip(t *testing.T) {
	ship := NewFighter(2, 1, 200, 200, nil)
	ship.Alive = false // Kill the ship

	cfg := MainGunConfig{
		X:         200,
		Y:         200,
		VelocityX: 1,
		VelocityY: 0,
		OwnerID:   1,
		FactionID: 0,
	}

	proj := NewMainGunProjectile(1, cfg)
	if proj.CheckCollision(ship) {
		t.Error("Should not collide with dead ship")
	}
}

// Test Dead Projectile Does Not Collide
func TestDeadProjectileNoCollision(t *testing.T) {
	ship := NewFighter(2, 1, 200, 200, nil)

	cfg := MainGunConfig{
		X:         200,
		Y:         200,
		VelocityX: 1,
		VelocityY: 0,
		OwnerID:   1,
		FactionID: 0,
	}

	proj := NewMainGunProjectile(1, cfg)
	proj.Alive = false

	if proj.CheckCollision(ship) {
		t.Error("Dead projectile should not collide")
	}
}

// Test Collision Detection Accounts for World Wrapping
func TestProjectileCollisionWrapping(t *testing.T) {
	// Ship near right edge
	ship := NewFighter(2, 1, float64(config.GameWidth)-5, 100, nil)

	// Projectile near left edge (wraps to be close to ship)
	// Distance across wrap = 10 pixels, within collision radius (14 + 2 = 16)
	cfg := MainGunConfig{
		X:         5,
		Y:         100,
		VelocityX: 1,
		VelocityY: 0,
		OwnerID:   1,
		FactionID: 0,
	}

	proj := NewMainGunProjectile(1, cfg)

	// Should detect collision across wrapping boundary
	// Distance across wrap = 10 pixels, within collision radius (14 + 2 = 16)
	if !proj.CheckCollision(ship) {
		shipX, shipY := ship.GetPosition()
		projX, projY := proj.GetPosition()
		actualDist := Distance(projX, projY, shipX, shipY)
		t.Errorf("Should collide across wrapping boundary (distance=%f, radius=%f)",
			actualDist, ship.GetCollisionRadius())
	}
}

// ============================================================================
// MissileProjectile Tests
// ============================================================================

// Test MissileProjectile Creation
func TestNewMissileProjectile(t *testing.T) {
	cfg := MissileConfig{
		X:         100,
		Y:         200,
		VelocityX: 5,
		VelocityY: 3,
		Rotation:  math.Pi / 4,
		OwnerID:   1,
		FactionID: 0,
		TargetID:  2,
		Sprite:    nil,
	}

	missile := NewMissileProjectile(42, cfg)

	// Verify identity
	if missile.GetID() != 42 {
		t.Errorf("Expected ID 42, got %d", missile.GetID())
	}
	if missile.GetFaction() != 0 {
		t.Errorf("Expected faction 0, got %d", missile.GetFaction())
	}
	if missile.GetOwnerID() != 1 {
		t.Errorf("Expected owner ID 1, got %d", missile.GetOwnerID())
	}
	if missile.TargetID != 2 {
		t.Errorf("Expected target ID 2, got %d", missile.TargetID)
	}

	// Verify position and rotation
	x, y := missile.GetPosition()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%f, %f)", x, y)
	}
	if missile.Rotation != math.Pi/4 {
		t.Errorf("Expected rotation %f, got %f", math.Pi/4, missile.Rotation)
	}

	// Verify starts alive
	if !missile.IsAlive() {
		t.Error("Missile should start alive")
	}

	// Verify damage
	if missile.GetDamage() != 1 {
		t.Errorf("Expected damage 1, got %d", missile.GetDamage())
	}
}

// Test Missile Tracking Valid Target
func TestMissileTrackingValidTarget(t *testing.T) {
	ctx := NewMockGameContext()

	// Create target ship at a different location
	target := NewFighter(2, 1, 100, 300, nil) // Target is below missile
	ctx.ships[target.GetID()] = target

	// Create missile heading right, but target is perpendicular (below)
	cfg := MissileConfig{
		X:         100,
		Y:         100,
		VelocityX: 4,
		VelocityY: 0,
		Rotation:  0, // Heading right (0 radians)
		OwnerID:   1,
		FactionID: 0,
		TargetID:  target.GetID(),
	}

	missile := NewMissileProjectile(1, cfg)

	// Update several times
	initialRotation := missile.Rotation
	for i := 0; i < 10; i++ {
		missile.Update(ctx)
	}

	// Rotation should have changed to track target
	// Target is at 90 degrees (π/2), so rotation should have increased
	if math.Abs(missile.Rotation-initialRotation) < 0.1 {
		t.Errorf("Missile rotation should change when tracking target, initial=%f, final=%f",
			initialRotation, missile.Rotation)
	}

	// Missile should be accelerating
	speed := math.Sqrt(missile.VelocityX*missile.VelocityX + missile.VelocityY*missile.VelocityY)
	if speed <= 4.0 {
		t.Errorf("Missile should accelerate, speed should be >4.0, got %f", speed)
	}
}

// Test Missile Continues Straight When Target Invalid
func TestMissileTrackingInvalidTarget(t *testing.T) {
	ctx := NewMockGameContext()

	// Create missile with no target ship in context
	cfg := MissileConfig{
		X:         100,
		Y:         100,
		VelocityX: 4,
		VelocityY: 0,
		Rotation:  0,
		OwnerID:   1,
		FactionID: 0,
		TargetID:  999, // Non-existent target
	}

	missile := NewMissileProjectile(1, cfg)

	initialVY := missile.VelocityY

	// Update once
	missile.Update(ctx)

	// Velocity should remain roughly the same (no tracking)
	// (May have slight acceleration but no direction change)
	if math.Abs(missile.VelocityY-initialVY) > 0.1 {
		t.Error("Missile should continue straight when target is invalid")
	}
}

// Test Missile Tracking When Target Dies
func TestMissileTrackingTargetDies(t *testing.T) {
	ctx := NewMockGameContext()

	// Create target ship
	target := NewFighter(2, 1, 300, 100, nil)
	ctx.ships[target.GetID()] = target

	// Create missile tracking the target
	cfg := MissileConfig{
		X:         100,
		Y:         100,
		VelocityX: 4,
		VelocityY: 0,
		Rotation:  0,
		OwnerID:   1,
		FactionID: 0,
		TargetID:  target.GetID(),
	}

	missile := NewMissileProjectile(1, cfg)

	// Update a few times while target is alive
	for i := 0; i < 5; i++ {
		missile.Update(ctx)
	}

	// Kill the target
	target.Alive = false

	// Record current velocity
	vxBeforeDeath := missile.VelocityX
	vyBeforeDeath := missile.VelocityY

	// Update again - should continue on last trajectory
	missile.Update(ctx)

	// Velocity direction should not change significantly (no more tracking)
	angleBefore := math.Atan2(vyBeforeDeath, vxBeforeDeath)
	angleAfter := math.Atan2(missile.VelocityY, missile.VelocityX)

	if math.Abs(NormalizeAngle(angleAfter-angleBefore)) > 0.1 {
		t.Error("Missile should continue on trajectory after target dies")
	}
}

// Test Missile Lifetime Expiration
func TestMissileLifetimeExpiration(t *testing.T) {
	ctx := NewMockGameContext()
	cfg := MissileConfig{
		X:         100,
		Y:         100,
		VelocityX: 1,
		VelocityY: 0,
		Rotation:  0,
		OwnerID:   1,
		FactionID: 0,
		TargetID:  -1,
	}

	missile := NewMissileProjectile(1, cfg)
	missile.Lifetime = 2

	// Update once
	missile.Update(ctx)
	if !missile.IsAlive() {
		t.Error("Missile should still be alive")
	}

	// Update again - should expire
	missile.Update(ctx)
	if missile.IsAlive() {
		t.Error("Missile should be dead after lifetime expires")
	}
}

// Test Missile Collision Detection
func TestMissileCollisionDetection(t *testing.T) {
	// Create enemy ship
	enemy := NewFighter(2, 1, 200, 200, nil)

	// Create missile near enemy
	cfg := MissileConfig{
		X:         200,
		Y:         200,
		VelocityX: 1,
		VelocityY: 0,
		Rotation:  0,
		OwnerID:   1,
		FactionID: 0,
		TargetID:  enemy.GetID(),
	}

	missile := NewMissileProjectile(1, cfg)

	if !missile.CheckCollision(enemy) {
		t.Error("Missile should collide with enemy at same position")
	}
}

// Test Missile Does Not Collide With Friendly
func TestMissileNoFriendlyFire(t *testing.T) {
	// Create friendly ship
	friendly := NewFighter(2, 0, 200, 200, nil)

	// Create missile from same faction
	cfg := MissileConfig{
		X:         200,
		Y:         200,
		VelocityX: 1,
		VelocityY: 0,
		Rotation:  0,
		OwnerID:   1,
		FactionID: 0, // Same faction as friendly
		TargetID:  -1,
	}

	missile := NewMissileProjectile(1, cfg)

	if missile.CheckCollision(friendly) {
		t.Error("Missile should not collide with friendly ship")
	}
}

// Test Missile Turn Rate Limiting
func TestMissileTurnRateLimiting(t *testing.T) {
	ctx := NewMockGameContext()

	// Create target directly behind missile
	target := NewFighter(2, 1, 0, 100, nil)
	ctx.ships[target.GetID()] = target

	// Create missile heading right (0 radians), target is behind (π radians)
	cfg := MissileConfig{
		X:         100,
		Y:         100,
		VelocityX: 4,
		VelocityY: 0,
		Rotation:  0,
		OwnerID:   1,
		FactionID: 0,
		TargetID:  target.GetID(),
	}

	missile := NewMissileProjectile(1, cfg)

	initialAngle := math.Atan2(missile.VelocityY, missile.VelocityX)

	// Update once
	missile.Update(ctx)

	newAngle := math.Atan2(missile.VelocityY, missile.VelocityX)
	turnAmount := math.Abs(NormalizeAngle(newAngle - initialAngle))

	// Turn should be limited by turn rate
	chars := config.GetProjectileCharacteristics(config.MissileProjectile)
	if turnAmount > chars.TurnRate+0.01 {
		t.Errorf("Turn amount %f exceeds max turn rate %f", turnAmount, chars.TurnRate)
	}

	// Should have turned in correct direction (CCW since target is at 180°)
	if turnAmount < 0.001 {
		t.Error("Missile should have turned toward target")
	}
}

// Test Missile World Wrapping
func TestMissileWrapping(t *testing.T) {
	ctx := NewMockGameContext()

	cfg := MissileConfig{
		X:         float64(config.GameWidth) - 5,
		Y:         100,
		VelocityX: 10,
		VelocityY: 0,
		Rotation:  0,
		OwnerID:   1,
		FactionID: 0,
		TargetID:  -1,
	}

	missile := NewMissileProjectile(1, cfg)
	missile.Update(ctx)

	x, y := missile.GetPosition()
	expectedX := float64(5)
	expectedY := float64(100)

	if math.Abs(x-expectedX) > 1e-9 || math.Abs(y-expectedY) > 1e-9 {
		t.Errorf("Expected position (%f, %f) after wrapping, got (%f, %f)", expectedX, expectedY, x, y)
	}
}

// Test Missile Tracks Across World Wrapping
func TestMissileTracksAcrossWrapping(t *testing.T) {
	ctx := NewMockGameContext()

	// Target near left edge
	target := NewFighter(2, 1, 10, 100, nil)
	ctx.ships[target.GetID()] = target

	// Missile near right edge, target is closer across wrap
	cfg := MissileConfig{
		X:         float64(config.GameWidth) - 10,
		Y:         100,
		VelocityX: 1,
		VelocityY: 0,
		Rotation:  0,
		OwnerID:   1,
		FactionID: 0,
		TargetID:  target.GetID(),
	}

	missile := NewMissileProjectile(1, cfg)

	// Update several times
	for i := 0; i < 5; i++ {
		missile.Update(ctx)
	}

	// Missile should turn to track across wrap (wrapped distance is ~20 pixels)
	// Should not turn around to go the long way
	// Verify missile is moving in positive X direction (toward right edge to wrap)
	if missile.VelocityX < 0 {
		t.Error("Missile should track across wrapping boundary, not turn around")
	}
}
