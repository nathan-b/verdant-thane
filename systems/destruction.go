package systems

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

// DestroyShip handles the complete ship destruction sequence:
// - Creates explosion at ship position
// - Handles player ship destruction (spectate mode)
// - Updates player score/kills if player made the kill
// - Removes ship entity from world
//
// This function should be called whenever a ship's health reaches 0,
// regardless of damage source (projectiles, beam weapons, etc.)
func DestroyShip(w donburi.World, shipEntry *donburi.Entry, killerEntity donburi.Entity, explosionSprite *ebiten.Image) {
	// Safety check - ensure ship is still valid
	if !w.Valid(shipEntry.Entity()) {
		return
	}

	shipPos := components.Position.Get(shipEntry)
	shipEntity := shipEntry.Entity()

	// Check if this is the player's ship and handle spectate mode
	playerStateQuery := donburi.NewQuery(filter.Contains(components.PlayerState))
	if playerStateEntry, ok := playerStateQuery.First(w); ok {
		state := components.PlayerState.Get(playerStateEntry)
		playerShip := state.ControlledShip

		// If player ship was destroyed, enter spectate mode before removing it
		if shipEntity == playerShip && w.Valid(playerShip) {
			// Get player faction
			faction := components.Faction.Get(shipEntry)

			// Enter spectate mode (will select an allied ship)
			EnterSpectateMode(w, playerStateEntry, faction.ID)
			// After spectate mode, update playerShip reference for kill tracking
			playerShip = state.ControlledShip
		}

		// Update player score and kills if player made the kill
		// Note: killerEntity is the ship that fired the projectile/beam
		if w.Valid(killerEntity) && killerEntity == playerShip {
			state.Kills++
			state.Score += 10 // 10 points per kill
		}
	}

	// Create explosion entity at ship position
	explosion := w.Create(
		components.IsExplosion,
		components.Position,
		components.Explosion,
		components.Sprite,
	)
	explosionEntry := w.Entry(explosion)
	components.Position.SetValue(explosionEntry, components.PositionData{X: shipPos.X, Y: shipPos.Y})
	components.Explosion.SetValue(explosionEntry, components.ExplosionData{
		CurrentFrame: 0,
		FrameTimer:   5, // 5 ticks per frame (~12 FPS)
	})
	components.Sprite.SetValue(explosionEntry, components.SpriteData{Image: explosionSprite})

	// Remove the ship entity
	w.Remove(shipEntity)
}
