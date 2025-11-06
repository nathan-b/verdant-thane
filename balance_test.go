package main

import (
	"math/rand"
	"testing"
)

// TestGetFleetComposition_BalanceEquivalence verifies the balance system
func TestGetFleetComposition_BalanceEquivalence(t *testing.T) {
	const (
		destroyerCost = 4 // fighters per destroyer
		testudonCost  = 8 // fighters per testudon
	)

	tests := []struct {
		name          string
		fighterBudget int
		maxDestroyers int
		maxTestudons  int
	}{
		{"Small fleet", 8, 5, 2},
		{"Medium fleet", 16, 10, 3},
		{"Large fleet", 32, 20, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create multiple compositions with same budget
			rng := rand.New(rand.NewSource(42))
			compositions := make([]FactionComposition, 5)

			for i := range compositions {
				compositions[i] = getFleetComposition(rng, tt.fighterBudget, tt.maxDestroyers, tt.maxTestudons)
			}

			// Verify each composition's total "cost" equals the budget
			for i, comp := range compositions {
				totalCost := comp.Fighters + (comp.Destroyers * destroyerCost) + (comp.Testudons * testudonCost)
				if totalCost != tt.fighterBudget {
					t.Errorf("Composition %d: cost mismatch. Budget=%d, Cost=%d (%d fighters + %d destroyers*%d + %d testudons*%d)",
						i, tt.fighterBudget, totalCost, comp.Fighters, comp.Destroyers, destroyerCost, comp.Testudons, testudonCost)
				}
			}

			// Log the compositions for manual verification of variety
			t.Logf("Budget: %d fighters", tt.fighterBudget)
			for i, comp := range compositions {
				t.Logf("  Faction %d: %d fighters, %d destroyers, %d testudons (total ships: %d)",
					i, comp.Fighters, comp.Destroyers, comp.Testudons, comp.Total())
			}
		})
	}
}

// TestGetFleetComposition_RespectsLimits verifies ship type limits are respected
func TestGetFleetComposition_RespectsLimits(t *testing.T) {
	tests := []struct {
		name          string
		maxDestroyers int
		maxTestudons  int
	}{
		{"No destroyers allowed", 0, 5},
		{"No testudons allowed", 5, 0},
		{"Limited destroyers", 2, 5},
		{"Limited testudons", 5, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(42))

			// Test multiple times to ensure limits are always respected
			for i := 0; i < 10; i++ {
				comp := getFleetComposition(rng, 32, tt.maxDestroyers, tt.maxTestudons)

				if comp.Destroyers > tt.maxDestroyers {
					t.Errorf("Destroyers exceed limit: got %d, max %d", comp.Destroyers, tt.maxDestroyers)
				}
				if comp.Testudons > tt.maxTestudons {
					t.Errorf("Testudons exceed limit: got %d, max %d", comp.Testudons, tt.maxTestudons)
				}
			}
		})
	}
}

// TestGetFleetComposition_Variety verifies different factions get varied compositions
func TestGetFleetComposition_Variety(t *testing.T) {
	const fighterBudget = 16
	rng := rand.New(rand.NewSource(123))

	// Generate 10 compositions
	compositions := make([]FactionComposition, 10)
	for i := range compositions {
		compositions[i] = getFleetComposition(rng, fighterBudget, 999, 999)
	}

	// Count how many unique compositions we got
	uniqueComps := make(map[FactionComposition]bool)
	for _, comp := range compositions {
		uniqueComps[comp] = true
	}

	// We should get at least some variety (not all identical)
	if len(uniqueComps) == 1 {
		t.Error("All factions got identical compositions - no variety!")
	}

	// Log variety for manual inspection
	t.Logf("Generated %d unique compositions from %d factions:", len(uniqueComps), len(compositions))
	for comp := range uniqueComps {
		t.Logf("  %d fighters, %d destroyers, %d testudons", comp.Fighters, comp.Destroyers, comp.Testudons)
	}
}
