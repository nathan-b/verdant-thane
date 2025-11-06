package systems

import (
	"math"
	"testing"

	"github.com/yohamta/donburi"

	"github.com/nathan/verdant-thane/components"
)

// TestBeamWeapon_DamageAccumulation verifies that beam damage accumulates correctly
func TestBeamWeapon_DamageAccumulation(t *testing.T) {
	world := donburi.NewWorld()

	// Create attacker with beam weapon (needs Position, Faction, Health, BeamWeapon)
	attacker := world.Create(
		components.Position,
		components.Faction,
		components.Health,
		components.BeamWeapon,
	)
	attackerEntry := world.Entry(attacker)
	components.Position.SetValue(attackerEntry, components.PositionData{X: 100, Y: 100})
	components.Faction.SetValue(attackerEntry, components.FactionData{ID: 0})
	components.Health.SetValue(attackerEntry, components.HealthData{Current: 100, Max: 100})

	// Create target
	target := world.Create(
		components.Position,
		components.Faction,
		components.Health,
	)
	targetEntry := world.Entry(target)
	components.Position.SetValue(targetEntry, components.PositionData{X: 150, Y: 100})
	components.Faction.SetValue(targetEntry, components.FactionData{ID: 1})
	components.Health.SetValue(targetEntry, components.HealthData{Current: 100, Max: 100})

	// Set beam weapon to target (use 0.5 damage/tick to avoid floating-point issues)
	components.BeamWeapon.SetValue(attackerEntry, components.BeamWeaponData{
		TargetEntity:      target,
		Range:             200.0,
		DamagePerTick:     0.5, // 1 damage per 2 ticks (avoids 0.1 float precision issues)
		DamageAccumulator: 0.0,
	})

	// Run for 2 ticks - should accumulate to 1.0 damage exactly
	for i := 0; i < 2; i++ {
		UpdateBeamWeapons(world)
	}

	// Check that exactly 1 damage was applied
	health := components.Health.Get(targetEntry)
	if health.Current != 99 {
		t.Errorf("Expected health 99, got %d", health.Current)
	}

	// Check accumulator is 0 (exact with 0.5 increments)
	beamWeapon := components.BeamWeapon.Get(attackerEntry)
	if beamWeapon.DamageAccumulator != 0.0 {
		t.Errorf("Accumulator should be 0, got %f", beamWeapon.DamageAccumulator)
	}
}

// TestBeamWeapon_RangeCheck verifies beam weapons stop firing when target is out of range
func TestBeamWeapon_RangeCheck(t *testing.T) {
	world := donburi.NewWorld()

	// Create attacker at origin
	attacker := world.Create(
		components.Position,
		components.Faction,
		components.Health,
		components.BeamWeapon,
	)
	attackerEntry := world.Entry(attacker)
	components.Position.SetValue(attackerEntry, components.PositionData{X: 100, Y: 100})
	components.Faction.SetValue(attackerEntry, components.FactionData{ID: 0})
	components.Health.SetValue(attackerEntry, components.HealthData{Current: 100, Max: 100})

	// Create target out of range (beam range is 200)
	target := world.Create(
		components.Position,
		components.Faction,
		components.Health,
	)
	targetEntry := world.Entry(target)
	components.Position.SetValue(targetEntry, components.PositionData{X: 400, Y: 100}) // 300 pixels away
	components.Faction.SetValue(targetEntry, components.FactionData{ID: 1})
	components.Health.SetValue(targetEntry, components.HealthData{Current: 100, Max: 100})

	// Set beam weapon to target (initially out of range)
	components.BeamWeapon.SetValue(attackerEntry, components.BeamWeaponData{
		TargetEntity:      target,
		Range:             200.0,
		DamagePerTick:     0.1,
		DamageAccumulator: 0.0,
	})

	// Run update - beam should clear target since out of range
	UpdateBeamWeapons(world)

	// Target should be cleared
	beamWeapon := components.BeamWeapon.Get(attackerEntry)
	if world.Valid(beamWeapon.TargetEntity) {
		t.Error("Target should be cleared when out of range")
	}

	// Health should be unchanged
	health := components.Health.Get(targetEntry)
	if health.Current != 100 {
		t.Errorf("Target health should be unchanged, got %d", health.Current)
	}
}

