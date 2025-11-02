package systems

import (
	"testing"

	"github.com/yohamta/donburi"

	"github.com/nathan/verdant-thane/components"
)

func TestUpdateMovement_BasicMovement(t *testing.T) {
	// Create test world
	world := donburi.NewWorld()

	// Create entity with position and velocity
	entity := world.Create(components.Position, components.Velocity)
	entry := world.Entry(entity)

	// Set initial position and velocity
	components.Position.SetValue(entry, components.PositionData{X: 100, Y: 200})
	components.Velocity.SetValue(entry, components.VelocityData{X: 5, Y: -3})

	// Run movement system
	UpdateMovement(world)

	// Check position was updated
	pos := components.Position.Get(entry)
	if pos.X != 105 {
		t.Errorf("Expected X=105, got %f", pos.X)
	}
	if pos.Y != 197 {
		t.Errorf("Expected Y=197, got %f", pos.Y)
	}
}

func TestUpdateMovement_WorldWrappingRight(t *testing.T) {
	world := donburi.NewWorld()

	entity := world.Create(components.Position, components.Velocity)
	entry := world.Entry(entity)

	// Position near right edge
	components.Position.SetValue(entry, components.PositionData{X: GameWidth - 1, Y: 100})
	components.Velocity.SetValue(entry, components.VelocityData{X: 5, Y: 0})

	UpdateMovement(world)

	pos := components.Position.Get(entry)
	// Should wrap to left side
	if pos.X >= GameWidth {
		t.Errorf("Expected X to wrap around, got %f", pos.X)
	}
	if pos.X != 4 { // -1 + 5 = 4, then wraps
		t.Errorf("Expected X=4 after wrapping, got %f", pos.X)
	}
}

func TestUpdateMovement_WorldWrappingLeft(t *testing.T) {
	world := donburi.NewWorld()

	entity := world.Create(components.Position, components.Velocity)
	entry := world.Entry(entity)

	// Position near left edge
	components.Position.SetValue(entry, components.PositionData{X: 2, Y: 100})
	components.Velocity.SetValue(entry, components.VelocityData{X: -5, Y: 0})

	UpdateMovement(world)

	pos := components.Position.Get(entry)
	// Should wrap to right side
	if pos.X < 0 {
		t.Errorf("Expected X to be positive after wrapping, got %f", pos.X)
	}
	expectedX := GameWidth - 3 // 2 - 5 = -3, wraps to GameWidth - 3
	if pos.X != float64(expectedX) {
		t.Errorf("Expected X=%f after wrapping, got %f", float64(expectedX), pos.X)
	}
}

func TestUpdateMovement_WorldWrappingBottom(t *testing.T) {
	world := donburi.NewWorld()

	entity := world.Create(components.Position, components.Velocity)
	entry := world.Entry(entity)

	// Position near bottom edge
	components.Position.SetValue(entry, components.PositionData{X: 100, Y: GameHeight - 1})
	components.Velocity.SetValue(entry, components.VelocityData{X: 0, Y: 5})

	UpdateMovement(world)

	pos := components.Position.Get(entry)
	// Should wrap to top
	if pos.Y >= GameHeight {
		t.Errorf("Expected Y to wrap around, got %f", pos.Y)
	}
}

func TestUpdateMovement_MultipleEntities(t *testing.T) {
	world := donburi.NewWorld()

	// Create three entities
	entity1 := world.Create(components.Position, components.Velocity)
	entity2 := world.Create(components.Position, components.Velocity)
	entity3 := world.Create(components.Position, components.Velocity)

	entry1 := world.Entry(entity1)
	entry2 := world.Entry(entity2)
	entry3 := world.Entry(entity3)

	// Set different positions and velocities
	components.Position.SetValue(entry1, components.PositionData{X: 10, Y: 10})
	components.Velocity.SetValue(entry1, components.VelocityData{X: 1, Y: 1})

	components.Position.SetValue(entry2, components.PositionData{X: 50, Y: 50})
	components.Velocity.SetValue(entry2, components.VelocityData{X: 2, Y: -2})

	components.Position.SetValue(entry3, components.PositionData{X: 100, Y: 100})
	components.Velocity.SetValue(entry3, components.VelocityData{X: 0, Y: 0})

	// Run movement system
	UpdateMovement(world)

	// Check all entities updated correctly
	pos1 := components.Position.Get(entry1)
	if pos1.X != 11 || pos1.Y != 11 {
		t.Errorf("Entity 1: Expected (11, 11), got (%f, %f)", pos1.X, pos1.Y)
	}

	pos2 := components.Position.Get(entry2)
	if pos2.X != 52 || pos2.Y != 48 {
		t.Errorf("Entity 2: Expected (52, 48), got (%f, %f)", pos2.X, pos2.Y)
	}

	pos3 := components.Position.Get(entry3)
	if pos3.X != 100 || pos3.Y != 100 {
		t.Errorf("Entity 3: Expected (100, 100), got (%f, %f)", pos3.X, pos3.Y)
	}
}

func TestUpdateMovement_ZeroVelocity(t *testing.T) {
	world := donburi.NewWorld()

	entity := world.Create(components.Position, components.Velocity)
	entry := world.Entry(entity)

	components.Position.SetValue(entry, components.PositionData{X: 123.456, Y: 789.012})
	components.Velocity.SetValue(entry, components.VelocityData{X: 0, Y: 0})

	UpdateMovement(world)

	pos := components.Position.Get(entry)
	if pos.X != 123.456 || pos.Y != 789.012 {
		t.Errorf("Position should not change with zero velocity, got (%f, %f)", pos.X, pos.Y)
	}
}
