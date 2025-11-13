package entity

import (
	"math"
	"testing"

	"github.com/nathan/verdant-thane/config"
)

// ============================================================================
// Basic Testudon Creation and Properties
// ============================================================================

func TestNewTestudon(t *testing.T) {
	testudon := NewTestudon(1, 0, 100, 200, nil)

	if testudon.ID != 1 {
		t.Errorf("Expected ID=1, got %d", testudon.ID)
	}
	if testudon.FactionID != 0 {
		t.Errorf("Expected FactionID=0, got %d", testudon.FactionID)
	}
	if testudon.Class != ClassTestudon {
		t.Errorf("Expected ClassTestudon, got %v", testudon.Class)
	}
	if testudon.X != 100 || testudon.Y != 200 {
		t.Errorf("Expected position (100,200), got (%f,%f)", testudon.X, testudon.Y)
	}
	if !testudon.Alive {
		t.Error("New testudon should be alive")
	}
	if testudon.PlayerControlled {
		t.Error("Testudons should never be player controlled")
	}
}

func TestTestudonCharacteristics(t *testing.T) {
	testudon := NewTestudon(1, 0, 0, 0, nil)
	chars := config.GetShipCharacteristics(ClassTestudon)

	// Verify testudon stats match config
	if testudon.MaxHealth != chars.MaxShield {
		t.Errorf("Expected MaxHealth=%d, got %d", chars.MaxShield, testudon.MaxHealth)
	}
	if testudon.Health != chars.MaxShield {
		t.Errorf("Expected Health=%d, got %d", chars.MaxShield, testudon.Health)
	}
	if testudon.MaxSpeed != chars.MaxSpeed {
		t.Errorf("Expected MaxSpeed=%f, got %f", chars.MaxSpeed, testudon.MaxSpeed)
	}
	if testudon.CollisionRadius != chars.CollisionRadius {
		t.Errorf("Expected CollisionRadius=%f, got %f", chars.CollisionRadius, testudon.CollisionRadius)
	}

	// Verify beam weapon characteristics
	if testudon.BeamRange != 200.0 {
		t.Errorf("Expected BeamRange=200.0, got %f", testudon.BeamRange)
	}
	// 0.125 = 1/8, exactly representable in binary (prevents FP accumulation errors)
	if testudon.BeamDamagePerTick != 0.125 {
		t.Errorf("Expected BeamDamagePerTick=0.125, got %f", testudon.BeamDamagePerTick)
	}
	if testudon.BeamTargetID != -1 {
		t.Errorf("Expected initial BeamTargetID=-1, got %d", testudon.BeamTargetID)
	}
	if testudon.BeamFiringAtID != -1 {
		t.Errorf("Expected initial BeamFiringAtID=-1, got %d", testudon.BeamFiringAtID)
	}

	// Verify no projectile weapons
	if len(testudon.Weapons) != 0 {
		t.Errorf("Testudons should not have projectile weapons, got %d weapons", len(testudon.Weapons))
	}
}

func TestTestudonNoProjectileWeapons(t *testing.T) {
	testudon := NewTestudon(1, 0, 0, 0, nil)
	ctx := NewMockGameContext()

	// FireWeapon should not spawn projectiles
	testudon.FireWeapon(100, 100, ctx)
	if len(ctx.spawnedProjectiles) > 0 {
		t.Error("Testudon should not spawn projectiles")
	}

	// FireMissile should not spawn missiles
	testudon.FireMissile(2, ctx)
	if len(ctx.spawnedMissiles) > 0 {
		t.Error("Testudon should not spawn missiles")
	}

	// CanFireWeapon should return false
	if testudon.CanFireWeapon() {
		t.Error("Testudon CanFireWeapon should return false")
	}

	// CanFireMissile should return false
	if testudon.CanFireMissile() {
		t.Error("Testudon CanFireMissile should return false")
	}
}

// ============================================================================
// Damage and Destruction
// ============================================================================

func TestTestudonTakeDamage(t *testing.T) {
	testudon := NewTestudon(1, 0, 100, 100, nil)
	ctx := NewMockGameContext()
	initialHealth := testudon.Health

	testudon.TakeDamage(3, 2, ctx)

	if testudon.Health != initialHealth-3 {
		t.Errorf("Expected health=%d, got %d", initialHealth-3, testudon.Health)
	}
	if !testudon.Alive {
		t.Error("Testudon should still be alive")
	}
}

