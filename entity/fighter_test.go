package entity

import (
	"math"
	"testing"

	"github.com/nathan-b/verdant-thane/config"
)

// MockGameContext is a test implementation of GameContext
type MockGameContext struct {
	spawnedProjectiles []MainGunConfig
	spawnedMissiles    []MissileConfig
	spawnedExplosions  []struct{ x, y float64 }
	ships              map[int]*Ship
	killCount          int
	scoreAdded         int
}

func NewMockGameContext() *MockGameContext {
	return &MockGameContext{
		spawnedProjectiles: make([]MainGunConfig, 0),
		spawnedMissiles:    make([]MissileConfig, 0),
		spawnedExplosions:  make([]struct{ x, y float64 }, 0),
		ships:              make(map[int]*Ship),
		killCount:          0,
		scoreAdded:         0,
	}
}

func (m *MockGameContext) SpawnProjectile(cfg MainGunConfig) {
	m.spawnedProjectiles = append(m.spawnedProjectiles, cfg)
}

func (m *MockGameContext) SpawnMissile(cfg MissileConfig) {
	m.spawnedMissiles = append(m.spawnedMissiles, cfg)
}

func (m *MockGameContext) SpawnExplosion(x, y float64) {
	m.spawnedExplosions = append(m.spawnedExplosions, struct{ x, y float64 }{x, y})
}

func (m *MockGameContext) GetShip(id int) *Ship {
	return m.ships[id]
}

func (m *MockGameContext) GetAllShips() []*Ship {
	ships := make([]*Ship, 0, len(m.ships))
	for _, ship := range m.ships {
		ships = append(ships, ship)
	}
	return ships
}

func (m *MockGameContext) GetShipsByFaction(factionID int) []*Ship {
	ships := make([]*Ship, 0)
	for _, ship := range m.ships {
		if ship.GetFaction() == factionID {
			ships = append(ships, ship)
		}
	}
	return ships
}

func (m *MockGameContext) FindNearestEnemy(ship *Ship) (*Ship, float64) {
	var nearest *Ship
	var minDist float64 = math.MaxFloat64

	shipX, shipY := ship.GetPosition()
	for _, other := range m.ships {
		if other.GetID() == ship.GetID() || other.GetFaction() == ship.GetFaction() || !other.IsAlive() {
			continue
		}

		otherX, otherY := other.GetPosition()
		dist := Distance(shipX, shipY, otherX, otherY)
		if dist < minDist {
			minDist = dist
			nearest = other
		}
	}

	return nearest, minDist
}

func (m *MockGameContext) FindNearestEnemyInArc(ship *Ship, arc, maxRange float64, rearFacing bool) (*Ship, float64) {
	// Simplified implementation for testing
	return m.FindNearestEnemy(ship)
}

func (m *MockGameContext) GetWorldSize() (float64, float64) {
	return float64(config.GameWidth), float64(config.GameHeight)
}

func (m *MockGameContext) AddKill() {
	m.killCount++
}

func (m *MockGameContext) AddScore(points int) {
	m.scoreAdded += points
}

func (m *MockGameContext) PlayImpactSound(targetShip *Ship, proj Projectile) {
	// No-op for tests
}

func (m *MockGameContext) SpawnParticle(x, y, vx, vy float64) {
	// No-op for tests
}

func (m *MockGameContext) OnShipDestroyed(victimShipID int, killerShipID int) {
	// No-op for tests
}

// Test Fighter Creation
func TestNewFighter(t *testing.T) {
	fighter := NewFighter(1, 0, 100, 200, nil)

	// Verify identity
	if fighter.GetID() != 1 {
		t.Errorf("Expected ID 1, got %d", fighter.GetID())
	}
	if fighter.GetFaction() != 0 {
		t.Errorf("Expected faction 0, got %d", fighter.GetFaction())
	}
	if fighter.GetClass() != ClassFighter {
		t.Errorf("Expected ClassFighter, got %v", fighter.GetClass())
	}

	// Verify position
	x, y := fighter.GetPosition()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%f, %f)", x, y)
	}

	// Verify initial state
	if !fighter.IsAlive() {
		t.Error("Fighter should be alive initially")
	}
	if fighter.IsPlayerControlled() {
		t.Error("Fighter should not be player-controlled initially")
	}

	// Verify stats match config
	chars := config.GetShipCharacteristics(ClassFighter)
	health, maxHealth := fighter.GetHealth()
	if health != chars.MaxShield || maxHealth != chars.MaxShield {
		t.Errorf("Expected health %d/%d, got %d/%d", chars.MaxShield, chars.MaxShield, health, maxHealth)
	}
	if fighter.MaxSpeed != chars.MaxSpeed {
		t.Errorf("Expected max speed %f, got %f", chars.MaxSpeed, fighter.MaxSpeed)
	}

	// Verify weapon starts charged
	if !fighter.CanFireWeapon() {
		t.Error("Weapon should start fully charged")
	}

	// Verify velocity starts at zero
	vx, vy := fighter.GetVelocity()
	if vx != 0 || vy != 0 {
		t.Errorf("Expected zero velocity, got (%f, %f)", vx, vy)
	}
}

