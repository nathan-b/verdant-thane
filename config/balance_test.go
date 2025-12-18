package config

import (
	"math/rand"
	"testing"
)

// TestGetRoundBalance_Round1 tests that round 1 has no destroyers or testudons
func TestGetRoundBalance_Round1(t *testing.T) {
	balance := GetRoundBalance(1)

	if balance.MinFighters != 7 {
		t.Errorf("Round 1 MinFighters = %d, want 7", balance.MinFighters)
	}
	if balance.MaxFighters != 12 {
		t.Errorf("Round 1 MaxFighters = %d, want 12", balance.MaxFighters)
	}
	if balance.DestroyerProbability != 0.0 {
		t.Errorf("Round 1 should have no destroyers, got probability %f", balance.DestroyerProbability)
	}
	if balance.TestudonProbability != 0.0 {
		t.Errorf("Round 1 should have no testudons, got probability %f", balance.TestudonProbability)
	}
	if balance.MaxDestroyers != 0 {
		t.Errorf("Round 1 MaxDestroyers = %d, want 0", balance.MaxDestroyers)
	}
	if balance.MaxTestudons != 0 {
		t.Errorf("Round 1 MaxTestudons = %d, want 0", balance.MaxTestudons)
	}
}

// TestGetRoundBalance_Round2 tests that round 2 introduces destroyers
func TestGetRoundBalance_Round2(t *testing.T) {
	balance := GetRoundBalance(2)

	if balance.MinFighters != 9 {
		t.Errorf("Round 2 MinFighters = %d, want 9", balance.MinFighters)
	}
	if balance.MaxFighters != 15 {
		t.Errorf("Round 2 MaxFighters = %d, want 15", balance.MaxFighters)
	}
	if balance.DestroyerProbability != 0.25 {
		t.Errorf("Round 2 DestroyerProbability = %f, want 0.25", balance.DestroyerProbability)
	}
	if balance.TestudonProbability != 0.0 {
		t.Errorf("Round 2 should have no testudons, got probability %f", balance.TestudonProbability)
	}
	if balance.MaxDestroyers != 1 {
		t.Errorf("Round 2 MaxDestroyers = %d, want 1", balance.MaxDestroyers)
	}
	if balance.MaxTestudons != 0 {
		t.Errorf("Round 2 MaxTestudons = %d, want 0", balance.MaxTestudons)
	}
}

// TestGetRoundBalance_Round3 tests that round 3 introduces testudons
func TestGetRoundBalance_Round3(t *testing.T) {
	balance := GetRoundBalance(3)

	if balance.MinFighters != 11 {
		t.Errorf("Round 3 MinFighters = %d, want 11", balance.MinFighters)
	}
	if balance.MaxFighters != 18 {
		t.Errorf("Round 3 MaxFighters = %d, want 18", balance.MaxFighters)
	}
	if balance.DestroyerProbability != 0.35 {
		t.Errorf("Round 3 DestroyerProbability = %f, want 0.35", balance.DestroyerProbability)
	}
	if balance.TestudonProbability != 0.25 {
		t.Errorf("Round 3 TestudonProbability = %f, want 0.25", balance.TestudonProbability)
	}
	if balance.MaxDestroyers != 2 {
		t.Errorf("Round 3 MaxDestroyers = %d, want 2", balance.MaxDestroyers)
	}
	if balance.MaxTestudons != 1 {
		t.Errorf("Round 3 MaxTestudons = %d, want 1", balance.MaxTestudons)
	}
}

// TestGetRoundBalance_Round4 tests round 4 difficulty increase
func TestGetRoundBalance_Round4(t *testing.T) {
	balance := GetRoundBalance(4)

	if balance.MinFighters != 13 {
		t.Errorf("Round 4 MinFighters = %d, want 13", balance.MinFighters)
	}
	if balance.MaxFighters != 21 {
		t.Errorf("Round 4 MaxFighters = %d, want 21", balance.MaxFighters)
	}
	if balance.DestroyerProbability != 0.50 {
		t.Errorf("Round 4 DestroyerProbability = %f, want 0.50", balance.DestroyerProbability)
	}
	if balance.TestudonProbability != 0.35 {
		t.Errorf("Round 4 TestudonProbability = %f, want 0.35", balance.TestudonProbability)
	}
	if balance.MaxDestroyers != 3 {
		t.Errorf("Round 4 MaxDestroyers = %d, want 3", balance.MaxDestroyers)
	}
	if balance.MaxTestudons != 1 {
		t.Errorf("Round 4 MaxTestudons = %d, want 1", balance.MaxTestudons)
	}
}

