package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"

	"github.com/nathan/verdant-thane/components"
	"github.com/nathan/verdant-thane/systems"
)

const (
	// Movement constants (for Fighter class)
	maxSpeed     = 6.0        // pixels per tick
	acceleration = 4.0 / 60.0 // pixels per second per tick

	// Firing constants (used for entity initialization)
	capacitorChargeRate = 1.0 / ((600.0 / 1000.0) * 60.0) // charge per tick (600ms charge time)
	firingConeAngle     = 30.0 * math.Pi / 180.0          // 30 degrees in radians

	// Starfield constants
	starDensity  = 0.0003 // stars per pixel
	starGridSize = 200    // grid size for deterministic star generation
)

// Game represents the main game state
type Game struct {
	// ECS World (interface, not pointer)
	world              donburi.World
	playerEntity       donburi.Entity
	playerStateEntity  donburi.Entity

	// Shared resources
	laserSprite    *ebiten.Image           // Shared sprite for all projectiles
	factionSprites *systems.FactionSprites // Ship sprites for all factions
	hudFont        *text.GoTextFace        // Font for HUD rendering

	// Camera (could be moved to ECS later)
	cameraX float64 // Camera position (follows player)
	cameraY float64
}

// modulo performs proper modulo operation (handles negatives correctly)
func modulo(a, b int) int {
	return ((a % b) + b) % b
}

// hashPosition creates a deterministic hash for a grid position
// Wraps grid coordinates to ensure consistent stars across world boundaries
func hashPosition(gridX, gridY int) int {
	// Calculate number of grid cells in the game world
	gridCountX := systems.GameWidth / starGridSize
	gridCountY := systems.GameHeight / starGridSize

	// Wrap grid coordinates to ensure tiling
	wrappedX := modulo(gridX, gridCountX)
	wrappedY := modulo(gridY, gridCountY)

	// Simple hash function for deterministic random generation
	h := wrappedX*73856093 ^ wrappedY*19349663
	if h < 0 {
		h = -h
	}
	return h
}

// generateStarsForGrid generates stars for a specific grid cell
func generateStarsForGrid(gridX, gridY int) []struct{ x, y float64 } {
	stars := []struct{ x, y float64 }{}

	// Use hash as seed for this grid cell (hash handles wrapping internally)
	seed := hashPosition(gridX, gridY)

	// Determine number of stars in this grid cell
	area := float64(starGridSize * starGridSize)
	numStars := int(area * starDensity)

	// Generate deterministic "random" positions within this grid
	for i := 0; i < numStars; i++ {
		// Simple LCG (Linear Congruential Generator) for deterministic randomness
		seed = (seed*1103515245 + 12345) & 0x7fffffff
		offsetX := float64(seed % starGridSize)

		seed = (seed*1103515245 + 12345) & 0x7fffffff
		offsetY := float64(seed % starGridSize)

		// Calculate base position for this grid cell
		baseX := float64(gridX * starGridSize)
		baseY := float64(gridY * starGridSize)

		// Star position in world coordinates
		x := baseX + offsetX
		y := baseY + offsetY

		stars = append(stars, struct{ x, y float64 }{x, y})
	}

	return stars
}