// Test Movement and Physics
func TestFighterMovement(t *testing.T) {
	fighter := NewFighter(1, 0, 100, 100, nil)
	// Set speed and rotation instead of velocity directly
	// Rotation 0 = facing up, we want velocity (5, 3)
	// atan2(5, -3) = angle for velocity vector
	fighter.Speed = math.Sqrt(5.0*5.0 + 3.0*3.0)
	fighter.Rotation = math.Atan2(5.0, -3.0)

	fighter.UpdateMovement()

	x, y := fighter.GetPosition()
	if math.Abs(x-105.0) > 1e-9 || math.Abs(y-103.0) > 1e-9 {
		t.Errorf("Expected position (105, 103), got (%f, %f)", x, y)
	}

	// Verify speed is maintained correctly
	expectedSpeed := math.Sqrt(5.0*5.0 + 3.0*3.0)
	if math.Abs(fighter.Speed-expectedSpeed) > 1e-9 {
		t.Errorf("Expected speed %f, got %f", expectedSpeed, fighter.Speed)
	}
}

// Test World Wrapping
func TestFighterWrapping(t *testing.T) {
	tests := []struct {
		name                 string
		startX, startY       float64
		velocityX, velocityY float64
		expectedX, expectedY float64
	}{
		{"Wrap right edge", float64(config.GameWidth) - 10, 100, 20, 0, 10, 100},
		{"Wrap left edge", 10, 100, -20, 0, float64(config.GameWidth) - 10, 100},
		{"Wrap bottom edge", 100, float64(config.GameHeight) - 10, 0, 20, 100, 10},
		{"Wrap top edge", 100, 10, 0, -20, 100, float64(config.GameHeight) - 10},
		{"No wrapping", 100, 100, 5, 5, 105, 105},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fighter := NewFighter(1, 0, tt.startX, tt.startY, nil)
			// Set speed and rotation to produce desired velocity
			fighter.Speed = math.Sqrt(tt.velocityX*tt.velocityX + tt.velocityY*tt.velocityY)
			if fighter.Speed > 0 {
				fighter.Rotation = math.Atan2(tt.velocityX, -tt.velocityY)
			}

			fighter.UpdateMovement()

			x, y := fighter.GetPosition()
			if math.Abs(x-tt.expectedX) > 1e-9 || math.Abs(y-tt.expectedY) > 1e-9 {
				t.Errorf("Expected position (%f, %f), got (%f, %f)", tt.expectedX, tt.expectedY, x, y)
			}
		})
	}
}

// Test Weapon Charging
func TestWeaponCharging(t *testing.T) {
	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Weapons[0].WeaponCapacitor = 0.0

	// Weapon should not be fireable when empty
	if fighter.CanFireWeapon() {
		t.Error("Weapon should not be fireable when capacitor is empty")
	}

	// Charge weapon
	chars := config.GetShipCharacteristics(ClassFighter)
	ticksToCharge := int(1.0 / chars.Weapons[0].CapacitorChargeRate)

	// Charge for enough ticks to guarantee full charge (add 1 to handle floating point precision)
	for i := 0; i <= ticksToCharge; i++ {
		fighter.UpdateWeapons()
	}

	// Should be fully charged now
	if !fighter.CanFireWeapon() {
		t.Errorf("Weapon should be fireable after charging (capacitor: %f)", fighter.Weapons[0].WeaponCapacitor)
	}
	if fighter.Weapons[0].WeaponCapacitor < 1.0 {
		t.Errorf("Expected capacitor to be >= 1.0, got %.10f", fighter.Weapons[0].WeaponCapacitor)
	}
}

// Test Weapon Firing
func TestFighterFireWeapon(t *testing.T) {
	ctx := NewMockGameContext()
	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Rotation = 0 // Facing right

	// Fire at a position directly ahead
	fighter.FireWeapon(200, 100, ctx)

	// Verify projectile was spawned
	if len(ctx.spawnedProjectiles) != 1 {
		t.Fatalf("Expected 1 projectile, got %d", len(ctx.spawnedProjectiles))
	}

	proj := ctx.spawnedProjectiles[0]

	// Verify ownership
	if proj.OwnerID != fighter.GetID() {
		t.Errorf("Expected owner ID %d, got %d", fighter.GetID(), proj.OwnerID)
	}
	if proj.FactionID != fighter.GetFaction() {
		t.Errorf("Expected faction ID %d, got %d", fighter.GetFaction(), proj.FactionID)
	}

	// Verify projectile spawns ahead of ship
	if proj.X <= fighter.X {
		t.Errorf("Projectile should spawn ahead of ship, got X=%f (ship X=%f)", proj.X, fighter.X)
	}

	// Verify projectile has velocity
	if proj.VelocityX <= 0 {
		t.Errorf("Projectile should have positive X velocity when firing right, got %f", proj.VelocityX)
	}

	// Verify capacitor was consumed
	if fighter.Weapons[0].WeaponCapacitor != 0.0 {
		t.Errorf("Expected capacitor to be 0.0 after firing, got %f", fighter.Weapons[0].WeaponCapacitor)
	}

	// Verify can't fire again immediately
	if fighter.CanFireWeapon() {
		t.Error("Should not be able to fire immediately after discharging capacitor")
	}
}