// TestGetRoundBalance_Round5Plus tests maximum difficulty for round 5+
func TestGetRoundBalance_Round5Plus(t *testing.T) {
	for round := 5; round <= 10; round++ {
		balance := GetRoundBalance(round)

		expectedMin := 7 + (round-1)*2
		expectedMax := 12 + (round-1)*3

		if balance.MinFighters != expectedMin {
			t.Errorf("Round %d MinFighters = %d, want %d", round, balance.MinFighters, expectedMin)
		}
		if balance.MaxFighters != expectedMax {
			t.Errorf("Round %d MaxFighters = %d, want %d", round, balance.MaxFighters, expectedMax)
		}
		if balance.DestroyerProbability != 0.65 {
			t.Errorf("Round %d DestroyerProbability = %f, want 0.65", round, balance.DestroyerProbability)
		}
		if balance.TestudonProbability != 0.45 {
			t.Errorf("Round %d TestudonProbability = %f, want 0.45", round, balance.TestudonProbability)
		}
		if balance.MaxDestroyers != 3 {
			t.Errorf("Round %d MaxDestroyers = %d, want 3", round, balance.MaxDestroyers)
		}
		if balance.MaxTestudons != 1 {
			t.Errorf("Round %d MaxTestudons = %d, want 1", round, balance.MaxTestudons)
		}
	}
}

// TestGetRoundBalance_InvalidRound tests that invalid rounds default to round 1
func TestGetRoundBalance_InvalidRound(t *testing.T) {
	testCases := []int{-5, -1, 0}

	for _, round := range testCases {
		balance := GetRoundBalance(round)
		expected := GetRoundBalance(1)

		if balance != expected {
			t.Errorf("Round %d should default to round 1 behavior, got different balance", round)
		}
	}
}

// TestGetRoundBalance_ProgressiveDifficulty tests that difficulty increases each round
func TestGetRoundBalance_ProgressiveDifficulty(t *testing.T) {
	var previousBalance RoundBalance
	for round := 1; round <= 10; round++ {
		balance := GetRoundBalance(round)

		if round > 1 {
			// Fighter budgets should increase
			if balance.MinFighters <= previousBalance.MinFighters {
				t.Errorf("Round %d MinFighters (%d) should be greater than round %d (%d)",
					round, balance.MinFighters, round-1, previousBalance.MinFighters)
			}
			if balance.MaxFighters <= previousBalance.MaxFighters {
				t.Errorf("Round %d MaxFighters (%d) should be greater than round %d (%d)",
					round, balance.MaxFighters, round-1, previousBalance.MaxFighters)
			}
		}

		previousBalance = balance
	}
}

// TestFactionComposition_Total tests the Total() method
func TestFactionComposition_Total(t *testing.T) {
	testCases := []struct {
		name      string
		comp      FactionComposition
		wantTotal int
	}{
		{
			name:      "all fighters",
			comp:      FactionComposition{Fighters: 10, Destroyers: 0, Testudons: 0},
			wantTotal: 10,
		},
		{
			name:      "mixed fleet",
			comp:      FactionComposition{Fighters: 5, Destroyers: 2, Testudons: 1},
			wantTotal: 8,
		},
		{
			name:      "empty fleet",
			comp:      FactionComposition{Fighters: 0, Destroyers: 0, Testudons: 0},
			wantTotal: 0,
		},
		{
			name:      "all destroyers",
			comp:      FactionComposition{Fighters: 0, Destroyers: 5, Testudons: 0},
			wantTotal: 5,
		},
		{
			name:      "all testudons",
			comp:      FactionComposition{Fighters: 0, Destroyers: 0, Testudons: 3},
			wantTotal: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.comp.Total()
			if got != tc.wantTotal {
				t.Errorf("Total() = %d, want %d", got, tc.wantTotal)
			}
		})
	}
}

// TestGetFleetComposition_AllFighters tests composition with 0 probabilities
func TestGetFleetComposition_AllFighters(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	comp := GetFleetComposition(rng, 20, 0, 0, 0.0, 0.0)

	if comp.Fighters != 20 {
		t.Errorf("Expected 20 fighters, got %d", comp.Fighters)
	}
	if comp.Destroyers != 0 {
		t.Errorf("Expected 0 destroyers, got %d", comp.Destroyers)
	}
	if comp.Testudons != 0 {
		t.Errorf("Expected 0 testudons, got %d", comp.Testudons)
	}
}

