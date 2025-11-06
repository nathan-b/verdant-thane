package systems

import (
	"math"
	"testing"

	"github.com/yohamta/donburi"

	"github.com/nathan/verdant-thane/components"
	"github.com/nathan/verdant-thane/config"
)

func TestMissileTurnRate_RespectsTurnLimit(t *testing.T) {
	world := donburi.NewWorld()

	// Create a target ship
	target := world.Create(components.IsShip, components.Position)
	targetEntry := world.Entry(target)
	components.Position.SetValue(targetEntry, components.PositionData{X: 100, Y: 100})

	// Create a missile heading right (90°), with target directly above (0°)
	// This requires a 90° turn, which should be constrained
	missile := world.Create(
		components.IsMissile,
		components.Missile,
		components.Position,
		components.Velocity,
		components.Rotation,
	)
	missileEntry := world.Entry(missile)

	// Missile at origin, velocity pointing right (east)
	components.Position.SetValue(missileEntry, components.PositionData{X: 100, Y: 200})
	components.Velocity.SetValue(missileEntry, components.VelocityData{X: 10.0, Y: 0.0}) // Heading right
	components.Rotation.SetValue(missileEntry, components.RotationData{Angle: math.Pi / 2})

	// Set missile to track the target
	missileChars := config.GetProjectileCharacteristics(config.MissileProjectile)
	components.Missile.SetValue(missileEntry, components.MissileData{
		TargetEntity: target,
		Acceleration: missileChars.Acceleration,
	})

	// Initial angle: pointing right (90°)
	initialVel := components.Velocity.Get(missileEntry)
	initialAngle := math.Atan2(initialVel.X, -initialVel.Y)

	// Run one tracking update
	UpdateMissileTracking(world)

	// Check that velocity changed
	newVel := components.Velocity.Get(missileEntry)
	newAngle := math.Atan2(newVel.X, -newVel.Y)

	// Calculate how much the angle changed
	angleDiff := math.Abs(NormalizeAngle(newAngle - initialAngle))

	// The angle should have changed, but not more than TurnRate
	maxTurnRate := missileChars.TurnRate
	if angleDiff > maxTurnRate+0.001 { // Small epsilon for floating point
		t.Errorf("Missile turned too fast: %.4f radians (max: %.4f)", angleDiff, maxTurnRate)
	}

	if angleDiff < 0.001 {
		t.Error("Missile didn't turn at all")
	}
}

func TestMissileTurnRate_GradualTurning(t *testing.T) {
	world := donburi.NewWorld()

	// Create a target ship directly behind the missile
	target := world.Create(components.IsShip, components.Position)
	targetEntry := world.Entry(target)
	components.Position.SetValue(targetEntry, components.PositionData{X: 100, Y: 300}) // Behind missile

	// Create a missile heading up (0°) with slower initial speed
	// (slower speed allows acceleration to have more effect on direction)
	missile := world.Create(
		components.IsMissile,
		components.Missile,
		components.Position,
		components.Velocity,
		components.Rotation,
	)
	missileEntry := world.Entry(missile)

	// Missile heading north with speed close to missile initial speed
	components.Position.SetValue(missileEntry, components.PositionData{X: 100, Y: 100})
	components.Velocity.SetValue(missileEntry, components.VelocityData{X: 0.0, Y: -4.0}) // Heading up
	components.Rotation.SetValue(missileEntry, components.RotationData{Angle: 0})

	missileChars := config.GetProjectileCharacteristics(config.MissileProjectile)
	components.Missile.SetValue(missileEntry, components.MissileData{
		TargetEntity: target,
		Acceleration: missileChars.Acceleration,
	})

	// Track angles over multiple updates
	angles := []float64{}
	vel := components.Velocity.Get(missileEntry)
	angles = append(angles, math.Atan2(vel.X, -vel.Y))

	// Run multiple tracking updates
	for i := 0; i < 50; i++ {
		UpdateMissileTracking(world)
		vel := components.Velocity.Get(missileEntry)
		angles = append(angles, math.Atan2(vel.X, -vel.Y))
	}

	// Verify that the missile is gradually turning (each step respects turn rate)
	for i := 1; i < len(angles); i++ {
		angleDiff := math.Abs(NormalizeAngle(angles[i] - angles[i-1]))
		if angleDiff > missileChars.TurnRate+0.001 {
			t.Errorf("Turn %d: Missile turned too fast: %.4f radians (max: %.4f)", i, angleDiff, missileChars.TurnRate)
		}
	}

	// Verify that the missile has turned significantly toward the target (should be turning around)
	finalAngle := angles[len(angles)-1]
	initialAngle := angles[0]
	totalTurn := math.Abs(NormalizeAngle(finalAngle - initialAngle))

	// With slower initial velocity, should turn noticeably over time
	// The acceleration (0.15) relative to velocity (4.0) means gradual turning
	// Over 50 ticks, expect at least a few degrees of turn
	if totalTurn < 0.05 { // ~3 degrees minimum
		t.Errorf("Missile didn't turn enough over 50 ticks: %.4f radians (%.1f degrees)", totalTurn, totalTurn*180/math.Pi)
	}

	// Verify missile is turning in the right direction (should be moving toward π radians)
	if finalAngle < initialAngle {
		t.Errorf("Missile turned in wrong direction: initial=%.4f, final=%.4f", initialAngle, finalAngle)
	}
}

