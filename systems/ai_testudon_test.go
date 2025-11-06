package systems

import (
	"testing"

	"github.com/yohamta/donburi"

	"github.com/nathan/verdant-thane/components"
)

// TestSelectTestudonTarget_PrioritizesTestudons verifies Testudons are highest priority targets
func TestSelectTestudonTarget_PrioritizesTestudons(t *testing.T) {
	world := donburi.NewWorld()

	// Create AI Testudon
	aiShip := world.Create(
		components.Position,
		components.Faction,
		components.Ship,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: 1000, Y: 1000})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})
	components.Ship.SetValue(aiEntry, components.ShipData{Class: components.Testudon})

	// Create enemy Fighter (close)
	enemyFighter := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
		components.Ship,
	)
	fighterEntry := world.Entry(enemyFighter)
	components.Position.SetValue(fighterEntry, components.PositionData{X: 1100, Y: 1000}) // 100 pixels away
	components.Faction.SetValue(fighterEntry, components.FactionData{ID: 1})
	components.Ship.SetValue(fighterEntry, components.ShipData{Class: components.Fighter})

	// Create enemy Testudon (far)
	enemyTestudon := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
		components.Ship,
	)
	testudonEntry := world.Entry(enemyTestudon)
	components.Position.SetValue(testudonEntry, components.PositionData{X: 1500, Y: 1000}) // 500 pixels away
	components.Faction.SetValue(testudonEntry, components.FactionData{ID: 1})
	components.Ship.SetValue(testudonEntry, components.ShipData{Class: components.Testudon})

	// Select target - should prioritize distant Testudon over close Fighter
	target := SelectTestudonTarget(world, aiShip)

	if target != enemyTestudon {
		t.Error("Should target enemy Testudon (higher priority) even though it's farther")
	}
}

// TestSelectTestudonTarget_PrioritizesDestroyers verifies Destroyers are higher priority than Fighters
func TestSelectTestudonTarget_PrioritizesDestroyers(t *testing.T) {
	world := donburi.NewWorld()

	// Create AI Testudon
	aiShip := world.Create(
		components.Position,
		components.Faction,
		components.Ship,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: 1000, Y: 1000})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})
	components.Ship.SetValue(aiEntry, components.ShipData{Class: components.Testudon})

	// Create enemy Fighter (close)
	enemyFighter := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
		components.Ship,
	)
	fighterEntry := world.Entry(enemyFighter)
	components.Position.SetValue(fighterEntry, components.PositionData{X: 1100, Y: 1000}) // 100 pixels
	components.Faction.SetValue(fighterEntry, components.FactionData{ID: 1})
	components.Ship.SetValue(fighterEntry, components.ShipData{Class: components.Fighter})

	// Create enemy Destroyer (far)
	enemyDestroyer := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
		components.Ship,
	)
	destroyerEntry := world.Entry(enemyDestroyer)
	components.Position.SetValue(destroyerEntry, components.PositionData{X: 1300, Y: 1000}) // 300 pixels
	components.Faction.SetValue(destroyerEntry, components.FactionData{ID: 1})
	components.Ship.SetValue(destroyerEntry, components.ShipData{Class: components.Destroyer})

	// Select target - should prioritize Destroyer over Fighter
	target := SelectTestudonTarget(world, aiShip)

	if target != enemyDestroyer {
		t.Error("Should target enemy Destroyer (higher priority) over Fighter")
	}
}

// TestSelectTestudonTarget_SamePriorityUsesDistance verifies distance tiebreaker
func TestSelectTestudonTarget_SamePriorityUsesDistance(t *testing.T) {
	world := donburi.NewWorld()

	// Create AI Testudon
	aiShip := world.Create(
		components.Position,
		components.Faction,
		components.Ship,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: 1000, Y: 1000})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})
	components.Ship.SetValue(aiEntry, components.ShipData{Class: components.Testudon})

	// Create two enemy Fighters at different distances
	farFighter := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
		components.Ship,
	)
	farEntry := world.Entry(farFighter)
	components.Position.SetValue(farEntry, components.PositionData{X: 1500, Y: 1000}) // 500 pixels
	components.Faction.SetValue(farEntry, components.FactionData{ID: 1})
	components.Ship.SetValue(farEntry, components.ShipData{Class: components.Fighter})

	closeFighter := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
		components.Ship,
	)
	closeEntry := world.Entry(closeFighter)
	components.Position.SetValue(closeEntry, components.PositionData{X: 1100, Y: 1000}) // 100 pixels
	components.Faction.SetValue(closeEntry, components.FactionData{ID: 1})
	components.Ship.SetValue(closeEntry, components.ShipData{Class: components.Fighter})

	// Select target - should choose closer Fighter
	target := SelectTestudonTarget(world, aiShip)

	if target != closeFighter {
		t.Error("Should target closer Fighter when both have same priority")
	}
}

