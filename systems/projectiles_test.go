package systems

import (
	"testing"

	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

func TestUpdateProjectileLifetime_DecrementsLifetime(t *testing.T) {
	world := donburi.NewWorld()

	projectile := world.Create(components.Projectile)
	entry := world.Entry(projectile)

	components.Projectile.SetValue(entry, components.ProjectileData{
		Lifetime:    10,
		MaxLifetime: 180,
	})

	// Run system once
	UpdateProjectileLifetime(world)

	proj := components.Projectile.Get(entry)
	if proj.Lifetime != 9 {
		t.Errorf("Expected lifetime 9, got %d", proj.Lifetime)
	}
}

func TestUpdateProjectileLifetime_RemovesExpiredProjectiles(t *testing.T) {
	world := donburi.NewWorld()

	projectile := world.Create(components.Projectile)
	entry := world.Entry(projectile)

	// Projectile with lifetime 1
	components.Projectile.SetValue(entry, components.ProjectileData{
		Lifetime:    1,
		MaxLifetime: 180,
	})

	// Store entity ID before removal
	entityID := entry.Entity()

	// Run system once - should decrement to 0 and remove
	UpdateProjectileLifetime(world)

	// Entity should be removed
	if world.Valid(entityID) {
		t.Error("Projectile should have been removed after lifetime reached 0")
	}
}

func TestUpdateProjectileLifetime_RemovesMultipleExpired(t *testing.T) {
	world := donburi.NewWorld()

	// Create 5 projectiles with different lifetimes
	proj1 := world.Create(components.Projectile)
	proj2 := world.Create(components.Projectile)
	proj3 := world.Create(components.Projectile)
	proj4 := world.Create(components.Projectile)
	proj5 := world.Create(components.Projectile)

	components.Projectile.SetValue(world.Entry(proj1), components.ProjectileData{Lifetime: 1, MaxLifetime: 180})
	components.Projectile.SetValue(world.Entry(proj2), components.ProjectileData{Lifetime: 5, MaxLifetime: 180})
	components.Projectile.SetValue(world.Entry(proj3), components.ProjectileData{Lifetime: 1, MaxLifetime: 180})
	components.Projectile.SetValue(world.Entry(proj4), components.ProjectileData{Lifetime: 10, MaxLifetime: 180})
	components.Projectile.SetValue(world.Entry(proj5), components.ProjectileData{Lifetime: 1, MaxLifetime: 180})

	// Run system
	UpdateProjectileLifetime(world)

	// Count remaining projectiles
	query := donburi.NewQuery(filter.Contains(components.Projectile))
	count := 0
	for range query.Iter(world) {
		count++
	}

	// Should have 2 remaining (proj2 and proj4)
	if count != 2 {
		t.Errorf("Expected 2 projectiles remaining, got %d", count)
	}

	// Verify the right ones survived
	if !world.Valid(proj2) {
		t.Error("Projectile 2 (lifetime 5) should still exist")
	}
	if !world.Valid(proj4) {
		t.Error("Projectile 4 (lifetime 10) should still exist")
	}

	// Verify the right ones were removed
	if world.Valid(proj1) {
		t.Error("Projectile 1 should have been removed")
	}
	if world.Valid(proj3) {
		t.Error("Projectile 3 should have been removed")
	}
	if world.Valid(proj5) {
		t.Error("Projectile 5 should have been removed")
	}
}

func TestUpdateProjectileLifetime_KeepsAlive(t *testing.T) {
	world := donburi.NewWorld()

	projectile := world.Create(components.Projectile)
	entry := world.Entry(projectile)

	components.Projectile.SetValue(entry, components.ProjectileData{
		Lifetime:    100,
		MaxLifetime: 180,
	})

	entityID := entry.Entity()

	// Run system multiple times
	for i := 0; i < 50; i++ {
		UpdateProjectileLifetime(world)
	}

	// Projectile should still exist
	if !world.Valid(entityID) {
		t.Error("Projectile should still exist after 50 ticks (started with 100)")
	}

	// Check lifetime decreased correctly
	proj := components.Projectile.Get(entry)
	if proj.Lifetime != 50 {
		t.Errorf("Expected lifetime 50, got %d", proj.Lifetime)
	}
}

func TestUpdateProjectileLifetime_HandlesEmptyWorld(t *testing.T) {
	world := donburi.NewWorld()

	// Run system on empty world - should not panic
	UpdateProjectileLifetime(world)

	// Verify world is still empty
	query := donburi.NewQuery(filter.Contains(components.Projectile))
	count := 0
	for range query.Iter(world) {
		count++
	}

	if count != 0 {
		t.Errorf("Expected 0 projectiles, got %d", count)
	}
}

func TestUpdateProjectileLifetime_NegativeLifetimeRemoved(t *testing.T) {
	world := donburi.NewWorld()

	projectile := world.Create(components.Projectile)
	entry := world.Entry(projectile)

	// Start with 0 lifetime (should be removed on next tick)
	components.Projectile.SetValue(entry, components.ProjectileData{
		Lifetime:    0,
		MaxLifetime: 180,
	})

	entityID := entry.Entity()

	UpdateProjectileLifetime(world)

	// Should be removed (0 - 1 = -1, which is <= 0)
	if world.Valid(entityID) {
		t.Error("Projectile with 0 lifetime should be removed")
	}
}

func TestUpdateProjectileLifetime_FullLifecycle(t *testing.T) {
	world := donburi.NewWorld()

	projectile := world.Create(components.Projectile)
	entry := world.Entry(projectile)

	initialLifetime := 10
	components.Projectile.SetValue(entry, components.ProjectileData{
		Lifetime:    initialLifetime,
		MaxLifetime: 180,
	})

	entityID := entry.Entity()

	// Run for exactly initialLifetime ticks
	for i := 0; i < initialLifetime; i++ {
		UpdateProjectileLifetime(world)

		// Should exist for first (initialLifetime - 1) ticks
		if i < initialLifetime-1 {
			if !world.Valid(entityID) {
				t.Errorf("Projectile removed too early at tick %d", i)
			}
		}
	}

	// After initialLifetime ticks, should be removed
	if world.Valid(entityID) {
		t.Error("Projectile should be removed after lifetime expires")
	}
}