// TestBeamWeapon_FriendlyFirePrevention verifies beams don't damage friendly ships
func TestBeamWeapon_FriendlyFirePrevention(t *testing.T) {
	world := donburi.NewWorld()

	// Create attacker
	attacker := world.Create(
		components.Position,
		components.Faction,
		components.Health,
		components.BeamWeapon,
	)
	attackerEntry := world.Entry(attacker)
	components.Position.SetValue(attackerEntry, components.PositionData{X: 100, Y: 100})
	components.Faction.SetValue(attackerEntry, components.FactionData{ID: 0}) // Green faction
	components.Health.SetValue(attackerEntry, components.HealthData{Current: 100, Max: 100})

	// Create friendly target (same faction)
	target := world.Create(
		components.Position,
		components.Faction,
		components.Health,
	)
	targetEntry := world.Entry(target)
	components.Position.SetValue(targetEntry, components.PositionData{X: 150, Y: 100})
	components.Faction.SetValue(targetEntry, components.FactionData{ID: 0}) // Same faction
	components.Health.SetValue(targetEntry, components.HealthData{Current: 100, Max: 100})

	// Set beam weapon to friendly target
	components.BeamWeapon.SetValue(attackerEntry, components.BeamWeaponData{
		TargetEntity:      target,
		Range:             200.0,
		DamagePerTick:     0.1,
		DamageAccumulator: 0.0,
	})

	// Run for 20 ticks
	for i := 0; i < 20; i++ {
		UpdateBeamWeapons(world)
	}

	// Friendly target should take no damage
	health := components.Health.Get(targetEntry)
	if health.Current != 100 {
		t.Errorf("Friendly fire should not occur, expected health 100, got %d", health.Current)
	}

	// Target should be cleared
	beamWeapon := components.BeamWeapon.Get(attackerEntry)
	if world.Valid(beamWeapon.TargetEntity) {
		t.Error("Friendly target should be cleared")
	}
}

// TestBeamWeapon_DeadTargetClearing verifies beam clears target when it dies
func TestBeamWeapon_DeadTargetClearing(t *testing.T) {
	world := donburi.NewWorld()

	// Create attacker
	attacker := world.Create(
		components.Position,
		components.Faction,
		components.Health,
		components.BeamWeapon,
	)
	attackerEntry := world.Entry(attacker)
	components.Position.SetValue(attackerEntry, components.PositionData{X: 100, Y: 100})
	components.Faction.SetValue(attackerEntry, components.FactionData{ID: 0})
	components.Health.SetValue(attackerEntry, components.HealthData{Current: 100, Max: 100})

	// Create target with 1 HP
	target := world.Create(
		components.Position,
		components.Faction,
		components.Health,
	)
	targetEntry := world.Entry(target)
	components.Position.SetValue(targetEntry, components.PositionData{X: 150, Y: 100})
	components.Faction.SetValue(targetEntry, components.FactionData{ID: 1})
	components.Health.SetValue(targetEntry, components.HealthData{Current: 1, Max: 100})

	// Set beam weapon to target (use 0.5 damage/tick to avoid floating-point issues)
	components.BeamWeapon.SetValue(attackerEntry, components.BeamWeaponData{
		TargetEntity:      target,
		Range:             200.0,
		DamagePerTick:     0.5, // Avoid 0.1 floating-point precision issues
		DamageAccumulator: 0.0,
	})

	// Run for 2 ticks - should kill target (1 HP, needs 1 damage)
	for i := 0; i < 2; i++ {
		UpdateBeamWeapons(world)
	}

	// Target should be dead
	health := components.Health.Get(targetEntry)
	if health.Current != 0 {
		t.Errorf("Target should be dead, got health %d", health.Current)
	}

	// Beam should clear dead target on next update
	UpdateBeamWeapons(world)
	beamWeapon := components.BeamWeapon.Get(attackerEntry)
	if world.Valid(beamWeapon.TargetEntity) {
		t.Error("Beam should clear dead target")
	}
}