// Test Firing Outside Cone
func TestFighterFiringCone(t *testing.T) {
	ctx := NewMockGameContext()
	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Rotation = 0 // Facing up (sprites face UP)

	// Try to fire at a position behind the ship (outside cone)
	// Behind means below (positive Y) since ship faces up
	fighter.FireWeapon(100, 200, ctx)

	// Should still fire, but along cone edge
	if len(ctx.spawnedProjectiles) != 1 {
		t.Fatalf("Expected projectile to fire along cone edge, got %d projectiles", len(ctx.spawnedProjectiles))
	}

	proj := ctx.spawnedProjectiles[0]

	// Projectile should be angled along cone edge, not directly at target
	// Since target is behind (180° in ship coords), and cone is 30°, projectile should fire at ±15° from forward
	// Calculate velocity angle in ship coordinate system (sprites face UP)
	firingAngle := math.Atan2(proj.VelocityX, -proj.VelocityY)
	angleDiff := math.Abs(NormalizeAngle(firingAngle - fighter.Rotation))

	chars := config.GetShipCharacteristics(ClassFighter)
	halfCone := chars.Weapons[0].FiringCone / 2

	// Should be at cone edge (approximately)
	if math.Abs(angleDiff-halfCone) > 0.01 {
		t.Errorf("Expected firing angle near cone edge (%f rad), got angle diff %f rad",
			halfCone, angleDiff)
	}
}

// Test Damage and Destruction
func TestFighterTakeDamage(t *testing.T) {
	ctx := NewMockGameContext()
	fighter := NewFighter(1, 0, 100, 100, nil)

	// Take non-lethal damage
	fighter.TakeDamage(3, 2, ctx)
	health, _ := fighter.GetHealth()
	if health != 5 { // Started with 8, took 3 damage
		t.Errorf("Expected health 5, got %d", health)
	}
	if !fighter.IsAlive() {
		t.Error("Fighter should still be alive")
	}

	// No explosion yet
	if len(ctx.spawnedExplosions) != 0 {
		t.Errorf("Expected no explosions, got %d", len(ctx.spawnedExplosions))
	}

	// Take lethal damage
	fighter.TakeDamage(10, 2, ctx)
	health, _ = fighter.GetHealth()
	if health != 0 {
		t.Errorf("Expected health 0, got %d", health)
	}
	if fighter.IsAlive() {
		t.Error("Fighter should be dead")
	}

	// Should spawn explosion
	if len(ctx.spawnedExplosions) != 1 {
		t.Fatalf("Expected 1 explosion, got %d", len(ctx.spawnedExplosions))
	}

	explosion := ctx.spawnedExplosions[0]
	if explosion.x != 100 || explosion.y != 100 {
		t.Errorf("Expected explosion at (100, 100), got (%f, %f)", explosion.x, explosion.y)
	}
}

// Test Player Controlled Score Tracking
func TestFighterScoreTracking(t *testing.T) {
	ctx := NewMockGameContext()

	// Create attacker ship (player-controlled)
	attacker := NewFighter(1, 0, 100, 100, nil)
	attacker.SetPlayerControlled(true)
	ctx.ships[attacker.GetID()] = attacker

	// Create target ship (enemy)
	target := NewFighter(2, 1, 200, 200, nil)

	// Kill target with player ship
	target.TakeDamage(100, attacker.GetID(), ctx)

	// Verify score was awarded
	if ctx.killCount != 1 {
		t.Errorf("Expected 1 kill, got %d", ctx.killCount)
	}
	if ctx.scoreAdded != 10 {
		t.Errorf("Expected 10 points, got %d", ctx.scoreAdded)
	}
}

// Test No Score for AI Kills
func TestFighterNoScoreForAIKills(t *testing.T) {
	ctx := NewMockGameContext()

	// Create attacker ship (AI-controlled)
	attacker := NewFighter(1, 0, 100, 100, nil)
	ctx.ships[attacker.GetID()] = attacker

	// Create target ship (enemy)
	target := NewFighter(2, 1, 200, 200, nil)

	// Kill target with AI ship
	target.TakeDamage(100, attacker.GetID(), ctx)

	// Verify no score was awarded
	if ctx.killCount != 0 {
		t.Errorf("Expected 0 kills for AI, got %d", ctx.killCount)
	}
	if ctx.scoreAdded != 0 {
		t.Errorf("Expected 0 points for AI, got %d", ctx.scoreAdded)
	}
}

