package config

import "math/rand"

// Ship balance costs (in fighter equivalents)
const (
	DestroyerCost = 4 // fighters per destroyer
	TestudonCost  = 8 // fighters per testudon
)

// Fleet generation probabilities
const (
	TestudonPurchaseProbability  = 0.33 // 33% chance to buy testudon if affordable
	DestroyerPurchaseProbability = 0.50 // 50% chance to buy destroyer if affordable
)

// FactionComposition defines the ship class breakdown for a single faction
type FactionComposition struct {
	Fighters   int
	Destroyers int
	Testudons  int
}

// Total returns the total number of ships in this composition
func (fc FactionComposition) Total() int {
	return fc.Fighters + fc.Destroyers + fc.Testudons
}

// GetFleetComposition converts a fighter budget into a mixed fleet composition
// using random exchanges at specified equivalence rates:
//   - 1 Destroyer = 4 Fighters
//   - 1 Testudon = 8 Fighters
//
// Each faction gets a different random composition from the same budget.
// Respects max limits for each ship type (e.g., no testudons until round 3).
func GetFleetComposition(rng *rand.Rand, numFighters int, maxDestroyers int, maxTestudons int) FactionComposition {
	remainingFighters := numFighters
	destroyers := 0
	testudons := 0

	// First, randomly purchase testudons (most expensive)
	// Each testudon has a probability of being purchased if we can afford it
	for testudons < maxTestudons && remainingFighters >= TestudonCost {
		if rng.Float64() < TestudonPurchaseProbability {
			testudons++
			remainingFighters -= TestudonCost
		} else {
			break // Stop trying to buy testudons
		}
	}

	// Next, randomly purchase destroyers
	// Each destroyer has a probability of being purchased if we can afford it
	for destroyers < maxDestroyers && remainingFighters >= DestroyerCost {
		if rng.Float64() < DestroyerPurchaseProbability {
			destroyers++
			remainingFighters -= DestroyerCost
		} else {
			break // Stop trying to buy destroyers
		}
	}

	return FactionComposition{
		Fighters:   remainingFighters,
		Destroyers: destroyers,
		Testudons:  testudons,
	}
}

// FleetConfig defines the composition of ships across factions
type FleetConfig struct {
	// Number of factions participating (2-4)
	NumFactions int
	// Ship composition per faction (indexed by faction ID)
	Compositions []FactionComposition
}

// GenerateFleetConfig creates a deterministic fleet configuration
// for testing specific scenarios (all factions get fighters only)
func GenerateFleetConfig(numFactions, shipsPerFaction int) FleetConfig {
	if numFactions < 2 {
		numFactions = 2
	}
	if numFactions > 4 {
		numFactions = 4
	}
	if shipsPerFaction < 1 {
		shipsPerFaction = 1
	}

	compositions := make([]FactionComposition, numFactions)
	for i := 0; i < numFactions; i++ {
		compositions[i] = FactionComposition{
			Fighters:   shipsPerFaction,
			Destroyers: 0,
			Testudons:  0,
		}
	}

	return FleetConfig{
		NumFactions:  numFactions,
		Compositions: compositions,
	}
}

// GenerateRandomFleetConfig creates a randomized fleet configuration
// using the provided seed for reproducibility
// Generates 2-4 factions with balanced but varied ship compositions
func GenerateRandomFleetConfig(seed int64) FleetConfig {
	rng := rand.New(rand.NewSource(seed))

	// Random number of factions (2-4)
	numFactions := rng.Intn(3) + 2 // 2, 3, or 4

	// Each faction gets a balanced fleet from a random fighter budget (7-16)
	// For now, we allow unlimited destroyers and testudons (TODO: add round limits)
	compositions := make([]FactionComposition, numFactions)
	for i := 0; i < numFactions; i++ {
		fighterBudget := rng.Intn(10) + 7 // 7 to 16 inclusive
		// Unlimited destroyers and testudons for now
		compositions[i] = GetFleetComposition(rng, fighterBudget, 999, 999)
	}

	return FleetConfig{
		NumFactions:  numFactions,
		Compositions: compositions,
	}
}
