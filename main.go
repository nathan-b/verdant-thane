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

	"github.com/nathan/verdant-thane/config"
	"github.com/nathan/verdant-thane/entity"
	"github.com/nathan/verdant-thane/persistence"
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
	useTestudon  = flag.Bool("testudon", false, "Guarantee each faction spawns with one AI-controlled testudon")
)

// GameState represents the current state of the game
type GameState int

const (
	TitleScreen GameState = iota
	InGame
	Victory
	GameOver
	Instructions
	HighScores
	Interstitial
)

// ProfileData tracks timing for performance profiling
type ProfileData struct {
	PlayerInput       time.Duration
	AIMovement        time.Duration
	WeaponsUpdate     time.Duration
	BeamWeapons       time.Duration
	Movement          time.Duration
	MissileTracking   time.Duration
	ProjectileLife    time.Duration
	Collisions        time.Duration
	EntityUpdate      time.Duration
	Explosions        time.Duration
	AIFiring          time.Duration
	RenderStars       time.Duration
	RenderShips       time.Duration
	RenderBeams       time.Duration
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
	p.BeamWeapons = 0
	p.Movement = 0
	p.MissileTracking = 0
	p.ProjectileLife = 0
	p.Collisions = 0
	p.Explosions = 0
	p.AIFiring = 0
	p.RenderStars = 0
	p.RenderShips = 0
	p.RenderBeams = 0
	p.RenderProjectiles = 0
	p.RenderExplosions = 0
	p.RenderMinimap = 0
	p.TotalUpdate = 0
	p.TotalDraw = 0
}

// Game represents the main game state
type Game struct {
	// Game state
	currentState       GameState
	titleDialog        *ui.Dialog              // Title screen dialog (only used in TitleScreen state)
	gameOverScreen     *ui.GameOverScreen      // Game over screen (only used in GameOver state)
	instructionsDialog *ui.Dialog              // Instructions screen dialog
	highScoresDialog   *ui.Dialog              // High scores screen dialog
	highScores         *persistence.HighScores // High scores loaded at game start
	battleNumber       int                     // Current battle number (1-indexed)
	nextFleetConfig    *config.FleetConfig     // Config for next battle (used by interstitial)

	// Entity system
	entityManager *EntityManager

	// Shared resources
	laserSprite     *ebiten.Image           // Shared sprite for all laser projectiles
	missileSprite   *ebiten.Image           // Shared sprite for all missile projectiles
	explosionSprite *ebiten.Image           // Sprite sheet for explosion animation
	factionSprites  *systems.FactionSprites // Ship sprites for all factions
	hudFont         *text.GoTextFace        // Font for HUD rendering

	// Camera
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

	// Input state tracking for spectate mode cycling
	prevKeyA        bool
	prevKeyD        bool
	prevKeySpace    bool
	lastProfileTime time.Time
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

	// Create instructions dialog
	instructionsDialog := ui.CreateInstructionsDialog()

	// Create high scores dialog
	highScoresDialog := ui.CreateHighScoresDialog()

	// Load high scores
	highScores, err := persistence.LoadHighScores()
	if err != nil {
		log.Printf("Warning: Failed to load high scores: %v", err)
		highScores = &persistence.HighScores{Entries: []persistence.HighScore{}}
	}

	// Create entity manager
	entityManager := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)

	return &Game{
		currentState:       TitleScreen,
		titleDialog:        titleDialog,
		gameOverScreen:     nil,
		instructionsDialog: instructionsDialog,
		highScoresDialog:   highScoresDialog,
		highScores:         highScores,
		entityManager:      entityManager,
		laserSprite:        laserSprite,
		missileSprite:      missileSprite,
		explosionSprite:    explosionSprite,
		factionSprites:     factionSprites,
		hudFont:            hudFont,
		lastFPSTime:        time.Now(),
		lastProfileTime:    time.Now(),
	}, nil
}

