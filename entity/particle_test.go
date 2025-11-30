package entity

import (
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// Helper function for floating point comparisons
func abs(x float64) float64 {
	return math.Abs(x)
}

// Test Afterburner Particle Creation
func TestNewAfterburnerParticle(t *testing.T) {
	p := NewAfterburnerParticle(42, 100, 200, 5.0, 3.0)

	if p == nil {
		t.Fatal("NewAfterburnerParticle should not return nil")
	}

	if p.ID != 42 {
		t.Errorf("Expected ID 42, got %d", p.ID)
	}

	if p.X != 100 || p.Y != 200 {
		t.Errorf("Expected position (100, 200), got (%f, %f)", p.X, p.Y)
	}

	// Particle velocity should be 30% of ship velocity (trailing effect)
	expectedVX := 5.0 * 0.3
	expectedVY := 3.0 * 0.3
	epsilon := 0.0001
	if abs(p.VX-expectedVX) > epsilon || abs(p.VY-expectedVY) > epsilon {
		t.Errorf("Expected velocity (%f, %f), got (%f, %f)",
			expectedVX, expectedVY, p.VX, p.VY)
	}

	if p.Lifetime != 15 {
		t.Errorf("Expected lifetime 15, got %d", p.Lifetime)
	}

	if p.MaxLife != 15 {
		t.Errorf("Expected max life 15, got %d", p.MaxLife)
	}

	if !p.Alive {
		t.Error("New particle should be alive")
	}

	// Verify orange color
	expectedColor := color.RGBA{255, 165, 0, 255}
	if p.Color != expectedColor {
		t.Errorf("Expected color %+v, got %+v", expectedColor, p.Color)
	}

	if p.Size != 2.0 {
		t.Errorf("Expected size 2.0, got %f", p.Size)
	}
}

// Test Particle GetID
func TestParticleGetID(t *testing.T) {
	p := NewAfterburnerParticle(123, 0, 0, 0, 0)

	if p.GetID() != 123 {
		t.Errorf("Expected ID 123, got %d", p.GetID())
	}
}

// Test Particle GetPosition
func TestParticleGetPosition(t *testing.T) {
	p := NewAfterburnerParticle(1, 456.7, 789.1, 0, 0)

	x, y := p.GetPosition()
	if x != 456.7 || y != 789.1 {
		t.Errorf("Expected position (456.7, 789.1), got (%f, %f)", x, y)
	}
}

// Test Particle IsAlive
func TestParticleIsAlive(t *testing.T) {
	p := NewAfterburnerParticle(1, 0, 0, 0, 0)

	if !p.IsAlive() {
		t.Error("New particle should be alive")
	}

	// Manually mark as dead
	p.Alive = false

	if p.IsAlive() {
		t.Error("Particle marked as dead should return false for IsAlive")
	}
}

// Test Particle Update Movement
func TestParticleUpdateMovement(t *testing.T) {
	p := NewAfterburnerParticle(1, 100, 200, 5.0, 3.0)

	// Create a mock GameContext (nil is fine for this test since particle doesn't use it)
	mockCtx := &mockGameContext{}

	err := p.Update(mockCtx)
	if err != nil {
		t.Fatalf("Update should not return error: %v", err)
	}

	// Particle should move by its velocity
	expectedX := 100 + (5.0 * 0.3)
	expectedY := 200 + (3.0 * 0.3)

	x, y := p.GetPosition()
	if x != expectedX || y != expectedY {
		t.Errorf("Expected position (%f, %f) after update, got (%f, %f)",
			expectedX, expectedY, x, y)
	}
}

// Test Particle Lifetime Decrement
func TestParticleLifetimeDecrement(t *testing.T) {
	p := NewAfterburnerParticle(1, 0, 0, 0, 0)
	mockCtx := &mockGameContext{}

	initialLifetime := p.Lifetime

	err := p.Update(mockCtx)
	if err != nil {
		t.Fatalf("Update should not return error: %v", err)
	}

	if p.Lifetime != initialLifetime-1 {
		t.Errorf("Expected lifetime %d after update, got %d", initialLifetime-1, p.Lifetime)
	}
}

// Test Particle Dies After Lifetime Expires
func TestParticleExpiresAfterLifetime(t *testing.T) {
	p := NewAfterburnerParticle(1, 0, 0, 0, 0)
	mockCtx := &mockGameContext{}

	// Run updates for lifetime + 1 ticks
	lifetime := p.Lifetime
	for i := 0; i < lifetime+1; i++ {
		p.Update(mockCtx)
	}

	if p.IsAlive() {
		t.Error("Particle should be dead after lifetime expires")
	}

	if p.Lifetime > 0 {
		t.Errorf("Expected lifetime <= 0 after expiration, got %d", p.Lifetime)
	}
}

// Test Particle Dies Exactly When Lifetime Reaches Zero
func TestParticleExpiresAtZero(t *testing.T) {
	p := NewAfterburnerParticle(1, 0, 0, 0, 0)
	mockCtx := &mockGameContext{}

	// Run updates until lifetime reaches 1
	for p.Lifetime > 1 {
		p.Update(mockCtx)
	}

	// Should still be alive with lifetime=1
	if !p.IsAlive() {
		t.Error("Particle should still be alive with lifetime=1")
	}

	// One more update should kill it
	p.Update(mockCtx)
	if p.IsAlive() {
		t.Error("Particle should be dead after lifetime reaches 0")
	}
}

// Test Particle Update Does Nothing When Dead
func TestParticleUpdateWhenDead(t *testing.T) {
	p := NewAfterburnerParticle(1, 100, 200, 5.0, 3.0)
	mockCtx := &mockGameContext{}

	// Kill the particle
	p.Alive = false

	initialX, initialY := p.GetPosition()
	initialLifetime := p.Lifetime

	// Update should do nothing
	err := p.Update(mockCtx)
	if err != nil {
		t.Fatalf("Update should not return error: %v", err)
	}

	x, y := p.GetPosition()
	if x != initialX || y != initialY {
		t.Error("Dead particle should not move")
	}

	if p.Lifetime != initialLifetime {
		t.Error("Dead particle lifetime should not change")
	}
}

// Test Particle Position Wrapping
func TestParticlePositionWrapping(t *testing.T) {
	// Create particle near edge of game board
	p := NewAfterburnerParticle(1, 10, 10, -100.0, -100.0) // Moving fast left and up
	mockCtx := &mockGameContext{}

	// Update should wrap position
	p.Update(mockCtx)

	x, y := p.GetPosition()

	// Position should have wrapped (WrapPosition is tested elsewhere, just verify it was called)
	// Particle should not be at negative coordinates
	if x < 0 || y < 0 {
		t.Errorf("Particle position should wrap, got negative coordinates (%f, %f)", x, y)
	}
}

// Test Particle Render with Fade
func TestParticleRenderFade(t *testing.T) {
	p := NewAfterburnerParticle(1, 100, 200, 0, 0)

	// Create a test screen
	screen := ebiten.NewImage(800, 600)

	// Render at full lifetime (should have full alpha)
	p.Render(screen, 0, 0)

	// Advance particle to half lifetime
	mockCtx := &mockGameContext{}
	halfLife := p.Lifetime / 2
	for i := 0; i < halfLife; i++ {
		p.Update(mockCtx)
	}

	// Render at half lifetime (should have ~50% alpha, but we can't easily test the visual output)
	// This test mainly verifies Render doesn't panic
	p.Render(screen, 0, 0)

	// Render after particle dies (should do nothing)
	p.Alive = false
	p.Render(screen, 0, 0) // Should not panic
}

// Test Particle Render with Camera Offset
func TestParticleRenderWithCamera(t *testing.T) {
	p := NewAfterburnerParticle(1, 1000, 1000, 0, 0)

	screen := ebiten.NewImage(800, 600)

	// Render with camera offset - should not panic
	// Particle at (1000, 1000) with camera at (500, 500) should render at screen position (500, 500)
	p.Render(screen, 500, 500)
}

// Test Particle Render When Dead Does Nothing
func TestParticleRenderWhenDead(t *testing.T) {
	p := NewAfterburnerParticle(1, 100, 200, 0, 0)
	p.Alive = false

	screen := ebiten.NewImage(800, 600)

	// Should not panic
	p.Render(screen, 0, 0)
}

// Test Multiple Particle Updates
func TestParticleMultipleUpdates(t *testing.T) {
	p := NewAfterburnerParticle(1, 100, 100, 2.0, 3.0)
	mockCtx := &mockGameContext{}

	// Expected velocity per tick
	vx := 2.0 * 0.3
	vy := 3.0 * 0.3

	// Update 5 times
	for i := 0; i < 5; i++ {
		p.Update(mockCtx)
	}

	expectedX := 100 + (vx * 5)
	expectedY := 100 + (vy * 5)

	x, y := p.GetPosition()

	// Account for potential wrapping by checking the values are reasonable
	// (more precise test would need to know exact game board size)
	if !p.IsAlive() {
		t.Error("Particle should still be alive after 5 updates (lifetime=15)")
	}

	if p.Lifetime != 10 {
		t.Errorf("Expected lifetime 10 after 5 updates, got %d", p.Lifetime)
	}

	// Verify position changed (either to expected or wrapped)
	if x == 100 && y == 100 {
		t.Error("Particle position should have changed after updates")
	}

	// If no wrapping occurred, position should match expected
	if x == expectedX && y != expectedY {
		t.Errorf("Y position mismatch: expected %f, got %f", expectedY, y)
	}
	if y == expectedY && x != expectedX {
		t.Errorf("X position mismatch: expected %f, got %f", expectedX, x)
	}
}

// ============================================================================
// Mock GameContext
// ============================================================================

type mockGameContext struct{}

func (m *mockGameContext) FindNearestEnemy(ship *Ship) (*Ship, float64) { return nil, 0 }
func (m *mockGameContext) FindNearestEnemyInArc(ship *Ship, arc, maxRange float64, rearFacing bool) (*Ship, float64) {
	return nil, 0
}
func (m *mockGameContext) GetShip(id int) *Ship                               { return nil }
func (m *mockGameContext) GetAllShips() []*Ship                               { return nil }
func (m *mockGameContext) GetShipsByFaction(factionID int) []*Ship            { return nil }
func (m *mockGameContext) GetWorldSize() (width, height float64)              { return 0, 0 }
func (m *mockGameContext) SpawnProjectile(config MainGunConfig)               {}
func (m *mockGameContext) SpawnMissile(config MissileConfig)                  {}
func (m *mockGameContext) SpawnExplosion(x, y float64)                        {}
func (m *mockGameContext) SpawnParticle(x, y, vx, vy float64)                 {}
func (m *mockGameContext) AddKill()                                           {}
func (m *mockGameContext) AddScore(points int)                                {}
func (m *mockGameContext) PlayImpactSound(targetShip *Ship, proj Projectile)  {}
func (m *mockGameContext) OnShipDestroyed(victimShipID int, killerShipID int) {}
