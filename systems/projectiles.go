package systems

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

// UpdateProjectileLifetime decrements projectile lifetimes and removes expired projectiles
func UpdateProjectileLifetime(w donburi.World) {
	query := donburi.NewQuery(filter.Contains(components.Projectile))

	toRemove := []donburi.Entity{}

	for entry := range query.Iter(w) {
		proj := components.Projectile.Get(entry)
		proj.Lifetime--

		if proj.Lifetime <= 0 {
			toRemove = append(toRemove, entry.Entity())
		}
	}

	// Remove expired projectiles
	for _, entity := range toRemove {
		w.Remove(entity)
	}
}
