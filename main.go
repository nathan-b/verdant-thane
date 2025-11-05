package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"

	"github.com/nathan/verdant-thane/components"
	"github.com/nathan/verdant-thane/systems"
	"github.com/nathan/verdant-thane/ui"
)

var (
	// Command-line flags for performance testing
	perfTest     = flag.Bool("perf", false, "Enable performance testing mode with large fleets")
	perfShips    = flag.Int("ships", 800, "Total number of ships for performance testing")
	perfFactions = flag.Int("factions", 4, "Number of factions for performance testing")
	showFPS      = flag.Bool("fps", false, "Show FPS/TPS counter")
	showProfile  = flag.Bool("profile", false, "Show detailed performance profiling data")
	useDestroyer = flag.Bool("destroyer", false, "Spawn player and opponents as destroyers instead of fighters")
)

const (
	// Starfield constants
	starDensity  = 0.0003 // stars per pixel
	starGridSize = 200    // grid size for deterministic star generation
)

// FleetConfig defines the composition of ships across factions
type FleetConfig struct {
	// Number of factions participating (2-4)
	NumFactions int
	// Ships per faction (indexed by faction ID)
	ShipsPerFaction []int
}

// GenerateFleetConfig creates a deterministic fleet configuration
// for testing specific scenarios
func GenerateFleetConfig(numFactions, shipsPerFaction int) FleetConfig {
	if numFactions < 2 {
		numFactions = 2
	}
	if numFactions > 4 {
		numFactions = 4
	}
	if shipsPerFaction < 1 {
		shipsPerFaction = 1
	}

	ships := make([]int, numFactions)
	for i := 0; i < numFactions; i++ {
		ships[i] = shipsPerFaction
	}

	return FleetConfig{
		NumFactions:     numFactions,
		ShipsPerFaction: ships,
	}
}

// GenerateRandomFleetConfig creates a randomized fleet configuration
// using the provided seed for reproducibility
// Generates 2-4 factions with 7-16 ships each
func GenerateRandomFleetConfig(seed int64) FleetConfig {
	rng := rand.New(rand.NewSource(seed))

	// Random number of factions (2-4)
	numFactions := rng.Intn(3) + 2 // 2, 3, or 4

	// Random ships per faction (7-16 per faction)
	ships := make([]int, numFactions)
	for i := 0; i < numFactions; i++ {
		ships[i] = rng.Intn(10) + 7 // 7 to 16 inclusive
	}

	return FleetConfig{
		NumFactions:     numFactions,
		ShipsPerFaction: ships,
	}
}

// GameState represents the current state of the game
type GameState int

const (
	TitleScreen GameState = iota
	InGame
)

// ProfileData tracks timing for performance profiling
type ProfileData struct {
	PlayerInput       time.Duration
	AIMovement        time.Duration
	WeaponsUpdate     time.Duration
	Movement          time.Duration
	MissileTracking   time.Duration
	ProjectileLife    time.Duration
	Collisions        time.Duration
	Explosions        time.Duration
	AIFiring          time.Duration
	RenderStars       time.Duration
	RenderShips       time.Duration
	RenderProjectiles time.Duration
	RenderExplosions  time.Duration
	RenderMinimap     time.Duration
	TotalUpdate       time.Duration
	TotalDraw         time.Duration
}

// Reset clears all timing data
func (p *ProfileData) Reset() {
	p.PlayerInput = 0
	p.AIMovement = 0
	p.WeaponsUpdate = 0
	p.Movement = 0
	p.MissileTracking = 0
	p.ProjectileLife = 0
	p.Collisions = 0
	p.Explosions = 0
	p.AIFiring = 0
	p.RenderStars = 0
	p.RenderShips = 0
	p.RenderProjectiles = 0
	p.RenderExplosions = 0
	p.RenderMinimap = 0
	p.TotalUpdate = 0
	p.TotalDraw = 0
}

