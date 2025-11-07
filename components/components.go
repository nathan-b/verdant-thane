package components

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

// Spatial components
type PositionData struct {
	X, Y float64
}

var Position = donburi.NewComponentType[PositionData]()

type VelocityData struct {
	X, Y float64
}

var Velocity = donburi.NewComponentType[VelocityData]()

type RotationData struct {
	Angle float64 // Radians
}

var Rotation = donburi.NewComponentType[RotationData]()

// Ship classes
type ShipClass int

const (
	Fighter ShipClass = iota
	Destroyer
	Testudon
	Mothership
)

// Ship-specific components
type ShipData struct {
	Class    ShipClass
	Speed    float64
	MaxSpeed float64
	Accel    float64
}

var Ship = donburi.NewComponentType[ShipData]()

type FactionData struct {
	ID int
}

var Faction = donburi.NewComponentType[FactionData]()

// FactionInfo stores information about a faction including its spawn point
type FactionInfoData struct {
	FactionID int
	SpawnX    float64
	SpawnY    float64
}

var FactionInfo = donburi.NewComponentType[FactionInfoData]()

type HealthData struct {
	Current int
	Max     int
}

var Health = donburi.NewComponentType[HealthData]()

type WeaponData struct {
	Capacitor  float64 // 0.0 to 1.0
	ChargeRate float64
	FiringCone float64 // Radians
}

var Weapon = donburi.NewComponentType[WeaponData]()

// SecondaryWeapon for destroyers (missiles)
type SecondaryWeaponData struct {
	Capacitor  float64 // 0.0 to 1.0
	ChargeRate float64
	FiringArc  float64 // Radians (rear 180° arc)
}

var SecondaryWeapon = donburi.NewComponentType[SecondaryWeaponData]()

// BeamWeapon for testudons (instantaneous beam attack)
type BeamWeaponData struct {
	TargetEntity      donburi.Entity // Current target being attacked
	Range             float64        // Maximum range of beam
	DamagePerTick     float64        // Damage dealt per tick (0.1 = 1 damage per 10 ticks)
	DamageAccumulator float64        // Accumulates partial damage
}

var BeamWeapon = donburi.NewComponentType[BeamWeaponData]()

// Rendering components
type SpriteData struct {
	Image *ebiten.Image
}

var Sprite = donburi.NewComponentType[SpriteData]()

// Projectile components
type ProjectileData struct {
	Lifetime    int
	MaxLifetime int
}

var Projectile = donburi.NewComponentType[ProjectileData]()

// Owner component tracks which entity created a projectile
type OwnerData struct {
	OwnerEntity donburi.Entity
}

var Owner = donburi.NewComponentType[OwnerData]()

// Missile component for tracking projectiles
type MissileData struct {
	TargetEntity donburi.Entity // Entity this missile is tracking
	Acceleration float64        // Pixels per tick² acceleration toward target
}

var Missile = donburi.NewComponentType[MissileData]()

// Player state (singleton component)
type PlayerStateData struct {
	ControlledShip donburi.Entity
	Score          int
	Kills          int
}

var PlayerState = donburi.NewComponentType[PlayerStateData]()

// AI state for decision making
type AIStateData struct {
	DecisionTimer int // Ticks until next decision (60 ticks ≈ 1 second)
	RetargetTimer int // Ticks until next target re-evaluation (180 ticks ≈ 3 seconds)
}

var AIState = donburi.NewComponentType[AIStateData]()

// AI targeting data
type AITargetData struct {
	TargetEntity donburi.Entity // The entity this AI is currently targeting
}

var AITarget = donburi.NewComponentType[AITargetData]()

// UnderAttack tracks entities that are currently attacking this ship (for Testudon defensive AI)
type UnderAttackData struct {
	Attackers []donburi.Entity // List of entities currently attacking
}

var UnderAttack = donburi.NewComponentType[UnderAttackData]()

// Explosion animation data
type ExplosionData struct {
	CurrentFrame int // Current animation frame (0-3)
	FrameTimer   int // Ticks until next frame
}

var Explosion = donburi.NewComponentType[ExplosionData]()

// Tags for entity classification
var PlayerControlled = donburi.NewTag("PlayerControlled")
var AIControlled = donburi.NewTag("AIControlled")
var IsShip = donburi.NewTag("IsShip")
var IsProjectile = donburi.NewTag("IsProjectile")
var IsMissile = donburi.NewTag("IsMissile")
var IsExplosion = donburi.NewTag("IsExplosion")