func TestTestudonDestruction(t *testing.T) {
	testudon := NewTestudon(1, 0, 100, 100, nil)
	ctx := NewMockGameContext()

	testudon.TakeDamage(testudon.MaxHealth+10, 2, ctx)

	if testudon.Health != 0 {
		t.Errorf("Expected health=0, got %d", testudon.Health)
	}
	if testudon.Alive {
		t.Error("Testudon should be dead")
	}
	if len(ctx.spawnedExplosions) != 1 {
		t.Errorf("Expected 1 explosion, got %d", len(ctx.spawnedExplosions))
	}
	if ctx.spawnedExplosions[0].x != 100 || ctx.spawnedExplosions[0].y != 100 {
		t.Error("Explosion position mismatch")
	}
}

func TestTestudonKillByPlayerAwardsPoints(t *testing.T) {
	testudon := NewTestudon(1, 1, 100, 100, nil) // Enemy faction
	ctx := NewMockGameContext()

	// Create player ship as attacker
	player := NewFighter(2, 0, 200, 200, nil)
	player.PlayerControlled = true
	ctx.ships[2] = player

	testudon.TakeDamage(testudon.MaxHealth, 2, ctx)

	if ctx.killCount != 1 {
		t.Errorf("Expected 1 kill, got %d", ctx.killCount)
	}
	if ctx.scoreAdded != 50 {
		t.Errorf("Expected 50 points for testudon kill, got %d", ctx.scoreAdded)
	}
}

func TestTestudonTracksAttacker(t *testing.T) {
	testudon := NewTestudon(1, 0, 100, 100, nil)
	ctx := NewMockGameContext()

	// Take damage from attacker ID 5
	testudon.TakeDamage(1, 5, ctx)

	if len(testudon.AttackerIDs) != 1 {
		t.Fatalf("Expected 1 attacker tracked, got %d", len(testudon.AttackerIDs))
	}
	if testudon.AttackerIDs[0] != 5 {
		t.Errorf("Expected attacker ID 5, got %d", testudon.AttackerIDs[0])
	}

	// Take more damage from same attacker - should not duplicate
	testudon.TakeDamage(1, 5, ctx)
	if len(testudon.AttackerIDs) != 1 {
		t.Errorf("Should not duplicate attacker, got %d attackers", len(testudon.AttackerIDs))
	}

	// Take damage from different attacker
	testudon.TakeDamage(1, 7, ctx)
	if len(testudon.AttackerIDs) != 2 {
		t.Errorf("Expected 2 attackers tracked, got %d", len(testudon.AttackerIDs))
	}
}

func TestTestudonNoAttackerTrackingWhenDead(t *testing.T) {
	testudon := NewTestudon(1, 0, 100, 100, nil)
	ctx := NewMockGameContext()

	// Kill the testudon
	testudon.TakeDamage(testudon.MaxHealth, 2, ctx)

	// Further damage should not track attackers
	initialAttackerCount := len(testudon.AttackerIDs)
	testudon.TakeDamage(10, 3, ctx)

	if len(testudon.AttackerIDs) != initialAttackerCount {
		t.Error("Dead testudon should not track new attackers")
	}
}

// ============================================================================
// Priority-Based Targeting
// ============================================================================

func TestTestudonTargetingPriorityTestudons(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	// Add enemy ships of different classes
	fighter := NewFighter(2, 1, 600, 500, nil)
	destroyer := NewDestroyer(3, 1, 700, 500, nil)
	enemyTestudon := NewTestudon(4, 1, 800, 500, nil)

	ctx.ships[2] = fighter
	ctx.ships[3] = destroyer
	ctx.ships[4] = enemyTestudon

	testudon.SelectTarget(ctx)

	// Should target enemy testudon (highest priority)
	if testudon.BeamTargetID != 4 {
		t.Errorf("Expected to target testudon (ID 4), got ID %d", testudon.BeamTargetID)
	}
}

func TestTestudonTargetingPriorityDestroyers(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	// Add enemy ships - no testudons
	fighter := NewFighter(2, 1, 600, 500, nil)
	destroyer := NewDestroyer(3, 1, 700, 500, nil)

	ctx.ships[2] = fighter
	ctx.ships[3] = destroyer

	testudon.SelectTarget(ctx)

	// Should target destroyer (higher priority than fighter)
	if testudon.BeamTargetID != 3 {
		t.Errorf("Expected to target destroyer (ID 3), got ID %d", testudon.BeamTargetID)
	}
}

