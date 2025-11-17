package config

import "math/rand"

// Ship balance costs (in fighter equivalents)
const (
	DestroyerCost = 4 // fighters per destroyer
	TestudonCost  = 8 // fighters per testudon
)

// RoundBalance defines the balance parameters for a specific round
type RoundBalance struct {
	MinFighters          int     // Minimum fighter budget for this round
	MaxFighters          int     // Maximum fighter budget for this round
	DestroyerProbability float64 // Probability to purchase each destroyer
	TestudonProbability  float64 // Probability to purchase each testudon
	MaxDestroyers        int     // Maximum destroyers allowed per faction
	MaxTestudons         int     // Maximum testudons allowed per faction
}

// GetRoundBalance returns the balance parameters for a given round number (1-indexed)
func GetRoundBalance(round int) RoundBalance {
	if round < 1 {
		round = 1
	}

	// Base values for round 1
	minFighters := 7
	maxFighters := 12

	// Progressive difficulty: each round increases fighter counts
	// Round 1: 7-12, Round 2: 9-15, Round 3: 11-18, etc.
	if round > 1 {
		minFighters += (round - 1) * 2
		maxFighters += (round - 1) * 3
	}

	// Destroyer and testudon probabilities and limits by round
	var destroyerProb, testudonProb float64
	var maxDestroyers, maxTestudons int

	switch {
	case round == 1:
		// Round 1: All fighters
		destroyerProb = 0.0
		testudonProb = 0.0
		maxDestroyers = 0
		maxTestudons = 0
	case round == 2:
		// Round 2: Introduce destroyers
		destroyerProb = 0.25
		testudonProb = 0.0
		maxDestroyers = 1
		maxTestudons = 0
	case round == 3:
		// Round 3: More destroyers, introduce testudons
		destroyerProb = 0.35
		testudonProb = 0.25
		maxDestroyers = 2
		maxTestudons = 1
	case round == 4:
		// Round 4: Even more destroyers, more testudons
		destroyerProb = 0.50
		testudonProb = 0.35
		maxDestroyers = 3
		maxTestudons = 1
	default:
		// Round 5+: Maximum difficulty
		destroyerProb = 0.65
		testudonProb = 0.45
		maxDestroyers = 3
		maxTestudons = 1
	}

	return RoundBalance{
		MinFighters:          minFighters,
		MaxFighters:          maxFighters,
		DestroyerProbability: destroyerProb,
		TestudonProbability:  testudonProb,
		MaxDestroyers:        maxDestroyers,
		MaxTestudons:         maxTestudons,
	}
}

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
func GetFleetComposition(rng *rand.Rand, numFighters int, maxDestroyers int, maxTestudons int, destroyerProb float64, testudonProb float64) FactionComposition {
	remainingFighters := numFighters
	destroyers := 0
	testudons := 0

	// First, randomly purchase testudons (most expensive)
	// Each testudon has a probability of being purchased if we can afford it
	for testudons < maxTestudons && remainingFighters >= TestudonCost {
		if rng.Float64() < testudonProb {
			testudons++
			remainingFighters -= TestudonCost
		} else {
			break // Stop trying to buy testudons
		}
	}

	// Next, randomly purchase destroyers
	// Each destroyer has a probability of being purchased if we can afford it
	for destroyers < maxDestroyers && remainingFighters >= DestroyerCost {
		if rng.Float64() < destroyerProb {
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
// using the provided seed and round number for reproducibility and progressive difficulty
// Generates 2-4 factions with balanced but varied ship compositions
// Round 1 is easiest (all fighters), with difficulty increasing each round
func GenerateRandomFleetConfig(seed int64, round int) FleetConfig {
	rng := rand.New(rand.NewSource(seed))

	// Get balance parameters for this round
	balance := GetRoundBalance(round)

	// Random number of factions (2-4)
	numFactions := rng.Intn(3) + 2 // 2, 3, or 4

	// Each faction gets a balanced fleet from a random fighter budget based on round
	compositions := make([]FactionComposition, numFactions)
	for i := 0; i < numFactions; i++ {
		// Fighter budget varies within round's min/max range
		fighterRange := balance.MaxFighters - balance.MinFighters + 1
		fighterBudget := rng.Intn(fighterRange) + balance.MinFighters

		// Use round-specific limits and probabilities
		compositions[i] = GetFleetComposition(
			rng,
			fighterBudget,
			balance.MaxDestroyers,
			balance.MaxTestudons,
			balance.DestroyerProbability,
			balance.TestudonProbability,
		)
	}

	// CRITICAL: Ensure faction 0 (player faction) has at least one controllable ship
	// Players cannot control testudons (AI-only), so faction 0 must have fighters or destroyers
	playerComp := &compositions[0]
	if playerComp.Fighters == 0 && playerComp.Destroyers == 0 && playerComp.Testudons > 0 {
		// Player faction has only testudons - convert one testudon to fighters
		playerComp.Testudons--
		playerComp.Fighters = TestudonCost // Give back the fighter equivalent (8 fighters)
	}

	return FleetConfig{
		NumFactions:  numFactions,
		Compositions: compositions,
	}
}
