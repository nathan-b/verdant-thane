package config

import (
	"math"
	"math/rand"
	"testing"
)

// ============================================================================
// Ship Characteristics Tests
// ============================================================================

func TestGetShipCharacteristicsFighter(t *testing.T) {
	chars := GetShipCharacteristics(ClassFighter)

	if chars.MaxSpeed != 6.0 {
		t.Errorf("Fighter MaxSpeed: expected 6.0, got %f", chars.MaxSpeed)
	}
	if chars.MaxShield != 8 {
		t.Errorf("Fighter MaxShield: expected 8, got %d", chars.MaxShield)
	}
	if chars.CollisionRadius != 14.0 {
		t.Errorf("Fighter CollisionRadius: expected 14.0, got %f", chars.CollisionRadius)
	}
	if chars.BaseSpritePath != "assets/fighter.png" {
		t.Errorf("Fighter sprite path: expected 'assets/fighter.png', got '%s'", chars.BaseSpritePath)
	}
}

func TestGetShipCharacteristicsDestroyer(t *testing.T) {
	chars := GetShipCharacteristics(ClassDestroyer)

	if chars.MaxSpeed != 5.0 {
		t.Errorf("Destroyer MaxSpeed: expected 5.0, got %f", chars.MaxSpeed)
	}
	if chars.MaxShield != 32 {
		t.Errorf("Destroyer MaxShield: expected 32, got %d", chars.MaxShield)
	}
	if chars.CollisionRadius != 30.0 {
		t.Errorf("Destroyer CollisionRadius: expected 30.0, got %f", chars.CollisionRadius)
	}
	if chars.BaseSpritePath != "assets/destroyer.png" {
		t.Errorf("Destroyer sprite path: expected 'assets/destroyer.png', got '%s'", chars.BaseSpritePath)
	}
}

func TestGetShipCharacteristicsTestudon(t *testing.T) {
	chars := GetShipCharacteristics(ClassTestudon)

	if chars.MaxSpeed != 3.0 {
		t.Errorf("Testudon MaxSpeed: expected 3.0, got %f", chars.MaxSpeed)
	}
	if chars.MaxShield != 100 {
		t.Errorf("Testudon MaxShield: expected 100, got %d", chars.MaxShield)
	}
	if chars.CollisionRadius != 50.0 {
		t.Errorf("Testudon CollisionRadius: expected 50.0, got %f", chars.CollisionRadius)
	}
	if len(chars.Weapons) != 0 {
		t.Errorf("Testudon Weapons: expected 0 (beam weapon only), got %d", len(chars.Weapons))
	}
	if chars.BaseSpritePath != "assets/testudon.png" {
		t.Errorf("Testudon sprite path: expected 'assets/testudon.png', got '%s'", chars.BaseSpritePath)
	}
}

func TestShipBalancing(t *testing.T) {
	fighter := GetShipCharacteristics(ClassFighter)
	destroyer := GetShipCharacteristics(ClassDestroyer)
	testudon := GetShipCharacteristics(ClassTestudon)

	// Speed progression: Fighter > Destroyer > Testudon
	if !(fighter.MaxSpeed > destroyer.MaxSpeed && destroyer.MaxSpeed > testudon.MaxSpeed) {
		t.Error("Ship speed should decrease: Fighter > Destroyer > Testudon")
	}

	// Armor progression: Testudon > Destroyer > Fighter
	if !(testudon.MaxShield > destroyer.MaxShield && destroyer.MaxShield > fighter.MaxShield) {
		t.Error("Ship armor should increase: Fighter < Destroyer < Testudon")
	}

	// Collision radius progression
	if !(testudon.CollisionRadius > destroyer.CollisionRadius && destroyer.CollisionRadius > fighter.CollisionRadius) {
		t.Error("Ship size should increase: Fighter < Destroyer < Testudon")
	}
}

func TestGetShipCharacteristicsInvalidClass(t *testing.T) {
	// Should return fighter as fallback
	chars := GetShipCharacteristics(ShipClass(999))

	if chars.BaseSpritePath != "assets/fighter.png" {
		t.Error("Invalid ship class should fallback to fighter")
	}
}

// ============================================================================
// Projectile Characteristics Tests
// ============================================================================