func TestTestudonTargetingPriorityFighters(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	// Add only enemy fighters
	fighter1 := NewFighter(2, 1, 600, 500, nil)
	fighter2 := NewFighter(3, 1, 700, 500, nil)

	ctx.ships[2] = fighter1
	ctx.ships[3] = fighter2

	testudon.SelectTarget(ctx)

	// Should target nearest fighter (ID 2 is closer)
	if testudon.BeamTargetID != 2 {
		t.Errorf("Expected to target nearest fighter (ID 2), got ID %d", testudon.BeamTargetID)
	}
}

func TestTestudonTargetingIgnoresFriendlies(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	// Add friendly ships only
	friendly1 := NewFighter(2, 0, 400, 500, nil)
	friendly2 := NewTestudon(3, 0, 300, 500, nil)

	ctx.ships[2] = friendly1
	ctx.ships[3] = friendly2

	testudon.SelectTarget(ctx)

	// Should not target friendlies
	if testudon.BeamTargetID != -1 {
		t.Errorf("Expected no target (ID -1), got ID %d", testudon.BeamTargetID)
	}
}

func TestTestudonTargetingIgnoresDead(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	// Add dead enemy
	enemy := NewFighter(2, 1, 600, 500, nil)
	enemy.Alive = false
	ctx.ships[2] = enemy

	testudon.SelectTarget(ctx)

	// Should not target dead ships
	if testudon.BeamTargetID != -1 {
		t.Errorf("Expected no target (ID -1), got ID %d", testudon.BeamTargetID)
	}
}

func TestTestudonDefensiveTargetingAttackers(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	// Add nearby high-priority target (testudon)
	enemyTestudon := NewTestudon(2, 1, 600, 500, nil)
	ctx.ships[2] = enemyTestudon

	// Add distant attacker (fighter)
	attacker := NewFighter(3, 1, 1000, 1000, nil)
	ctx.ships[3] = attacker

	// Take damage from distant fighter
	testudon.TakeDamage(1, 3, ctx)

	// Select target - should prioritize attacker over closer testudon
	testudon.SelectTarget(ctx)

	if testudon.BeamTargetID != 3 {
		t.Errorf("Expected to target attacker (ID 3), got ID %d", testudon.BeamTargetID)
	}
}

func TestTestudonDefensiveTargetingNearestAttacker(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	// Add multiple attackers at different distances
	attacker1 := NewFighter(2, 1, 1000, 1000, nil) // Far
	attacker2 := NewFighter(3, 1, 600, 500, nil)   // Near

	ctx.ships[2] = attacker1
	ctx.ships[3] = attacker2

	// Take damage from both
	testudon.TakeDamage(1, 2, ctx)
	testudon.TakeDamage(1, 3, ctx)

	testudon.SelectTarget(ctx)

	// Should target nearest attacker
	if testudon.BeamTargetID != 3 {
		t.Errorf("Expected to target nearest attacker (ID 3), got ID %d", testudon.BeamTargetID)
	}
}

func TestTestudonCleansUpDeadAttackers(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	// Add attackers
	attacker1 := NewFighter(2, 1, 600, 500, nil)
	attacker2 := NewFighter(3, 1, 700, 500, nil)
	ctx.ships[2] = attacker1
	ctx.ships[3] = attacker2

	// Take damage from both
	testudon.TakeDamage(1, 2, ctx)
	testudon.TakeDamage(1, 3, ctx)

	if len(testudon.AttackerIDs) != 2 {
		t.Fatalf("Expected 2 attackers, got %d", len(testudon.AttackerIDs))
	}

	// Kill one attacker
	attacker1.Alive = false

	// Select target - should clean up dead attacker
	testudon.SelectTarget(ctx)

	if len(testudon.AttackerIDs) != 1 {
		t.Errorf("Expected 1 attacker after cleanup, got %d", len(testudon.AttackerIDs))
	}
	if testudon.AttackerIDs[0] != 3 {
		t.Errorf("Expected remaining attacker ID 3, got %d", testudon.AttackerIDs[0])
	}
}

// ============================================================================
// Beam Weapon Mechanics
// ============================================================================