// Game represents the main game state
type Game struct {
	// Game state
	currentState GameState
	titleDialog  *ui.Dialog // Title screen dialog (only used in TitleScreen state)

	// ECS World (interface, not pointer)
	world             donburi.World
	playerEntity      donburi.Entity
	playerStateEntity donburi.Entity

	// Shared resources
	laserSprite     *ebiten.Image           // Shared sprite for all laser projectiles
	missileSprite   *ebiten.Image           // Shared sprite for all missile projectiles
	explosionSprite *ebiten.Image           // Sprite sheet for explosion animation
	factionSprites  *systems.FactionSprites // Ship sprites for all factions
	hudFont         *text.GoTextFace        // Font for HUD rendering

	// Camera (could be moved to ECS later)
	cameraX float64 // Camera position (follows player)
	cameraY float64

	// Performance monitoring
	frameCount  int
	lastFPSTime time.Time
	currentFPS  float64
	currentTPS  float64

	// Performance profiling
	profileData       ProfileData
	profileFrameCount int
	lastProfileTime   time.Time
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

// NewGame creates and initializes a new game, starting at the title screen
func NewGame() (*Game, error) {
	// Load shared assets
	laserSprite, _, err := ebitenutil.NewImageFromFile("assets/laser.png")
	if err != nil {
		return nil, err
	}

	missileSprite, _, err := ebitenutil.NewImageFromFile("assets/missile.png")
	if err != nil {
		return nil, err
	}

	// Load explosion sprite sheet (400x70, 4 frames of 100x70 each)
	explosionSprite, _, err := ebitenutil.NewImageFromFile("assets/explosion.png")
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

	// Create title screen dialog
	titleDialog := ui.CreateTitleScreen()

	return &Game{
		currentState:    TitleScreen,
		titleDialog:     titleDialog,
		laserSprite:     laserSprite,
		missileSprite:   missileSprite,
		explosionSprite: explosionSprite,
		factionSprites:  factionSprites,
		hudFont:         hudFont,
		lastFPSTime:     time.Now(),
		lastProfileTime: time.Now(),
	}, nil
}

// StartGame transitions from title screen to in-game state by spawning ships
func (g *Game) StartGame(fleetConfig FleetConfig) error {
	// Create ECS world
	g.world = donburi.NewWorld()

	// Initialize factions and spawn points
	systems.InitializeFactions(g.world)

	// Spawn ships according to fleet configuration
	const spawnRadius = 75.0 // Radius for circular spawn pattern
	var playerShip donburi.Entity
	var playerFound bool

	for factionID := 0; factionID < fleetConfig.NumFactions; factionID++ {
		numShips := fleetConfig.ShipsPerFaction[factionID]

		for shipIndex := 0; shipIndex < numShips; shipIndex++ {
			// First ship of faction 0 is player-controlled
			isPlayerControlled := (factionID == 0 && shipIndex == 0)

			// Determine ship class based on -destroyer flag
			shipClass := components.Fighter
			if *useDestroyer {
				shipClass = components.Destroyer
			}

			// Get ship characteristics from database
			shipChars := GetShipCharacteristics(shipClass)

			// Spawn ship at faction spawn point
			ship, err := systems.SpawnShip(g.world, systems.ShipConfig{
				Class:              shipClass,
				FactionID:          factionID,
				MaxSpeed:           shipChars.MaxSpeed,
				Acceleration:       shipChars.Acceleration,
				MaxHealth:          shipChars.MaxShield,
				CapacitorRate:      shipChars.CapacitorChargeRate,
				FiringCone:         shipChars.FiringCone,
				FactionSprites:     g.factionSprites,
				IsPlayerControlled: isPlayerControlled,
			})
			if err != nil {
				return fmt.Errorf("failed to spawn ship for faction %d: %w", factionID, err)
			}

			// Apply position offset for ships after the first in this faction
			// Arrange in circular pattern around spawn point
			if shipIndex > 0 {
				entry := g.world.Entry(ship)
				pos := components.Position.Get(entry)

				// Calculate angle for this ship in the circle
				angleStep := 2.0 * math.Pi / float64(numShips-1)
				angle := float64(shipIndex-1) * angleStep

				// Apply offset
				pos.X += spawnRadius * math.Cos(angle)
				pos.Y += spawnRadius * math.Sin(angle)
			}

			// Track player ship
			if isPlayerControlled {
				playerShip = ship
				playerFound = true
			}
		}
	}

	// Ensure we found a player ship
	if !playerFound {
		return fmt.Errorf("no player ship spawned (faction 0 must have at least 1 ship)")
	}

	// Create player state entity (singleton for score/kills tracking)
	playerState := g.world.Create(components.PlayerState)
	playerStateEntry := g.world.Entry(playerState)
	components.PlayerState.SetValue(playerStateEntry, components.PlayerStateData{
		ControlledShip: playerShip,
		Score:          0,
		Kills:          0,
	})

	// Get player position for camera initialization
	playerEntry := g.world.Entry(playerShip)
	playerPos := components.Position.Get(playerEntry)

	// Set game state
	g.playerEntity = playerShip
	g.playerStateEntity = playerState
	g.cameraX = playerPos.X - float64(systems.ScreenWidth)/2
	g.cameraY = playerPos.Y - float64(systems.ScreenHeight)/2
	g.currentState = InGame

	return nil
}

// Update updates the game logic using ECS systems
// This is called 60 times per second
func (g *Game) Update() error {
	// Update performance counters
	g.frameCount++
	if time.Since(g.lastFPSTime) >= time.Second {
		g.currentTPS = float64(g.frameCount) / time.Since(g.lastFPSTime).Seconds()
		g.currentFPS = ebiten.ActualFPS()
		g.frameCount = 0
		g.lastFPSTime = time.Now()
	}

	switch g.currentState {
	case TitleScreen:
		// Handle title screen button clicks
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			mouseX, mouseY := ebiten.CursorPosition()
			buttonIndex := ui.CheckButtonClick(g.titleDialog, mouseX, mouseY)

			if buttonIndex == 0 { // "Play Game" button
				// Generate random fleet configuration
				fleetConfig := GenerateRandomFleetConfig(rand.Int63())
				if err := g.StartGame(fleetConfig); err != nil {
					return fmt.Errorf("failed to start game: %w", err)
				}
			}
			// TODO: Handle other buttons (Settings, Instructions, High Scores)
		}

	case InGame:
		updateStart := time.Now()

		// Run systems in sequence with timing
		t := time.Now()
		systems.UpdatePlayerInput(g.world)
		g.profileData.PlayerInput += time.Since(t)

		t = time.Now()
		systems.UpdateAIMovement(g.world)
		g.profileData.AIMovement += time.Since(t)

		t = time.Now()
		systems.UpdateWeapons(g.world)
		g.profileData.WeaponsUpdate += time.Since(t)

		t = time.Now()
		systems.UpdateMovement(g.world)
		g.profileData.Movement += time.Since(t)

		t = time.Now()
		systems.UpdateMissileTracking(g.world)
		g.profileData.MissileTracking += time.Since(t)

		t = time.Now()
		systems.UpdateProjectileLifetime(g.world)
		g.profileData.ProjectileLife += time.Since(t)

		t = time.Now()
		systems.UpdateCollisions(g.world, g.explosionSprite)
		g.profileData.Collisions += time.Since(t)

		t = time.Now()
		systems.UpdateExplosions(g.world)
		g.profileData.Explosions += time.Since(t)

		// Handle player firing
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			// Get mouse position in world coordinates
			mouseX, mouseY := ebiten.CursorPosition()
			worldMouseX := float64(mouseX) + g.cameraX
			worldMouseY := float64(mouseY) + g.cameraY

			// Get player ship entry and attempt to fire
			if g.world.Valid(g.playerEntity) {
				playerEntry := g.world.Entry(g.playerEntity)

				// Check if cursor is within main weapon firing arc
				if systems.IsTargetInFiringArc(playerEntry, worldMouseX, worldMouseY) {
					// Cursor in front arc - fire main weapon (laser)
					systems.FireWeapon(g.world, playerEntry, worldMouseX, worldMouseY, g.laserSprite)
				} else if playerEntry.HasComponent(components.SecondaryWeapon) {
					// Cursor outside front arc and ship has missiles
					// Try to fire missile at nearest enemy in rear arc
					if nearestEnemy, found := systems.FindNearestEnemyInRearArc(g.world, playerEntry); found {
						systems.FireMissile(g.world, playerEntry, nearestEnemy, g.missileSprite)
					}
				}
			}
		}

		// Handle AI firing
		t = time.Now()
		systems.UpdateAIFiring(g.world, g.playerEntity, g.laserSprite, g.missileSprite)
		g.profileData.AIFiring += time.Since(t)

		// Update camera to follow player
		if g.world.Valid(g.playerEntity) {
			playerEntry := g.world.Entry(g.playerEntity)
			pos := components.Position.Get(playerEntry)
			g.cameraX = pos.X - float64(systems.ScreenWidth)/2
			g.cameraY = pos.Y - float64(systems.ScreenHeight)/2
		}

		g.profileData.TotalUpdate += time.Since(updateStart)
		g.profileFrameCount++
	}

	return nil
}

