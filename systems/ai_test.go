package systems

import (
	"testing"

	"github.com/yohamta/donburi"

	"github.com/nathan/verdant-thane/components"
	"github.com/nathan/verdant-thane/config"
)

func TestSelectNearestEnemy_SingleEnemy(t *testing.T) {
	world := donburi.NewWorld()

	// Create AI ship (faction 0)
	aiShip := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: 100, Y: 100})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})

	// Create single enemy (faction 1)
	enemy := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
	)
	enemyEntry := world.Entry(enemy)
	components.Position.SetValue(enemyEntry, components.PositionData{X: 200, Y: 200})
	components.Faction.SetValue(enemyEntry, components.FactionData{ID: 1})

	// Select nearest enemy
	target := SelectNearestEnemy(world, aiShip)

	if target != enemy {
		t.Error("SelectNearestEnemy should return the single enemy")
	}
}

func TestSelectNearestEnemy_MultipleEnemies(t *testing.T) {
	world := donburi.NewWorld()

	// Create AI ship (faction 0) at origin
	aiShip := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: 0, Y: 0})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})

	// Create far enemy
	farEnemy := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
	)
	farEntry := world.Entry(farEnemy)
	components.Position.SetValue(farEntry, components.PositionData{X: 1000, Y: 1000})
	components.Faction.SetValue(farEntry, components.FactionData{ID: 1})

	// Create close enemy
	closeEnemy := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
	)
	closeEntry := world.Entry(closeEnemy)
	components.Position.SetValue(closeEntry, components.PositionData{X: 50, Y: 50})
	components.Faction.SetValue(closeEntry, components.FactionData{ID: 1})

	// Select nearest enemy
	target := SelectNearestEnemy(world, aiShip)

	if target != closeEnemy {
		t.Error("SelectNearestEnemy should return the closest enemy")
	}
}

func TestSelectNearestEnemy_NoEnemies(t *testing.T) {
	world := donburi.NewWorld()

	// Create AI ship (faction 0)
	aiShip := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: 100, Y: 100})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})

	// Create friendly ship (same faction - should be ignored)
	friendly := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
	)
	friendlyEntry := world.Entry(friendly)
	components.Position.SetValue(friendlyEntry, components.PositionData{X: 150, Y: 150})
	components.Faction.SetValue(friendlyEntry, components.FactionData{ID: 0})

	// Select nearest enemy
	target := SelectNearestEnemy(world, aiShip)

	if world.Valid(target) {
		t.Error("SelectNearestEnemy should return invalid entity when no enemies")
	}
}

func TestSelectNearestEnemy_WorldWrapping(t *testing.T) {
	world := donburi.NewWorld()

	// Create AI ship at far right edge
	aiShip := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: float64(config.GameWidth - 100), Y: 100})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})

	// Create enemy at far left edge (should be close due to wrapping)
	wrappedEnemy := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
	)
	wrappedEntry := world.Entry(wrappedEnemy)
	components.Position.SetValue(wrappedEntry, components.PositionData{X: 100, Y: 100})
	components.Faction.SetValue(wrappedEntry, components.FactionData{ID: 1})

	// Create enemy in middle (farther despite direct distance)
	middleEnemy := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
	)
	middleEntry := world.Entry(middleEnemy)
	components.Position.SetValue(middleEntry, components.PositionData{X: float64(config.GameWidth / 2), Y: 100})
	components.Faction.SetValue(middleEntry, components.FactionData{ID: 1})

	// Select nearest enemy (should prefer wrapped enemy)
	target := SelectNearestEnemy(world, aiShip)

	if target != wrappedEnemy {
		t.Error("SelectNearestEnemy should handle world wrapping correctly")
	}
}

func TestAIRetargeting_TimerExpires(t *testing.T) {
	world := donburi.NewWorld()
	InitializeFactions(world)

	// Create AI ship with AIState and AITarget
	aiShip := world.Create(
		components.AIControlled,
		components.AIState,
		components.AITarget,
		components.Position,
		components.Ship,
		components.Rotation,
		components.Velocity,
		components.Faction,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: 100, Y: 100})
	components.Ship.SetValue(aiEntry, components.ShipData{MaxSpeed: 6.0})
	components.Rotation.SetValue(aiEntry, components.RotationData{Angle: 0})
	components.Velocity.SetValue(aiEntry, components.VelocityData{X: 0, Y: 0})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})

	// Set retarget timer to expire next update
	aiState := components.AIState.Get(aiEntry)
	aiState.RetargetTimer = 1

	// Create enemy
	enemy := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
	)
	enemyEntry := world.Entry(enemy)
	components.Position.SetValue(enemyEntry, components.PositionData{X: 200, Y: 200})
	components.Faction.SetValue(enemyEntry, components.FactionData{ID: 1})

	// Run AI movement (should trigger retargeting)
	UpdateAIMovement(world)

	// Check that target was set
	aiTarget := components.AITarget.Get(aiEntry)
	if aiTarget.TargetEntity != enemy {
		t.Error("AI should retarget when timer expires")
	}

	// Check that timer was reset
	aiStateAfter := components.AIState.Get(aiEntry)
	if aiStateAfter.RetargetTimer != config.AIRetargetInterval {
		t.Errorf("RetargetTimer should be reset to %d, got %d", config.AIRetargetInterval, aiStateAfter.RetargetTimer)
	}
}

