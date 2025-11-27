package entity

import (
	"testing"

	"github.com/nathan-b/verdant-thane/config"
)

// Test Impact Creation
func TestNewImpact(t *testing.T) {
	impact := NewImpact(42, 100, 200, nil)

	// Verify identity
	if impact.GetID() != 42 {
		t.Errorf("Expected ID 42, got %d", impact.GetID())
	}

	// Verify position
	x, y := impact.GetPosition()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%f, %f)", x, y)
	}

	// Verify starts alive
	if !impact.IsAlive() {
		t.Error("Impact should start alive")
	}

	// Verify starts at frame 0
	if impact.CurrentFrame != 0 {
		t.Errorf("Expected current frame 0, got %d", impact.CurrentFrame)
	}

	// Verify timer starts at 0
	if impact.FrameTimer != 0 {
		t.Errorf("Expected frame timer 0, got %d", impact.FrameTimer)
	}
}

// Test Impact Frame Advancement
func TestImpactFrameAdvancement(t *testing.T) {
	ctx := NewMockGameContext()
	impact := NewImpact(1, 100, 100, nil)

	// Update 2 times (less than 3 ticks per frame)
	for i := 0; i < 2; i++ {
		impact.Update(ctx)
	}

	// Should still be on frame 0
	if impact.CurrentFrame != 0 {
		t.Errorf("Expected frame 0 after 2 updates, got %d", impact.CurrentFrame)
	}
	if !impact.IsAlive() {
		t.Error("Impact should still be alive")
	}

	// Update once more (3 total) - should advance to frame 1
	impact.Update(ctx)

	if impact.CurrentFrame != 1 {
		t.Errorf("Expected frame 1 after 3 updates, got %d", impact.CurrentFrame)
	}
	if impact.FrameTimer != 0 {
		t.Errorf("Expected frame timer reset to 0, got %d", impact.FrameTimer)
	}
}

// Test Impact Animation Completes
func TestImpactAnimationCompletion(t *testing.T) {
	ctx := NewMockGameContext()
	impact := NewImpact(1, 100, 100, nil)

	// Calculate total ticks needed for animation to complete
	// 5 frames, 3 ticks per frame = 15 ticks total
	ticksPerFrame := 3
	totalTicks := config.ImpactFrameCount * ticksPerFrame

	// Update until just before completion
	for i := 0; i < totalTicks-1; i++ {
		impact.Update(ctx)
		if !impact.IsAlive() {
			t.Fatalf("Impact died early at tick %d", i+1)
		}
	}

	// Should be on last frame
	if impact.CurrentFrame != config.ImpactFrameCount-1 {
		t.Errorf("Expected current frame %d, got %d",
			config.ImpactFrameCount-1, impact.CurrentFrame)
	}

	// One more update should complete the animation
	impact.Update(ctx)

	if impact.IsAlive() {
		t.Error("Impact should be dead after animation completes")
	}
}

// Test Dead Impact Does Not Update
func TestDeadImpactDoesNotUpdate(t *testing.T) {
	ctx := NewMockGameContext()
	impact := NewImpact(1, 100, 100, nil)
	impact.Alive = false
	impact.CurrentFrame = 2

	// Try to update
	impact.Update(ctx)

	// Frame should not advance
	if impact.CurrentFrame != 2 {
		t.Errorf("Dead impact frame should not change, got %d", impact.CurrentFrame)
	}

	// Should still be dead
	if impact.IsAlive() {
		t.Error("Impact should remain dead")
	}
}

// Test Impact Frame Progression
func TestImpactFrameProgression(t *testing.T) {
	ctx := NewMockGameContext()
	impact := NewImpact(1, 100, 100, nil)

	expectedFrames := []int{0, 1, 2, 3, 4}

	for i, expectedFrame := range expectedFrames {
		// Verify current frame
		if impact.CurrentFrame != expectedFrame {
			t.Errorf("At step %d: expected frame %d, got %d", i, expectedFrame, impact.CurrentFrame)
		}

		// Update 3 times to advance to next frame
		for j := 0; j < 3; j++ {
			impact.Update(ctx)
		}
	}

	// After all frames, should be dead
	if impact.IsAlive() {
		t.Error("Impact should be dead after showing all frames")
	}
}

// Test Impact Position Does Not Change
func TestImpactPositionStatic(t *testing.T) {
	ctx := NewMockGameContext()
	impact := NewImpact(1, 123.45, 678.90, nil)

	initialX, initialY := impact.GetPosition()

	// Update multiple times
	for i := 0; i < 10; i++ {
		impact.Update(ctx)
	}

	// Position should not change
	x, y := impact.GetPosition()
	if x != initialX || y != initialY {
		t.Errorf("Impact position should not change, expected (%f, %f), got (%f, %f)",
			initialX, initialY, x, y)
	}
}