// TestSelectTestudonTarget_DefensivePriority verifies attackers are highest priority
func TestSelectTestudonTarget_DefensivePriority(t *testing.T) {
	world := donburi.NewWorld()

	// Create AI Testudon with UnderAttack component
	aiShip := world.Create(
		components.Position,
		components.Faction,
		components.Ship,
		components.UnderAttack,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: 1000, Y: 1000})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})
	components.Ship.SetValue(aiEntry, components.ShipData{Class: components.Testudon})

	// Create attacker (Fighter, far away)
	attacker := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
		components.Ship,
	)
	attackerEntry := world.Entry(attacker)
	components.Position.SetValue(attackerEntry, components.PositionData{X: 2000, Y: 1000}) // 1000 pixels away
	components.Faction.SetValue(attackerEntry, components.FactionData{ID: 1})
	components.Ship.SetValue(attackerEntry, components.ShipData{Class: components.Fighter})

	// Track attacker
	components.UnderAttack.SetValue(aiEntry, components.UnderAttackData{
		Attackers: []donburi.Entity{attacker},
	})

	// Create enemy Testudon (close, higher ship priority)
	enemyTestudon := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
		components.Ship,
	)
	testudonEntry := world.Entry(enemyTestudon)
	components.Position.SetValue(testudonEntry, components.PositionData{X: 1100, Y: 1000}) // 100 pixels
	components.Faction.SetValue(testudonEntry, components.FactionData{ID: 1})
	components.Ship.SetValue(testudonEntry, components.ShipData{Class: components.Testudon})

	// Select target - should prioritize attacker (defensive) over closer Testudon
	target := SelectTestudonTarget(world, aiShip)

	if target != attacker {
		t.Error("Should prioritize attacker (defensive behavior) over all other targets")
	}
}

// TestSelectTestudonTarget_IgnoresFriendlies verifies friendly ships are not targeted
func TestSelectTestudonTarget_IgnoresFriendlies(t *testing.T) {
	world := donburi.NewWorld()

	// Create AI Testudon
	aiShip := world.Create(
		components.Position,
		components.Faction,
		components.Ship,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: 1000, Y: 1000})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})
	components.Ship.SetValue(aiEntry, components.ShipData{Class: components.Testudon})

	// Create friendly Testudon (same faction)
	friendlyTestudon := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
		components.Ship,
	)
	friendlyEntry := world.Entry(friendlyTestudon)
	components.Position.SetValue(friendlyEntry, components.PositionData{X: 1100, Y: 1000})
	components.Faction.SetValue(friendlyEntry, components.FactionData{ID: 0}) // Same faction
	components.Ship.SetValue(friendlyEntry, components.ShipData{Class: components.Testudon})

	// Create enemy Fighter (far)
	enemyFighter := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
		components.Ship,
	)
	fighterEntry := world.Entry(enemyFighter)
	components.Position.SetValue(fighterEntry, components.PositionData{X: 2000, Y: 1000})
	components.Faction.SetValue(fighterEntry, components.FactionData{ID: 1})
	components.Ship.SetValue(fighterEntry, components.ShipData{Class: components.Fighter})

	// Select target - should skip friendly and target enemy
	target := SelectTestudonTarget(world, aiShip)

	if target != enemyFighter {
		t.Error("Should ignore friendly ships and target enemy")
	}
}

// TestSelectTestudonTarget_NoEnemies verifies returns invalid entity when no enemies
func TestSelectTestudonTarget_NoEnemies(t *testing.T) {
	world := donburi.NewWorld()

	// Create AI Testudon
	aiShip := world.Create(
		components.Position,
		components.Faction,
		components.Ship,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: 1000, Y: 1000})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})
	components.Ship.SetValue(aiEntry, components.ShipData{Class: components.Testudon})

	// Create only friendly ships
	friendly := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
		components.Ship,
	)
	friendlyEntry := world.Entry(friendly)
	components.Position.SetValue(friendlyEntry, components.PositionData{X: 1100, Y: 1000})
	components.Faction.SetValue(friendlyEntry, components.FactionData{ID: 0})
	components.Ship.SetValue(friendlyEntry, components.ShipData{Class: components.Fighter})

	// Select target - should return invalid entity
	target := SelectTestudonTarget(world, aiShip)

	if world.Valid(target) {
		t.Error("Should return invalid entity when no enemies exist")
	}
}

// TestSelectTestudonTarget_InvalidAttackerCleanup verifies invalid attackers are skipped
func TestSelectTestudonTarget_InvalidAttackerCleanup(t *testing.T) {
	world := donburi.NewWorld()

	// Create AI Testudon with UnderAttack
	aiShip := world.Create(
		components.Position,
		components.Faction,
		components.Ship,
		components.UnderAttack,
	)
	aiEntry := world.Entry(aiShip)
	components.Position.SetValue(aiEntry, components.PositionData{X: 1000, Y: 1000})
	components.Faction.SetValue(aiEntry, components.FactionData{ID: 0})
	components.Ship.SetValue(aiEntry, components.ShipData{Class: components.Testudon})

	// Create attacker then destroy it
	attacker := world.Create(components.Position)
	components.UnderAttack.SetValue(aiEntry, components.UnderAttackData{
		Attackers: []donburi.Entity{attacker},
	})
	world.Remove(attacker) // Attacker is now invalid

	// Create valid enemy
	enemy := world.Create(
		components.IsShip,
		components.Position,
		components.Faction,
		components.Ship,
	)
	enemyEntry := world.Entry(enemy)
	components.Position.SetValue(enemyEntry, components.PositionData{X: 1100, Y: 1000})
	components.Faction.SetValue(enemyEntry, components.FactionData{ID: 1})
	components.Ship.SetValue(enemyEntry, components.ShipData{Class: components.Fighter})

	// Select target - should skip invalid attacker and target normal enemy
	target := SelectTestudonTarget(world, aiShip)

	if target != enemy {
		t.Error("Should skip invalid attacker and fall back to normal targeting")
	}
}