// NewGame creates and initializes a new game with ECS
func NewGame() (*Game, error) {
	// Create ECS world
	world := donburi.NewWorld()

	// Load shared assets
	laserSprite, _, err := ebitenutil.NewImageFromFile("assets/laser.png")
	if err != nil {
		return nil, err
	}

	// Load faction sprites
	factionSprites, err := systems.LoadFactionSprites()
	if err != nil {
		return nil, err
	}

	// Load font for HUD
	fontBytes, err := os.ReadFile("assets/orbitron.ttf")
	if err != nil {
		return nil, fmt.Errorf("failed to load font: %w", err)
	}

	fontSource, err := text.NewGoTextFaceSource(bytes.NewReader(fontBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	hudFont := &text.GoTextFace{
		Source: fontSource,
		Size:   14,
	}

	// Initialize factions and spawn points
	systems.InitializeFactions(world)

	// Spawn player ship at faction 0 (green) spawn point
	playerShip, err := systems.SpawnShip(world, systems.ShipConfig{
		Class:              components.Fighter,
		FactionID:          0, // Green team
		MaxSpeed:           maxSpeed,
		Acceleration:       acceleration,
		MaxHealth:          8,
		CapacitorRate:      capacitorChargeRate,
		FiringCone:         firingConeAngle,
		FactionSprites:     factionSprites,
		IsPlayerControlled: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to spawn player ship: %w", err)
	}

	// Create player state entity (singleton for score/kills tracking)
	playerState := world.Create(components.PlayerState)
	playerStateEntry := world.Entry(playerState)
	components.PlayerState.SetValue(playerStateEntry, components.PlayerStateData{
		ControlledShip: playerShip,
		Score:          0,
		Kills:          0,
	})

	// Get player position for camera initialization
	playerEntry := world.Entry(playerShip)
	playerPos := components.Position.Get(playerEntry)

	return &Game{
		world:             world,
		playerEntity:      playerShip,
		playerStateEntity: playerState,
		laserSprite:       laserSprite,
		factionSprites:    factionSprites,
		hudFont:           hudFont,
		cameraX:           playerPos.X - float64(systems.ScreenWidth)/2,
		cameraY:           playerPos.Y - float64(systems.ScreenHeight)/2,
	}, nil
}

// Update updates the game logic using ECS systems
// This is called 60 times per second
func (g *Game) Update() error {
	// Run systems in sequence
	systems.UpdatePlayerInput(g.world)
	systems.UpdateWeapons(g.world)
	systems.UpdateMovement(g.world)
	systems.UpdateProjectileLifetime(g.world)

	// Handle player firing
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		// Get mouse position in world coordinates
		mouseX, mouseY := ebiten.CursorPosition()
		worldMouseX := float64(mouseX) + g.cameraX
		worldMouseY := float64(mouseY) + g.cameraY

		// Get player ship entry and attempt to fire
		if g.world.Valid(g.playerEntity) {
			playerEntry := g.world.Entry(g.playerEntity)
			systems.FireWeapon(g.world, playerEntry, worldMouseX, worldMouseY, g.laserSprite)
		}
	}

	// Update camera to follow player
	if g.world.Valid(g.playerEntity) {
		playerEntry := g.world.Entry(g.playerEntity)
		pos := components.Position.Get(playerEntry)
		g.cameraX = pos.X - float64(systems.ScreenWidth)/2
		g.cameraY = pos.Y - float64(systems.ScreenHeight)/2
	}

	return nil
}

// Draw renders the game screen
func (g *Game) Draw(screen *ebiten.Image) {
	// Fill the screen with black
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Draw stars
	// Determine which grid cells are visible
	minGridX := int(g.cameraX) / starGridSize
	maxGridX := int(g.cameraX+float64(systems.ScreenWidth)) / starGridSize
	minGridY := int(g.cameraY) / starGridSize
	maxGridY := int(g.cameraY+float64(systems.ScreenHeight)) / starGridSize

	// Draw stars for visible grid cells
	for gridX := minGridX; gridX <= maxGridX; gridX++ {
		for gridY := minGridY; gridY <= maxGridY; gridY++ {
			stars := generateStarsForGrid(gridX, gridY)
			for _, star := range stars {
				// Convert star position to screen coordinates
				// We need to handle wrapping: stars might need to be drawn at wrapped positions
				drawStarAtPosition := func(worldX, worldY float64) {
					screenX := worldX - g.cameraX
					screenY := worldY - g.cameraY

					// Only draw if on screen
					if screenX >= 0 && screenX < float64(systems.ScreenWidth) && screenY >= 0 && screenY < float64(systems.ScreenHeight) {
						// TODO: Repalce DrawFilledRect with FillRect
						vector.DrawFilledRect(screen, float32(screenX), float32(screenY), 1, 1, color.White, false)
					}
				}

				// Draw star at its primary position
				drawStarAtPosition(star.x, star.y)

				// Also check if we should draw the star at wrapped positions
				// This handles the case where the camera is near world boundaries
				if star.x < g.cameraX {
					// Star is to the left of camera, try drawing wrapped to the right
					drawStarAtPosition(star.x+float64(systems.GameWidth), star.y)
				}
				if star.x > g.cameraX+float64(systems.ScreenWidth) {
					// Star is to the right of camera, try drawing wrapped to the left
					drawStarAtPosition(star.x-float64(systems.GameWidth), star.y)
				}
				if star.y < g.cameraY {
					// Star is above camera, try drawing wrapped below
					drawStarAtPosition(star.x, star.y+float64(systems.GameHeight))
				}
				if star.y > g.cameraY+float64(systems.ScreenHeight) {
					// Star is below camera, try drawing wrapped above
					drawStarAtPosition(star.x, star.y-float64(systems.GameHeight))
				}
			}
		}
	}

	// Draw ships using ECS render system
	systems.RenderShips(g.world, screen, g.cameraX, g.cameraY)

	// Draw projectiles using ECS render system
	systems.RenderProjectiles(g.world, screen, g.cameraX, g.cameraY)

	// Draw HUD
	textColor := color.White

	// Get player state from ECS
	var playerScore, playerKills, playerShield int
	if g.world.Valid(g.playerStateEntity) {
		stateEntry := g.world.Entry(g.playerStateEntity)
		state := components.PlayerState.Get(stateEntry)
		playerScore = state.Score
		playerKills = state.Kills

		// Get shield from player ship
		if g.world.Valid(state.ControlledShip) {
			shipEntry := g.world.Entry(state.ControlledShip)
			health := components.Health.Get(shipEntry)
			playerShield = health.Current
		}
	}

	// Upper left: Score
	scoreText := fmt.Sprintf("Score: %d", playerScore)
	scoreOp := &text.DrawOptions{}
	scoreOp.GeoM.Translate(10, 10)
	scoreOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, scoreText, g.hudFont, scoreOp)

	// Upper right: Shield
	shieldText := fmt.Sprintf("Shield: %d", playerShield)
	shieldWidth, _ := text.Measure(shieldText, g.hudFont, 0)
	shieldOp := &text.DrawOptions{}
	shieldOp.GeoM.Translate(float64(systems.ScreenWidth)-shieldWidth-10, 10)
	shieldOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, shieldText, g.hudFont, shieldOp)

	// Upper right: Kills
	killsText := fmt.Sprintf("Kills: %d", playerKills)
	killsWidth, _ := text.Measure(killsText, g.hudFont, 0)
	killsOp := &text.DrawOptions{}
	killsOp.GeoM.Translate(float64(systems.ScreenWidth)-killsWidth-10, 27)
	killsOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, killsText, g.hudFont, killsOp)
}

// Layout returns the game's screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return systems.ScreenWidth, systems.ScreenHeight
}

func main() {
	ebiten.SetWindowSize(systems.ScreenWidth, systems.ScreenHeight)
	ebiten.SetWindowTitle("Verdant Thane")

	game, err := NewGame()
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
