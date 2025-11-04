package systems

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

const (
	explosionFrameCount = 4 // Total number of explosion animation frames
)

// UpdateExplosions advances explosion animations and removes completed explosions
func UpdateExplosions(w donburi.World) {
	query := donburi.NewQuery(filter.Contains(
		components.IsExplosion,
		components.Explosion,
	))

	explosionsToRemove := []donburi.Entity{}

	for entry := range query.Iter(w) {
		explosion := components.Explosion.Get(entry)

		// Countdown frame timer
		explosion.FrameTimer--

		// When timer reaches zero, advance to next frame
		if explosion.FrameTimer <= 0 {
			explosion.CurrentFrame++

			// Check if animation is complete
			if explosion.CurrentFrame >= explosionFrameCount {
				// Mark for removal
				explosionsToRemove = append(explosionsToRemove, entry.Entity())
			} else {
				// Reset timer for next frame
				explosion.FrameTimer = 5 // 5 ticks per frame
			}
		}
	}

	// Remove completed explosions
	for _, entity := range explosionsToRemove {
		if w.Valid(entity) {
			w.Remove(entity)
		}
	}
}