func TestMissileTurnRate_NoInstantSnap(t *testing.T) {
	world := donburi.NewWorld()

	// Create a target perpendicular to missile path
	target := world.Create(components.IsShip, components.Position)
	targetEntry := world.Entry(target)
	components.Position.SetValue(targetEntry, components.PositionData{X: 200, Y: 100}) // To the right

	// Create a missile heading up
	missile := world.Create(
		components.IsMissile,
		components.Missile,
		components.Position,
		components.Velocity,
		components.Rotation,
	)
	missileEntry := world.Entry(missile)

	components.Position.SetValue(missileEntry, components.PositionData{X: 100, Y: 100})
	components.Velocity.SetValue(missileEntry, components.VelocityData{X: 0.0, Y: -10.0}) // Heading up
	components.Rotation.SetValue(missileEntry, components.RotationData{Angle: 0})

	missileChars := config.GetProjectileCharacteristics(config.MissileProjectile)
	components.Missile.SetValue(missileEntry, components.MissileData{
		TargetEntity: target,
		Acceleration: missileChars.Acceleration,
	})

	// Get initial velocity direction
	initialVel := components.Velocity.Get(missileEntry)
	initialAngle := math.Atan2(initialVel.X, -initialVel.Y)

	// Run one update
	UpdateMissileTracking(world)

	// Get new velocity direction
	newVel := components.Velocity.Get(missileEntry)
	newAngle := math.Atan2(newVel.X, -newVel.Y)

	// Calculate angle to target
	missilePos := components.Position.Get(missileEntry)
	targetPos := components.Position.Get(targetEntry)
	dx := targetPos.X - missilePos.X
	dy := targetPos.Y - missilePos.Y
	targetAngle := math.Atan2(dx, -dy)

	// The missile should NOT be pointing directly at the target after one tick
	angleToTarget := math.Abs(NormalizeAngle(newAngle - targetAngle))

	// Should still be significantly off from target direction
	if angleToTarget < 0.1 { // Less than ~6 degrees off
		t.Errorf("Missile snapped too close to target direction in one tick: %.4f radians off", angleToTarget)
	}

	// But it should be closer than before
	initialAngleToTarget := math.Abs(NormalizeAngle(initialAngle - targetAngle))
	if angleToTarget >= initialAngleToTarget {
		t.Error("Missile didn't turn toward target at all")
	}
}

func TestMissileTurnRate_NoTurningWhenNoTarget(t *testing.T) {
	world := donburi.NewWorld()

	// Create a missile with no valid target
	missile := world.Create(
		components.IsMissile,
		components.Missile,
		components.Position,
		components.Velocity,
		components.Rotation,
	)
	missileEntry := world.Entry(missile)

	components.Position.SetValue(missileEntry, components.PositionData{X: 100, Y: 100})
	components.Velocity.SetValue(missileEntry, components.VelocityData{X: 5.0, Y: -5.0})
	components.Rotation.SetValue(missileEntry, components.RotationData{Angle: math.Pi / 4})

	// Invalid target entity
	var invalidTarget donburi.Entity
	missileChars := config.GetProjectileCharacteristics(config.MissileProjectile)
	components.Missile.SetValue(missileEntry, components.MissileData{
		TargetEntity: invalidTarget,
		Acceleration: missileChars.Acceleration,
	})

	// Get initial velocity
	initialVel := components.Velocity.Get(missileEntry)
	initialVX := initialVel.X
	initialVY := initialVel.Y

	// Run tracking update
	UpdateMissileTracking(world)

	// Velocity should be unchanged (no target to track)
	newVel := components.Velocity.Get(missileEntry)
	if math.Abs(newVel.X-initialVX) > 0.001 || math.Abs(newVel.Y-initialVY) > 0.001 {
		t.Error("Missile velocity changed without valid target")
	}
}

func TestMissileTurnRate_AccelerationInTurnDirection(t *testing.T) {
	world := donburi.NewWorld()

	// Create a target to the right
	target := world.Create(components.IsShip, components.Position)
	targetEntry := world.Entry(target)
	components.Position.SetValue(targetEntry, components.PositionData{X: 200, Y: 100})

	// Create a missile heading up with initial speed close to missile specs
	missile := world.Create(
		components.IsMissile,
		components.Missile,
		components.Position,
		components.Velocity,
		components.Rotation,
	)
	missileEntry := world.Entry(missile)

	components.Position.SetValue(missileEntry, components.PositionData{X: 100, Y: 100})
	components.Velocity.SetValue(missileEntry, components.VelocityData{X: 0.0, Y: -4.0}) // Heading up
	components.Rotation.SetValue(missileEntry, components.RotationData{Angle: 0})

	missileChars := config.GetProjectileCharacteristics(config.MissileProjectile)
	components.Missile.SetValue(missileEntry, components.MissileData{
		TargetEntity: target,
		Acceleration: missileChars.Acceleration,
	})

	// Get initial velocity
	initialVel := components.Velocity.Get(missileEntry)
	initialSpeed := math.Sqrt(initialVel.X*initialVel.X + initialVel.Y*initialVel.Y)
	initialX := initialVel.X

	// Run multiple updates to see cumulative effect
	for i := 0; i < 10; i++ {
		UpdateMissileTracking(world)
	}

	// Check that velocity has changed
	newVel := components.Velocity.Get(missileEntry)
	newSpeed := math.Sqrt(newVel.X*newVel.X + newVel.Y*newVel.Y)

	// Speed should have increased (acceleration applied)
	if newSpeed <= initialSpeed {
		t.Errorf("Missile speed didn't increase: initial=%.4f, new=%.4f", initialSpeed, newSpeed)
	}

	// X velocity should have increased significantly after 10 ticks (turning right toward target)
	if newVel.X <= initialX+0.1 {
		t.Errorf("Missile didn't turn right toward target enough: initial X=%.4f, new X=%.4f", initialX, newVel.X)
	}
}