func TestTestudonBeamFiringAtInRangeTarget(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	// Add enemy in range
	enemy := NewFighter(2, 1, 550, 500, nil) // 50 pixels away, well within 200 range
	ctx.ships[2] = enemy

	testudon.BeamTargetID = 2
	testudon.UpdateBeamWeapon(ctx)

	if testudon.BeamFiringAtID != 2 {
		t.Errorf("Expected BeamFiringAtID=2, got %d", testudon.BeamFiringAtID)
	}
	if testudon.BeamDamageAccumulator == 0 {
		t.Error("Expected damage accumulator to increase")
	}
}

func TestTestudonBeamNotFiringAtOutOfRangeTarget(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	// Add enemy out of range
	enemy := NewFighter(2, 1, 1000, 1000, nil) // Far away, > 200 range
	ctx.ships[2] = enemy

	testudon.BeamTargetID = 2
	testudon.UpdateBeamWeapon(ctx)

	if testudon.BeamFiringAtID != -1 {
		t.Errorf("Expected BeamFiringAtID=-1 (not firing), got %d", testudon.BeamFiringAtID)
	}
	if testudon.BeamDamageAccumulator != 0 {
		t.Error("Expected damage accumulator to reset when not firing")
	}
}

func TestTestudonBeamDamageAccumulation(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	enemy := NewFighter(2, 1, 550, 500, nil)
	ctx.ships[2] = enemy // Add to context so UpdateBeamWeapon can find it

	testudon.BeamTargetID = 2
	initialHealth := enemy.Health

	// Fire beam for 8 ticks (0.125 damage/tick = 1.0 accumulated, exactly)
	for i := 0; i < 8; i++ {
		testudon.UpdateBeamWeapon(ctx)
	}

	// After 8 ticks at 0.125 damage/tick (1.0 accumulated), should have dealt 1 damage
	if enemy.Health != initialHealth-1 {
		t.Errorf("Expected health=%d after 8 ticks, got %d", initialHealth-1, enemy.Health)
	}

	// Fire for 8 more ticks (total 16 ticks = 2.0 damage)
	for i := 0; i < 8; i++ {
		testudon.UpdateBeamWeapon(ctx)
	}

	// Should have dealt 2 total damage
	if enemy.Health != initialHealth-2 {
		t.Errorf("Expected health=%d after 16 ticks, got %d", initialHealth-2, enemy.Health)
	}
}

func TestTestudonBeamKillsTarget(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	enemy := NewFighter(2, 1, 550, 500, nil)
	ctx.ships[2] = enemy

	testudon.BeamTargetID = 2

	// Fire beam until target dies (fighter has 8 HP = 64 ticks at 0.125 damage/tick)
	for i := 0; i < 100; i++ {
		if !enemy.IsAlive() {
			break
		}
		testudon.UpdateBeamWeapon(ctx)
	}

	if enemy.IsAlive() {
		t.Error("Beam should have killed target")
	}
	if testudon.BeamTargetID != -1 {
		t.Error("BeamTargetID should be cleared when target dies")
	}
	if testudon.BeamFiringAtID != -1 {
		t.Error("BeamFiringAtID should be cleared when target dies")
	}
}

func TestTestudonBeamOpportunisticFiring(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	// Primary target out of range
	primaryTarget := NewFighter(2, 1, 1000, 1000, nil)
	ctx.ships[2] = primaryTarget

	// Different enemy in range
	closeEnemy := NewFighter(3, 1, 550, 500, nil)
	ctx.ships[3] = closeEnemy

	testudon.BeamTargetID = 2 // Targeting out-of-range ship

	testudon.UpdateBeamWeapon(ctx)

	// Should opportunistically fire at closer enemy
	if testudon.BeamFiringAtID != 3 {
		t.Errorf("Expected opportunistic firing at ID 3, got ID %d", testudon.BeamFiringAtID)
	}
	if testudon.BeamDamageAccumulator == 0 {
		t.Error("Should be accumulating damage on opportunistic target")
	}
}

func TestTestudonBeamIgnoresFriendlies(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	// Friendly ship in range
	friendly := NewFighter(2, 0, 550, 500, nil)
	ctx.ships[2] = friendly

	testudon.BeamTargetID = 2

	testudon.UpdateBeamWeapon(ctx)

	// Should not fire at friendly
	if testudon.BeamFiringAtID != -1 {
		t.Error("Should not fire at friendly target")
	}
}

