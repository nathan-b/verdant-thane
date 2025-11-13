package main

import (
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/nathan/verdant-thane/audio"
	"github.com/nathan/verdant-thane/config"
	"github.com/nathan/verdant-thane/entity"
	"github.com/nathan/verdant-thane/systems"
)

// Profiler interface for collecting performance timing data
type Profiler interface {
	RecordAIMovement(d time.Duration)
	RecordWeaponsUpdate(d time.Duration)
	RecordBeamWeapons(d time.Duration)
	RecordMovement(d time.Duration)
	RecordAIFiring(d time.Duration)
	RecordMissileTracking(d time.Duration)
	RecordProjectileLife(d time.Duration)
	RecordCollisions(d time.Duration)
	RecordExplosions(d time.Duration)
}

// EntityManager manages all game entities using ID-based lookups
type EntityManager struct {
	// ID generation
	nextID int

	// Entity collections
	ships       map[int]entity.Ship
	projectiles map[int]entity.Projectile
	explosions  map[int]*entity.Explosion
	particles   map[int]*entity.Particle

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

	// Performance profiling (optional)
	profiler Profiler

	// Audio manager (optional)
	audioManager *audio.Manager

	// Chat window (optional)
	chatWindow ChatWindowInterface

	// Spatial grid for optimized queries (rebuilt each frame)
	spatialGrid *spatialGrid
}

// ChatWindowInterface defines the interface for chat window callbacks
type ChatWindowInterface interface {
	OnKillFighter(killerFactionID, killerShipID int)
	OnKillDestroyer(killerFactionID, killerShipID int)
	OnKillTestudon(killerFactionID, killerShipID int)
	OnFriendlyDestroyerDestroyed(observerFactionID, observerShipID int)
	OnFriendlyTestudonDestroyed(observerFactionID, observerShipID int)
}

