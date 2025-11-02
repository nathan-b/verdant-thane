package systems

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

const (
	gameWidth  = 5040
	gameHeight = 5040
)

// UpdateMovement applies velocity to position and handles world wrapping
func UpdateMovement(w donburi.World) {
	query := donburi.NewQuery(
		filter.Contains(
			components.Position,
			components.Velocity,
		),
	)

	for entry := range query.Iter(w) {
		pos := components.Position.Get(entry)
		vel := components.Velocity.Get(entry)

		// Apply velocity
		pos.X += vel.X
		pos.Y += vel.Y

		// World wrapping (toroidal topology)
		if pos.X < 0 {
			pos.X += gameWidth
		} else if pos.X >= gameWidth {
			pos.X -= gameWidth
		}
		if pos.Y < 0 {
			pos.Y += gameHeight
		} else if pos.Y >= gameHeight {
			pos.Y -= gameHeight
		}
	}
}