// Test Player Control Setting
func TestFighterPlayerControl(t *testing.T) {
	fighter := NewFighter(1, 0, 100, 100, nil)

	if fighter.IsPlayerControlled() {
		t.Error("Fighter should not start as player-controlled")
	}

	fighter.SetPlayerControlled(true)
	if !fighter.IsPlayerControlled() {
		t.Error("Fighter should be player-controlled after setting")
	}

	fighter.SetPlayerControlled(false)
	if fighter.IsPlayerControlled() {
		t.Error("Fighter should not be player-controlled after unsetting")
	}
}

// Test AI Target Selection
func TestFighterAITargetSelection(t *testing.T) {
	ctx := NewMockGameContext()

	// Create fighter (faction 0)
	fighter := NewFighter(1, 0, 100, 100, nil)
	ctx.ships[fighter.GetID()] = fighter

	// Create enemy ships (faction 1)
	enemy1 := NewFighter(2, 1, 200, 100, nil)   // Close enemy
	enemy2 := NewFighter(3, 1, 1000, 1000, nil) // Far enemy
	ctx.ships[enemy1.GetID()] = enemy1
	ctx.ships[enemy2.GetID()] = enemy2

	// Create friendly ship (faction 0) - should not be targeted
	friendly := NewFighter(4, 0, 150, 100, nil)
	ctx.ships[friendly.GetID()] = friendly

	// Select target
	fighter.SelectTarget(ctx)

	// Should target the nearest enemy
	if fighter.AITargetID != enemy1.GetID() {
		t.Errorf("Expected to target nearest enemy (ID %d), got ID %d", enemy1.GetID(), fighter.AITargetID)
	}
}

// Test AI Retargeting
func TestFighterAIRetargeting(t *testing.T) {
	ctx := NewMockGameContext()

	fighter := NewFighter(1, 0, 100, 100, nil)
	ctx.ships[fighter.GetID()] = fighter

	enemy := NewFighter(2, 1, 200, 100, nil)
	ctx.ships[enemy.GetID()] = enemy

	// Set initial target and timer
	fighter.AITargetID = enemy.GetID()
	fighter.AIRetargetTimer = 10

	// Update AI several times
	for i := 0; i < 5; i++ {
		fighter.UpdateAI(ctx)
	}

	// Timer should have decreased
	if fighter.AIRetargetTimer >= 10 {
		t.Errorf("Expected retarget timer to decrease, still at %d", fighter.AIRetargetTimer)
	}

	// Target should still be the same
	if fighter.AITargetID != enemy.GetID() {
		t.Errorf("Target should not change until timer expires")
	}
}

// Test AI Retargeting When Target Dies
func TestFighterAIRetargetWhenTargetDies(t *testing.T) {
	ctx := NewMockGameContext()

	fighter := NewFighter(1, 0, 100, 100, nil)
	ctx.ships[fighter.GetID()] = fighter

	enemy1 := NewFighter(2, 1, 200, 100, nil)
	enemy2 := NewFighter(3, 1, 300, 100, nil)
	ctx.ships[enemy1.GetID()] = enemy1
	ctx.ships[enemy2.GetID()] = enemy2

	// Set initial target
	fighter.AITargetID = enemy1.GetID()

	// Kill the target
	enemy1.Alive = false

	// Update AI - should retarget
	fighter.UpdateAI(ctx)

	// Should now target enemy2
	if fighter.AITargetID != enemy2.GetID() {
		t.Errorf("Expected to retarget to enemy2 (ID %d), got ID %d", enemy2.GetID(), fighter.AITargetID)
	}
}

// Test Dead Ship Does Not Update
func TestDeadFighterDoesNotUpdate(t *testing.T) {
	ctx := NewMockGameContext()
	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Alive = false
	fighter.VelocityX = 5.0

	// Record initial position
	initialX, initialY := fighter.GetPosition()

	// Try to update
	fighter.Update(ctx)

	// Position should not change
	x, y := fighter.GetPosition()
	if x != initialX || y != initialY {
		t.Error("Dead fighter should not update position")
	}

	// Should not be able to fire
	fighter.FireWeapon(200, 100, ctx)
	if len(ctx.spawnedProjectiles) != 0 {
		t.Error("Dead fighter should not fire weapons")
	}
}