// StartGame transitions from title screen to in-game state by spawning ships
func (g *Game) StartGame(fleetConfig config.FleetConfig) error {
	// Clear entity manager for new game
	g.entityManager.Clear()

	// Initialize factions and spawn points
	g.entityManager.InitializeFactions()

	// Spawn ships according to fleet configuration
	const spawnRadius = 75.0 // Radius for circular spawn pattern
	var playerShip entity.Ship

	for factionID := 0; factionID < fleetConfig.NumFactions; factionID++ {
		comp := fleetConfig.Compositions[factionID]
		totalShips := comp.Total()

		// Build list of ship classes to spawn for this faction
		shipClasses := make([]entity.ShipClass, 0, totalShips)

		// Add fighters first (player will be first fighter of faction 0)
		for i := 0; i < comp.Fighters; i++ {
			shipClasses = append(shipClasses, entity.ClassFighter)
		}
		// Add destroyers
		for i := 0; i < comp.Destroyers; i++ {
			shipClasses = append(shipClasses, entity.ClassDestroyer)
		}
		// Add testudons
		for i := 0; i < comp.Testudons; i++ {
			shipClasses = append(shipClasses, entity.ClassTestudon)
		}

		// Debug flags can override ship classes
		if *useDestroyer {
			// Convert all fighters to destroyers (but not testudons)
			for i := range shipClasses {
				if shipClasses[i] == entity.ClassFighter {
					shipClasses[i] = entity.ClassDestroyer
				}
			}
		}
		if *useTestudon {
			// Add one testudon at index 1 if there are enough ships
			if len(shipClasses) > 1 {
				// Insert testudon at position 1 (second ship)
				shipClasses = append(shipClasses[:1], append([]entity.ShipClass{entity.ClassTestudon}, shipClasses[1:]...)...)
			}
		}

		// Get faction spawn point for offset calculations
		spawnX, spawnY, _ := g.entityManager.GetFactionSpawnPoint(factionID)

		// Spawn all ships for this faction
		for shipIndex, shipClass := range shipClasses {
			// First ship of faction 0 is player-controlled
			isPlayerControlled := (factionID == 0 && shipIndex == 0)

			// Calculate position with circular offset
			x, y := spawnX, spawnY
			if shipIndex > 0 {
				angleStep := 2.0 * math.Pi / float64(len(shipClasses)-1)
				angle := float64(shipIndex-1) * angleStep
				x += spawnRadius * math.Cos(angle)
				y += spawnRadius * math.Sin(angle)
			}

			// Spawn ship
			ship := g.entityManager.SpawnShip(shipClass, factionID, x, y)
			if isPlayerControlled {
				ship.SetPlayerControlled(true)
				g.entityManager.SetPlayerShip(ship.GetID())
				playerShip = ship
			}
		}
	}

	// Ensure we found a player ship
	if playerShip == nil {
		return fmt.Errorf("no player ship spawned (faction 0 must have at least 1 ship)")
	}

	// Initialize camera to follow player ship
	newX, newY := playerShip.GetPosition()
	g.cameraX = newX - float64(config.ScreenWidth)/2
	g.cameraY = newY - float64(config.ScreenHeight)/2

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
				fleetConfig := config.GenerateRandomFleetConfig(rand.Int63())
				if err := g.StartGame(fleetConfig); err != nil {
					return fmt.Errorf("failed to start game: %w", err)
				}
			} else if buttonIndex == 2 { // "Instructions" button
				g.currentState = Instructions
			} else if buttonIndex == 3 { // "High Scores" button
				g.currentState = HighScores
			}
			// TODO: Handle other button (Settings)
		}

	case InGame:
		updateStart := time.Now()

		// Update all entities (ships, projectiles, explosions, collisions)
		t := time.Now()
		g.entityManager.UpdateAll()
		g.profileData.EntityUpdate += time.Since(t)

		// Handle spectate mode controls
		if g.entityManager.IsSpectating() {
			// Update spectate mode validation
			g.entityManager.UpdateSpectateMode()

			// Handle input
			keyA := ebiten.IsKeyPressed(ebiten.KeyA)
			keyD := ebiten.IsKeyPressed(ebiten.KeyD)
			keySpace := ebiten.IsKeyPressed(ebiten.KeySpace)

			if keyD && !g.prevKeyD {
				// D key was just pressed - cycle to next allied ship
				g.entityManager.CycleSpectateNext()
			} else if keyA && !g.prevKeyA {
				// A key was just pressed - cycle to previous allied ship
				g.entityManager.CycleSpectatePrevious()
			}

			// Handle spacebar respawn (only on key press, not held)
			if keySpace && !g.prevKeySpace {
				// Space was just pressed - try to respawn into spectated ship
				g.entityManager.RespawnIntoSpectatedShip()
			}

			// Update previous key states
			g.prevKeyA = keyA
			g.prevKeyD = keyD
			g.prevKeySpace = keySpace
		} else {
			// Handle player firing
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				playerShip := g.entityManager.GetPlayerShip()
				if playerShip != nil && playerShip.IsAlive() {
					// Get mouse position in world coordinates
					mouseX, mouseY := ebiten.CursorPosition()
					worldMouseX := float64(mouseX) + g.cameraX
					worldMouseY := float64(mouseY) + g.cameraY

					// Fire weapon
					// (For destroyers, this may fire either the main gun or the missile)
					playerShip.FireWeapon(worldMouseX, worldMouseY, g.entityManager)
				}
			}
		}

		// Check for battle end conditions
		battleResult := g.entityManager.CheckBattleEnd()
		switch battleResult {
		case PlayerVictory:
			g.currentState = Victory
		case PlayerDefeat:
			g.currentState = GameOver
		}

		// Update camera to follow player or spectated ship
		var shipToFollow entity.Ship
		if g.entityManager.IsSpectating() {
			shipToFollow = g.entityManager.GetSpectatedShip()
		} else {
			shipToFollow = g.entityManager.GetPlayerShip()
		}

		if shipToFollow != nil {
			x, y := shipToFollow.GetPosition()
			g.cameraX = x - float64(config.ScreenWidth)/2
			g.cameraY = y - float64(config.ScreenHeight)/2
		}

		g.profileData.TotalUpdate += time.Since(updateStart)
		g.profileFrameCount++

	case Victory:
		// Generate next battle configuration on first entry to Victory state
		if g.nextFleetConfig == nil {
			// Increment battle number and generate harder fleet
			g.battleNumber++
			nextConfig := config.GenerateRandomFleetConfig(rand.Int63())
			g.nextFleetConfig = &nextConfig
		}

		// Transition to interstitial on key press
		if ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.currentState = Interstitial
		}

	case GameOver:
		// Create game over screen if not already created
		if g.gameOverScreen == nil {
			// Get player stats from EntityManager
			score, kills, deaths := g.entityManager.GetPlayerStats()

			// Check if it's a high score
			isHighScore := g.highScores.IsHighScore(score)

			// Get default player name
			playerName := persistence.GetCurrentUsername()

			// Create game over screen with ebitenui and button callback
			g.gameOverScreen = ui.NewGameOverScreen(playerName, score, kills, deaths, isHighScore, g.hudFont.Source, func() {
				// Save high score
				playerName := g.gameOverScreen.GetPlayerName()
				score, kills, deaths := g.entityManager.GetPlayerStats()
				g.highScores.AddScore(playerName, score, kills, deaths)
				if err := persistence.SaveHighScores(g.highScores); err != nil {
					log.Printf("Warning: Failed to save high scores: %v", err)
				}

				// Return to title screen and clear game over screen
				g.gameOverScreen = nil
				g.currentState = TitleScreen
			})
		}

		// Update UI (handles text input and button interactions)
		g.gameOverScreen.Update()

	case Instructions:
		// Handle instructions screen close button
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			mouseX, mouseY := ebiten.CursorPosition()
			if ui.CheckCloseButtonClick(g.instructionsDialog, mouseX, mouseY) {
				g.currentState = TitleScreen
			}
		}

		// Also handle ESC key
		if ebiten.IsKeyPressed(ebiten.KeyEscape) {
			g.currentState = TitleScreen
		}

	case HighScores:
		// Handle high scores screen close button
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			mouseX, mouseY := ebiten.CursorPosition()
			if ui.CheckCloseButtonClick(g.highScoresDialog, mouseX, mouseY) {
				g.currentState = TitleScreen
			}
		}

		// Also handle ESC key
		if ebiten.IsKeyPressed(ebiten.KeyEscape) {
			g.currentState = TitleScreen
		}

	case Interstitial:
		// Handle interstitial screen - start next battle
		if ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace) {
			// Start next battle
			if g.nextFleetConfig != nil {
				if err := g.StartGame(*g.nextFleetConfig); err != nil {
					log.Printf("Error starting next battle: %v", err)
					g.currentState = TitleScreen
				} else {
					// Clear next fleet config after using it
					g.nextFleetConfig = nil
				}
			} else {
				// Fallback: return to title if no config
				g.currentState = TitleScreen
			}
		}
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
		minGridX := int(cameraX) / config.StarGridSize
		maxGridX := int(cameraX+float64(config.ScreenWidth)) / config.StarGridSize
		minGridY := int(cameraY) / config.StarGridSize
		maxGridY := int(cameraY+float64(config.ScreenHeight)) / config.StarGridSize

		for gridX := minGridX; gridX <= maxGridX; gridX++ {
			for gridY := minGridY; gridY <= maxGridY; gridY++ {
				stars := systems.GenerateStarsForGrid(gridX, gridY)
				for _, star := range stars {
					screenX := star.X - cameraX
					screenY := star.Y - cameraY
					if screenX >= 0 && screenX < float64(config.ScreenWidth) && screenY >= 0 && screenY < float64(config.ScreenHeight) {
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
		titleX := (float64(config.ScreenWidth) - titleWidth) / 2
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
		minGridX := int(g.cameraX) / config.StarGridSize
		maxGridX := int(g.cameraX+float64(config.ScreenWidth)) / config.StarGridSize
		minGridY := int(g.cameraY) / config.StarGridSize
		maxGridY := int(g.cameraY+float64(config.ScreenHeight)) / config.StarGridSize

		// Draw stars for visible grid cells
		for gridX := minGridX; gridX <= maxGridX; gridX++ {
			for gridY := minGridY; gridY <= maxGridY; gridY++ {
				stars := systems.GenerateStarsForGrid(gridX, gridY)
				for _, star := range stars {
					// Convert star position to screen coordinates
					// We need to handle wrapping: stars might need to be drawn at wrapped positions
					drawStarAtPosition := func(worldX, worldY float64) {
						screenX := worldX - g.cameraX
						screenY := worldY - g.cameraY

						// Only draw if on screen
						if screenX >= 0 && screenX < float64(config.ScreenWidth) && screenY >= 0 && screenY < float64(config.ScreenHeight) {
							vector.FillRect(screen, float32(screenX), float32(screenY), 1, 1, color.White, false)
						}
					}

					// Draw star at its primary position
					drawStarAtPosition(star.X, star.Y)

					// Also check if we should draw the star at wrapped positions
					// This handles the case where the camera is near world boundaries
					if star.X < g.cameraX {
						// Star is to the left of camera, try drawing wrapped to the right
						drawStarAtPosition(star.X+float64(config.GameWidth), star.Y)
					}
					if star.X > g.cameraX+float64(config.ScreenWidth) {
						// Star is to the right of camera, try drawing wrapped to the left
						drawStarAtPosition(star.X-float64(config.GameWidth), star.Y)
					}
					if star.Y < g.cameraY {
						// Star is above camera, try drawing wrapped below
						drawStarAtPosition(star.X, star.Y+float64(config.GameHeight))
					}
					if star.Y > g.cameraY+float64(config.ScreenHeight) {
						// Star is below camera, try drawing wrapped above
						drawStarAtPosition(star.X, star.Y-float64(config.GameHeight))
					}
				}
			}
		}
		g.profileData.RenderStars += time.Since(starsStart)

		// Draw ships
		t := time.Now()
		for _, ship := range g.entityManager.ships {
			ship.Render(screen, g.cameraX, g.cameraY)
		}
		g.profileData.RenderShips += time.Since(t)

		// Draw projectiles
		t = time.Now()
		for _, proj := range g.entityManager.projectiles {
			proj.Render(screen, g.cameraX, g.cameraY)
		}
		g.profileData.RenderProjectiles += time.Since(t)

		// Draw explosions
		t = time.Now()
		for _, explosion := range g.entityManager.explosions {
			explosion.Render(screen, g.cameraX, g.cameraY)
		}
		g.profileData.RenderExplosions += time.Since(t)

		// Draw minimap
		t = time.Now()
		g.renderMinimap(screen)
		g.profileData.RenderMinimap += time.Since(t)

		// Draw HUD
		textColor := color.White

		// Get player state from EntityManager
		playerScore, playerKills, _ := g.entityManager.GetPlayerStats()
		var playerShield int
		playerShip := g.entityManager.GetPlayerShip()
		if playerShip != nil {
			playerShield, _ = playerShip.GetHealth()
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
		shieldOp.GeoM.Translate(float64(config.ScreenWidth)-shieldWidth-10, 10)
		shieldOp.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, shieldText, g.hudFont, shieldOp)

		// Upper right: Kills
		killsText := fmt.Sprintf("Kills: %d", playerKills)
		killsWidth, _ := text.Measure(killsText, g.hudFont, 0)
		killsOp := &text.DrawOptions{}
		killsOp.GeoM.Translate(float64(config.ScreenWidth)-killsWidth-10, 27)
		killsOp.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, killsText, g.hudFont, killsOp)

		// Spectate mode instructions (centered at bottom)
		if g.entityManager.IsSpectating() {
			instructionsLine1 := "A / D to switch ships"
			instructionsLine2 := "SPACE to take control"

			// Check if current ship is a testudon (can't respawn into it)
			spectatedShip := g.entityManager.GetSpectatedShip()
			if spectatedShip != nil && spectatedShip.GetClass() == entity.ClassTestudon {
				instructionsLine2 = "Cannot take control of Testudon"
			}

			// Draw first line
			line1Width, _ := text.Measure(instructionsLine1, g.hudFont, 0)
			line1Op := &text.DrawOptions{}
			line1Op.GeoM.Translate(float64(config.ScreenWidth/2)-line1Width/2, float64(config.ScreenHeight)-50)
			line1Op.ColorScale.ScaleWithColor(textColor)
			text.Draw(screen, instructionsLine1, g.hudFont, line1Op)

			// Draw second line
			line2Width, _ := text.Measure(instructionsLine2, g.hudFont, 0)
			line2Op := &text.DrawOptions{}
			line2Op.GeoM.Translate(float64(config.ScreenWidth/2)-line2Width/2, float64(config.ScreenHeight)-35)
			line2Op.ColorScale.ScaleWithColor(textColor)
			text.Draw(screen, instructionsLine2, g.hudFont, line2Op)
		}

		// FPS/TPS counter (if enabled)
		if *showFPS {
			fpsText := fmt.Sprintf("FPS: %.1f  TPS: %.1f", g.currentFPS, g.currentTPS)
			fpsOp := &text.DrawOptions{}
			fpsOp.GeoM.Translate(10, float64(config.ScreenHeight)-20)
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

	case Victory:
		// Draw stars background
		cameraX, cameraY := 0.0, 0.0
		minGridX := int(cameraX) / config.StarGridSize
		maxGridX := int(cameraX+float64(config.ScreenWidth)) / config.StarGridSize
		minGridY := int(cameraY) / config.StarGridSize
		maxGridY := int(cameraY+float64(config.ScreenHeight)) / config.StarGridSize
		centerX := float64(config.ScreenWidth) / 2.0
		centerY := float64(config.ScreenHeight) / 2.0

		for gridX := minGridX; gridX <= maxGridX; gridX++ {
			for gridY := minGridY; gridY <= maxGridY; gridY++ {
				stars := systems.GenerateStarsForGrid(gridX, gridY)
				for _, star := range stars {
					screenX := star.X - cameraX
					screenY := star.Y - cameraY
					if screenX >= 0 && screenX < float64(config.ScreenWidth) && screenY >= 0 && screenY < float64(config.ScreenHeight) {
						vector.FillRect(screen, float32(screenX), float32(screenY), 1, 1, color.White, false)
					}
				}
			}
		}

		// Display victory message
		textColor := color.RGBA{255, 255, 255, 255}
		goldColor := color.RGBA{255, 215, 0, 255}

		victoryText := fmt.Sprintf("BATTLE %d VICTORY!", g.battleNumber)
		victoryWidth, _ := text.Measure(victoryText, g.hudFont, 0)
		victoryOp := &text.DrawOptions{}
		victoryOp.GeoM.Translate(centerX-victoryWidth/2, centerY-40)
		victoryOp.ColorScale.ScaleWithColor(goldColor)
		text.Draw(screen, victoryText, g.hudFont, victoryOp)

		continueText := "Press ENTER to continue"
		continueWidth, _ := text.Measure(continueText, g.hudFont, 0)
		continueOp := &text.DrawOptions{}
		continueOp.GeoM.Translate(centerX-continueWidth/2, centerY+20)
		continueOp.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, continueText, g.hudFont, continueOp)

		// Stats
		statsY := centerY + 50
		playerScore, playerKills, playerDeaths := g.entityManager.GetPlayerStats()

		// Score
		scoreText := fmt.Sprintf("Score: %d", playerScore)
		scoreWidth, _ := text.Measure(scoreText, g.hudFont, 0)
		scoreOp := &text.DrawOptions{}
		scoreOp.GeoM.Translate(centerX-scoreWidth/2, statsY)
		scoreOp.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, scoreText, g.hudFont, scoreOp)

		// Kills
		killsText := fmt.Sprintf("Kills: %d", playerKills)
		killsWidth, _ := text.Measure(killsText, g.hudFont, 0)
		killsOp := &text.DrawOptions{}
		killsOp.GeoM.Translate(centerX-killsWidth/2, statsY+30)
		killsOp.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, killsText, g.hudFont, killsOp)

		// Deaths
		deathsText := fmt.Sprintf("Deaths: %d", playerDeaths)
		deathsWidth, _ := text.Measure(deathsText, g.hudFont, 0)
		deathsOp := &text.DrawOptions{}
		deathsOp.GeoM.Translate(centerX-deathsWidth/2, statsY+60)
		deathsOp.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, deathsText, g.hudFont, deathsOp)

	case GameOver:
		// Draw stars background
		cameraX, cameraY := 0.0, 0.0
		minGridX := int(cameraX) / config.StarGridSize
		maxGridX := int(cameraX+float64(config.ScreenWidth)) / config.StarGridSize
		minGridY := int(cameraY) / config.StarGridSize
		maxGridY := int(cameraY+float64(config.ScreenHeight)) / config.StarGridSize

		for gridX := minGridX; gridX <= maxGridX; gridX++ {
			for gridY := minGridY; gridY <= maxGridY; gridY++ {
				stars := systems.GenerateStarsForGrid(gridX, gridY)
				for _, star := range stars {
					screenX := star.X - cameraX
					screenY := star.Y - cameraY
					if screenX >= 0 && screenX < float64(config.ScreenWidth) && screenY >= 0 && screenY < float64(config.ScreenHeight) {
						vector.FillRect(screen, float32(screenX), float32(screenY), 1, 1, color.White, false)
					}
				}
			}
		}

		// Draw game over screen
		if g.gameOverScreen != nil {
			g.gameOverScreen.Draw(screen)
		}

	case Instructions:
		// Draw stars background
		cameraX, cameraY := 0.0, 0.0
		minGridX := int(cameraX) / config.StarGridSize
		maxGridX := int(cameraX+float64(config.ScreenWidth)) / config.StarGridSize
		minGridY := int(cameraY) / config.StarGridSize
		maxGridY := int(cameraY+float64(config.ScreenHeight)) / config.StarGridSize

		for gridX := minGridX; gridX <= maxGridX; gridX++ {
			for gridY := minGridY; gridY <= maxGridY; gridY++ {
				stars := systems.GenerateStarsForGrid(gridX, gridY)
				for _, star := range stars {
					screenX := star.X - cameraX
					screenY := star.Y - cameraY
					if screenX >= 0 && screenX < float64(config.ScreenWidth) && screenY >= 0 && screenY < float64(config.ScreenHeight) {
						vector.FillRect(screen, float32(screenX), float32(screenY), 1, 1, color.White, false)
					}
				}
			}
		}

		// Draw instructions dialog and content
		ui.RenderDialog(screen, g.instructionsDialog, g.hudFont)
		ui.DrawCloseButton(screen, g.instructionsDialog)
		ui.DrawInstructionsText(screen, g.hudFont.Source)

	case HighScores:
		// Draw stars background
		cameraX, cameraY := 0.0, 0.0
		minGridX := int(cameraX) / config.StarGridSize
		maxGridX := int(cameraX+float64(config.ScreenWidth)) / config.StarGridSize
		minGridY := int(cameraY) / config.StarGridSize
		maxGridY := int(cameraY+float64(config.ScreenHeight)) / config.StarGridSize

		for gridX := minGridX; gridX <= maxGridX; gridX++ {
			for gridY := minGridY; gridY <= maxGridY; gridY++ {
				stars := systems.GenerateStarsForGrid(gridX, gridY)
				for _, star := range stars {
					screenX := star.X - cameraX
					screenY := star.Y - cameraY
					if screenX >= 0 && screenX < float64(config.ScreenWidth) && screenY >= 0 && screenY < float64(config.ScreenHeight) {
						vector.FillRect(screen, float32(screenX), float32(screenY), 1, 1, color.White, false)
					}
				}
			}
		}

		// Draw high scores dialog and table
		ui.RenderDialog(screen, g.highScoresDialog, g.hudFont)
		ui.DrawCloseButton(screen, g.highScoresDialog)
		ui.DrawHighScoresTable(screen, g.hudFont.Source, g.highScores)

	case Interstitial:
		// Draw stars background
		cameraX, cameraY := 0.0, 0.0
		minGridX := int(cameraX) / config.StarGridSize
		maxGridX := int(cameraX+float64(config.ScreenWidth)) / config.StarGridSize
		minGridY := int(cameraY) / config.StarGridSize
		maxGridY := int(cameraY+float64(config.ScreenHeight)) / config.StarGridSize

		for gridX := minGridX; gridX <= maxGridX; gridX++ {
			for gridY := minGridY; gridY <= maxGridY; gridY++ {
				stars := systems.GenerateStarsForGrid(gridX, gridY)
				for _, star := range stars {
					screenX := star.X - cameraX
					screenY := star.Y - cameraY
					if screenX >= 0 && screenX < float64(config.ScreenWidth) && screenY >= 0 && screenY < float64(config.ScreenHeight) {
						vector.FillRect(screen, float32(screenX), float32(screenY), 1, 1, color.White, false)
					}
				}
			}
		}

		// Display interstitial information
		textColor := color.RGBA{255, 255, 255, 255}

		// Get player stats from EntityManager
		score, kills, _ := g.entityManager.GetPlayerStats()

		// Display current stats
		statsY := float64(config.ScreenHeight)/2 - 80
		statsTitle := fmt.Sprintf("Battle %d Complete!", g.battleNumber)
		statsTitleWidth, _ := text.Measure(statsTitle, g.hudFont, 0)
		statsTitleOp := &text.DrawOptions{}
		statsTitleOp.GeoM.Translate(float64(config.ScreenWidth/2)-statsTitleWidth/2, statsY)
		statsTitleOp.ColorScale.ScaleWithColor(color.RGBA{255, 215, 0, 255})
		text.Draw(screen, statsTitle, g.hudFont, statsTitleOp)

		scoreText := fmt.Sprintf("Current Score: %d", score)
		scoreWidth, _ := text.Measure(scoreText, g.hudFont, 0)
		scoreOp := &text.DrawOptions{}
		scoreOp.GeoM.Translate(float64(config.ScreenWidth/2)-scoreWidth/2, statsY+40)
		scoreOp.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, scoreText, g.hudFont, scoreOp)

		killsText := fmt.Sprintf("Total Kills: %d", kills)
		killsWidth, _ := text.Measure(killsText, g.hudFont, 0)
		killsOp := &text.DrawOptions{}
		killsOp.GeoM.Translate(float64(config.ScreenWidth/2)-killsWidth/2, statsY+70)
		killsOp.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, killsText, g.hudFont, killsOp)

		// Next battle info
		if g.nextFleetConfig != nil {
			nextBattleText := fmt.Sprintf("Next Battle: %d factions", g.nextFleetConfig.NumFactions)
			nextBattleWidth, _ := text.Measure(nextBattleText, g.hudFont, 0)
			nextBattleOp := &text.DrawOptions{}
			nextBattleOp.GeoM.Translate(float64(config.ScreenWidth/2)-nextBattleWidth/2, statsY+110)
			nextBattleOp.ColorScale.ScaleWithColor(textColor)
			text.Draw(screen, nextBattleText, g.hudFont, nextBattleOp)
		}

		continueText := "Press ENTER to continue"
		continueWidth, _ := text.Measure(continueText, g.hudFont, 0)
		continueOp := &text.DrawOptions{}
		continueOp.GeoM.Translate(float64(config.ScreenWidth/2)-continueWidth/2, statsY+150)
		continueOp.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, continueText, g.hudFont, continueOp)
	}
}

