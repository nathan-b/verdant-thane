package main

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/nathan/verdant-thane/config"
	"github.com/nathan/verdant-thane/entity"
	"github.com/nathan/verdant-thane/systems"
)

// EntityManager manages all game entities using ID-based lookups
type EntityManager struct {
	// ID generation
	nextID int

	// Entity collections
	ships       map[int]entity.Ship
	projectiles map[int]entity.Projectile
	explosions  map[int]*entity.Explosion

	// Faction spawn points
	factionSpawnPoints map[int]struct{ x, y float64 }

	// Player state
	playerShipID    int // ID of player-controlled ship (-1 if dead/spectating)
	isSpectating    bool
	spectatedShipID int
	score           int
	kills           int
	deaths          int

	// Sprites (shared resources)
	laserSprite     *ebiten.Image
	missileSprite   *ebiten.Image
	explosionSprite *ebiten.Image
	factionSprites  *systems.FactionSprites // Ship sprites for all classes and factions
}

// NewEntityManager creates a new entity manager
func NewEntityManager(laserSprite, missileSprite, explosionSprite *ebiten.Image, factionSprites *systems.FactionSprites) *EntityManager {
	return &EntityManager{
		nextID:             1,
		ships:              make(map[int]entity.Ship),
		projectiles:        make(map[int]entity.Projectile),
		explosions:         make(map[int]*entity.Explosion),
		factionSpawnPoints: make(map[int]struct{ x, y float64 }),
		playerShipID:       -1,
		isSpectating:       false,
		spectatedShipID:    -1,
		score:              0,
		kills:              0,
		deaths:             0,
		laserSprite:        laserSprite,
		missileSprite:      missileSprite,
		explosionSprite:    explosionSprite,
		factionSprites:     factionSprites,
	}
}

// ============================================================================
// GameContext Interface Implementation
// ============================================================================

// SpawnProjectile creates a new laser projectile
func (em *EntityManager) SpawnProjectile(cfg entity.MainGunConfig) {
	id := em.nextID
	em.nextID++

	// Set sprite from manager
	cfg.Sprite = em.laserSprite

	laser := entity.NewMainGunProjectile(id, cfg)
	em.projectiles[id] = laser
}

// SpawnMissile creates a new missile projectile
func (em *EntityManager) SpawnMissile(cfg entity.MissileConfig) {
	id := em.nextID
	em.nextID++

	// Set sprite from manager
	cfg.Sprite = em.missileSprite

	missile := entity.NewMissileProjectile(id, cfg)
	em.projectiles[id] = missile
}

// SpawnExplosion creates a new explosion animation
func (em *EntityManager) SpawnExplosion(x, y float64) {
	id := em.nextID
	em.nextID++

	explosion := entity.NewExplosion(id, x, y, em.explosionSprite)
	em.explosions[id] = explosion
}

// GetShip returns a ship by ID
func (em *EntityManager) GetShip(id int) entity.Ship {
	return em.ships[id]
}

// GetAllShips returns a slice of all ships
func (em *EntityManager) GetAllShips() []entity.Ship {
	ships := make([]entity.Ship, 0, len(em.ships))
	for _, ship := range em.ships {
		ships = append(ships, ship)
	}
	return ships
}

// GetShipsByFaction returns all ships belonging to a faction
func (em *EntityManager) GetShipsByFaction(factionID int) []entity.Ship {
	ships := make([]entity.Ship, 0)
	for _, ship := range em.ships {
		if ship.GetFaction() == factionID {
			ships = append(ships, ship)
		}
	}
	return ships
}

// FindNearestEnemy finds the nearest enemy ship to the given ship
func (em *EntityManager) FindNearestEnemy(ship entity.Ship) (entity.Ship, float64) {
	var nearestShip entity.Ship
	minDistance := math.MaxFloat64

	shipX, shipY := ship.GetPosition()
	shipFaction := ship.GetFaction()

	for _, otherShip := range em.ships {
		// Skip same faction
		if otherShip.GetFaction() == shipFaction {
			continue
		}

		// Skip dead ships
		if !otherShip.IsAlive() {
			continue
		}

		// Skip self (shouldn't happen, but safety check)
		if otherShip.GetID() == ship.GetID() {
			continue
		}

		otherX, otherY := otherShip.GetPosition()
		dist := entity.Distance(shipX, shipY, otherX, otherY)

		if dist < minDistance {
			minDistance = dist
			nearestShip = otherShip
		}
	}

	return nearestShip, minDistance
}