// Draw renders the game screen
func (g *Game) Draw(screen *ebiten.Image) {
	// Fill the screen with black
	screen.Fill(color.RGBA{0, 0, 0, 255})

	switch g.currentState {
	case TitleScreen:
		// Draw stars background (static, camera at origin)
		cameraX, cameraY := 0.0, 0.0
		minGridX := int(cameraX) / starGridSize
		maxGridX := int(cameraX+float64(systems.ScreenWidth)) / starGridSize
		minGridY := int(cameraY) / starGridSize
		maxGridY := int(cameraY+float64(systems.ScreenHeight)) / starGridSize

		for gridX := minGridX; gridX <= maxGridX; gridX++ {
			for gridY := minGridY; gridY <= maxGridY; gridY++ {
				stars := generateStarsForGrid(gridX, gridY)
				for _, star := range stars {
					screenX := star.x - cameraX
					screenY := star.y - cameraY
					if screenX >= 0 && screenX < float64(systems.ScreenWidth) && screenY >= 0 && screenY < float64(systems.ScreenHeight) {
						vector.FillRect(screen, float32(screenX), float32(screenY), 1, 1, color.White, false)
					}
				}
			}
		}

		// Draw title text
		titleText := "Verdant Thane"
		titleFont := &text.GoTextFace{
			Source: g.hudFont.Source,
			Size:   36,
		}
		titleWidth, _ := text.Measure(titleText, titleFont, 0)
		titleX := (float64(systems.ScreenWidth) - titleWidth) / 2
		titleY := ui.GetTitleY()

		titleOp := &text.DrawOptions{}
		titleOp.GeoM.Translate(titleX, titleY)
		titleOp.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, titleText, titleFont, titleOp)

		// Draw title screen dialog
		ui.RenderDialog(screen, g.titleDialog, g.hudFont)

	case InGame:
		drawStart := time.Now()

		// Draw stars
		starsStart := time.Now()
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
							vector.FillRect(screen, float32(screenX), float32(screenY), 1, 1, color.White, false)
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
		g.profileData.RenderStars += time.Since(starsStart)

		// Draw ships using ECS render system
		t := time.Now()
		systems.RenderShips(g.world, screen, g.cameraX, g.cameraY)
		g.profileData.RenderShips += time.Since(t)

		// Draw projectiles using ECS render system
		t = time.Now()
		systems.RenderProjectiles(g.world, screen, g.cameraX, g.cameraY)
		g.profileData.RenderProjectiles += time.Since(t)

		// Draw explosions using ECS render system
		t = time.Now()
		systems.RenderExplosions(g.world, screen, g.cameraX, g.cameraY)
		g.profileData.RenderExplosions += time.Since(t)

		// Draw minimap
		t = time.Now()
		systems.RenderMinimap(g.world, screen, g.playerEntity)
		g.profileData.RenderMinimap += time.Since(t)

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

		// FPS/TPS counter (if enabled)
		if *showFPS {
			fpsText := fmt.Sprintf("FPS: %.1f  TPS: %.1f", g.currentFPS, g.currentTPS)
			fpsOp := &text.DrawOptions{}
			fpsOp.GeoM.Translate(10, float64(systems.ScreenHeight)-20)
			fpsOp.ColorScale.ScaleWithColor(textColor)
			text.Draw(screen, fpsText, g.hudFont, fpsOp)
		}

		// Performance profiling
		g.profileData.TotalDraw += time.Since(drawStart)

		// Display profiling data every second (if enabled)
		if time.Since(g.lastProfileTime) >= time.Second {
			if *showProfile {
				// Calculate averages
				frameCount := float64(g.profileFrameCount)
				if frameCount == 0 {
					frameCount = 1
				}

				// Print profiling data to console
				fmt.Printf("\n=== Performance Profile (avg per frame) ===\n")
				fmt.Printf("Update Systems:\n")
				fmt.Printf("  Player Input:    %6.2f ms\n", float64(g.profileData.PlayerInput.Microseconds())/frameCount/1000.0)
				fmt.Printf("  AI Movement:     %6.2f ms\n", float64(g.profileData.AIMovement.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Weapons Update:  %6.2f ms\n", float64(g.profileData.WeaponsUpdate.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Movement:        %6.2f ms\n", float64(g.profileData.Movement.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Projectile Life: %6.2f ms\n", float64(g.profileData.ProjectileLife.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Collisions:      %6.2f ms\n", float64(g.profileData.Collisions.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Explosions:      %6.2f ms\n", float64(g.profileData.Explosions.Microseconds())/frameCount/1000.0)
				fmt.Printf("  AI Firing:       %6.2f ms\n", float64(g.profileData.AIFiring.Microseconds())/frameCount/1000.0)
				fmt.Printf("Render Systems:\n")
				fmt.Printf("  Stars:           %6.2f ms\n", float64(g.profileData.RenderStars.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Ships:           %6.2f ms\n", float64(g.profileData.RenderShips.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Projectiles:     %6.2f ms\n", float64(g.profileData.RenderProjectiles.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Explosions:      %6.2f ms\n", float64(g.profileData.RenderExplosions.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Minimap:         %6.2f ms\n", float64(g.profileData.RenderMinimap.Microseconds())/frameCount/1000.0)
				fmt.Printf("Totals:\n")
				fmt.Printf("  Total Update:    %6.2f ms\n", float64(g.profileData.TotalUpdate.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Total Draw:      %6.2f ms\n", float64(g.profileData.TotalDraw.Microseconds())/frameCount/1000.0)
				totalFrame := g.profileData.TotalUpdate + g.profileData.TotalDraw
				fmt.Printf("  Total Frame:     %6.2f ms\n", float64(totalFrame.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Frame Budget:    %6.2f ms (60 FPS)\n", 16.67)
				fmt.Printf("==========================================\n")
			}

			// Reset profiling data
			g.profileData.Reset()
			g.profileFrameCount = 0
			g.lastProfileTime = time.Now()
		}
	}
}

// Layout returns the game's screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return systems.ScreenWidth, systems.ScreenHeight
}

func main() {
	flag.Parse()

	ebiten.SetWindowSize(systems.ScreenWidth, systems.ScreenHeight)
	ebiten.SetWindowTitle("Verdant Thane")

	game, err := NewGame()
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	// If performance testing mode is enabled, skip title screen and spawn large fleet
	if *perfTest {
		log.Printf("Performance test mode enabled: %d ships across %d factions", *perfShips, *perfFactions)

		// Calculate ships per faction
		shipsPerFaction := make([]int, *perfFactions)
		baseShips := *perfShips / *perfFactions
		remainder := *perfShips % *perfFactions

		for i := 0; i < *perfFactions; i++ {
			shipsPerFaction[i] = baseShips
			if i < remainder {
				shipsPerFaction[i]++
			}
		}

		fleetConfig := FleetConfig{
			NumFactions:     *perfFactions,
			ShipsPerFaction: shipsPerFaction,
		}

		if err := game.StartGame(fleetConfig); err != nil {
			log.Fatalf("Failed to start performance test: %v", err)
		}

		log.Printf("Fleet spawned successfully. Ships per faction: %v", shipsPerFaction)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
