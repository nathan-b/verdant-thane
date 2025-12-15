//go:build !js && !wasm

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/nathan-b/verdant-thane/config"
)

var (
	// Command-line flags (for testing)
	perfTest       *bool
	perfShips      *int
	perfFactions   *int
	showFPS        *bool
	showProfile    *bool
	profileStartup *bool
	profileNewGame *bool
	useDestroyer   *bool
	useTestudon    *bool

	// Gameplay tweaking flags
	firingCone      *int
	firingRate      *float64
	turnRate        *float64
	projectileLife  *float64
	projectileSpeed *float64
	maxSpeed        *float64
	acceleration    *float64
	aiFire          *float64

	// Version flag
	showVersion *bool
)

// initFlags initializes all command-line flags for desktop builds
func initFlags() {
	perfTest = flag.Bool("perf", false, "Enable performance testing mode with large fleets")
	perfShips = flag.Int("ships", 800, "Total number of ships for performance testing")
	perfFactions = flag.Int("factions", 4, "Number of factions for performance testing")
	showFPS = flag.Bool("fps", false, "Show FPS/TPS counter")
	showProfile = flag.Bool("profile", false, "Show detailed in-game performance profiling data")
	profileStartup = flag.Bool("profile-startup", false, "Profile startup time from main() to title screen")
	profileNewGame = flag.Bool("profile-newgame", false, "Profile new game initialization time")
	useDestroyer = flag.Bool("destroyer", false, "Spawn player and opponents as destroyers instead of fighters")
	useTestudon = flag.Bool("testudon", false, "Guarantee each faction spawns with one AI-controlled testudon")

	firingCone = flag.Int("fc", 0, "Fighter firing cone in degrees (1-180, 0=use default)")
	firingRate = flag.Float64("fr", 0, "Fighter firing rate in shots per second (1-10, 0=use default)")
	turnRate = flag.Float64("tr", 0, "Fighter turn rate in degrees per tick (1-10, 0=use default)")
	projectileLife = flag.Float64("pl", 0, "Main gun projectile lifetime in seconds (0.5-10, 0=use default)")
	projectileSpeed = flag.Float64("ps", 0, "Main gun projectile speed in pixels per tick (8-16, 0=use default)")
	maxSpeed = flag.Float64("ms", 0, "Fighter max speed in pixels per tick (4-10, 0=use default)")
	acceleration = flag.Float64("ac", 0, "Fighter acceleration in pixels per second (2-10, 0=use default)")
	aiFire = flag.Float64("af", 0, "AI firing probability (0.1-1.0, 0=use default)")

	showVersion = flag.Bool("version", false, "Show version information and exit")
}

// parseFlags parses command-line flags and applies configuration overrides
func parseFlags() {
	initFlags()
	flag.Parse()

	// Show version if requested
	if *showVersion {
		fmt.Printf("Verdant Thane %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
		os.Exit(0)
	}

	// Apply gameplay tweaks if specified
	if *firingCone > 0 || *firingRate > 0 || *turnRate > 0 || *projectileLife > 0 || *projectileSpeed > 0 || *maxSpeed > 0 || *acceleration > 0 || *aiFire > 0 {
		if *firingCone > 0 {
			if *firingCone < 1 || *firingCone > 180 {
				log.Fatalf("Invalid firing cone: %d (must be 1-180)", *firingCone)
			}
			config.OverrideFighterFiringCone(*firingCone)
			log.Printf("Fighter firing cone override: %d degrees", *firingCone)
		}
		if *firingRate > 0 {
			if *firingRate < 1 || *firingRate > 10 {
				log.Fatalf("Invalid firing rate: %.2f (must be 1-10)", *firingRate)
			}
			config.OverrideFighterFiringRate(*firingRate)
			log.Printf("Fighter firing rate override: %.2f shots/sec", *firingRate)
		}
		if *turnRate > 0 {
			if *turnRate < 1 || *turnRate > 10 {
				log.Fatalf("Invalid turn rate: %.2f (must be 1-10)", *turnRate)
			}
			config.OverrideFighterTurnRate(*turnRate)
			log.Printf("Fighter turn rate override: %.2f degrees/tick", *turnRate)
		}
		if *projectileLife > 0 {
			if *projectileLife < 0.5 || *projectileLife > 10 {
				log.Fatalf("Invalid projectile lifetime: %.2f (must be 0.5-10)", *projectileLife)
			}
			config.OverrideMainGunProjectileLifetime(*projectileLife)
			log.Printf("Main gun projectile lifetime override: %.2f seconds", *projectileLife)
		}
		if *projectileSpeed > 0 {
			if *projectileSpeed < 8 || *projectileSpeed > 16 {
				log.Fatalf("Invalid projectile speed: %.2f (must be 8-16)", *projectileSpeed)
			}
			config.OverrideMainGunProjectileSpeed(*projectileSpeed)
			log.Printf("Main gun projectile speed override: %.2f pixels/tick", *projectileSpeed)
		}
		if *maxSpeed > 0 {
			if *maxSpeed < 4 || *maxSpeed > 10 {
				log.Fatalf("Invalid max speed: %.2f (must be 4-10)", *maxSpeed)
			}
			config.OverrideFighterMaxSpeed(*maxSpeed)
			log.Printf("Fighter max speed override: %.2f pixels/tick", *maxSpeed)
		}
		if *acceleration > 0 {
			if *acceleration < 2 || *acceleration > 10 {
				log.Fatalf("Invalid acceleration: %.2f (must be 2-10)", *acceleration)
			}
			config.OverrideFighterAcceleration(*acceleration)
			log.Printf("Fighter acceleration override: %.2f pixels/sec", *acceleration)
		}
		if *aiFire > 0 {
			if *aiFire < 0.1 || *aiFire > 1.0 {
				log.Fatalf("Invalid AI fire probability: %.2f (must be 0.1-1.0)", *aiFire)
			}
			config.OverrideAIFireProbability(*aiFire)
			log.Printf("AI fire probability override: %.2f", *aiFire)
		}
	}
}
