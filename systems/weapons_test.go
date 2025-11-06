package systems

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
	"github.com/nathan/verdant-thane/config"
)

func TestUpdateWeapons_ChargesCapacitor(t *testing.T) {
	world := donburi.NewWorld()

	entity := world.Create(components.Weapon)
	entry := world.Entry(entity)

	// Start with empty capacitor
	components.Weapon.SetValue(entry, components.WeaponData{
		Capacitor:  0.0,
		ChargeRate: 1.0 / (config.GetProjectileCharacteristics(config.LaserProjectile).ChargeTime * 60.0),
		FiringCone: 30 * math.Pi / 180,
	})

	// Run weapon system multiple times
	for i := 0; i < 60; i++ { // 1 second worth of ticks
		UpdateWeapons(world)
	}

	weapon := components.Weapon.Get(entry)
	// Should be at least partially charged
	if weapon.Capacitor <= 0.0 {
		t.Error("Capacitor should have charged after 60 ticks")
	}
	if weapon.Capacitor > 1.0 {
		t.Errorf("Capacitor should not exceed 1.0, got %f", weapon.Capacitor)
	}
}

func TestUpdateWeapons_DoesNotOvercharge(t *testing.T) {
	world := donburi.NewWorld()

	entity := world.Create(components.Weapon)
	entry := world.Entry(entity)

	// Start nearly full
	components.Weapon.SetValue(entry, components.WeaponData{
		Capacitor:  0.99,
		ChargeRate: 1.0 / (config.GetProjectileCharacteristics(config.LaserProjectile).ChargeTime * 60.0),
		FiringCone: 30 * math.Pi / 180,
	})

	// Run many times
	for i := 0; i < 1000; i++ {
		UpdateWeapons(world)
	}

	weapon := components.Weapon.Get(entry)
	if weapon.Capacitor > 1.0 {
		t.Errorf("Capacitor overcharged to %f", weapon.Capacitor)
	}
	if weapon.Capacitor != 1.0 {
		t.Errorf("Capacitor should be exactly 1.0, got %f", weapon.Capacitor)
	}
}

func TestFireWeapon_FailsWhenNotCharged(t *testing.T) {
	world := donburi.NewWorld()

	ship := world.Create(
		components.Position,
		components.Rotation,
		components.Faction,
		components.Weapon,
	)
	entry := world.Entry(ship)

	components.Position.SetValue(entry, components.PositionData{X: 100, Y: 100})
	components.Rotation.SetValue(entry, components.RotationData{Angle: 0})
	components.Faction.SetValue(entry, components.FactionData{ID: 0})
	components.Weapon.SetValue(entry, components.WeaponData{
		Capacitor:  0.5, // Not fully charged
		ChargeRate: 1.0 / (config.GetProjectileCharacteristics(config.LaserProjectile).ChargeTime * 60.0),
		FiringCone: 30 * math.Pi / 180,
	})

	// Try to fire
	fired := FireWeapon(world, entry, 150, 50, nil)

	if fired {
		t.Error("Should not be able to fire with capacitor at 0.5")
	}
}

func TestFireWeapon_SucceedsWhenCharged(t *testing.T) {
	world := donburi.NewWorld()

	// Create a dummy sprite (1x1 image)
	dummySprite := ebiten.NewImage(1, 1)

	ship := world.Create(
		components.Position,
		components.Rotation,
		components.Faction,
		components.Weapon,
	)
	entry := world.Entry(ship)

	components.Position.SetValue(entry, components.PositionData{X: 100, Y: 100})
	components.Rotation.SetValue(entry, components.RotationData{Angle: 0})
	components.Faction.SetValue(entry, components.FactionData{ID: 0})
	components.Weapon.SetValue(entry, components.WeaponData{
		Capacitor:  1.0, // Fully charged
		ChargeRate: 1.0 / (config.GetProjectileCharacteristics(config.LaserProjectile).ChargeTime * 60.0),
		FiringCone: 30 * math.Pi / 180,
	})

	// Fire weapon
	fired := FireWeapon(world, entry, 150, 50, dummySprite)

	if !fired {
		t.Error("Should be able to fire with full capacitor")
	}

	// Check capacitor was drained
	weapon := components.Weapon.Get(entry)
	if weapon.Capacitor != 0.0 {
		t.Errorf("Capacitor should be drained to 0.0, got %f", weapon.Capacitor)
	}
}

func TestFireWeapon_CreatesProjectile(t *testing.T) {
	world := donburi.NewWorld()

	dummySprite := ebiten.NewImage(1, 1)

	ship := world.Create(
		components.Position,
		components.Rotation,
		components.Faction,
		components.Weapon,
	)
	entry := world.Entry(ship)

	components.Position.SetValue(entry, components.PositionData{X: 100, Y: 100})
	components.Rotation.SetValue(entry, components.RotationData{Angle: 0})
	components.Faction.SetValue(entry, components.FactionData{ID: 0})
	components.Weapon.SetValue(entry, components.WeaponData{
		Capacitor:  1.0,
		ChargeRate: 1.0 / (config.GetProjectileCharacteristics(config.LaserProjectile).ChargeTime * 60.0),
		FiringCone: 30 * math.Pi / 180,
	})

	// Count projectiles before
	query := donburi.NewQuery(filter.Contains(components.IsProjectile))
	countBefore := 0
	for range query.Iter(world) {
		countBefore++
	}

	// Fire weapon
	FireWeapon(world, entry, 150, 50, dummySprite)

	// Count projectiles after
	countAfter := 0
	for range query.Iter(world) {
		countAfter++
	}

	if countAfter != countBefore+1 {
		t.Errorf("Expected 1 new projectile, got %d before and %d after", countBefore, countAfter)
	}
}