func TestGetProjectileCharacteristicsLaser(t *testing.T) {
	chars := GetProjectileCharacteristics(LaserProjectile)

	if chars.Type != LaserProjectile {
		t.Errorf("Laser Type: expected %d, got %d", LaserProjectile, chars.Type)
	}
	if chars.Speed != 12.0 {
		t.Errorf("Laser Speed: expected 12.0, got %f", chars.Speed)
	}
	if chars.Acceleration != 0.0 {
		t.Error("Laser should have no acceleration")
	}
	if chars.TurnRate != 0.0 {
		t.Error("Laser should have no turn rate")
	}
	if chars.Lifetime != 180 {
		t.Errorf("Laser Lifetime: expected 180, got %d", chars.Lifetime)
	}
	if chars.ChargeTime != 0.6 {
		t.Errorf("Laser ChargeTime: expected 0.6, got %f", chars.ChargeTime)
	}
	if chars.SpritePath != "assets/laser.png" {
		t.Errorf("Laser sprite path: expected 'assets/laser.png', got '%s'", chars.SpritePath)
	}
}

func TestGetProjectileCharacteristicsMissile(t *testing.T) {
	chars := GetProjectileCharacteristics(MissileProjectile)

	if chars.Type != MissileProjectile {
		t.Errorf("Missile Type: expected %d, got %d", MissileProjectile, chars.Type)
	}
	if chars.Speed != 4.0 {
		t.Errorf("Missile Speed: expected 4.0, got %f", chars.Speed)
	}
	if chars.Acceleration != 0.15 {
		t.Errorf("Missile Acceleration: expected 0.15, got %f", chars.Acceleration)
	}
	if chars.TurnRate <= 0 {
		t.Error("Missile should have turn rate")
	}
	if chars.Lifetime != 150 {
		t.Errorf("Missile Lifetime: expected 150, got %d", chars.Lifetime)
	}
	if chars.ChargeTime != 1.2 {
		t.Errorf("Missile ChargeTime: expected 1.2, got %f", chars.ChargeTime)
	}
	if chars.SpritePath != "assets/missile.png" {
		t.Errorf("Missile sprite path: expected 'assets/missile.png', got '%s'", chars.SpritePath)
	}
}

func TestProjectileBalancing(t *testing.T) {
	laser := GetProjectileCharacteristics(LaserProjectile)
	missile := GetProjectileCharacteristics(MissileProjectile)

	// Laser should be faster
	if laser.Speed <= missile.Speed {
		t.Error("Laser should be faster than missile")
	}

	// Missile should have longer charge time (lower fire rate)
	if missile.ChargeTime <= laser.ChargeTime {
		t.Error("Missile should have longer charge time than laser")
	}

	// Missile should have tracking (acceleration and turn rate)
	if missile.Acceleration == 0 || missile.TurnRate == 0 {
		t.Error("Missile should have tracking capabilities")
	}

	// Laser should not track
	if laser.Acceleration != 0 || laser.TurnRate != 0 {
		t.Error("Laser should not have tracking capabilities")
	}
}

func TestGetProjectileCharacteristicsInvalidType(t *testing.T) {
	// Should return laser as fallback
	chars := GetProjectileCharacteristics(ProjectileType(999))

	if chars.Type != LaserProjectile {
		t.Error("Invalid projectile type should fallback to laser")
	}
}

// ============================================================================
// Constants Tests
// ============================================================================

func TestScreenDimensions(t *testing.T) {
	if ScreenWidth != 1024 {
		t.Errorf("ScreenWidth: expected 1024, got %d", ScreenWidth)
	}
	if ScreenHeight != 768 {
		t.Errorf("ScreenHeight: expected 768, got %d", ScreenHeight)
	}
}

func TestGameWorldDimensions(t *testing.T) {
	if GameWidth != 5040 {
		t.Errorf("GameWidth: expected 5040, got %d", GameWidth)
	}
	if GameHeight != 5040 {
		t.Errorf("GameHeight: expected 5040, got %d", GameHeight)
	}

	// World should be square
	if GameWidth != GameHeight {
		t.Error("Game world should be square")
	}
}

func TestMinimapConfiguration(t *testing.T) {
	if MinimapSize != 205 {
		t.Errorf("MinimapSize: expected 205, got %d", MinimapSize)
	}

	// Minimap should be positioned in bottom-right corner
	expectedX := ScreenWidth - MinimapSize - 10
	if MinimapX != expectedX {
		t.Errorf("MinimapX: expected %d, got %d", expectedX, MinimapX)
	}

	expectedY := ScreenHeight - MinimapSize - 5
	if MinimapY != expectedY {
		t.Errorf("MinimapY: expected %d, got %d", expectedY, MinimapY)
	}

	// Minimap should fit on screen
	if MinimapX < 0 || MinimapY < 0 {
		t.Error("Minimap should be fully on screen")
	}
	if MinimapX+MinimapSize > ScreenWidth || MinimapY+MinimapSize > ScreenHeight {
		t.Error("Minimap should not extend beyond screen bounds")
	}
}