// TestTrackAttacker_AddsNewAttacker verifies attackers are tracked correctly
func TestTrackAttacker_AddsNewAttacker(t *testing.T) {
	world := donburi.NewWorld()

	// Create target with UnderAttack component
	target := world.Create(
		components.UnderAttack,
	)
	targetEntry := world.Entry(target)
	components.UnderAttack.SetValue(targetEntry, components.UnderAttackData{
		Attackers: []donburi.Entity{},
	})

	// Create attacker
	attacker := world.Create(
		components.Position,
	)

	// Track the attacker
	TrackAttacker(world, targetEntry, attacker)

	// Verify attacker was added
	underAttack := components.UnderAttack.Get(targetEntry)
	if len(underAttack.Attackers) != 1 {
		t.Fatalf("Expected 1 attacker, got %d", len(underAttack.Attackers))
	}
	if underAttack.Attackers[0] != attacker {
		t.Error("Attacker entity mismatch")
	}
}

// TestTrackAttacker_NoDuplicates verifies same attacker isn't added twice
func TestTrackAttacker_NoDuplicates(t *testing.T) {
	world := donburi.NewWorld()

	// Create target
	target := world.Create(
		components.UnderAttack,
	)
	targetEntry := world.Entry(target)
	components.UnderAttack.SetValue(targetEntry, components.UnderAttackData{
		Attackers: []donburi.Entity{},
	})

	// Create attacker
	attacker := world.Create(
		components.Position,
	)

	// Track the attacker multiple times
	TrackAttacker(world, targetEntry, attacker)
	TrackAttacker(world, targetEntry, attacker)
	TrackAttacker(world, targetEntry, attacker)

	// Should only be tracked once
	underAttack := components.UnderAttack.Get(targetEntry)
	if len(underAttack.Attackers) != 1 {
		t.Errorf("Expected 1 attacker (no duplicates), got %d", len(underAttack.Attackers))
	}
}

// TestTrackAttacker_RemovesInvalidAttackers verifies cleanup of destroyed attackers
func TestTrackAttacker_RemovesInvalidAttackers(t *testing.T) {
	world := donburi.NewWorld()

	// Create target
	target := world.Create(
		components.UnderAttack,
	)
	targetEntry := world.Entry(target)
	components.UnderAttack.SetValue(targetEntry, components.UnderAttackData{
		Attackers: []donburi.Entity{},
	})

	// Create two attackers
	attacker1 := world.Create(components.Position)
	attacker2 := world.Create(components.Position)

	// Track both
	TrackAttacker(world, targetEntry, attacker1)
	TrackAttacker(world, targetEntry, attacker2)

	// Verify both tracked
	underAttack := components.UnderAttack.Get(targetEntry)
	if len(underAttack.Attackers) != 2 {
		t.Fatalf("Expected 2 attackers, got %d", len(underAttack.Attackers))
	}

	// Destroy first attacker
	world.Remove(attacker1)

	// Add a new attacker (this should trigger cleanup)
	attacker3 := world.Create(components.Position)
	TrackAttacker(world, targetEntry, attacker3)

	// Should have 2 valid attackers (attacker2 and attacker3)
	underAttack = components.UnderAttack.Get(targetEntry)
	if len(underAttack.Attackers) != 2 {
		t.Errorf("Expected 2 valid attackers after cleanup, got %d", len(underAttack.Attackers))
	}

	// Verify attacker1 is not in list
	for _, a := range underAttack.Attackers {
		if a == attacker1 {
			t.Error("Destroyed attacker should be removed from list")
		}
	}
}

// TestIsInRange_BasicDistance verifies simple distance checking
func TestIsInRange_BasicDistance(t *testing.T) {
	// Within range
	if !IsInRange(0, 0, 100, 0, 150) {
		t.Error("Should be in range (distance 100, range 150)")
	}

	// Out of range
	if IsInRange(0, 0, 200, 0, 150) {
		t.Error("Should be out of range (distance 200, range 150)")
	}

	// Exact range boundary
	if !IsInRange(0, 0, 150, 0, 150) {
		t.Error("Should be in range at exact boundary")
	}
}

