package systems

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/nathan/verdant-thane/components"
)

func TestUpdateCollisions_PlayerScoreAndKills(t *testing.T) {
	world := donburi.NewWorld()

	// Create player ship (faction 0)
	playerShip := world.Create(
		components.IsShip,
		components.Position,
		components.Health,
		components.Faction,
	)
	playerEntry := world.Entry(playerShip)
	components.Position.SetValue(playerEntry, components.PositionData{X: 100, Y: 100})
	components.Health.SetValue(playerEntry, components.HealthData{Current: 8, Max: 8})
	components.Faction.SetValue(playerEntry, components.FactionData{ID: 0})

	// Create enemy ship (faction 1)
	enemyShip := world.Create(
		components.IsShip,
		components.Position,
		components.Health,
		components.Faction,
	)
	enemyEntry := world.Entry(enemyShip)
	components.Position.SetValue(enemyEntry, components.PositionData{X: 105, Y: 100})
	components.Health.SetValue(enemyEntry, components.HealthData{Current: 1, Max: 8}) // Only 1 HP left
	components.Faction.SetValue(enemyEntry, components.FactionData{ID: 1})

	// Create player projectile (faction 0) near enemy
	projectile := world.Create(
		components.IsProjectile,
		components.Position,
		components.Faction,
		components.Sprite,
	)
	projEntry := world.Entry(projectile)
	components.Position.SetValue(projEntry, components.PositionData{X: 105, Y: 100}) // Same position as enemy
	components.Faction.SetValue(projEntry, components.FactionData{ID: 0})
	components.Sprite.SetValue(projEntry, components.SpriteData{Image: ebiten.NewImage(1, 1)})

	// Create player state
	playerState := world.Create(components.PlayerState)
	playerStateEntry := world.Entry(playerState)
	components.PlayerState.SetValue(playerStateEntry, components.PlayerStateData{
		ControlledShip: playerShip,
		Score:          0,
		Kills:          0,
	})

	// Run collision system
	dummyExplosionSprite := ebiten.NewImage(400, 70)
	UpdateCollisions(world, dummyExplosionSprite)

	// Check that enemy ship was destroyed
	if world.Valid(enemyShip) {
		t.Error("Enemy ship should have been destroyed")
	}

	// Check that projectile was removed
	if world.Valid(projectile) {
		t.Error("Projectile should have been removed")
	}

	// Check that player score increased by 10
	state := components.PlayerState.Get(playerStateEntry)
	if state.Score != 10 {
		t.Errorf("Expected score 10, got %d", state.Score)
	}

	// Check that player kills increased by 1
	if state.Kills != 1 {
		t.Errorf("Expected kills 1, got %d", state.Kills)
	}
}

func TestUpdateCollisions_NoScoreForFriendlyFire(t *testing.T) {
	world := donburi.NewWorld()

	// Create two ships from same faction
	ship1 := world.Create(
		components.IsShip,
		components.Position,
		components.Health,
		components.Faction,
	)
	ship1Entry := world.Entry(ship1)
	components.Position.SetValue(ship1Entry, components.PositionData{X: 100, Y: 100})
	components.Health.SetValue(ship1Entry, components.HealthData{Current: 1, Max: 8})
	components.Faction.SetValue(ship1Entry, components.FactionData{ID: 0})

	// Create player state
	playerState := world.Create(components.PlayerState)
	playerStateEntry := world.Entry(playerState)
	components.PlayerState.SetValue(playerStateEntry, components.PlayerStateData{
		ControlledShip: ship1,
		Score:          0,
		Kills:          0,
	})

	// Create projectile from same faction
	projectile := world.Create(
		components.IsProjectile,
		components.Position,
		components.Faction,
		components.Sprite,
	)
	projEntry := world.Entry(projectile)
	components.Position.SetValue(projEntry, components.PositionData{X: 100, Y: 100})
	components.Faction.SetValue(projEntry, components.FactionData{ID: 0}) // Same faction!
	components.Sprite.SetValue(projEntry, components.SpriteData{Image: ebiten.NewImage(1, 1)})

	// Run collision system
	dummyExplosionSprite := ebiten.NewImage(400, 70)
	UpdateCollisions(world, dummyExplosionSprite)

	// Friendly fire should not happen - ship should still exist
	if !world.Valid(ship1) {
		t.Error("Ship should not be destroyed by friendly fire")
	}

	// Score should remain 0
	state := components.PlayerState.Get(playerStateEntry)
	if state.Score != 0 {
		t.Errorf("Expected score 0 (no friendly fire), got %d", state.Score)
	}
	if state.Kills != 0 {
		t.Errorf("Expected kills 0 (no friendly fire), got %d", state.Kills)
	}
}

func TestUpdateCollisions_MultipleKills(t *testing.T) {
	world := donburi.NewWorld()

	// Create player state
	playerState := world.Create(components.PlayerState)
	playerStateEntry := world.Entry(playerState)
	components.PlayerState.SetValue(playerStateEntry, components.PlayerStateData{
		Score: 0,
		Kills: 0,
	})

	// Create 3 enemy ships at same location
	for i := 0; i < 3; i++ {
		enemyShip := world.Create(
			components.IsShip,
			components.Position,
			components.Health,
			components.Faction,
		)
		enemyEntry := world.Entry(enemyShip)
		components.Position.SetValue(enemyEntry, components.PositionData{X: 100, Y: 100})
		components.Health.SetValue(enemyEntry, components.HealthData{Current: 1, Max: 8})
		components.Faction.SetValue(enemyEntry, components.FactionData{ID: 1})
	}

	// Create 3 player projectiles at same location
	for i := 0; i < 3; i++ {
		projectile := world.Create(
			components.IsProjectile,
			components.Position,
			components.Faction,
			components.Sprite,
		)
		projEntry := world.Entry(projectile)
		components.Position.SetValue(projEntry, components.PositionData{X: 100, Y: 100})
		components.Faction.SetValue(projEntry, components.FactionData{ID: 0})
		components.Sprite.SetValue(projEntry, components.SpriteData{Image: ebiten.NewImage(1, 1)})
	}

	// Run collision system
	dummyExplosionSprite := ebiten.NewImage(400, 70)
	UpdateCollisions(world, dummyExplosionSprite)

	// Check that player got 3 kills and 30 score
	state := components.PlayerState.Get(playerStateEntry)
	if state.Kills != 3 {
		t.Errorf("Expected kills 3, got %d", state.Kills)
	}
	if state.Score != 30 {
		t.Errorf("Expected score 30 (3 kills * 10), got %d", state.Score)
	}
}