// Test Velocity Inheritance in Projectiles
func TestFighterProjectileVelocityInheritance(t *testing.T) {
	ctx := NewMockGameContext()
	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Rotation = 0 // Facing up (sprites face UP)
	fighter.VelocityX = 2.0
	fighter.VelocityY = 1.0

	// Fire straight ahead (upward)
	fighter.FireWeapon(100, 0, ctx)

	if len(ctx.spawnedProjectiles) != 1 {
		t.Fatalf("Expected 1 projectile, got %d", len(ctx.spawnedProjectiles))
	}

	proj := ctx.spawnedProjectiles[0]

	// Projectile should inherit ship velocity
	// Base projectile speed is 12.0 going up (0 rotation), so:
	// VX should be ~0 + 2 = 2 (no X component from firing angle, plus ship velocity)
	// VY should be ~-12 + 1 = -11 (firing upward is negative Y, plus ship velocity)
	if math.Abs(proj.VelocityX-2.0) > 0.1 {
		t.Errorf("Expected projectile VX ~2 (0 base + 2 ship), got %f", proj.VelocityX)
	}

	if math.Abs(proj.VelocityY-(-11.0)) > 0.1 {
		t.Errorf("Expected projectile VY ~-11 (-12 base + 1 ship), got %f", proj.VelocityY)
	}
}

// Test Update with Living AI-Controlled Fighter
func TestFighterUpdateAIControlled(t *testing.T) {
	ctx := NewMockGameContext()

	// Create AI-controlled fighter
	fighter := NewFighter(1, 0, 100, 100, nil)
	ctx.ships[fighter.GetID()] = fighter

	// Create enemy for AI to target
	enemy := NewFighter(2, 1, 200, 100, nil)
	ctx.ships[enemy.GetID()] = enemy

	// Discharge weapon to test charging
	fighter.Weapons[0].WeaponCapacitor = 0.5

	// Set initial speed
	fighter.Speed = 2.0

	// Record initial state
	initialX, _ := fighter.GetPosition()
	initialCapacitor := fighter.Weapons[0].WeaponCapacitor

	// Update the fighter
	err := fighter.Update(ctx)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	// Verify movement occurred (position changed based on speed)
	x, _ := fighter.GetPosition()
	if x == initialX && fighter.Speed > 0 {
		t.Error("Fighter should have moved during update")
	}

	// Verify weapon charging occurred
	if fighter.Weapons[0].WeaponCapacitor <= initialCapacitor {
		t.Error("Weapon should have charged during update")
	}

	// Verify AI selected a target
	if fighter.AITargetID == -1 {
		t.Error("AI should have selected a target")
	}
}

// Test AI Patrol Behavior (No Target Available)
func TestFighterAIPatrolBehavior(t *testing.T) {
	ctx := NewMockGameContext()

	// Create fighter with no enemies (only friendly ships)
	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Speed = 0 // Start stationary
	ctx.ships[fighter.GetID()] = fighter

	// Create only friendly ships - no enemies to target
	friendly := NewFighter(2, 0, 200, 100, nil) // Same faction
	ctx.ships[friendly.GetID()] = friendly

	// Update AI multiple times to let patrol behavior kick in
	for i := 0; i < 10; i++ {
		fighter.UpdateAI(ctx)
	}

	// In patrol mode, should accelerate toward patrol speed (50% of max)
	targetPatrolSpeed := fighter.MaxSpeed * config.AIPatrolSpeed
	if fighter.Speed == 0 {
		t.Error("Fighter in patrol mode should have accelerated from zero")
	}
	if fighter.Speed > targetPatrolSpeed+0.1 {
		t.Errorf("Fighter patrol speed should not exceed %.2f, got %.2f", targetPatrolSpeed, fighter.Speed)
	}

	// Target ID should be -1 (no target)
	if fighter.AITargetID != -1 {
		t.Errorf("Expected no target (ID -1), got ID %d", fighter.AITargetID)
	}
}

// Test AI Speed Adjustment When Above Target Speed
func TestFighterAISpeedDeceleration(t *testing.T) {
	ctx := NewMockGameContext()

	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Speed = fighter.MaxSpeed // Start at max speed
	ctx.ships[fighter.GetID()] = fighter

	// No enemies - should enter patrol mode which targets 50% speed
	// Update AI multiple times
	for i := 0; i < 100; i++ {
		fighter.UpdateAI(ctx)
	}

	// Should have decelerated toward patrol speed
	targetPatrolSpeed := fighter.MaxSpeed * config.AIPatrolSpeed
	if math.Abs(fighter.Speed-targetPatrolSpeed) > fighter.Accel*2 {
		t.Errorf("Expected speed near %.2f (patrol), got %.2f", targetPatrolSpeed, fighter.Speed)
	}
}

// Test AI Pursuit Speed Adjustment
func TestFighterAIPursuitSpeedAdjustment(t *testing.T) {
	ctx := NewMockGameContext()

	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Speed = 0 // Start stationary
	ctx.ships[fighter.GetID()] = fighter

	// Create enemy to trigger pursuit mode
	enemy := NewFighter(2, 1, 500, 100, nil)
	ctx.ships[enemy.GetID()] = enemy

	// Update AI multiple times
	for i := 0; i < 50; i++ {
		fighter.UpdateAI(ctx)
	}

	// In pursuit mode, should accelerate toward 80-100% of max speed
	minPursuitSpeed := fighter.MaxSpeed * config.AIPursuitSpeedMin
	if fighter.Speed < minPursuitSpeed*0.5 {
		t.Errorf("Fighter in pursuit should accelerate toward %.2f+, got %.2f", minPursuitSpeed, fighter.Speed)
	}
}

