package config

import "math"

// Screen dimensions
const (
	ScreenWidth  = 1024
	ScreenHeight = 768
)

// Game world dimensions
const (
	GameWidth  = 5040
	GameHeight = 5040
)

// Minimap dimensions and position
const (
	MinimapSize = 205                            // 20% of screen width (1024 * 0.2 ≈ 205)
	MinimapX    = ScreenWidth - MinimapSize - 10 // Bottom right corner with 10px margin
	MinimapY    = ScreenHeight - MinimapSize - 5 // Bottom with small margin
)

// Collision detection
const (
	ShipCollisionRadius       = 16.0 // Half the approximate sprite size
	ProjectileCollisionRadius = 2.0  // Small collision for projectiles
	CollisionGridSize         = 128  // Spatial grid cell size for broad-phase collision
)

// Player input
const (
	RotationSpeed = 3.0 * math.Pi / 180.0 // 3 degrees per tick
)

// AI behavior parameters
const (
	AIDecisionInterval = 60                    // Ticks between decisions (~1 second at 60 TPS)
	AIRetargetInterval = 60                    // Ticks between target re-evaluation (~1 second at 60 TPS)
	AIRotationSpeed    = 3.0 * math.Pi / 180.0 // 3 degrees per tick rotation toward target
	AIPursuitSpeedMin  = 0.80                  // Minimum speed when pursuing (80% of max)
	AIPursuitSpeedMax  = 1.00                  // Maximum speed when pursuing (100% of max)
	AIPatrolSpeed      = 0.50                  // Speed when no target (50% of max)

	// AI firing behavior (range-dependent)
	AIMaxFiringProbability = 1.0 / 15.0 // Max firing probability (1 in 15 frames when weapon charged)
	AIPreferredRange       = 350.0      // Range where AI fires at max probability (pixels)
	AIMaxRange             = 1000.0     // Range where AI firing probability drops to near zero (pixels)
)

// Explosion animation
const (
	ExplosionFrameCount = 4 // Total number of explosion animation frames
)

// Starfield rendering
const (
	StarDensity  = 0.0003 // stars per pixel
	StarGridSize = 200    // grid size for deterministic star generation
)

// Title screen layout constants
const (
	TitleY         = 150.0 // Y position of title text
	DialogWidth    = 300.0
	ButtonWidth    = 250.0
	ButtonHeight   = 40.0
	ButtonSpacing  = 15.0              // Vertical spacing between buttons
	InternalMargin = 2 * ButtonSpacing // Distance from top/bottom of dialog to first/last button
)