func TestCollisionConfiguration(t *testing.T) {
	if CollisionGridSize != 128 {
		t.Errorf("CollisionGridSize: expected 128, got %d", CollisionGridSize)
	}
	if ShipCollisionRadius != 16.0 {
		t.Errorf("ShipCollisionRadius: expected 16.0, got %f", ShipCollisionRadius)
	}
	if ProjectileCollisionRadius != 2.0 {
		t.Errorf("ProjectileCollisionRadius: expected 2.0, got %f", ProjectileCollisionRadius)
	}
}

func TestRotationSpeed(t *testing.T) {
	expectedRadians := 3.0 * math.Pi / 180.0

	if math.Abs(RotationSpeed-expectedRadians) > 0.0001 {
		t.Errorf("RotationSpeed: expected %f radians, got %f", expectedRadians, RotationSpeed)
	}
}

func TestAIParameters(t *testing.T) {
	if AIDecisionInterval != 60 {
		t.Errorf("AIDecisionInterval: expected 60, got %d", AIDecisionInterval)
	}
	if AIRetargetInterval != 60 {
		t.Errorf("AIRetargetInterval: expected 60, got %d", AIRetargetInterval)
	}

	expectedAIRotation := 3.0 * math.Pi / 180.0
	if math.Abs(AIRotationSpeed-expectedAIRotation) > 0.0001 {
		t.Errorf("AIRotationSpeed: expected %f, got %f", expectedAIRotation, AIRotationSpeed)
	}

	if AIPursuitSpeedMin != 0.80 {
		t.Errorf("AIPursuitSpeedMin: expected 0.80, got %f", AIPursuitSpeedMin)
	}
	if AIPursuitSpeedMax != 1.00 {
		t.Errorf("AIPursuitSpeedMax: expected 1.00, got %f", AIPursuitSpeedMax)
	}
	if AIPatrolSpeed != 0.50 {
		t.Errorf("AIPatrolSpeed: expected 0.50, got %f", AIPatrolSpeed)
	}
}

func TestExplosionConfiguration(t *testing.T) {
	if ExplosionFrameCount != 4 {
		t.Errorf("ExplosionFrameCount: expected 4, got %d", ExplosionFrameCount)
	}
}

func TestStarfieldConfiguration(t *testing.T) {
	if StarDensity != 0.0003 {
		t.Errorf("StarDensity: expected 0.0003, got %f", StarDensity)
	}
	if StarGridSize != 200 {
		t.Errorf("StarGridSize: expected 200, got %d", StarGridSize)
	}
}

// ============================================================================
// Fleet Balance Tests
// ============================================================================

func TestFleetBalanceCosts(t *testing.T) {
	if DestroyerCost != 4 {
		t.Errorf("DestroyerCost: expected 4, got %d", DestroyerCost)
	}
	if TestudonCost != 8 {
		t.Errorf("TestudonCost: expected 8, got %d", TestudonCost)
	}

	// Testudon should cost more than destroyer
	if TestudonCost <= DestroyerCost {
		t.Error("Testudon should cost more than destroyer")
	}
}

func TestFleetPurchaseProbabilities(t *testing.T) {
	if TestudonPurchaseProbability != 0.33 {
		t.Errorf("TestudonPurchaseProbability: expected 0.33, got %f", TestudonPurchaseProbability)
	}
	if DestroyerPurchaseProbability != 0.50 {
		t.Errorf("DestroyerPurchaseProbability: expected 0.50, got %f", DestroyerPurchaseProbability)
	}

	// Probabilities should be in valid range
	if TestudonPurchaseProbability < 0 || TestudonPurchaseProbability > 1 {
		t.Error("TestudonPurchaseProbability should be between 0 and 1")
	}
	if DestroyerPurchaseProbability < 0 || DestroyerPurchaseProbability > 1 {
		t.Error("DestroyerPurchaseProbability should be between 0 and 1")
	}
}

func TestFactionCompositionTotal(t *testing.T) {
	comp := FactionComposition{
		Fighters:   10,
		Destroyers: 3,
		Testudons:  2,
	}

	if comp.Total() != 15 {
		t.Errorf("Total: expected 15, got %d", comp.Total())
	}
}