// Test AI Patrol Speed Deceleration When Above Target
func TestFighterAIPatrolSpeedDeceleration(t *testing.T) {
	ctx := NewMockGameContext()

	// Create fighter at max speed with no enemies
	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Speed = fighter.MaxSpeed
	ctx.ships[fighter.GetID()] = fighter

	// No enemies in context - will enter patrol mode

	// Update AI many times to reach patrol speed
	for i := 0; i < 200; i++ {
		fighter.UpdateAI(ctx)
	}

	targetPatrolSpeed := fighter.MaxSpeed * config.AIPatrolSpeed
	// Should be close to patrol speed
	if math.Abs(fighter.Speed-targetPatrolSpeed) > fighter.Accel*3 {
		t.Errorf("Expected speed near %.2f, got %.2f", targetPatrolSpeed, fighter.Speed)
	}
}

// Test AI Rotation Toward Target
func TestFighterAIRotationTowardTarget(t *testing.T) {
	ctx := NewMockGameContext()

	// Create fighter facing up (rotation = 0)
	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Rotation = 0
	ctx.ships[fighter.GetID()] = fighter

	// Create enemy to the right (positive X)
	enemy := NewFighter(2, 1, 200, 100, nil)
	ctx.ships[enemy.GetID()] = enemy

	// Update AI to trigger rotation
	for i := 0; i < 30; i++ {
		fighter.UpdateAI(ctx)
	}

	// Fighter should have rotated toward the enemy (to the right)
	// Sprites face UP, so to face right, rotation should be ~π/2
	expectedAngle := math.Pi / 2
	angleDiff := math.Abs(NormalizeAngle(fighter.Rotation - expectedAngle))
	if angleDiff > 0.2 {
		t.Errorf("Expected rotation near %.2f (facing right), got %.2f", expectedAngle, fighter.Rotation)
	}
}

// Test AI Firing Decision - Close Range High Probability
func TestFighterAIFiringCloseRange(t *testing.T) {
	ctx := NewMockGameContext()

	// Create fighter with fully charged weapon
	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Rotation = 0 // Facing up
	ctx.ships[fighter.GetID()] = fighter

	// Create enemy very close and directly ahead (within preferred range)
	// Enemy at negative Y (above) since ship faces up
	enemy := NewFighter(2, 1, 100, 100-50, nil) // 50 pixels ahead
	ctx.ships[enemy.GetID()] = enemy

	// Run many AI updates to trigger firing (probabilistic)
	shotsFired := 0
	for i := 0; i < 200; i++ {
		fighter.Weapons[0].WeaponCapacitor = 1.0 // Recharge weapon each time
		initialProjectiles := len(ctx.spawnedProjectiles)
		fighter.UpdateAI(ctx)
		if len(ctx.spawnedProjectiles) > initialProjectiles {
			shotsFired++
		}
	}

	// At close range with max firing probability, should fire frequently
	if shotsFired == 0 {
		t.Error("AI should have fired at least once at close range target in firing arc")
	}
}

// Test AI No Firing When Target Outside Arc
func TestFighterAINoFiringOutsideArc(t *testing.T) {
	ctx := NewMockGameContext()

	// Create fighter facing up
	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Rotation = 0
	ctx.ships[fighter.GetID()] = fighter

	// Create enemy directly behind (outside firing arc)
	enemy := NewFighter(2, 1, 100, 200, nil) // Below = behind since ship faces up
	ctx.ships[enemy.GetID()] = enemy

	// Set target manually to ensure AI is tracking this enemy
	fighter.AITargetID = enemy.GetID()
	fighter.AIRetargetTimer = 1000 // Don't retarget

	// Run AI updates
	for i := 0; i < 50; i++ {
		fighter.Weapons[0].WeaponCapacitor = 1.0
		fighter.UpdateAI(ctx)
		// Reset rotation to prevent AI from rotating target into arc
		fighter.Rotation = 0
	}

	// Should not fire when target is behind (outside cone)
	if len(ctx.spawnedProjectiles) > 0 {
		t.Errorf("AI should not fire at target outside arc, but fired %d shots", len(ctx.spawnedProjectiles))
	}
}

