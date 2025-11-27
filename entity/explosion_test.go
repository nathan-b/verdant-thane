package entity

import (
	"testing"

	"github.com/nathan-b/verdant-thane/config"
)

// Test Explosion Creation
func TestNewExplosion(t *testing.T) {
	explosion := NewExplosion(42, 100, 200, nil)

	// Verify identity
	if explosion.GetID() != 42 {
		t.Errorf("Expected ID 42, got %d", explosion.GetID())
	}

	// Verify position
	x, y := explosion.GetPosition()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%f, %f)", x, y)
	}

	// Verify starts alive
	if !explosion.IsAlive() {
		t.Error("Explosion should start alive")
	}

	// Verify starts at frame 0
	if explosion.CurrentFrame != 0 {
		t.Errorf("Expected current frame 0, got %d", explosion.CurrentFrame)
	}

	// Verify timer starts at 0
	if explosion.FrameTimer != 0 {
		t.Errorf("Expected frame timer 0, got %d", explosion.FrameTimer)
	}
}

// Test Explosion Frame Advancement
func TestExplosionFrameAdvancement(t *testing.T) {
	ctx := NewMockGameContext()
	explosion := NewExplosion(1, 100, 100, nil)

	// Update 4 times (less than 5 ticks per frame)
	for i := 0; i < 4; i++ {
		explosion.Update(ctx)
	}

	// Should still be on frame 0
	if explosion.CurrentFrame != 0 {
		t.Errorf("Expected frame 0 after 4 updates, got %d", explosion.CurrentFrame)
	}
	if !explosion.IsAlive() {
		t.Error("Explosion should still be alive")
	}

	// Update once more (5 total) - should advance to frame 1
	explosion.Update(ctx)

	if explosion.CurrentFrame != 1 {
		t.Errorf("Expected frame 1 after 5 updates, got %d", explosion.CurrentFrame)
	}
	if explosion.FrameTimer != 0 {
		t.Errorf("Expected frame timer reset to 0, got %d", explosion.FrameTimer)
	}
}

// Test Explosion Animation Completes
func TestExplosionAnimationCompletion(t *testing.T) {
	ctx := NewMockGameContext()
	explosion := NewExplosion(1, 100, 100, nil)

	// Calculate total ticks needed for animation to complete
	// 4 frames, 5 ticks per frame = 20 ticks total
	ticksPerFrame := 5
	totalTicks := config.ExplosionFrameCount * ticksPerFrame

	// Update until just before completion
	for i := 0; i < totalTicks-1; i++ {
		explosion.Update(ctx)
		if !explosion.IsAlive() {
			t.Fatalf("Explosion died early at tick %d", i+1)
		}
	}

	// Should be on last frame
	if explosion.CurrentFrame != config.ExplosionFrameCount-1 {
		t.Errorf("Expected current frame %d, got %d",
			config.ExplosionFrameCount-1, explosion.CurrentFrame)
	}

	// One more update should complete the animation
	explosion.Update(ctx)

	if explosion.IsAlive() {
		t.Error("Explosion should be dead after animation completes")
	}
}

// Test Dead Explosion Does Not Update
func TestDeadExplosionDoesNotUpdate(t *testing.T) {
	ctx := NewMockGameContext()
	explosion := NewExplosion(1, 100, 100, nil)
	explosion.Alive = false
	explosion.CurrentFrame = 2

	// Try to update
	explosion.Update(ctx)

	// Frame should not advance
	if explosion.CurrentFrame != 2 {
		t.Errorf("Dead explosion frame should not change, got %d", explosion.CurrentFrame)
	}

	// Should still be dead
	if explosion.IsAlive() {
		t.Error("Explosion should remain dead")
	}
}

// Test Explosion Frame Progression
func TestExplosionFrameProgression(t *testing.T) {
	ctx := NewMockGameContext()
	explosion := NewExplosion(1, 100, 100, nil)

	expectedFrames := []int{0, 1, 2, 3}

	for i, expectedFrame := range expectedFrames {
		// Verify current frame
		if explosion.CurrentFrame != expectedFrame {
			t.Errorf("At step %d: expected frame %d, got %d", i, expectedFrame, explosion.CurrentFrame)
		}

		// Update 5 times to advance to next frame
		for j := 0; j < 5; j++ {
			explosion.Update(ctx)
		}
	}

	// After all frames, should be dead
	if explosion.IsAlive() {
		t.Error("Explosion should be dead after showing all frames")
	}
}

// Test Explosion Position Does Not Change
func TestExplosionPositionStatic(t *testing.T) {
	ctx := NewMockGameContext()
	explosion := NewExplosion(1, 123.45, 678.90, nil)

	initialX, initialY := explosion.GetPosition()

	// Update multiple times
	for i := 0; i < 10; i++ {
		explosion.Update(ctx)
	}

	// Position should not change
	x, y := explosion.GetPosition()
	if x != initialX || y != initialY {
		t.Errorf("Explosion position should not change, expected (%f, %f), got (%f, %f)",
			initialX, initialY, x, y)
	}
}