func TestFactionCompositionTotalZero(t *testing.T) {
	comp := FactionComposition{
		Fighters:   0,
		Destroyers: 0,
		Testudons:  0,
	}

	if comp.Total() != 0 {
		t.Errorf("Total: expected 0, got %d", comp.Total())
	}
}

func TestGetFleetCompositionOnlyFighters(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	comp := GetFleetComposition(rng, 10, 0, 0) // No destroyers or testudons allowed

	if comp.Fighters != 10 {
		t.Errorf("Expected 10 fighters, got %d", comp.Fighters)
	}
	if comp.Destroyers != 0 {
		t.Errorf("Expected 0 destroyers, got %d", comp.Destroyers)
	}
	if comp.Testudons != 0 {
		t.Errorf("Expected 0 testudons, got %d", comp.Testudons)
	}
}

func TestGetFleetCompositionRespectsBudget(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	budget := 20
	comp := GetFleetComposition(rng, budget, 999, 999)

	// Calculate actual cost
	actualCost := comp.Fighters + (comp.Destroyers * DestroyerCost) + (comp.Testudons * TestudonCost)

	if actualCost != budget {
		t.Errorf("Fleet should use full budget. Budget=%d, Cost=%d, Composition=%+v",
			budget, actualCost, comp)
	}
}

func TestGetFleetCompositionRespectsMaxLimits(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	comp := GetFleetComposition(rng, 100, 2, 1) // Limited destroyers and testudons

	if comp.Destroyers > 2 {
		t.Errorf("Should not exceed max destroyers: got %d, max 2", comp.Destroyers)
	}
	if comp.Testudons > 1 {
		t.Errorf("Should not exceed max testudons: got %d, max 1", comp.Testudons)
	}
}

func TestGetFleetCompositionDeterminism(t *testing.T) {
	// Same seed should produce same composition
	rng1 := rand.New(rand.NewSource(123))
	rng2 := rand.New(rand.NewSource(123))

	comp1 := GetFleetComposition(rng1, 20, 10, 10)
	comp2 := GetFleetComposition(rng2, 20, 10, 10)

	if comp1.Fighters != comp2.Fighters || comp1.Destroyers != comp2.Destroyers || comp1.Testudons != comp2.Testudons {
		t.Error("Same seed should produce same composition")
	}
}

func TestGetFleetCompositionVariation(t *testing.T) {
	// Different seeds should (usually) produce different compositions
	rng1 := rand.New(rand.NewSource(42))
	rng2 := rand.New(rand.NewSource(123))

	comp1 := GetFleetComposition(rng1, 20, 10, 10)
	comp2 := GetFleetComposition(rng2, 20, 10, 10)

	// At least one value should differ (not guaranteed, but highly likely with these budgets)
	if comp1.Fighters == comp2.Fighters && comp1.Destroyers == comp2.Destroyers && comp1.Testudons == comp2.Testudons {
		t.Log("Warning: Different seeds produced identical compositions (unlikely but possible)")
	}
}

func TestGenerateFleetConfigBasic(t *testing.T) {
	config := GenerateFleetConfig(2, 10)

	if config.NumFactions != 2 {
		t.Errorf("NumFactions: expected 2, got %d", config.NumFactions)
	}
	if len(config.Compositions) != 2 {
		t.Errorf("Compositions length: expected 2, got %d", len(config.Compositions))
	}

	for i, comp := range config.Compositions {
		if comp.Fighters != 10 {
			t.Errorf("Faction %d: expected 10 fighters, got %d", i, comp.Fighters)
		}
		if comp.Destroyers != 0 || comp.Testudons != 0 {
			t.Errorf("Faction %d: should only have fighters", i)
		}
	}
}

func TestGenerateFleetConfigBoundaries(t *testing.T) {
	// Test minimum factions
	config := GenerateFleetConfig(1, 10) // Should clamp to 2
	if config.NumFactions != 2 {
		t.Errorf("Should clamp to minimum 2 factions, got %d", config.NumFactions)
	}

	// Test maximum factions
	config = GenerateFleetConfig(10, 10) // Should clamp to 4
	if config.NumFactions != 4 {
		t.Errorf("Should clamp to maximum 4 factions, got %d", config.NumFactions)
	}

	// Test minimum ships
	config = GenerateFleetConfig(2, 0) // Should clamp to 1
	if config.Compositions[0].Fighters != 1 {
		t.Errorf("Should clamp to minimum 1 ship per faction, got %d", config.Compositions[0].Fighters)
	}
}