// FindNearestEnemyInArc finds the nearest enemy within a firing arc
func (em *EntityManager) FindNearestEnemyInArc(ship entity.Ship, arc, maxRange float64, rearFacing bool) (entity.Ship, float64) {
	var nearestShip entity.Ship
	minDistance := math.MaxFloat64

	shipX, shipY := ship.GetPosition()
	shipRotation := ship.GetRotation()
	shipFaction := ship.GetFaction()

	// Calculate arc center (forward or rear)
	arcCenter := shipRotation
	if rearFacing {
		arcCenter = entity.NormalizeAngle(shipRotation + math.Pi)
	}

	for _, otherShip := range em.ships {
		// Skip same faction
		if otherShip.GetFaction() == shipFaction {
			continue
		}

		// Skip dead ships
		if !otherShip.IsAlive() {
			continue
		}

		otherX, otherY := otherShip.GetPosition()

		// Check range
		dist := entity.Distance(shipX, shipY, otherX, otherY)
		if dist > maxRange {
			continue
		}

		// Check arc
		dx, dy := entity.GetWrappedDistance(shipX, shipY, otherX, otherY)
		angleToTarget := math.Atan2(dy, dx)
		angleFromArcCenter := math.Abs(entity.NormalizeAngle(angleToTarget - arcCenter))

		if angleFromArcCenter > arc/2 {
			continue
		}

		if dist < minDistance {
			minDistance = dist
			nearestShip = otherShip
		}
	}

	return nearestShip, minDistance
}

// GetWorldSize returns the world dimensions
func (em *EntityManager) GetWorldSize() (float64, float64) {
	return float64(config.GameWidth), float64(config.GameHeight)
}

// AddKill increments the player's kill count
func (em *EntityManager) AddKill() {
	em.kills++
}

// AddScore adds points to the player's score
func (em *EntityManager) AddScore(points int) {
	em.score += points
}

// ============================================================================
// Entity Management Methods
// ============================================================================

// SpawnShip creates a new ship of the given class at a specific position
func (em *EntityManager) SpawnShip(class entity.ShipClass, factionID int, x, y float64) entity.Ship {
	id := em.nextID
	em.nextID++

	// Get faction sprite for this ship class
	sprite := em.getSpriteForShip(class, factionID)

	var ship entity.Ship
	switch class {
	case entity.ClassFighter:
		ship = entity.NewFighter(id, factionID, x, y, sprite)
	case entity.ClassDestroyer:
		ship = entity.NewDestroyer(id, factionID, x, y, sprite)
	case entity.ClassTestudon:
		ship = entity.NewTestudon(id, factionID, x, y, sprite)
	default:
		ship = entity.NewFighter(id, factionID, x, y, sprite)
	}

	em.ships[id] = ship
	return ship
}

// SpawnShipAtFactionPoint spawns a ship at the faction's spawn point
func (em *EntityManager) SpawnShipAtFactionPoint(class entity.ShipClass, factionID int) entity.Ship {
	x, y, ok := em.GetFactionSpawnPoint(factionID)
	if !ok {
		// Fallback to center if spawn point not found
		x = float64(config.GameWidth) / 2
		y = float64(config.GameHeight) / 2
	}
	return em.SpawnShip(class, factionID, x, y)
}

// getSpriteForShip returns the appropriate sprite for a ship class and faction
func (em *EntityManager) getSpriteForShip(class entity.ShipClass, factionID int) *ebiten.Image {
	var classSprites *systems.ShipClassSprites

	switch class {
	case entity.ClassFighter:
		classSprites = em.factionSprites.Fighter
	case entity.ClassDestroyer:
		classSprites = em.factionSprites.Destroyer
	case entity.ClassTestudon:
		classSprites = em.factionSprites.Testudon
	default:
		classSprites = em.factionSprites.Fighter
	}

	// Map faction ID to sprite
	switch factionID {
	case 0:
		return classSprites.Green
	case 1:
		return classSprites.Blue
	case 2:
		return classSprites.Red
	case 3:
		return classSprites.Yellow
	default:
		return classSprites.Green
	}
}

// RemoveShip removes a ship from the manager
func (em *EntityManager) RemoveShip(id int) {
	delete(em.ships, id)
}

// RemoveProjectile removes a projectile from the manager
func (em *EntityManager) RemoveProjectile(id int) {
	delete(em.projectiles, id)
}

// RemoveExplosion removes an explosion from the manager
func (em *EntityManager) RemoveExplosion(id int) {
	delete(em.explosions, id)
}

// UpdateAll updates all entities and handles collisions
func (em *EntityManager) UpdateAll() {
	// Update ships
	for id, ship := range em.ships {
		ship.Update(em)
		if !ship.IsAlive() {
			// Handle player death
			if ship.IsPlayerControlled() {
				em.handlePlayerDeath()
			}
			delete(em.ships, id)
		}
	}

	// Update projectiles
	for id, proj := range em.projectiles {
		proj.Update(em)
		if !proj.IsAlive() {
			delete(em.projectiles, id)
		}
	}

	// Update explosions
	for id, explosion := range em.explosions {
		explosion.Update(em)
		if !explosion.IsAlive() {
			delete(em.explosions, id)
		}
	}

	// Handle collisions (projectiles vs ships)
	em.updateCollisions()
}