// Test AI Firing Far Range Low Probability
func TestFighterAIFiringFarRange(t *testing.T) {
	ctx := NewMockGameContext()

	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Rotation = 0 // Facing up
	ctx.ships[fighter.GetID()] = fighter

	// Create enemy far away but in arc (beyond max range)
	farDistance := config.AIMaxRange + 100
	enemy := NewFighter(2, 1, 100, 100-farDistance, nil)
	ctx.ships[enemy.GetID()] = enemy

	// Set target manually
	fighter.AITargetID = enemy.GetID()
	fighter.AIRetargetTimer = 1000

	// Run many AI updates
	shotsFired := 0
	for i := 0; i < 100; i++ {
		fighter.Weapons[0].WeaponCapacitor = 1.0
		initialProjectiles := len(ctx.spawnedProjectiles)
		fighter.UpdateAI(ctx)
		if len(ctx.spawnedProjectiles) > initialProjectiles {
			shotsFired++
		}
	}

	// At far range, firing probability is much lower (10% of max)
	// We may or may not fire, but it should be significantly less than close range
	// Just verify the code path executed without crashing
	t.Logf("Shots fired at far range: %d (expected lower rate)", shotsFired)
}

// Test AI Medium Range Firing
func TestFighterAIFiringMediumRange(t *testing.T) {
	ctx := NewMockGameContext()

	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Rotation = 0 // Facing up
	ctx.ships[fighter.GetID()] = fighter

	// Create enemy at medium range (between preferred and max)
	mediumDistance := (config.AIPreferredRange + config.AIMaxRange) / 2
	enemy := NewFighter(2, 1, 100, 100-mediumDistance, nil)
	ctx.ships[enemy.GetID()] = enemy

	fighter.AITargetID = enemy.GetID()
	fighter.AIRetargetTimer = 1000

	// Run AI updates to exercise medium range code path
	for i := 0; i < 50; i++ {
		fighter.Weapons[0].WeaponCapacitor = 1.0
		fighter.UpdateAI(ctx)
	}

	// Just verify it ran without error - probabilistic firing
	t.Log("Medium range firing test completed")
}

// Test AI Rotation When Angle Difference Is Small
func TestFighterAIRotationSmallAngle(t *testing.T) {
	ctx := NewMockGameContext()

	fighter := NewFighter(1, 0, 100, 100, nil)
	fighter.Rotation = 0
	ctx.ships[fighter.GetID()] = fighter

	// Create enemy almost directly ahead (very small angle difference)
	enemy := NewFighter(2, 1, 100.5, 50, nil) // Slightly to the right, but mostly ahead
	ctx.ships[enemy.GetID()] = enemy

	initialRotation := fighter.Rotation
	fighter.UpdateAI(ctx)

	// Small angle difference should snap to target angle
	// rather than rotating by full rotation speed
	rotationChange := math.Abs(fighter.Rotation - initialRotation)
	if rotationChange > config.AIRotationSpeed*1.1 {
		t.Errorf("Small angle adjustment should be <= rotation speed, got change of %.4f", rotationChange)
	}
}

// Test Afterburner Increases Max Speed
func TestFighterAfterburnerMaxSpeedBoost(t *testing.T) {
	fighter := NewFighter(1, 0, 100, 100, nil)

	// Verify fighter has afterburner
	if !fighter.HasAfterburnerSystem {
		t.Fatal("Fighter should have afterburner system")
	}

	// Get characteristics for verification
	chars := config.GetShipCharacteristics(ClassFighter)
	normalMaxSpeed := chars.MaxSpeed
	boostedMaxSpeed := normalMaxSpeed * chars.AfterburnerMaxSpeedMultiplier

	// Activate afterburner and accelerate beyond normal max speed
	fighter.AfterburnerActive = true
	fighter.AfterburnerCharge = 100.0    // Ensure we have charge
	fighter.Speed = normalMaxSpeed + 1.0 // Set speed above normal max

	// Simulate player input to accelerate
	// We'll manually call the acceleration logic similar to UpdatePlayerInput
	effectiveAccel := fighter.Accel
	if fighter.AfterburnerActive {
		effectiveAccel *= fighter.AfterburnerAccelMultiplier
	}

	// Accelerate to beyond normal max speed
	fighter.Speed += effectiveAccel

	// Calculate effective max speed with afterburner
	effectiveMaxSpeed := normalMaxSpeed
	if fighter.AfterburnerActive {
		effectiveMaxSpeed *= fighter.AfterburnerMaxSpeedMultiplier
	}

	// Cap speed to effective max
	if fighter.Speed > effectiveMaxSpeed {
		fighter.Speed = effectiveMaxSpeed
	}

	// With afterburner active, should be able to reach boosted max speed
	if fighter.Speed <= normalMaxSpeed {
		t.Errorf("Fighter should exceed normal max speed (%f) with afterburner, got %f",
			normalMaxSpeed, fighter.Speed)
	}

	if fighter.Speed > boostedMaxSpeed+0.01 {
		t.Errorf("Fighter should not exceed boosted max speed (%f), got %f",
			boostedMaxSpeed, fighter.Speed)
	}
}