func TestGenerateRandomFleetConfigDeterminism(t *testing.T) {
	// Same seed should produce same configuration
	config1 := GenerateRandomFleetConfig(12345)
	config2 := GenerateRandomFleetConfig(12345)

	if config1.NumFactions != config2.NumFactions {
		t.Error("Same seed should produce same number of factions")
	}

	for i := range config1.Compositions {
		if config1.Compositions[i].Fighters != config2.Compositions[i].Fighters {
			t.Errorf("Faction %d fighters mismatch", i)
		}
		if config1.Compositions[i].Destroyers != config2.Compositions[i].Destroyers {
			t.Errorf("Faction %d destroyers mismatch", i)
		}
		if config1.Compositions[i].Testudons != config2.Compositions[i].Testudons {
			t.Errorf("Faction %d testudons mismatch", i)
		}
	}
}

func TestGenerateRandomFleetConfigVariation(t *testing.T) {
	// Different seeds should produce different configurations
	config1 := GenerateRandomFleetConfig(42)
	config2 := GenerateRandomFleetConfig(123)

	// At least something should be different
	identical := config1.NumFactions == config2.NumFactions
	if identical && len(config1.Compositions) == len(config2.Compositions) {
		for i := range config1.Compositions {
			if config1.Compositions[i].Fighters != config2.Compositions[i].Fighters ||
				config1.Compositions[i].Destroyers != config2.Compositions[i].Destroyers ||
				config1.Compositions[i].Testudons != config2.Compositions[i].Testudons {
				identical = false
				break
			}
		}
	} else {
		identical = false
	}

	if identical {
		t.Log("Warning: Different seeds produced identical configurations (unlikely but possible)")
	}
}

func TestGenerateRandomFleetConfigPlayerFactionHasControllableShips(t *testing.T) {
	// Test many random configurations to ensure player always has controllable ships
	for seed := int64(0); seed < 100; seed++ {
		config := GenerateRandomFleetConfig(seed)

		playerComp := config.Compositions[0]

		// Player faction (0) must have fighters or destroyers (testudons are AI-only)
		if playerComp.Fighters == 0 && playerComp.Destroyers == 0 {
			t.Errorf("Seed %d: Player faction has no controllable ships: %+v", seed, playerComp)
		}
	}
}

func TestGenerateRandomFleetConfigFactionCount(t *testing.T) {
	// Verify faction count is in valid range (2-4)
	for seed := int64(0); seed < 100; seed++ {
		config := GenerateRandomFleetConfig(seed)

		if config.NumFactions < 2 || config.NumFactions > 4 {
			t.Errorf("Seed %d: Invalid faction count %d (should be 2-4)", seed, config.NumFactions)
		}

		if len(config.Compositions) != config.NumFactions {
			t.Errorf("Seed %d: Compositions length mismatch: %d vs %d",
				seed, len(config.Compositions), config.NumFactions)
		}
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestShipAndProjectileConsistency(t *testing.T) {
	fighter := GetShipCharacteristics(ClassFighter)
	destroyer := GetShipCharacteristics(ClassDestroyer)
	laser := GetProjectileCharacteristics(LaserProjectile)
	missile := GetProjectileCharacteristics(MissileProjectile)

	// Laser should be faster than any ship
	if laser.Speed <= fighter.MaxSpeed {
		t.Error("Laser should be faster than fighter")
	}

	// Missile charge time should match destroyer's weapon system
	// (approximately - some variance expected)
	if math.Abs(missile.ChargeTime-laser.ChargeTime) > 1.0 {
		t.Log("Missile and laser have very different charge times")
	}

	// Fighter and destroyer should have same firing cone for main gun
	if len(fighter.Weapons) > 0 && len(destroyer.Weapons) > 0 {
		if fighter.Weapons[0].FiringCone != destroyer.Weapons[0].FiringCone {
			t.Error("Fighter and destroyer should have same firing cone for main gun")
		}
	}
}

func TestConfigurationBalance(t *testing.T) {
	// Verify game balance makes sense

	// World should be significantly larger than screen
	if GameWidth <= ScreenWidth*2 || GameHeight <= ScreenHeight*2 {
		t.Error("Game world should be much larger than screen")
	}

	// Collision grid should be smaller than world
	if CollisionGridSize >= GameWidth || CollisionGridSize >= GameHeight {
		t.Error("Collision grid cells should be smaller than world")
	}

	// Ship collision radius should be larger than projectile
	if ShipCollisionRadius <= ProjectileCollisionRadius {
		t.Error("Ships should have larger collision radius than projectiles")
	}
}