func TestTestudonGetBeamTargetID(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	enemy := NewFighter(2, 1, 550, 500, nil)
	ctx.ships[2] = enemy

	testudon.BeamTargetID = 2
	testudon.UpdateBeamWeapon(ctx)

	// GetBeamTargetID should return who we're firing at
	if testudon.GetBeamTargetID() != 2 {
		t.Errorf("Expected GetBeamTargetID()=2, got %d", testudon.GetBeamTargetID())
	}
}

// ============================================================================
// AI Movement and Rotation
// ============================================================================

func TestTestudonAIMovesTowardTarget(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	testudon.Rotation = 0 // Facing up
	ctx := NewMockGameContext()

	// Add enemy to the right
	enemy := NewFighter(2, 1, 700, 500, nil)
	ctx.ships[2] = enemy

	testudon.BeamTargetID = 2
	testudon.UpdateAI(ctx)

	// Should start moving
	if testudon.Speed == 0 {
		t.Error("Testudon should accelerate toward target")
	}

	// Should rotate toward target (to the right = positive rotation)
	expectedAngle := math.Pi / 2 // 90 degrees right
	angleDiff := math.Abs(NormalizeAngle(testudon.Rotation - expectedAngle))

	// Should be rotating toward target (not there instantly)
	if angleDiff < 0.01 {
		// Already perfectly aligned (could happen on first frame)
	} else if angleDiff > math.Pi/2 {
		t.Error("Testudon should be rotating toward target")
	}
}

func TestTestudonAIPatrolBehaviorNoTarget(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	testudon.BeamTargetID = -1
	initialRotation := testudon.Rotation

	testudon.UpdateAI(ctx)

	// Should patrol at reduced speed
	expectedPatrolSpeed := testudon.MaxSpeed * config.AIPatrolSpeed
	if testudon.Speed > expectedPatrolSpeed+0.01 {
		t.Errorf("Expected patrol speed ~%f, got %f", expectedPatrolSpeed, testudon.Speed)
	}

	// Should maintain heading (rotation unchanged)
	if testudon.Rotation != initialRotation {
		t.Error("Patrol behavior should maintain heading")
	}
}

func TestTestudonAIRetargeting(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	enemy := NewFighter(2, 1, 600, 500, nil)
	ctx.ships[2] = enemy

	// Set retarget timer to trigger
	testudon.AIRetargetTimer = 0
	testudon.UpdateAI(ctx)

	// Should have selected a target
	if testudon.BeamTargetID != 2 {
		t.Errorf("Expected to target enemy (ID 2), got ID %d", testudon.BeamTargetID)
	}

	// Retarget timer should be reset
	if testudon.AIRetargetTimer != config.AIRetargetInterval {
		t.Errorf("Expected retarget timer reset to %d, got %d",
			config.AIRetargetInterval, testudon.AIRetargetTimer)
	}
}

func TestTestudonAIRetargetsWhenTargetDies(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	ctx := NewMockGameContext()

	enemy1 := NewFighter(2, 1, 600, 500, nil)
	enemy2 := NewFighter(3, 1, 700, 500, nil)
	ctx.ships[2] = enemy1
	ctx.ships[3] = enemy2

	testudon.BeamTargetID = 2
	testudon.AIRetargetTimer = 30 // Not ready to retarget yet

	// Kill current target
	enemy1.Alive = false

	testudon.UpdateAI(ctx)

	// Should immediately retarget to other enemy
	if testudon.BeamTargetID != 3 {
		t.Errorf("Expected to retarget to enemy 3, got ID %d", testudon.BeamTargetID)
	}
}

// ============================================================================
// Movement and Physics
// ============================================================================

func TestTestudonMovement(t *testing.T) {
	testudon := NewTestudon(1, 0, 500, 500, nil)
	testudon.Rotation = 0 // Facing up
	testudon.Speed = 1.0
	testudon.VelocityY = -1.0

	initialY := testudon.Y
	testudon.UpdateMovement()

	// Should move upward (negative Y)
	if testudon.Y >= initialY {
		t.Error("Testudon should have moved upward")
	}
}

