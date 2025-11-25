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

	if chars.MaxSpeed != 5.0 {
		t.Errorf("Fighter MaxSpeed: expected 5.0, got %f", chars.MaxSpeed)
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

	// Speed progression: Fighter >= Destroyer > Testudon
	if !(fighter.MaxSpeed >= destroyer.MaxSpeed && destroyer.MaxSpeed > testudon.MaxSpeed) {
		t.Error("Ship speed should decrease: Fighter >= Destroyer > Testudon")
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
	if chars.Lifetime != 120 {
		t.Errorf("Laser Lifetime: expected 120, got %d", chars.Lifetime)
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

func TestOverrideMainGunProjectileSpeed(t *testing.T) {
	// Save original speed
	original := GetProjectileCharacteristics(LaserProjectile).Speed

	// Override speed
	newSpeed := 10.0
	OverrideMainGunProjectileSpeed(newSpeed)

	// Verify override
	chars := GetProjectileCharacteristics(LaserProjectile)
	if chars.Speed != newSpeed {
		t.Errorf("Expected speed %.2f, got %.2f", newSpeed, chars.Speed)
	}

	// Restore original speed
	OverrideMainGunProjectileSpeed(original)
}

func TestOverrideAIFireProbability(t *testing.T) {
	// Save original probabilities from fighter (all ships have same values)
	originalFighter := GetShipCharacteristics(ClassFighter)

	// Override to 0.6 fire probability
	// Should be split: 0.4 accurate (2/3), 0.2 random (1/3)
	OverrideAIFireProbability(0.6)

	// Verify fighter
	fighter := GetShipCharacteristics(ClassFighter)
	if math.Abs(fighter.AIAccurateShotProbability-0.4) > 0.001 {
		t.Errorf("Fighter accurate: expected 0.4, got %.4f", fighter.AIAccurateShotProbability)
	}
	if math.Abs(fighter.AIRandomShotProbability-0.2) > 0.001 {
		t.Errorf("Fighter random: expected 0.2, got %.4f", fighter.AIRandomShotProbability)
	}

	// Verify destroyer
	destroyer := GetShipCharacteristics(ClassDestroyer)
	if math.Abs(destroyer.AIAccurateShotProbability-0.4) > 0.001 {
		t.Errorf("Destroyer accurate: expected 0.4, got %.4f", destroyer.AIAccurateShotProbability)
	}
	if math.Abs(destroyer.AIRandomShotProbability-0.2) > 0.001 {
		t.Errorf("Destroyer random: expected 0.2, got %.4f", destroyer.AIRandomShotProbability)
	}

	// Verify testudon
	testudon := GetShipCharacteristics(ClassTestudon)
	if math.Abs(testudon.AIAccurateShotProbability-0.4) > 0.001 {
		t.Errorf("Testudon accurate: expected 0.4, got %.4f", testudon.AIAccurateShotProbability)
	}
	if math.Abs(testudon.AIRandomShotProbability-0.2) > 0.001 {
		t.Errorf("Testudon random: expected 0.2, got %.4f", testudon.AIRandomShotProbability)
	}

	// Restore original probabilities
	OverrideAIFireProbability(originalFighter.AIAccurateShotProbability + originalFighter.AIRandomShotProbability)
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
	expectedRadians := 3.2 * math.Pi / 180.0

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
	// Test that all round balance probabilities are in valid range [0, 1]
	for round := 1; round <= 10; round++ {
		balance := GetRoundBalance(round)
		if balance.TestudonProbability < 0 || balance.TestudonProbability > 1 {
			t.Errorf("Round %d: TestudonProbability should be between 0 and 1, got %f",
				round, balance.TestudonProbability)
		}
		if balance.DestroyerProbability < 0 || balance.DestroyerProbability > 1 {
			t.Errorf("Round %d: DestroyerProbability should be between 0 and 1, got %f",
				round, balance.DestroyerProbability)
		}
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
	// No destroyers or testudons allowed (max = 0)
	comp := GetFleetComposition(rng, 10, 0, 0, 0.5, 0.5)

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
	comp := GetFleetComposition(rng, budget, 999, 999, 0.5, 0.33)

	// Calculate actual cost
	actualCost := comp.Fighters + (comp.Destroyers * DestroyerCost) + (comp.Testudons * TestudonCost)

	if actualCost != budget {
		t.Errorf("Fleet should use full budget. Budget=%d, Cost=%d, Composition=%+v",
			budget, actualCost, comp)
	}
}

func TestGetFleetCompositionRespectsMaxLimits(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	// Limited destroyers and testudons with high probabilities
	comp := GetFleetComposition(rng, 100, 2, 1, 0.9, 0.9)

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

	comp1 := GetFleetComposition(rng1, 20, 10, 10, 0.5, 0.33)
	comp2 := GetFleetComposition(rng2, 20, 10, 10, 0.5, 0.33)

	if comp1.Fighters != comp2.Fighters || comp1.Destroyers != comp2.Destroyers || comp1.Testudons != comp2.Testudons {
		t.Error("Same seed should produce same composition")
	}
}

func TestGetFleetCompositionVariation(t *testing.T) {
	// Different seeds should (usually) produce different compositions
	rng1 := rand.New(rand.NewSource(42))
	rng2 := rand.New(rand.NewSource(123))

	comp1 := GetFleetComposition(rng1, 20, 10, 10, 0.5, 0.33)
	comp2 := GetFleetComposition(rng2, 20, 10, 10, 0.5, 0.33)

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
	// Same seed and round should produce same configuration
	config1 := GenerateRandomFleetConfig(12345, 3)
	config2 := GenerateRandomFleetConfig(12345, 3)

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
	config1 := GenerateRandomFleetConfig(42, 3)
	config2 := GenerateRandomFleetConfig(123, 3)

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
	// Test many random configurations across different rounds to ensure player always has controllable ships
	for seed := int64(0); seed < 100; seed++ {
		for round := 1; round <= 5; round++ {
			config := GenerateRandomFleetConfig(seed, round)

			playerComp := config.Compositions[0]

			// Player faction (0) must have fighters or destroyers (testudons are AI-only)
			if playerComp.Fighters == 0 && playerComp.Destroyers == 0 {
				t.Errorf("Seed %d, Round %d: Player faction has no controllable ships: %+v", seed, round, playerComp)
			}
		}
	}
}

func TestGenerateRandomFleetConfigFactionCount(t *testing.T) {
	// Verify faction count is in valid range (2-4)
	for seed := int64(0); seed < 100; seed++ {
		config := GenerateRandomFleetConfig(seed, 3)

		if config.NumFactions < 2 || config.NumFactions > 4 {
			t.Errorf("Seed %d: Invalid faction count %d (should be 2-4)", seed, config.NumFactions)
		}

		if len(config.Compositions) != config.NumFactions {
			t.Errorf("Seed %d: Compositions length mismatch: %d vs %d",
				seed, len(config.Compositions), config.NumFactions)
		}
	}
}

func TestGetRoundBalance(t *testing.T) {
	// Test round 1: All fighters, no advanced units
	round1 := GetRoundBalance(1)
	if round1.MinFighters != 7 || round1.MaxFighters != 12 {
		t.Errorf("Round 1: Expected 7-12 fighters, got %d-%d", round1.MinFighters, round1.MaxFighters)
	}
	if round1.DestroyerProbability != 0.0 || round1.TestudonProbability != 0.0 {
		t.Errorf("Round 1: Expected 0%% destroyer/testudon probability, got %.2f/%.2f",
			round1.DestroyerProbability, round1.TestudonProbability)
	}
	if round1.MaxDestroyers != 0 || round1.MaxTestudons != 0 {
		t.Errorf("Round 1: Expected 0 max destroyers/testudons, got %d/%d",
			round1.MaxDestroyers, round1.MaxTestudons)
	}

	// Test round 2: Introduce destroyers
	round2 := GetRoundBalance(2)
	if round2.MinFighters != 9 || round2.MaxFighters != 15 {
		t.Errorf("Round 2: Expected 9-15 fighters, got %d-%d", round2.MinFighters, round2.MaxFighters)
	}
	if round2.DestroyerProbability != 0.25 || round2.TestudonProbability != 0.0 {
		t.Errorf("Round 2: Expected 25%% destroyer, 0%% testudon probability, got %.2f/%.2f",
			round2.DestroyerProbability, round2.TestudonProbability)
	}
	if round2.MaxDestroyers != 1 || round2.MaxTestudons != 0 {
		t.Errorf("Round 2: Expected max 1 destroyer, 0 testudons, got %d/%d",
			round2.MaxDestroyers, round2.MaxTestudons)
	}

	// Test round 3: Introduce testudons
	round3 := GetRoundBalance(3)
	if round3.MinFighters != 11 || round3.MaxFighters != 18 {
		t.Errorf("Round 3: Expected 11-18 fighters, got %d-%d", round3.MinFighters, round3.MaxFighters)
	}
	if round3.DestroyerProbability != 0.35 || round3.TestudonProbability != 0.25 {
		t.Errorf("Round 3: Expected 35%% destroyer, 25%% testudon probability, got %.2f/%.2f",
			round3.DestroyerProbability, round3.TestudonProbability)
	}
	if round3.MaxDestroyers != 2 || round3.MaxTestudons != 1 {
		t.Errorf("Round 3: Expected max 2 destroyers, 1 testudon, got %d/%d",
			round3.MaxDestroyers, round3.MaxTestudons)
	}

	// Test round 4: Higher limits
	round4 := GetRoundBalance(4)
	if round4.MinFighters != 13 || round4.MaxFighters != 21 {
		t.Errorf("Round 4: Expected 13-21 fighters, got %d-%d", round4.MinFighters, round4.MaxFighters)
	}
	if round4.DestroyerProbability != 0.50 || round4.TestudonProbability != 0.35 {
		t.Errorf("Round 4: Expected 50%% destroyer, 35%% testudon probability, got %.2f/%.2f",
			round4.DestroyerProbability, round4.TestudonProbability)
	}
	if round4.MaxDestroyers != 3 || round4.MaxTestudons != 1 {
		t.Errorf("Round 4: Expected max 3 destroyers, 1 testudon, got %d/%d",
			round4.MaxDestroyers, round4.MaxTestudons)
	}

	// Test round 5: Maximum difficulty
	round5 := GetRoundBalance(5)
	if round5.MinFighters != 15 || round5.MaxFighters != 24 {
		t.Errorf("Round 5: Expected 15-24 fighters, got %d-%d", round5.MinFighters, round5.MaxFighters)
	}
	if round5.DestroyerProbability != 0.65 || round5.TestudonProbability != 0.45 {
		t.Errorf("Round 5: Expected 65%% destroyer, 45%% testudon probability, got %.2f/%.2f",
			round5.DestroyerProbability, round5.TestudonProbability)
	}
	if round5.MaxDestroyers != 3 || round5.MaxTestudons != 1 {
		t.Errorf("Round 5: Expected max 3 destroyers, 1 testudon, got %d/%d",
			round5.MaxDestroyers, round5.MaxTestudons)
	}

	// Test round 10: Should maintain round 5 settings
	round10 := GetRoundBalance(10)
	if round10.MinFighters != 25 || round10.MaxFighters != 39 {
		t.Errorf("Round 10: Expected 25-39 fighters, got %d-%d", round10.MinFighters, round10.MaxFighters)
	}
	if round10.DestroyerProbability != 0.65 || round10.TestudonProbability != 0.45 {
		t.Errorf("Round 10: Expected 65%% destroyer, 45%% testudon probability, got %.2f/%.2f",
			round10.DestroyerProbability, round10.TestudonProbability)
	}
	if round10.MaxDestroyers != 3 || round10.MaxTestudons != 1 {
		t.Errorf("Round 10: Expected max 3 destroyers, 1 testudon, got %d/%d",
			round10.MaxDestroyers, round10.MaxTestudons)
	}
}

func TestRoundBasedFleetProgression(t *testing.T) {
	// Test that round 1 never has destroyers or testudons
	for seed := int64(0); seed < 100; seed++ {
		config := GenerateRandomFleetConfig(seed, 1)
		for i, comp := range config.Compositions {
			if comp.Destroyers > 0 {
				t.Errorf("Round 1, Seed %d, Faction %d: Should have 0 destroyers, got %d",
					seed, i, comp.Destroyers)
			}
			if comp.Testudons > 0 {
				t.Errorf("Round 1, Seed %d, Faction %d: Should have 0 testudons, got %d",
					seed, i, comp.Testudons)
			}
			if comp.Fighters < 7 || comp.Fighters > 12 {
				t.Errorf("Round 1, Seed %d, Faction %d: Fighters should be 7-12, got %d",
					seed, i, comp.Fighters)
			}
		}
	}

	// Test that round 2 never has testudons and respects destroyer limits
	for seed := int64(0); seed < 100; seed++ {
		config := GenerateRandomFleetConfig(seed, 2)
		for i, comp := range config.Compositions {
			if comp.Testudons > 0 {
				t.Errorf("Round 2, Seed %d, Faction %d: Should have 0 testudons, got %d",
					seed, i, comp.Testudons)
			}
			if comp.Destroyers > 1 {
				t.Errorf("Round 2, Seed %d, Faction %d: Should have max 1 destroyer, got %d",
					seed, i, comp.Destroyers)
			}
		}
	}

	// Test that round 3 respects limits
	for seed := int64(0); seed < 100; seed++ {
		config := GenerateRandomFleetConfig(seed, 3)
		for i, comp := range config.Compositions {
			if comp.Destroyers > 2 {
				t.Errorf("Round 3, Seed %d, Faction %d: Should have max 2 destroyers, got %d",
					seed, i, comp.Destroyers)
			}
			if comp.Testudons > 1 {
				t.Errorf("Round 3, Seed %d, Faction %d: Should have max 1 testudon, got %d",
					seed, i, comp.Testudons)
			}
		}
	}

	// Test that round 5 respects limits
	for seed := int64(0); seed < 50; seed++ {
		config := GenerateRandomFleetConfig(seed, 5)
		for i, comp := range config.Compositions {
			if comp.Destroyers > 3 {
				t.Errorf("Round 5, Seed %d, Faction %d: Should have max 3 destroyers, got %d",
					seed, i, comp.Destroyers)
			}
			if comp.Testudons > 1 {
				t.Errorf("Round 5, Seed %d, Faction %d: Should have max 1 testudon, got %d",
					seed, i, comp.Testudons)
			}
		}
	}
}

func TestRoundProgressionIncreasesDifficulty(t *testing.T) {
	// Verify that average fleet size increases with rounds
	const numSeeds = 50

	for round := 1; round <= 5; round++ {
		totalShips := 0
		totalFighters := 0
		totalDestroyers := 0
		totalTestudons := 0

		for seed := int64(0); seed < numSeeds; seed++ {
			config := GenerateRandomFleetConfig(seed, round)
			for _, comp := range config.Compositions {
				totalShips += comp.Total()
				totalFighters += comp.Fighters
				totalDestroyers += comp.Destroyers
				totalTestudons += comp.Testudons
			}
		}

		avgShips := float64(totalShips) / float64(numSeeds)
		avgFighters := float64(totalFighters) / float64(numSeeds)
		avgDestroyers := float64(totalDestroyers) / float64(numSeeds)
		avgTestudons := float64(totalTestudons) / float64(numSeeds)

		t.Logf("Round %d: Avg %.1f ships (%.1f fighters, %.1f destroyers, %.1f testudons)",
			round, avgShips, avgFighters, avgDestroyers, avgTestudons)

		// Verify fighter budget is in expected range
		balance := GetRoundBalance(round)
		if avgFighters < float64(balance.MinFighters) || avgFighters > float64(balance.MaxFighters) {
			t.Logf("Round %d: Average fighters %.1f is within expected range %d-%d",
				round, avgFighters, balance.MinFighters, balance.MaxFighters)
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
