package systems

import (
	"math"
	"testing"

	"github.com/yohamta/donburi"

	"github.com/nathan/verdant-thane/components"
)

func TestMissile_CanDoUTurnInOneSecond(t *testing.T) {
	world := donburi.NewWorld()

	// Create a target directly behind the missile (180° turn needed)
	target := world.Create(components.IsShip, components.Position)
	targetEntry := world.Entry(target)
	components.Position.SetValue(targetEntry, components.PositionData{X: 100, Y: 200}) // Behind

	// Create a missile heading north
	missile := world.Create(
		components.IsMissile,
		components.Missile,
		components.Position,
		components.Velocity,
		components.Rotation,
	)
	missileEntry := world.Entry(missile)

	components.Position.SetValue(missileEntry, components.PositionData{X: 100, Y: 100})
	components.Velocity.SetValue(missileEntry, components.VelocityData{X: 0.0, Y: -4.0}) // Heading up/north
	components.Rotation.SetValue(missileEntry, components.RotationData{Angle: 0})

	missileChars := GetProjectileCharacteristics(MissileProjectile)
	components.Missile.SetValue(missileEntry, components.MissileData{
		TargetEntity: target,
		Acceleration: missileChars.Acceleration,
	})

	initialVel := components.Velocity.Get(missileEntry)
	initialAngle := math.Atan2(initialVel.X, -initialVel.Y)

	t.Logf("Turn rate: %.4f radians/tick (%.1f degrees/tick)", missileChars.TurnRate, missileChars.TurnRate*180/math.Pi)
	t.Logf("At 60 TPS: %.1f degrees/second", missileChars.TurnRate*180/math.Pi*60)
	t.Logf("Expected time for 180° turn: %.1f ticks", math.Pi/(missileChars.TurnRate))

	// Run for 60 ticks (1 second at 60 TPS)
	for i := 0; i < 60; i++ {
		UpdateMissileTracking(world)
	}

	finalVel := components.Velocity.Get(missileEntry)
	finalAngle := math.Atan2(finalVel.X, -finalVel.Y)

	totalTurn := math.Abs(normalizeAngle(finalAngle - initialAngle))
	t.Logf("Initial angle: %.4f radians (%.1f degrees)", initialAngle, initialAngle*180/math.Pi)
	t.Logf("Final angle: %.4f radians (%.1f degrees)", finalAngle, finalAngle*180/math.Pi)
	t.Logf("Total turn: %.4f radians (%.1f degrees)", totalTurn, totalTurn*180/math.Pi)

	// At 6 degrees per tick, 60 ticks should turn 360 degrees
	// For a 180 degree turn, should only need 30 ticks
	// So after 60 ticks, should have completed the 180 degree turn
	expectedMinimumTurn := math.Pi // 180 degrees
	tolerance := 0.01              // ~0.6 degrees tolerance for floating-point precision

	if totalTurn < expectedMinimumTurn-tolerance {
		t.Errorf("Missile did NOT complete 180° turn in 1 second: only turned %.1f degrees (expected at least %.1f degrees)",
			totalTurn*180/math.Pi, expectedMinimumTurn*180/math.Pi)
	}
}