// TestGetFleetComposition_MaxLimits tests that max limits are respected
func TestGetFleetComposition_MaxLimits(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	// High probability but low max limits
	comp := GetFleetComposition(rng, 100, 2, 1, 1.0, 1.0)

	if comp.Destroyers > 2 {
		t.Errorf("Destroyers exceeded max limit: got %d, want <= 2", comp.Destroyers)
	}
	if comp.Testudons > 1 {
		t.Errorf("Testudons exceeded max limit: got %d, want <= 1", comp.Testudons)
	}

	// Total should not exceed budget
	totalCost := comp.Fighters + comp.Destroyers*DestroyerCost + comp.Testudons*TestudonCost
	if totalCost > 100 {
		t.Errorf("Total cost %d exceeds budget 100", totalCost)
	}
}

// TestGetFleetComposition_BudgetSpending tests that budget is properly consumed
func TestGetFleetComposition_BudgetSpending(t *testing.T) {
	testCases := []struct {
		name          string
		budget        int
		maxDestroyers int
		maxTestudons  int
		destroyerProb float64
		testudonProb  float64
	}{
		{"small budget", 10, 2, 1, 0.5, 0.5},
		{"medium budget", 50, 3, 2, 0.5, 0.5},
		{"large budget", 100, 5, 3, 0.5, 0.5},
		{"no advanced ships", 20, 0, 0, 0.0, 0.0},
		{"guaranteed purchase", 50, 5, 2, 1.0, 1.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(42))
			comp := GetFleetComposition(rng, tc.budget, tc.maxDestroyers, tc.maxTestudons, tc.destroyerProb, tc.testudonProb)

			// Calculate actual cost
			actualCost := comp.Fighters + comp.Destroyers*DestroyerCost + comp.Testudons*TestudonCost

			// Total cost should not exceed budget
			if actualCost > tc.budget {
				t.Errorf("Cost %d exceeds budget %d (F:%d D:%d T:%d)",
					actualCost, tc.budget, comp.Fighters, comp.Destroyers, comp.Testudons)
			}

			// Should always have at least 1 ship
			if comp.Total() == 0 {
				t.Errorf("Fleet composition resulted in 0 ships with budget %d", tc.budget)
			}

			// Verify max limits
			if comp.Destroyers > tc.maxDestroyers {
				t.Errorf("Destroyers %d exceeds max %d", comp.Destroyers, tc.maxDestroyers)
			}
			if comp.Testudons > tc.maxTestudons {
				t.Errorf("Testudons %d exceeds max %d", comp.Testudons, tc.maxTestudons)
			}
		})
	}
}

// TestGetFleetComposition_InsufficientBudget tests behavior with very small budgets
func TestGetFleetComposition_InsufficientBudget(t *testing.T) {
	testCases := []struct {
		name   string
		budget int
	}{
		{"1 fighter", 1},
		{"2 fighters", 2},
		{"3 fighters (can't afford destroyer)", 3},
		{"7 fighters (can't afford testudon)", 7},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(42))
			// Try to buy advanced ships with high probability
			comp := GetFleetComposition(rng, tc.budget, 5, 5, 1.0, 1.0)

			// With budget < DestroyerCost (4), should get all fighters
			if tc.budget < DestroyerCost {
				if comp.Destroyers > 0 || comp.Testudons > 0 {
					t.Errorf("Budget %d too small for advanced ships, got D:%d T:%d",
						tc.budget, comp.Destroyers, comp.Testudons)
				}
				if comp.Fighters != tc.budget {
					t.Errorf("Expected %d fighters with budget %d, got %d",
						tc.budget, tc.budget, comp.Fighters)
				}
			}

			// With budget < TestudonCost (8), should not buy testudons
			if tc.budget < TestudonCost && comp.Testudons > 0 {
				t.Errorf("Budget %d too small for testudons, got %d", tc.budget, comp.Testudons)
			}
		})
	}
}