// renderMinimap renders the minimap using EntityManager
func (g *Game) renderMinimap(screen *ebiten.Image) {
	// Minimap constants
	const minimapSize = 120
	const minimapMargin = 10
	const minimapX = config.ScreenWidth - minimapSize - minimapMargin
	const minimapY = config.ScreenHeight - minimapSize - minimapMargin

	// Draw minimap background
	vector.DrawFilledRect(screen, minimapX, minimapY, minimapSize, minimapSize, color.RGBA{20, 20, 20, 200}, false)

	// Calculate scaling factors
	scaleX := float64(minimapSize) / float64(config.GameWidth)
	scaleY := float64(minimapSize) / float64(config.GameHeight)

	// Faction colors
	factionColors := []color.RGBA{
		{0, 255, 0, 255},   // Green (player faction)
		{0, 128, 255, 255}, // Blue
		{255, 0, 0, 255},   // Red
		{255, 255, 0, 255}, // Yellow
	}

	// Draw all ships as colored dots
	for _, ship := range g.entityManager.ships {
		if !ship.IsAlive() {
			continue
		}

		x, y := ship.GetPosition()
		minimapDotX := minimapX + float32(x*scaleX)
		minimapDotY := minimapY + float32(y*scaleY)

		factionID := ship.GetFaction()
		shipColor := factionColors[factionID%len(factionColors)]

		// Draw ship dot
		vector.FillRect(screen, minimapDotX-1, minimapDotY-1, 3, 3, shipColor, false)
	}

	// Draw player marker (light green plus sign)
	playerShip := g.entityManager.GetPlayerShip()
	if playerShip != nil && playerShip.IsAlive() {
		x, y := playerShip.GetPosition()
		playerX := minimapX + float32(x*scaleX)
		playerY := minimapY + float32(y*scaleY)

		playerMarkerColor := color.RGBA{100, 255, 100, 255}
		// Horizontal line
		vector.FillRect(screen, playerX-3, playerY, 7, 1, playerMarkerColor, false)
		// Vertical line
		vector.FillRect(screen, playerX, playerY-3, 1, 7, playerMarkerColor, false)
	}
}

// Layout returns the game's screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

func main() {
	flag.Parse()

	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle("Verdant Thane")

	game, err := NewGame()
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	// If performance testing mode is enabled, skip title screen and spawn large fleet
	if *perfTest {
		log.Printf("Performance test mode enabled: %d ships across %d factions", *perfShips, *perfFactions)

		// Calculate ships per faction
		baseShips := *perfShips / *perfFactions
		remainder := *perfShips % *perfFactions

		compositions := make([]config.FactionComposition, *perfFactions)
		for i := 0; i < *perfFactions; i++ {
			numShips := baseShips
			if i < remainder {
				numShips++
			}
			// For performance tests, all ships are fighters
			compositions[i] = config.FactionComposition{
				Fighters:   numShips,
				Destroyers: 0,
				Testudons:  0,
			}
		}

		fleetConfig := config.FleetConfig{
			NumFactions:  *perfFactions,
			Compositions: compositions,
		}

		if err := game.StartGame(fleetConfig); err != nil {
			log.Fatalf("Failed to start performance test: %v", err)
		}

		log.Printf("Fleet spawned successfully. Compositions: %+v", compositions)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