func TestFireWeapon_ProjectileHasCorrectComponents(t *testing.T) {
	world := donburi.NewWorld()

	dummySprite := ebiten.NewImage(1, 1)

	ship := world.Create(
		components.Position,
		components.Rotation,
		components.Faction,
		components.Weapon,
	)
	entry := world.Entry(ship)

	components.Position.SetValue(entry, components.PositionData{X: 100, Y: 200})
	components.Rotation.SetValue(entry, components.RotationData{Angle: 0})
	components.Faction.SetValue(entry, components.FactionData{ID: 1})
	components.Weapon.SetValue(entry, components.WeaponData{
		Capacitor:  1.0,
		ChargeRate: 1.0 / (config.GetProjectileCharacteristics(config.LaserProjectile).ChargeTime * 60.0),
		FiringCone: 30 * math.Pi / 180,
	})

	// Fire weapon
	FireWeapon(world, entry, 150, 50, dummySprite)

	// Find the projectile
	query := donburi.NewQuery(filter.Contains(components.IsProjectile))
	projectileEntry, ok := query.First(world)
	if !ok {
		t.Fatal("No projectile created")
	}

	// Check it has required components
	if !projectileEntry.HasComponent(components.Position) {
		t.Error("Projectile missing Position component")
	}
	if !projectileEntry.HasComponent(components.Velocity) {
		t.Error("Projectile missing Velocity component")
	}
	if !projectileEntry.HasComponent(components.Faction) {
		t.Error("Projectile missing Faction component")
	}
	if !projectileEntry.HasComponent(components.Projectile) {
		t.Error("Projectile missing Projectile component")
	}

	// Check projectile starts at ship position
	projPos := components.Position.Get(projectileEntry)
	if projPos.X != 100 || projPos.Y != 200 {
		t.Errorf("Projectile should start at ship position (100, 200), got (%f, %f)", projPos.X, projPos.Y)
	}

	// Check projectile has same faction
	projFaction := components.Faction.Get(projectileEntry)
	if projFaction.ID != 1 {
		t.Errorf("Projectile should have faction 1, got %d", projFaction.ID)
	}

	// Check projectile has velocity (should be moving)
	projVel := components.Velocity.Get(projectileEntry)
	velocityMagnitude := math.Sqrt(projVel.X*projVel.X + projVel.Y*projVel.Y)
	if velocityMagnitude < 1.0 {
		t.Error("Projectile should have non-zero velocity")
	}
}

func TestFireWeapon_FiringConeConstraint(t *testing.T) {
	world := donburi.NewWorld()

	dummySprite := ebiten.NewImage(1, 1)

	ship := world.Create(
		components.Position,
		components.Rotation,
		components.Faction,
		components.Weapon,
	)
	entry := world.Entry(ship)

	components.Position.SetValue(entry, components.PositionData{X: 100, Y: 100})
	components.Rotation.SetValue(entry, components.RotationData{Angle: 0}) // Facing up
	components.Faction.SetValue(entry, components.FactionData{ID: 0})
	components.Weapon.SetValue(entry, components.WeaponData{
		Capacitor:  1.0,
		ChargeRate: 1.0 / (config.GetProjectileCharacteristics(config.LaserProjectile).ChargeTime * 60.0),
		FiringCone: 30 * math.Pi / 180, // 30 degree cone
	})

	// Try to fire directly behind (should be constrained to cone edge)
	FireWeapon(world, entry, 100, 200, dummySprite) // Target below (behind)

	// Find projectile
	query := donburi.NewQuery(filter.Contains(components.IsProjectile))
	projectileEntry, ok := query.First(world)
	if !ok {
		t.Fatal("No projectile created")
	}

	// Projectile should not be going straight down (180 degrees from facing)
	// It should be constrained to the firing cone
	projVel := components.Velocity.Get(projectileEntry)

	// Calculate angle of projectile velocity
	velAngle := math.Atan2(projVel.X, -projVel.Y)
	angleDiff := math.Abs(NormalizeAngle(velAngle - 0)) // Difference from ship facing (0)

	// Should be within or at cone edge (15 degrees = 30/2)
	maxAllowedAngle := (30 * math.Pi / 180) / 2
	if angleDiff > maxAllowedAngle+0.01 { // Small epsilon for floating point
		t.Errorf("Projectile angle should be constrained to cone (max %.2f rad), got %.2f rad", maxAllowedAngle, angleDiff)
	}
}