// TestGetFleetComposition_Deterministic tests that same seed produces same result
func TestGetFleetComposition_Deterministic(t *testing.T) {
	seed := int64(12345)
	budget := 50
	maxDestroyers := 3
	maxTestudons := 2
	destroyerProb := 0.5
	testudonProb := 0.4

	// Generate composition twice with same seed
	rng1 := rand.New(rand.NewSource(seed))
	comp1 := GetFleetComposition(rng1, budget, maxDestroyers, maxTestudons, destroyerProb, testudonProb)

	rng2 := rand.New(rand.NewSource(seed))
	comp2 := GetFleetComposition(rng2, budget, maxDestroyers, maxTestudons, destroyerProb, testudonProb)

	if comp1 != comp2 {
		t.Errorf("Same seed should produce same composition:\n  First:  %+v\n  Second: %+v", comp1, comp2)
	}
}

// TestGenerateFleetConfig_ValidInput tests basic fleet config generation
func TestGenerateFleetConfig_ValidInput(t *testing.T) {
	config := GenerateFleetConfig(3, 10)

	if config.NumFactions != 3 {
		t.Errorf("NumFactions = %d, want 3", config.NumFactions)
	}
	if len(config.Compositions) != 3 {
		t.Errorf("len(Compositions) = %d, want 3", len(config.Compositions))
	}

	for i, comp := range config.Compositions {
		if comp.Fighters != 10 {
			t.Errorf("Faction %d: Fighters = %d, want 10", i, comp.Fighters)
		}
		if comp.Destroyers != 0 {
			t.Errorf("Faction %d: Destroyers = %d, want 0", i, comp.Destroyers)
		}
		if comp.Testudons != 0 {
			t.Errorf("Faction %d: Testudons = %d, want 0", i, comp.Testudons)
		}
	}
}

// TestGenerateFleetConfig_BoundaryValidation tests input validation
func TestGenerateFleetConfig_BoundaryValidation(t *testing.T) {
	testCases := []struct {
		name            string
		numFactions     int
		shipsPerFaction int
		wantFactions    int
		wantShips       int
	}{
		{"too few factions", 1, 10, 2, 10},
		{"too many factions", 5, 10, 4, 10},
		{"zero ships", 3, 0, 3, 1},
		{"negative ships", 3, -5, 3, 1},
		{"valid config", 3, 15, 3, 15},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := GenerateFleetConfig(tc.numFactions, tc.shipsPerFaction)

			if config.NumFactions != tc.wantFactions {
				t.Errorf("NumFactions = %d, want %d", config.NumFactions, tc.wantFactions)
			}
			if len(config.Compositions) != tc.wantFactions {
				t.Errorf("len(Compositions) = %d, want %d", len(config.Compositions), tc.wantFactions)
			}
			if config.Compositions[0].Fighters != tc.wantShips {
				t.Errorf("Ships per faction = %d, want %d", config.Compositions[0].Fighters, tc.wantShips)
			}
		})
	}
}

// TestGenerateRandomFleetConfig_Deterministic tests reproducibility with same seed
func TestGenerateRandomFleetConfig_Deterministic(t *testing.T) {
	seed := int64(99999)
	round := 3

	config1 := GenerateRandomFleetConfig(seed, round)
	config2 := GenerateRandomFleetConfig(seed, round)

	if config1.NumFactions != config2.NumFactions {
		t.Errorf("NumFactions differs: %d vs %d", config1.NumFactions, config2.NumFactions)
	}

	if len(config1.Compositions) != len(config2.Compositions) {
		t.Fatalf("Composition count differs: %d vs %d", len(config1.Compositions), len(config2.Compositions))
	}

	for i := range config1.Compositions {
		if config1.Compositions[i] != config2.Compositions[i] {
			t.Errorf("Faction %d composition differs:\n  First:  %+v\n  Second: %+v",
				i, config1.Compositions[i], config2.Compositions[i])
		}
	}
}