// updateCollisions checks for projectile-ship collisions
// Uses spatial grid partitioning to reduce collision checks from O(n×m) to O(n×k)
func (em *EntityManager) updateCollisions() {
	// Build spatial grid with all ships
	grid := newSpatialGrid()
	for _, ship := range em.ships {
		if ship.IsAlive() {
			grid.insert(ship)
		}
	}

	// Check each projectile against nearby ships only
	toDelete := make([]int, 0)

	for projID, proj := range em.projectiles {
		if !proj.IsAlive() {
			continue
		}

		// Get ships in 3×3 cell neighborhood around projectile
		projX, projY := proj.GetPosition()
		nearbyShips := grid.getNearbyShips(projX, projY)

		// Check collision against nearby ships only (not all ships)
		for _, ship := range nearbyShips {
			if !ship.IsAlive() {
				continue
			}

			// Check collision
			if proj.CheckCollision(ship) {
				// Apply damage
				ship.TakeDamage(proj.GetDamage(), proj.GetOwnerID(), em)

				// Mark projectile for deletion
				toDelete = append(toDelete, projID)
				break // Projectile can only hit one ship
			}
		}
	}

	// Delete collided projectiles
	for _, projID := range toDelete {
		delete(em.projectiles, projID)
	}
}

// handlePlayerDeath transitions player to spectate mode
func (em *EntityManager) handlePlayerDeath() {
	em.playerShipID = -1
	em.isSpectating = true
	em.deaths++

	// Find a friendly ship to spectate
	playerFaction := 0 // Player is always faction 0
	for _, ship := range em.ships {
		if ship.GetFaction() == playerFaction && ship.IsAlive() {
			em.spectatedShipID = ship.GetID()
			return
		}
	}

	// No friendly ships found
	em.spectatedShipID = -1
}

// SetPlayerShip sets the player-controlled ship
func (em *EntityManager) SetPlayerShip(shipID int) {
	// Clear old player ship
	if em.playerShipID >= 0 {
		if oldShip := em.ships[em.playerShipID]; oldShip != nil {
			oldShip.SetPlayerControlled(false)
		}
	}

	// Set new player ship
	em.playerShipID = shipID
	if shipID >= 0 {
		if newShip := em.ships[shipID]; newShip != nil {
			newShip.SetPlayerControlled(true)
		}
	}
}

// GetPlayerShip returns the player-controlled ship
func (em *EntityManager) GetPlayerShip() entity.Ship {
	if em.playerShipID < 0 {
		return nil
	}
	return em.ships[em.playerShipID]
}

// GetSpectatedShip returns the currently spectated ship
func (em *EntityManager) GetSpectatedShip() entity.Ship {
	if em.spectatedShipID < 0 {
		return nil
	}
	return em.ships[em.spectatedShipID]
}

// IsSpectating returns whether the player is in spectate mode
func (em *EntityManager) IsSpectating() bool {
	return em.isSpectating
}

// GetPlayerStats returns the player's score, kills, and deaths
func (em *EntityManager) GetPlayerStats() (score, kills, deaths int) {
	return em.score, em.kills, em.deaths
}

// CycleSpectateNext cycles to the next allied ship
func (em *EntityManager) CycleSpectateNext() {
	if !em.isSpectating {
		return
	}

	playerFaction := 0
	var allies []entity.Ship
	for _, ship := range em.ships {
		if ship.GetFaction() == playerFaction && ship.IsAlive() {
			allies = append(allies, ship)
		}
	}

	if len(allies) == 0 {
		em.spectatedShipID = -1
		return
	}

	// Find current index
	currentIndex := -1
	for i, ship := range allies {
		if ship.GetID() == em.spectatedShipID {
			currentIndex = i
			break
		}
	}

	// Cycle to next
	nextIndex := (currentIndex + 1) % len(allies)
	em.spectatedShipID = allies[nextIndex].GetID()
}

// CycleSpectatePrevious cycles to the previous allied ship
func (em *EntityManager) CycleSpectatePrevious() {
	if !em.isSpectating {
		return
	}

	playerFaction := 0
	var allies []entity.Ship
	for _, ship := range em.ships {
		if ship.GetFaction() == playerFaction && ship.IsAlive() {
			allies = append(allies, ship)
		}
	}

	if len(allies) == 0 {
		em.spectatedShipID = -1
		return
	}

	// Find current index
	currentIndex := -1
	for i, ship := range allies {
		if ship.GetID() == em.spectatedShipID {
			currentIndex = i
			break
		}
	}

	// Cycle to previous
	prevIndex := currentIndex - 1
	if prevIndex < 0 {
		prevIndex = len(allies) - 1
	}
	em.spectatedShipID = allies[prevIndex].GetID()
}