// TestIsInRange_ToroidalWrapping verifies range checking across world boundaries
func TestIsInRange_ToroidalWrapping(t *testing.T) {
	// Ships near opposite edges should be close (wrapping)
	// World is 5040x5040
	// Ship 1 at (10, 2500), Ship 2 at (5030, 2500)
	// Wrapped distance should be 20 pixels (not 5020)

	if !IsInRange(10, 2500, 5030, 2500, 50) {
		t.Error("Ships near opposite X edges should be in range due to wrapping")
	}

	// Similar test for Y wrapping
	if !IsInRange(2500, 10, 2500, 5030, 50) {
		t.Error("Ships near opposite Y edges should be in range due to wrapping")
	}
}

// TestGetWrappedDistance_Wrapping verifies wrapped distance calculation
func TestGetWrappedDistance_Wrapping(t *testing.T) {
	// Test X wrapping (World is 5040 wide)
	// Ship 1 at x=10, Ship 2 at x=5030
	// Direct distance: 5020, Wrapped distance: -20 (5030 - 5040 + 10 = 0, so 10 - 5030 = -5020, wrapped = 10 - (5030 - 5040) = 20)
	dx, dy := GetWrappedDistance(10, 2500, 5030, 2500)
	expectedDx := -20.0 // Should wrap to shorter distance
	if math.Abs(dx-expectedDx) > 0.1 {
		t.Errorf("Expected dx=%.1f (wrapped), got %.1f", expectedDx, dx)
	}
	if math.Abs(dy) > 0.1 {
		t.Errorf("Expected dy=0, got %.1f", dy)
	}

	// Test Y wrapping (World is 5040 tall)
	dx, dy = GetWrappedDistance(2500, 10, 2500, 5030)
	expectedDy := -20.0
	if math.Abs(dy-expectedDy) > 0.1 {
		t.Errorf("Expected dy=%.1f (wrapped), got %.1f", expectedDy, dy)
	}
	if math.Abs(dx) > 0.1 {
		t.Errorf("Expected dx=0, got %.1f", dx)
	}
}

// TestSetBeamTarget_SetsTarget verifies beam target assignment
func TestSetBeamTarget_SetsTarget(t *testing.T) {
	world := donburi.NewWorld()

	// Create ship with beam weapon
	ship := world.Create(
		components.BeamWeapon,
	)
	shipEntry := world.Entry(ship)
	var emptyEntity donburi.Entity
	components.BeamWeapon.SetValue(shipEntry, components.BeamWeaponData{
		TargetEntity:      emptyEntity,
		Range:             200.0,
		DamagePerTick:     0.1,
		DamageAccumulator: 0.5, // Non-zero accumulator
	})

	// Create target
	target := world.Create(components.Position)

	// Set beam target
	SetBeamTarget(shipEntry, target)

	// Verify target set and accumulator reset
	beamWeapon := components.BeamWeapon.Get(shipEntry)
	if beamWeapon.TargetEntity != target {
		t.Error("Target not set correctly")
	}
	if beamWeapon.DamageAccumulator != 0.0 {
		t.Errorf("Accumulator should be reset to 0, got %f", beamWeapon.DamageAccumulator)
	}
}

// TestClearBeamTarget_ClearsTarget verifies beam target clearing
func TestClearBeamTarget_ClearsTarget(t *testing.T) {
	world := donburi.NewWorld()

	// Create ship with beam weapon and a target
	ship := world.Create(
		components.BeamWeapon,
	)
	shipEntry := world.Entry(ship)
	target := world.Create(components.Position)

	components.BeamWeapon.SetValue(shipEntry, components.BeamWeaponData{
		TargetEntity:      target,
		Range:             200.0,
		DamagePerTick:     0.1,
		DamageAccumulator: 0.5,
	})

	// Clear target
	ClearBeamTarget(shipEntry)

	// Verify target cleared and accumulator reset
	beamWeapon := components.BeamWeapon.Get(shipEntry)
	if world.Valid(beamWeapon.TargetEntity) {
		t.Error("Target should be cleared")
	}
	if beamWeapon.DamageAccumulator != 0.0 {
		t.Errorf("Accumulator should be reset, got %f", beamWeapon.DamageAccumulator)
	}
}