func TestAIRetargeting_DeadTarget(t *testing.T) {
	world := donburi.NewWorld()
	InitializeFactions(world)

	// Create AI ship
	aiShip := world.Create(
		components.AIControlled,
		components.AIState,
		components.AITarget,
		components.Position,
		components.Ship,
		components.Rotation,
		components.Velocity,
		components.Faction,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: 100, Y: 100})
	components.Ship.SetValue(aiEntry, components.ShipData{MaxSpeed: 6.0})
	components.Rotation.SetValue(aiEntry, components.RotationData{Angle: 0})
	components.Velocity.SetValue(aiEntry, components.VelocityData{X: 0, Y: 0})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})

	// Create first enemy and set as target
	firstEnemy := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
	)
	firstEntry := world.Entry(firstEnemy)
	components.Position.SetValue(firstEntry, components.PositionData{X: 150, Y: 150})
	components.Faction.SetValue(firstEntry, components.FactionData{ID: 1})

	aiTarget := components.AITarget.Get(aiEntry)
	aiTarget.TargetEntity = firstEnemy

	// Create second enemy
	secondEnemy := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
	)
	secondEntry := world.Entry(secondEnemy)
	components.Position.SetValue(secondEntry, components.PositionData{X: 200, Y: 200})
	components.Faction.SetValue(secondEntry, components.FactionData{ID: 1})

	// "Kill" the first enemy
	world.Remove(firstEnemy)

	// Run AI movement (should detect dead target and retarget)
	UpdateAIMovement(world)

	// Check that target was updated to second enemy
	aiTargetAfter := components.AITarget.Get(aiEntry)
	if aiTargetAfter.TargetEntity != secondEnemy {
		t.Error("AI should retarget when current target is destroyed")
	}
}

func TestAIMovement_PursuitBehavior(t *testing.T) {
	world := donburi.NewWorld()
	InitializeFactions(world)

	// Create AI ship facing north (angle 0)
	aiShip := world.Create(
		components.AIControlled,
		components.AIState,
		components.AITarget,
		components.Position,
		components.Ship,
		components.Rotation,
		components.Velocity,
		components.Faction,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: 100, Y: 100})
	components.Ship.SetValue(aiEntry, components.ShipData{Speed: 0, MaxSpeed: 6.0})
	components.Rotation.SetValue(aiEntry, components.RotationData{Angle: 0})
	components.Velocity.SetValue(aiEntry, components.VelocityData{X: 0, Y: 0})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})
	components.AIState.SetValue(aiEntry, components.AIStateData{RetargetTimer: 1})

	// Create enemy to the east
	enemy := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
	)
	enemyEntry := world.Entry(enemy)
	components.Position.SetValue(enemyEntry, components.PositionData{X: 300, Y: 100})
	components.Faction.SetValue(enemyEntry, components.FactionData{ID: 1})

	// Run AI movement
	UpdateAIMovement(world)

	// AI should rotate toward enemy (to the east, angle ≈ π/2)
	rotation := components.Rotation.Get(aiEntry)
	// After one update, should have rotated toward east
	// Since enemy is directly east and ship faces north, target angle is π/2
	// Ship should rotate by aiRotationSpeed toward that angle
	if rotation.Angle <= 0 {
		t.Error("AI should rotate toward enemy to the east (positive angle)")
	}

	// AI should increase speed for pursuit
	shipData := components.Ship.Get(aiEntry)
	minExpected := shipData.MaxSpeed * config.AIPursuitSpeedMin
	if shipData.Speed < minExpected {
		t.Errorf("AI speed should be at least %f for pursuit, got %f", minExpected, shipData.Speed)
	}
}

func TestAIMovement_NoTarget(t *testing.T) {
	world := donburi.NewWorld()
	InitializeFactions(world)

	// Create AI ship with no enemies
	aiShip := world.Create(
		components.AIControlled,
		components.AIState,
		components.AITarget,
		components.Position,
		components.Ship,
		components.Rotation,
		components.Velocity,
		components.Faction,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: 100, Y: 100})
	components.Ship.SetValue(aiEntry, components.ShipData{Speed: 6.0, MaxSpeed: 6.0})
	components.Rotation.SetValue(aiEntry, components.RotationData{Angle: 0})
	components.Velocity.SetValue(aiEntry, components.VelocityData{X: 0, Y: 0})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})
	components.AIState.SetValue(aiEntry, components.AIStateData{RetargetTimer: 1})

	var emptyEntity donburi.Entity
	components.AITarget.SetValue(aiEntry, components.AITargetData{TargetEntity: emptyEntity})

	// Run AI movement
	UpdateAIMovement(world)

	// AI should reduce to patrol speed when no target
	shipData := components.Ship.Get(aiEntry)
	expectedSpeed := shipData.MaxSpeed * config.AIPatrolSpeed
	if shipData.Speed != expectedSpeed {
		t.Errorf("AI speed should be patrol speed %f when no target, got %f", expectedSpeed, shipData.Speed)
	}
}