// RespawnIntoSpectatedShip attempts to take control of the spectated ship
func (em *EntityManager) RespawnIntoSpectatedShip() bool {
	if !em.isSpectating || em.spectatedShipID < 0 {
		return false
	}

	ship := em.ships[em.spectatedShipID]
	if ship == nil || !ship.IsAlive() {
		return false
	}

	// Can't respawn into testudons (AI-only ships)
	if ship.GetClass() == entity.ClassTestudon {
		return false
	}

	// Take control
	em.SetPlayerShip(em.spectatedShipID)
	em.isSpectating = false
	em.spectatedShipID = -1

	return true
}

// UpdateSpectateMode validates spectated ship and auto-switches if needed
func (em *EntityManager) UpdateSpectateMode() {
	if !em.isSpectating {
		return
	}

	// Check if spectated ship is still valid
	spectatedShip := em.ships[em.spectatedShipID]
	if spectatedShip == nil || !spectatedShip.IsAlive() {
		// Find another ship to spectate
		em.handlePlayerDeath() // Reuses logic to find new spectate target
	}
}

// BattleResult represents the outcome of a battle
type BattleResult int

const (
	BattleOngoing BattleResult = iota
	PlayerVictory
	PlayerDefeat
)

// CheckBattleEnd checks if the battle has ended
func (em *EntityManager) CheckBattleEnd() BattleResult {
	playerFaction := 0

	// Count ships per faction
	factionCounts := make(map[int]int)
	for _, ship := range em.ships {
		if ship.IsAlive() {
			factionCounts[ship.GetFaction()]++
		}
	}

	// Check if player's faction has any ships left
	playerFactionAlive := factionCounts[playerFaction] > 0

	// Count how many factions are still alive
	aliveFactions := 0
	for _, count := range factionCounts {
		if count > 0 {
			aliveFactions++
		}
	}

	// Player defeated if their faction has no ships
	if !playerFactionAlive {
		return PlayerDefeat
	}

	// Player victory if only their faction remains
	if aliveFactions == 1 {
		return PlayerVictory
	}

	// Battle ongoing
	return BattleOngoing
}

// SetFactionSpawnPoint sets the spawn point for a faction
func (em *EntityManager) SetFactionSpawnPoint(factionID int, x, y float64) {
	em.factionSpawnPoints[factionID] = struct{ x, y float64 }{x, y}
}

// GetFactionSpawnPoint returns the spawn point for a faction
func (em *EntityManager) GetFactionSpawnPoint(factionID int) (float64, float64, bool) {
	sp, ok := em.factionSpawnPoints[factionID]
	return sp.x, sp.y, ok
}

// Clear removes all entities
func (em *EntityManager) Clear() {
	em.ships = make(map[int]entity.Ship)
	em.projectiles = make(map[int]entity.Projectile)
	em.explosions = make(map[int]*entity.Explosion)
	em.factionSpawnPoints = make(map[int]struct{ x, y float64 })
	em.playerShipID = -1
	em.isSpectating = false
	em.spectatedShipID = -1
	em.score = 0
	em.kills = 0
	em.deaths = 0
	em.nextID = 1
}

// InitializeFactions creates faction spawn points at cardinal directions
// Spawn points are randomly shuffled to assign them to factions
func (em *EntityManager) InitializeFactions() {
	// Calculate spawn point positions
	// Each spawn is halfway between center and edge in each cardinal direction
	spawnPoints := []struct{ x, y float64 }{
		{float64(config.GameWidth) / 2, float64(config.GameHeight) / 4},     // North
		{float64(config.GameWidth) / 2, 3 * float64(config.GameHeight) / 4}, // South
		{3 * float64(config.GameWidth) / 4, float64(config.GameHeight) / 2}, // East
		{float64(config.GameWidth) / 4, float64(config.GameHeight) / 2},     // West
	}

	// Shuffle spawn points to randomly assign them to factions
	// Note: Using math/rand which should be seeded by the game
	for i := len(spawnPoints) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		spawnPoints[i], spawnPoints[j] = spawnPoints[j], spawnPoints[i]
	}

	// Assign spawn points to factions
	// Faction IDs: 0 = Green (player), 1 = Blue, 2 = Red, 3 = Yellow
	for factionID := 0; factionID < 4; factionID++ {
		em.SetFactionSpawnPoint(factionID, spawnPoints[factionID].x, spawnPoints[factionID].y)
	}
}