// TestGenerateRandomFleetConfig_ProgressiveDifficulty tests difficulty scaling
func TestGenerateRandomFleetConfig_ProgressiveDifficulty(t *testing.T) {
	seed := int64(42)

	// Round 1: All fighters
	config1 := GenerateRandomFleetConfig(seed, 1)
	for i, comp := range config1.Compositions {
		if comp.Destroyers > 0 || comp.Testudons > 0 {
			t.Errorf("Round 1 faction %d should have only fighters, got D:%d T:%d",
				i, comp.Destroyers, comp.Testudons)
		}
		if comp.Fighters < 7 || comp.Fighters > 12 {
			t.Errorf("Round 1 faction %d fighters %d outside range [7, 12]",
				i, comp.Fighters)
		}
	}

	// Round 3: Should have testudons available
	config3 := GenerateRandomFleetConfig(seed+1, 3)
	balance3 := GetRoundBalance(3)
	for i, comp := range config3.Compositions {
		// Budget should be in round 3 range
		fighterEquiv := comp.Fighters + comp.Destroyers*DestroyerCost + comp.Testudons*TestudonCost
		if fighterEquiv < balance3.MinFighters || fighterEquiv > balance3.MaxFighters {
			t.Errorf("Round 3 faction %d budget %d outside range [%d, %d]",
				i, fighterEquiv, balance3.MinFighters, balance3.MaxFighters)
		}

		// Verify max limits
		if comp.Destroyers > balance3.MaxDestroyers {
			t.Errorf("Round 3 faction %d has %d destroyers, max is %d",
				i, comp.Destroyers, balance3.MaxDestroyers)
		}
		if comp.Testudons > balance3.MaxTestudons {
			t.Errorf("Round 3 faction %d has %d testudons, max is %d",
				i, comp.Testudons, balance3.MaxTestudons)
		}
	}
}

// TestGenerateRandomFleetConfig_PlayerFactionValidation tests critical player faction bug prevention
func TestGenerateRandomFleetConfig_PlayerFactionValidation(t *testing.T) {
	// Test many random seeds to ensure player faction always has controllable ships
	for seed := int64(0); seed < 1000; seed++ {
		for round := 1; round <= 10; round++ {
			config := GenerateRandomFleetConfig(seed, round)

			// Player is always faction 0
			playerComp := config.Compositions[0]

			// Player must have at least one controllable ship (fighter or destroyer)
			// Testudons are AI-only
			controllableShips := playerComp.Fighters + playerComp.Destroyers
			if controllableShips == 0 {
				t.Errorf("Seed %d Round %d: Player faction has no controllable ships (F:%d D:%d T:%d)",
					seed, round, playerComp.Fighters, playerComp.Destroyers, playerComp.Testudons)
			}

			// Player faction should always have at least one ship total
			if playerComp.Total() == 0 {
				t.Errorf("Seed %d Round %d: Player faction has no ships at all", seed, round)
			}
		}
	}
}

// TestGenerateRandomFleetConfig_FactionCount tests that faction count is in valid range
func TestGenerateRandomFleetConfig_FactionCount(t *testing.T) {
	for seed := int64(0); seed < 100; seed++ {
		config := GenerateRandomFleetConfig(seed, 3)

		if config.NumFactions < 2 || config.NumFactions > 4 {
			t.Errorf("Seed %d: NumFactions = %d, want 2-4", seed, config.NumFactions)
		}

		if len(config.Compositions) != config.NumFactions {
			t.Errorf("Seed %d: len(Compositions) = %d, want %d",
				seed, len(config.Compositions), config.NumFactions)
		}
	}
}

// TestGenerateRandomFleetConfig_FleetSizeVariation tests that fleets vary in size
func TestGenerateRandomFleetConfig_FleetSizeVariation(t *testing.T) {
	seed := int64(12345)
	round := 4

	config := GenerateRandomFleetConfig(seed, round)

	// With random budgets, factions should have different total ship counts
	// (This is probabilistic, but with different budgets it's very likely)
	sizes := make(map[int]bool)
	for i, comp := range config.Compositions {
		size := comp.Total()
		if size == 0 {
			t.Errorf("Faction %d has no ships", i)
		}
		sizes[size] = true
	}

	// With 3-4 factions and random budgets, we should see some variation
	// (Not enforced strictly, but good to check)
	t.Logf("Round %d generated %d factions with %d different fleet sizes",
		round, config.NumFactions, len(sizes))
}

// TestGenerateRandomFleetConfig_BudgetRespect tests that generated fleets respect round budgets
func TestGenerateRandomFleetConfig_BudgetRespect(t *testing.T) {
	for round := 1; round <= 5; round++ {
		balance := GetRoundBalance(round)
		config := GenerateRandomFleetConfig(int64(round*100), round)

		for i, comp := range config.Compositions {
			fighterEquiv := comp.Fighters + comp.Destroyers*DestroyerCost + comp.Testudons*TestudonCost

			if fighterEquiv < balance.MinFighters || fighterEquiv > balance.MaxFighters {
				t.Errorf("Round %d faction %d: budget %d outside range [%d, %d]",
					round, i, fighterEquiv, balance.MinFighters, balance.MaxFighters)
			}
		}
	}
}
