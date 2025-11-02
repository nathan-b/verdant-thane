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

// Player state (singleton component)
type PlayerStateData struct {
	ControlledShip donburi.Entity
	Score          int
	Kills          int
}

var PlayerState = donburi.NewComponentType[PlayerStateData]()

// Tags for entity classification
var PlayerControlled = donburi.NewTag("PlayerControlled")
var AIControlled = donburi.NewTag("AIControlled")
var IsShip = donburi.NewTag("IsShip")
var IsProjectile = donburi.NewTag("IsProjectile")
