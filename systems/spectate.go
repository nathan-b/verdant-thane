package systems

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

// SelectAlliedShipToSpectate finds an allied ship to spectate (excluding the given ship)
// Returns the entity of the allied ship, or an invalid entity if none found
func SelectAlliedShipToSpectate(w donburi.World, factionID int, excludeEntity donburi.Entity) donburi.Entity {
	query := donburi.NewQuery(filter.Contains(
		components.IsShip,
		components.Faction,
		components.Position,
	))

	// Find any ship from the same faction
	for entry := range query.Iter(w) {
		if entry.Entity() == excludeEntity {
			continue
		}

		faction := components.Faction.Get(entry)
		if faction.ID == factionID {
			return entry.Entity()
		}
	}

	// No allied ship found
	var emptyEntity donburi.Entity
	return emptyEntity
}

// GetAllAlliedShips returns a list of all ships from the specified faction
func GetAllAlliedShips(w donburi.World, factionID int) []donburi.Entity {
	query := donburi.NewQuery(filter.Contains(
		components.IsShip,
		components.Faction,
		components.Position,
	))

	allies := make([]donburi.Entity, 0)
	for entry := range query.Iter(w) {
		faction := components.Faction.Get(entry)
		if faction.ID == factionID {
			allies = append(allies, entry.Entity())
		}
	}

	return allies
}

// GetNextAlliedShip cycles to the next allied ship in the list
// If currentShip is invalid or not in the list, returns the first allied ship
func GetNextAlliedShip(w donburi.World, factionID int, currentShip donburi.Entity) donburi.Entity {
	allies := GetAllAlliedShips(w, factionID)

	if len(allies) == 0 {
		var emptyEntity donburi.Entity
		return emptyEntity
	}

	// If current ship is invalid, return first allied ship
	if !w.Valid(currentShip) {
		return allies[0]
	}

	// Find current ship in list
	currentIndex := -1
	for i, ship := range allies {
		if ship == currentShip {
			currentIndex = i
			break
		}
	}

	// If not found, return first ship
	if currentIndex == -1 {
		return allies[0]
	}

	// Return next ship (wrapping around)
	nextIndex := (currentIndex + 1) % len(allies)
	return allies[nextIndex]
}

// GetPreviousAlliedShip cycles to the previous allied ship in the list
// If currentShip is invalid or not in the list, returns the last allied ship
func GetPreviousAlliedShip(w donburi.World, factionID int, currentShip donburi.Entity) donburi.Entity {
	allies := GetAllAlliedShips(w, factionID)

	if len(allies) == 0 {
		var emptyEntity donburi.Entity
		return emptyEntity
	}

	// If current ship is invalid, return last allied ship
	if !w.Valid(currentShip) {
		return allies[len(allies)-1]
	}

	// Find current ship in list
	currentIndex := -1
	for i, ship := range allies {
		if ship == currentShip {
			currentIndex = i
			break
		}
	}

	// If not found, return last ship
	if currentIndex == -1 {
		return allies[len(allies)-1]
	}

	// Return previous ship (wrapping around)
	prevIndex := (currentIndex - 1 + len(allies)) % len(allies)
	return allies[prevIndex]
}

// EnterSpectateMode puts the player into spectate mode, selecting an allied ship
func EnterSpectateMode(w donburi.World, playerStateEntry *donburi.Entry, factionID int) {
	state := components.PlayerState.Get(playerStateEntry)

	// Select an allied ship to spectate (excluding the dead player ship)
	spectateShip := SelectAlliedShipToSpectate(w, factionID, state.ControlledShip)

	if w.Valid(spectateShip) {
		// Enter spectate mode
		state.IsSpectating = true
		state.SpectatedShip = spectateShip
		state.Deaths++

		// Clear controlled ship
		var emptyEntity donburi.Entity
		state.ControlledShip = emptyEntity
	}
	// If no allied ships, player is defeated (will handle in Update)
}

// UpdateSpectateMode checks if spectated ship is still valid and switches if needed
func UpdateSpectateMode(w donburi.World, playerStateEntry *donburi.Entry, factionID int) {
	state := components.PlayerState.Get(playerStateEntry)

	if !state.IsSpectating {
		return
	}

	// Check if spectated ship is still valid
	if !w.Valid(state.SpectatedShip) {
		// Spectated ship died, find another allied ship
		newSpectateShip := SelectAlliedShipToSpectate(w, factionID, state.SpectatedShip)

		if w.Valid(newSpectateShip) {
			state.SpectatedShip = newSpectateShip
		} else {
			// No more allied ships - player is defeated
			state.IsSpectating = false
			var emptyEntity donburi.Entity
			state.SpectatedShip = emptyEntity
		}
	}
}
