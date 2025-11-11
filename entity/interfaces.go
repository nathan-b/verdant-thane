package entity

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/nathan/verdant-thane/config"
)

// ShipClass is an alias to avoid import cycles
type ShipClass = config.ShipClass

// Re-export ship class constants for convenience
const (
	ClassFighter   = config.ClassFighter
	ClassDestroyer = config.ClassDestroyer
	ClassTestudon  = config.ClassTestudon
)

// Entity is the base interface that all game entities implement
type Entity interface {
	Update(ctx GameContext) error
	Render(screen *ebiten.Image, cameraX, cameraY float64)
	GetID() int
	GetPosition() (x, y float64)
	IsAlive() bool
}

// Ship extends Entity with ship-specific methods
type Ship interface {
	Entity

	// Combat
	TakeDamage(amount int, attackerID int, ctx GameContext)
	GetHealth() (current, max int)
	GetCollisionRadius() float64

	// Identity
	GetFaction() int
	GetClass() ShipClass

	// Physics
	GetVelocity() (vx, vy float64)
	GetRotation() float64

	// Control
	SetPlayerControlled(controlled bool)
	IsPlayerControlled() bool

	// Weapons
	FireWeapon(mouseX, mouseY float64, ctx GameContext)
	CanFireWeapon() bool

	// Update subsystems (for profiling)
	UpdateWeapons()
	UpdateAI(ctx GameContext)
	UpdatePlayerInput(ctx GameContext)
	UpdateMovement()
}

// Projectile extends Entity with projectile-specific methods
type Projectile interface {
	Entity
	GetOwnerID() int
	GetDamage() int
	CheckCollision(ship Ship) bool
	GetFaction() int // Inherited from owner
}

// GameContext provides entities with access to game state and operations
// This interface allows entities to interact with the game world without
// creating circular dependencies between packages
type GameContext interface {
	// Entity spawning
	SpawnProjectile(config MainGunConfig)
	SpawnMissile(config MissileConfig)
	SpawnExplosion(x, y float64)

	// Entity queries
	GetShip(id int) Ship
	GetAllShips() []Ship
	GetShipsByFaction(factionID int) []Ship
	FindNearestEnemy(ship Ship) (nearestShip Ship, distance float64)
	FindNearestEnemyInArc(ship Ship, arc, maxRange float64, rearFacing bool) (nearestShip Ship, distance float64)

	// World info
	GetWorldSize() (width, height float64)

	// Player state (for score tracking)
	AddKill()
	AddScore(points int)
}

// MainGunConfig contains parameters for spawning a main gun projectile
type MainGunConfig struct {
	X, Y                 float64
	VelocityX, VelocityY float64
	OwnerID              int
	FactionID            int
	Sprite               *ebiten.Image
}

// MissileConfig contains parameters for spawning a missile
type MissileConfig struct {
	X, Y                 float64
	VelocityX, VelocityY float64
	Rotation             float64
	OwnerID              int
	FactionID            int
	TargetID             int
	Sprite               *ebiten.Image
}