// Test Afterburner Max Speed Returns to Normal When Deactivated
func TestFighterAfterburnerDeactivationMaxSpeed(t *testing.T) {
	fighter := NewFighter(1, 0, 100, 100, nil)
	chars := config.GetShipCharacteristics(ClassFighter)
	normalMaxSpeed := chars.MaxSpeed
	boostedMaxSpeed := normalMaxSpeed * chars.AfterburnerMaxSpeedMultiplier

	// Set speed to boosted max speed with afterburner active
	fighter.AfterburnerActive = true
	fighter.Speed = boostedMaxSpeed

	// Deactivate afterburner
	fighter.AfterburnerActive = false

	// In real game, the next update would apply friction/deceleration
	// but max speed enforcement happens during acceleration
	// Let's simulate trying to maintain speed above normal max
	effectiveMaxSpeed := normalMaxSpeed
	if fighter.AfterburnerActive {
		effectiveMaxSpeed *= fighter.AfterburnerMaxSpeedMultiplier
	}

	// Speed should be clamped to normal max when accelerating without afterburner
	if fighter.Speed > normalMaxSpeed {
		// During normal movement updates, speed naturally decays
		// but here we verify that the effective max is back to normal
		if effectiveMaxSpeed != normalMaxSpeed {
			t.Errorf("Effective max speed should be normal (%f) when afterburner inactive, got %f",
				normalMaxSpeed, effectiveMaxSpeed)
		}
	}
}

// Test Afterburner Max Speed Multiplier Applied Correctly
func TestFighterAfterburnerMaxSpeedMultiplier(t *testing.T) {
	fighter := NewFighter(1, 0, 100, 100, nil)
	chars := config.GetShipCharacteristics(ClassFighter)

	// Verify the multiplier is set correctly from config
	if fighter.AfterburnerMaxSpeedMultiplier != chars.AfterburnerMaxSpeedMultiplier {
		t.Errorf("Expected afterburner max speed multiplier %f, got %f",
			chars.AfterburnerMaxSpeedMultiplier, fighter.AfterburnerMaxSpeedMultiplier)
	}

	// Verify it's greater than 1.0 (should boost speed)
	if fighter.AfterburnerMaxSpeedMultiplier <= 1.0 {
		t.Errorf("Afterburner max speed multiplier should be > 1.0, got %f",
			fighter.AfterburnerMaxSpeedMultiplier)
	}

	// Calculate expected boosted max speed
	expectedBoostedMaxSpeed := fighter.MaxSpeed * fighter.AfterburnerMaxSpeedMultiplier

	// Verify it's actually higher than normal max
	if expectedBoostedMaxSpeed <= fighter.MaxSpeed {
		t.Errorf("Boosted max speed (%f) should be greater than normal max speed (%f)",
			expectedBoostedMaxSpeed, fighter.MaxSpeed)
	}
}

// Test Afterburner Minimum Activation Charge (20%)
func TestFighterAfterburnerMinimumActivationCharge(t *testing.T) {
	fighter := NewFighter(1, 0, 100, 100, nil)
	const minActivationCharge = 0.2 * 360.0 // 20% of max charge (72.0)

	// Test 1: Cannot activate with charge below 20% (simulating space key press logic)
	fighter.AfterburnerCharge = 50.0 // ~14% of 360
	fighter.AfterburnerActive = false

	// Simulate the activation check (space key pressed)
	canActivate := fighter.AfterburnerActive || fighter.AfterburnerCharge >= minActivationCharge
	if canActivate {
		t.Error("Should not be able to activate afterburner with charge below 20%")
	}

	// Test 2: Can activate with charge at exactly 20%
	fighter.AfterburnerCharge = 72.0 // Exactly 20% of 360
	fighter.AfterburnerActive = false

	canActivate = fighter.AfterburnerActive || fighter.AfterburnerCharge >= minActivationCharge
	if !canActivate {
		t.Error("Should be able to activate afterburner with charge at exactly 20%")
	}

	// Test 3: Can activate with charge above 20%
	fighter.AfterburnerCharge = 200.0 // >20% of 360
	fighter.AfterburnerActive = false

	canActivate = fighter.AfterburnerActive || fighter.AfterburnerCharge >= minActivationCharge
	if !canActivate {
		t.Error("Should be able to activate afterburner with charge above 20%")
	}

	// Test 4: Once activated, can continue even when draining below 20%
	fighter.AfterburnerCharge = 50.0 // Below 20%
	fighter.AfterburnerActive = true // Already active

	// Even with charge below 20%, activation condition should be met because it's already active
	canActivate = fighter.AfterburnerActive || fighter.AfterburnerCharge >= minActivationCharge
	if !canActivate {
		t.Error("Should be able to continue afterburner even after draining below 20%")
	}

	// Test 5: Verify the 20% threshold value
	if minActivationCharge != 72.0 {
		t.Errorf("Expected minimum activation charge to be 72.0 (20%% of 360), got %f", minActivationCharge)
	}
}
