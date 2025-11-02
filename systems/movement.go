package systems

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
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
			pos.X += GameWidth
		} else if pos.X >= GameWidth {
			pos.X -= GameWidth
		}
		if pos.Y < 0 {
			pos.Y += GameHeight
		} else if pos.Y >= GameHeight {
			pos.Y -= GameHeight
		}
	}
}