func TestTestudonWorldWrapping(t *testing.T) {
	testudon := NewTestudon(1, 0, 10, 500, nil)
	testudon.Rotation = -math.Pi / 2 // Facing left (west, 270 degrees)
	testudon.Speed = 20              // Moving at high speed

	testudon.UpdateMovement()

	// Should wrap to right side (X should be near GameWidth)
	// VelocityX = sin(-π/2) * 20 = -1 * 20 = -20 (moving left)
	// New X = 10 + (-20) = -10, which wraps to GameWidth - 10
	expectedX := float64(config.GameWidth) - 10
	if math.Abs(testudon.X-expectedX) > 1.0 {
		t.Errorf("Testudon should have wrapped to X≈%f, got X=%f",
			expectedX, testudon.X)
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestTestudonFullCombatScenario(t *testing.T) {
	ctx := NewMockGameContext()

	// Create testudon and enemies
	testudon := NewTestudon(1, 0, 500, 500, nil)
	fighter := NewFighter(2, 1, 550, 500, nil) // In beam range
	destroyer := NewDestroyer(3, 1, 600, 500, nil)

	ctx.ships[1] = testudon
	ctx.ships[2] = fighter
	ctx.ships[3] = destroyer

	// Run simulation for 100 ticks
	for tick := 0; tick < 100; tick++ {
		testudon.Update(ctx)
	}

	// Testudon should have:
	// 1. Selected a target (priority: destroyer > fighter)
	if testudon.BeamTargetID != 3 {
		t.Errorf("Expected testudon to target destroyer (ID 3), got ID %d", testudon.BeamTargetID)
	}

	// 2. Been moving and rotating
	if testudon.Speed == 0 {
		t.Error("Testudon should be moving")
	}

	// 3. Possibly damaged enemies if in range
	// (This depends on exact distances and movement, so we just verify
	// the beam weapon was attempting to fire)
	if testudon.BeamFiringAtID == -1 && Distance(testudon.X, testudon.Y, 600, 500) < 200 {
		t.Error("Testudon should be firing at in-range target")
	}
}

func TestTestudonVersusPlayer(t *testing.T) {
	ctx := NewMockGameContext()

	testudon := NewTestudon(1, 1, 500, 500, nil) // Enemy testudon
	player := NewFighter(2, 0, 550, 500, nil)    // Player in range
	player.PlayerControlled = true

	ctx.ships[1] = testudon
	ctx.ships[2] = player

	initialPlayerHealth := player.Health

	// Testudon attacks player
	testudon.SelectTarget(ctx)
	if testudon.BeamTargetID != 2 {
		t.Fatal("Testudon should target player")
	}

	// Fire beam for enough ticks to damage player
	for i := 0; i < 20; i++ {
		testudon.UpdateBeamWeapon(ctx)
	}

	// Player should have taken damage
	if player.Health >= initialPlayerHealth {
		t.Error("Player should have taken damage from testudon beam")
	}

	// Player destroys testudon
	testudon.TakeDamage(testudon.MaxHealth, 2, ctx)

	// Player should get points
	if ctx.scoreAdded != 50 {
		t.Errorf("Expected 50 points for killing testudon, got %d", ctx.scoreAdded)
	}
	if ctx.killCount != 1 {
		t.Errorf("Expected 1 kill, got %d", ctx.killCount)
	}
}

func TestTestudonDefensiveCombat(t *testing.T) {
	ctx := NewMockGameContext()

	testudon := NewTestudon(1, 0, 500, 500, nil)
	// Distant high-priority target
	enemyTestudon := NewTestudon(2, 1, 1000, 1000, nil)
	// Close attacker
	attacker := NewFighter(3, 1, 550, 500, nil)

	ctx.ships[1] = testudon
	ctx.ships[2] = enemyTestudon
	ctx.ships[3] = attacker

	// Initially targets high-priority testudon
	testudon.SelectTarget(ctx)
	if testudon.BeamTargetID != 2 {
		t.Fatal("Expected initial target to be enemy testudon")
	}

	// Get attacked by fighter
	testudon.TakeDamage(1, 3, ctx)

	// Should switch to defensive targeting
	testudon.SelectTarget(ctx)
	if testudon.BeamTargetID != 3 {
		t.Errorf("Expected defensive retarget to attacker (ID 3), got ID %d", testudon.BeamTargetID)
	}

	// Should be firing at attacker (in range)
	testudon.UpdateBeamWeapon(ctx)
	if testudon.BeamFiringAtID != 3 {
		t.Error("Should be firing at attacker")
	}
}