// NewEntityManager creates a new entity manager
func NewEntityManager(laserSprite, missileSprite, explosionSprite *ebiten.Image, factionSprites *systems.FactionSprites) *EntityManager {
	return &EntityManager{
		nextID:             1,
		ships:              make(map[int]entity.Ship),
		projectiles:        make(map[int]entity.Projectile),
		explosions:         make(map[int]*entity.Explosion),
		particles:          make(map[int]*entity.Particle),
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

	// Play laser sound effect
	if em.audioManager != nil {
		em.audioManager.PlaySound("laser")
	}
}

// SpawnMissile creates a new missile projectile
func (em *EntityManager) SpawnMissile(cfg entity.MissileConfig) {
	id := em.nextID
	em.nextID++

	// Set sprite from manager
	cfg.Sprite = em.missileSprite

	missile := entity.NewMissileProjectile(id, cfg)
	em.projectiles[id] = missile

	// Play laser sound effect (missiles use same sound as main gun for now)
	if em.audioManager != nil {
		em.audioManager.PlaySound("laser")
	}
}

// SpawnExplosion creates a new explosion animation
func (em *EntityManager) SpawnExplosion(x, y float64) {
	id := em.nextID
	em.nextID++

	explosion := entity.NewExplosion(id, x, y, em.explosionSprite)
	em.explosions[id] = explosion

	// Play explosion sound effect
	if em.audioManager != nil {
		em.audioManager.PlaySound("explosion")
	}
}

// SpawnParticle creates a new particle (e.g., afterburner exhaust)
func (em *EntityManager) SpawnParticle(x, y, vx, vy float64) {
	id := em.nextID
	em.nextID++

	particle := entity.NewAfterburnerParticle(id, x, y, vx, vy)
	em.particles[id] = particle
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
	// Use spatial grid for optimized O(k) search instead of O(n)
	// Grid is rebuilt each frame in UpdateAll()
	if em.spatialGrid != nil {
		return em.spatialGrid.findNearestEnemy(ship)
	}

	// Fallback to linear search if grid not available (shouldn't happen in normal gameplay)
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
		// Sprites face UP (Y-axis), so use atan2(dx, -dy) instead of atan2(dy, dx)
		angleToTarget := math.Atan2(dx, -dy)
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
	// Handle nil factionSprites (for unit tests)
	if em.factionSprites == nil {
		return nil
	}

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

	// Handle nil classSprites (for unit tests)
	if classSprites == nil {
		return nil
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

// SetProfiler sets the profiler for performance tracking
func (em *EntityManager) SetProfiler(p Profiler) {
	em.profiler = p
}

// SetAudioManager sets the audio manager for sound effects
func (em *EntityManager) SetAudioManager(am *audio.Manager) {
	em.audioManager = am
}

// SetChatWindow sets the chat window for event notifications
func (em *EntityManager) SetChatWindow(cw ChatWindowInterface) {
	em.chatWindow = cw
}

// PlayImpactSound plays impact sound only if the target ship is player-controlled
func (em *EntityManager) PlayImpactSound(targetShip entity.Ship) {
	if em.audioManager != nil && targetShip.IsPlayerControlled() {
		em.audioManager.PlaySound("impact")
	}
}

// OnShipDestroyed handles chat events when a ship is destroyed
func (em *EntityManager) OnShipDestroyed(victimShipID int, killerShipID int) {
	if em.chatWindow == nil {
		return
	}

	victimShip := em.GetShip(victimShipID)
	killerShip := em.GetShip(killerShipID)

	if victimShip == nil || killerShip == nil {
		return
	}

	killerFaction := killerShip.GetFaction()
	killerID := killerShip.GetID()
	victimClass := victimShip.GetClass()
	victimFaction := victimShip.GetFaction()

	// Killer announces their kill
	switch victimClass {
	case entity.ClassFighter:
		em.chatWindow.OnKillFighter(killerFaction, killerID)
	case entity.ClassDestroyer:
		em.chatWindow.OnKillDestroyer(killerFaction, killerID)
	case entity.ClassTestudon:
		em.chatWindow.OnKillTestudon(killerFaction, killerID)
	}

	// Friendly ships react to capital ship losses
	if killerFaction != victimFaction {
		// Find a random friendly ship to comment on the loss
		friendlyShips := em.GetShipsByFaction(victimFaction)
		if len(friendlyShips) > 0 {
			observer := friendlyShips[rand.Intn(len(friendlyShips))]
			observerID := observer.GetID()
			observerFaction := observer.GetFaction()

			switch victimClass {
			case entity.ClassDestroyer:
				em.chatWindow.OnFriendlyDestroyerDestroyed(observerFaction, observerID)
			case entity.ClassTestudon:
				em.chatWindow.OnFriendlyTestudonDestroyed(observerFaction, observerID)
			}
		}
	}
}

// UpdateAll updates all entities and handles collisions
// Updates are broken into discrete passes to enable accurate per-subsystem profiling
func (em *EntityManager) UpdateAll() {
	// Build spatial grid once for this frame (used by AI targeting and collision detection)
	gridBuildStart := time.Now()
	em.spatialGrid = newSpatialGrid()
	for _, ship := range em.ships {
		if ship.IsAlive() {
			em.spatialGrid.insert(ship)
		}
	}
	gridBuildTime := time.Since(gridBuildStart)

	// Pass 1: Weapon capacitor charging for all ships
	weaponsStart := time.Now()
	for _, ship := range em.ships {
		if ship.IsAlive() {
			ship.UpdateWeapons()

			// Destroyers have a second weapon (missiles) that also needs charging
			if destroyer, ok := ship.(*entity.Destroyer); ok {
				destroyer.UpdateMissileWeapon()
			}
		}
	}
	if em.profiler != nil {
		em.profiler.RecordWeaponsUpdate(time.Since(weaponsStart) + gridBuildTime)
	}

	// Pass 2: AI/Player control updates (includes targeting, rotation, velocity, firing)
	aiStart := time.Now()
	for _, ship := range em.ships {
		if !ship.IsAlive() {
			continue
		}

		if ship.IsPlayerControlled() {
			// Player input is typically very fast
			ship.UpdatePlayerInput(em)
		} else {
			// AI updates include targeting, movement, and firing logic all together
			ship.UpdateAI(em)
		}
	}
	aiDuration := time.Since(aiStart)
	if em.profiler != nil {
		// Record as both AI movement and firing since UpdateAI does both
		// This allows the profiler output to show the combined metric twice
		// for compatibility with the old format, but they're the same value
		em.profiler.RecordAIMovement(aiDuration)
		em.profiler.RecordAIFiring(time.Duration(0)) // Not separately measurable
	}

	// Pass 3: Beam weapon updates for Testudons
	beamStart := time.Now()
	for _, ship := range em.ships {
		if !ship.IsAlive() {
			continue
		}

		// Only Testudons have beam weapons
		if testudon, ok := ship.(*entity.Testudon); ok {
			testudon.UpdateBeamWeapon(em)
		}
	}
	if em.profiler != nil {
		em.profiler.RecordBeamWeapons(time.Since(beamStart))
	}

	// Pass 4: Movement updates for all ships and cleanup of dead ships
	movementStart := time.Now()
	toDeleteShips := make([]int, 0)
	for id, ship := range em.ships {
		// Remove ships that died in previous updates (e.g., from TakeDamage)
		if !ship.IsAlive() {
			toDeleteShips = append(toDeleteShips, id)
			if ship.IsPlayerControlled() {
				em.handlePlayerDeath()
			}
			continue
		}

		ship.UpdateMovement()

		// Check if ship died during movement (shouldn't happen, but be safe)
		if !ship.IsAlive() {
			toDeleteShips = append(toDeleteShips, id)
			if ship.IsPlayerControlled() {
				em.handlePlayerDeath()
			}
		}
	}

	// Delete dead ships
	for _, id := range toDeleteShips {
		delete(em.ships, id)
	}

	if em.profiler != nil {
		em.profiler.RecordMovement(time.Since(movementStart))
	}

	// Update projectiles (includes missile tracking and lifetime)
	projUpdateStart := time.Now()
	missileTrackingDuration := time.Duration(0)
	lifetimeDuration := time.Duration(0)

	for id, proj := range em.projectiles {
		// Time missile-specific tracking
		trackStart := time.Now()
		proj.Update(em)
		trackDuration := time.Since(trackStart)

		// Missiles spend more time in Update() due to tracking
		// Check if this is a missile (has non-zero target tracking time)
		if missile, ok := proj.(*entity.MissileProjectile); ok && missile.Lifetime < missile.MaxLifetime {
			// Missile tracking takes majority of update time
			missileTrackingDuration += trackDuration * 70 / 100
			lifetimeDuration += trackDuration * 30 / 100
		} else {
			// Simple projectiles just update lifetime and position
			lifetimeDuration += trackDuration
		}

		if !proj.IsAlive() {
			delete(em.projectiles, id)
		}
	}

	if em.profiler != nil {
		em.profiler.RecordMissileTracking(missileTrackingDuration)
		em.profiler.RecordProjectileLife(time.Since(projUpdateStart) - missileTrackingDuration)
	}

	// Update explosions
	explosionStart := time.Now()
	for id, explosion := range em.explosions {
		explosion.Update(em)
		if !explosion.IsAlive() {
			delete(em.explosions, id)
		}
	}
	if em.profiler != nil {
		em.profiler.RecordExplosions(time.Since(explosionStart))
	}

	// Update particles (afterburner exhaust, etc.)
	for id, particle := range em.particles {
		particle.Update(em)
		if !particle.IsAlive() {
			delete(em.particles, id)
		}
	}

	// Handle collisions (projectiles vs ships)
	collisionStart := time.Now()
	em.updateCollisions()
	if em.profiler != nil {
		em.profiler.RecordCollisions(time.Since(collisionStart))
	}
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

				// Play impact sound effect only for player ship
				em.PlayImpactSound(ship)

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

// Clear removes all entities but preserves player stats (score, kills, deaths)
// This allows stats to accumulate across multiple battles in a game session
func (em *EntityManager) Clear() {
	em.ships = make(map[int]entity.Ship)
	em.projectiles = make(map[int]entity.Projectile)
	em.explosions = make(map[int]*entity.Explosion)
	em.factionSpawnPoints = make(map[int]struct{ x, y float64 })
	em.playerShipID = -1
	em.isSpectating = false
	em.spectatedShipID = -1
	// NOTE: Do NOT reset score, kills, deaths - they accumulate across battles
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
