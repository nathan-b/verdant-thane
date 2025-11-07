package systems

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

// BattleEndResult indicates the outcome of a battle for the player
type BattleEndResult int

const (
	BattleOngoing BattleEndResult = iota
	PlayerVictory
	PlayerDefeat
)

// CheckBattleEnd determines if the battle has ended (victory or defeat)
// Returns:
// - PlayerVictory if only the player's faction remains
// - PlayerDefeat if the player's faction has been eliminated
// - BattleOngoing if multiple factions still exist
func CheckBattleEnd(w donburi.World, playerFactionID int) BattleEndResult {
	query := donburi.NewQuery(filter.Contains(
		components.IsShip,
		components.Faction,
	))

	// Count factions with living ships
	factionsAlive := make(map[int]bool)
	playerFactionAlive := false

	for entry := range query.Iter(w) {
		faction := components.Faction.Get(entry)
		factionsAlive[faction.ID] = true
		if faction.ID == playerFactionID {
			playerFactionAlive = true
		}
	}

	// Check if player faction is eliminated
	if !playerFactionAlive {
		return PlayerDefeat
	}

	// Check if only player faction remains
	if len(factionsAlive) == 1 && playerFactionAlive {
		return PlayerVictory
	}

	// Battle ongoing
	return BattleOngoing
}
