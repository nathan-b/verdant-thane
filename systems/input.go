package systems

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
	"github.com/nathan/verdant-thane/config"
)

// UpdatePlayerInput handles keyboard input for the player-controlled ship
func UpdatePlayerInput(w donburi.World) {
	query := donburi.NewQuery(
		filter.Contains(
			components.PlayerControlled,
			components.Ship,
			components.Rotation,
			components.Velocity,
		),
	)

	// There should only be one player ship
	playerEntry, ok := query.First(w)
	if !ok {
		return // No player ship exists
	}

	ship := components.Ship.Get(playerEntry)
	rot := components.Rotation.Get(playerEntry)
	vel := components.Velocity.Get(playerEntry)

	// Handle rotation
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		rot.Angle -= config.RotationSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		rot.Angle += config.RotationSpeed
	}

	// Handle acceleration/deceleration
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		ship.Speed += ship.Accel
		if ship.Speed > ship.MaxSpeed {
			ship.Speed = ship.MaxSpeed
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		ship.Speed -= ship.Accel
		if ship.Speed < 0 {
			ship.Speed = 0
		}
	}

	// Update velocity based on rotation and speed
	// Sprite faces upward at angle 0
	vel.X = math.Sin(rot.Angle) * ship.Speed
	vel.Y = -math.Cos(rot.Angle) * ship.Speed
}
