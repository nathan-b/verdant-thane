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

	"github.com/nathan/verdant-thane/audio"
	"github.com/nathan/verdant-thane/config"
	"github.com/nathan/verdant-thane/entity"
	"github.com/nathan/verdant-thane/persistence"
	"github.com/nathan/verdant-thane/systems"
	"github.com/nathan/verdant-thane/ui"
)

var (
	// Command-line flags for performance testing
	perfTest       = flag.Bool("perf", false, "Enable performance testing mode with large fleets")
	perfShips      = flag.Int("ships", 800, "Total number of ships for performance testing")
	perfFactions   = flag.Int("factions", 4, "Number of factions for performance testing")
	showFPS        = flag.Bool("fps", false, "Show FPS/TPS counter")
	showProfile    = flag.Bool("profile", false, "Show detailed in-game performance profiling data")
	profileStartup = flag.Bool("profile-startup", false, "Profile startup time from main() to title screen")
	profileNewGame = flag.Bool("profile-newgame", false, "Profile new game initialization time")
	useDestroyer   = flag.Bool("destroyer", false, "Spawn player and opponents as destroyers instead of fighters")
	useTestudon    = flag.Bool("testudon", false, "Guarantee each faction spawns with one AI-controlled testudon")
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
	Settings
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

// Profiler interface implementation
func (p *ProfileData) RecordAIMovement(d time.Duration)      { p.AIMovement += d }
func (p *ProfileData) RecordWeaponsUpdate(d time.Duration)   { p.WeaponsUpdate += d }
func (p *ProfileData) RecordBeamWeapons(d time.Duration)     { p.BeamWeapons += d }
func (p *ProfileData) RecordMovement(d time.Duration)        { p.Movement += d }
func (p *ProfileData) RecordAIFiring(d time.Duration)        { p.AIFiring += d }
func (p *ProfileData) RecordMissileTracking(d time.Duration) { p.MissileTracking += d }
func (p *ProfileData) RecordProjectileLife(d time.Duration)  { p.ProjectileLife += d }
func (p *ProfileData) RecordCollisions(d time.Duration)      { p.Collisions += d }
func (p *ProfileData) RecordExplosions(d time.Duration)      { p.Explosions += d }

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

// StartupProfileData tracks timing for game startup (main() to title screen)
type StartupProfileData struct {
	LoadLaserSprite     time.Duration
	LoadMissileSprite   time.Duration
	LoadExplosionSprite time.Duration
	LoadFactionSprites  time.Duration
	LoadFont            time.Duration
	CreateUIDialogs     time.Duration
	LoadPersistence     time.Duration
	CreateEntityManager time.Duration
	CreateAudioManager  time.Duration
	LoadSoundEffects    time.Duration
	LoadMenuMusic       time.Duration
	CreateChatWindow    time.Duration
	LoadBGMTracks       time.Duration
	Total               time.Duration
}

// NewGameProfileData tracks timing for new game initialization (Play Game to game loaded)
type NewGameProfileData struct {
	ClearEntities      time.Duration
	InitializeFactions time.Duration
	SpawnShips         time.Duration
	InitializeCamera   time.Duration
	LoadGameMusic      time.Duration
	Total              time.Duration
}

// Game represents the main game state
type Game struct {
	// Game state
	currentState       GameState
	titleDialog        *ui.Dialog              // Title screen dialog
	gameOverScreen     *ui.GameOverScreen      // Game over screen
	instructionsDialog *ui.Dialog              // Instructions screen dialog
	highScoresDialog   *ui.Dialog              // High scores screen dialog
	settingsScreen     *ui.SettingsScreen      // Settings screen
	highScores         *persistence.HighScores // High scores loaded at game start
	settings           *persistence.Settings   // Game settings loaded at game start
	battleNumber       int                     // Current battle number (1-indexed)
	currentFleetConfig *config.FleetConfig     // Config for current battle (used for quick restart)
	nextFleetConfig    *config.FleetConfig     // Config for next battle (used by interstitial)

	// System managers
	entityManager *EntityManager
	chatWindow    *ChatWindow
	audioManager  *audio.Manager

	// Shared resources
	laserSprite     *ebiten.Image           // Shared sprite for all laser projectiles
	missileSprite   *ebiten.Image           // Shared sprite for all missile projectiles
	explosionSprite *ebiten.Image           // Sprite sheet for explosion animation
	factionSprites  *systems.FactionSprites // Ship sprites for all factions
	hudFont         *text.GoTextFace        // Font for HUD rendering
	factionColors   []color.RGBA            // Faction colors for rendering (beams, minimap, etc.)

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
	lastProfileTime   time.Time
	newGameProfile    NewGameProfileData

	// Pause state
	paused bool

	// Debounce control to avoid multiple toggles per key press
	prevKeyA     bool
	prevKeyD     bool
	prevKeyP     bool
	prevKeyN     bool
	prevKeyM     bool
	prevKeySpace bool
}

// NewGame creates and initializes a new game, starting at the title screen
func NewGame() (*Game, error) {
	var startupProfile StartupProfileData
	totalStart := time.Now()

	// Load shared assets
	t := time.Now()
	laserSprite, _, err := ebitenutil.NewImageFromFile("assets/laser.png")
	if err != nil {
		return nil, err
	}
	startupProfile.LoadLaserSprite = time.Since(t)

	t = time.Now()
	missileSprite, _, err := ebitenutil.NewImageFromFile("assets/missile.png")
	if err != nil {
		return nil, err
	}
	startupProfile.LoadMissileSprite = time.Since(t)

	// Load explosion sprite sheet (400x70, 4 frames of 100x70 each)
	t = time.Now()
	explosionSprite, _, err := ebitenutil.NewImageFromFile("assets/explosion.png")
	if err != nil {
		return nil, err
	}
	startupProfile.LoadExplosionSprite = time.Since(t)

	// Load faction sprites
	t = time.Now()
	factionSprites, err := systems.LoadFactionSprites()
	if err != nil {
		return nil, err
	}
	startupProfile.LoadFactionSprites = time.Since(t)

	// Load font for HUD
	t = time.Now()
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
	startupProfile.LoadFont = time.Since(t)

	// Create title screen dialog
	t = time.Now()
	titleDialog := ui.CreateTitleScreen()

	// Create instructions dialog
	instructionsDialog := ui.CreateInstructionsDialog()

	// Create high scores dialog
	highScoresDialog := ui.CreateHighScoresDialog()
	startupProfile.CreateUIDialogs = time.Since(t)

	// Load high scores
	t = time.Now()
	highScores, err := persistence.LoadHighScores()
	if err != nil {
		log.Printf("Warning: Failed to load high scores: %v", err)
		highScores = &persistence.HighScores{Entries: []persistence.HighScore{}}
	}

	// Load settings
	settings, err := persistence.LoadSettings()
	if err != nil {
		log.Printf("Warning: Failed to load settings: %v", err)
		settings = &persistence.Settings{SoundMuted: false, MusicMuted: false}
	}
	startupProfile.LoadPersistence = time.Since(t)

	// Create entity manager
	t = time.Now()
	entityManager := NewEntityManager(laserSprite, missileSprite, explosionSprite, factionSprites)
	startupProfile.CreateEntityManager = time.Since(t)

	// Create audio manager and load sound effects
	t = time.Now()
	audioManager, err := audio.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create audio manager: %w", err)
	}

	// Apply saved settings (mute and volume)
	audioManager.SetSoundMuted(settings.SoundMuted)
	audioManager.SetMusicMuted(settings.MusicMuted)
	audioManager.SetSoundVolume(settings.SoundVolume)
	audioManager.SetMusicVolume(settings.MusicVolume)
	startupProfile.CreateAudioManager = time.Since(t)

	// Load sound effects
	t = time.Now()
	if err := audioManager.LoadSound("laser", "assets/laser.wav"); err != nil {
		log.Printf("Warning: Failed to load laser sound: %v", err)
	}
	if err := audioManager.LoadSound("impact", "assets/impact.wav"); err != nil {
		log.Printf("Warning: Failed to load impact sound: %v", err)
	}
	if err := audioManager.LoadSound("explosion", "assets/explosion.wav"); err != nil {
		log.Printf("Warning: Failed to load explosion sound: %v", err)
	}
	startupProfile.LoadSoundEffects = time.Since(t)

	// Load menu music (needed immediately)
	t = time.Now()
	if err := audioManager.LoadMusic("menu", "assets/energy-electrowave.mp3"); err != nil {
		log.Printf("Warning: Failed to load menu music: %v", err)
	}
	startupProfile.LoadMenuMusic = time.Since(t)

	// Create chat window
	t = time.Now()
	chatWindow, err := NewChatWindow("assets/chatter.json")
	if err != nil {
		log.Printf("Warning: Failed to create chat window: %v", err)
		chatWindow = nil // Continue without chat
	}
	startupProfile.CreateChatWindow = time.Since(t)

	t = time.Now()
	game := &Game{
		currentState:       TitleScreen,
		titleDialog:        titleDialog,
		gameOverScreen:     nil,
		instructionsDialog: instructionsDialog,
		highScoresDialog:   highScoresDialog,
		settingsScreen:     nil,
		highScores:         highScores,
		settings:           settings,
		entityManager:      entityManager,
		chatWindow:         chatWindow,
		audioManager:       audioManager,
		laserSprite:        laserSprite,
		missileSprite:      missileSprite,
		explosionSprite:    explosionSprite,
		factionSprites:     factionSprites,
		hudFont:            hudFont,
		factionColors: []color.RGBA{
			{0, 255, 0, 255},   // Green (player faction)
			{0, 128, 255, 255}, // Blue
			{255, 0, 0, 255},   // Red
			{255, 255, 0, 255}, // Yellow
		},
		lastFPSTime:     time.Now(),
		lastProfileTime: time.Now(),
	}
	createGameStruct := time.Since(t)

	t = time.Now()
	// Set profiler on entity manager for performance tracking
	entityManager.SetProfiler(&game.profileData)

	// Set audio manager on entity manager for sound effects
	entityManager.SetAudioManager(audioManager)

	// Set chat window on entity manager for chat events
	if chatWindow != nil {
		entityManager.SetChatWindow(chatWindow)
	}
	wireUpGame := time.Since(t)

	// Load in-game music tracks
	t = time.Now()
	bgmTracks := []string{
		"assets/bgm/0-top.mp3",
		"assets/bgm/adrenaline-rush.mp3",
		"assets/bgm/crazy-bad.mp3",
		"assets/bgm/dance-with-demons.mp3",
		"assets/bgm/tank-metal.mp3",
	}
	for i, track := range bgmTracks {
		musicName := fmt.Sprintf("bgm%d", i)
		if err := audioManager.LoadMusic(musicName, track); err != nil {
			log.Printf("Warning: Failed to load BGM track %s: %v", track, err)
		}
	}
	startupProfile.LoadBGMTracks = time.Since(t)

	// Start menu music
	t = time.Now()
	if err := audioManager.PlayMusic("menu"); err != nil {
		log.Printf("Warning: Failed to play menu music: %v", err)
	}
	startMenuMusic := time.Since(t)

	// Calculate total startup time
	startupProfile.Total = time.Since(totalStart)

	// Print startup profiling data if enabled
	if *profileStartup {
		// Calculate total measured time
		measured := startupProfile.LoadLaserSprite +
			startupProfile.LoadMissileSprite +
			startupProfile.LoadExplosionSprite +
			startupProfile.LoadFactionSprites +
			startupProfile.LoadFont +
			startupProfile.CreateUIDialogs +
			startupProfile.LoadPersistence +
			startupProfile.CreateEntityManager +
			startupProfile.CreateAudioManager +
			startupProfile.LoadSoundEffects +
			startupProfile.LoadMenuMusic +
			startupProfile.CreateChatWindow +
			startupProfile.LoadBGMTracks +
			createGameStruct +
			wireUpGame +
			startMenuMusic

		unaccounted := startupProfile.Total - measured

		fmt.Printf("\n=== Startup Profiling (main() to title screen) ===\n")
		fmt.Printf("  Load Laser Sprite:      %6.2f ms\n", float64(startupProfile.LoadLaserSprite.Microseconds())/1000.0)
		fmt.Printf("  Load Missile Sprite:    %6.2f ms\n", float64(startupProfile.LoadMissileSprite.Microseconds())/1000.0)
		fmt.Printf("  Load Explosion Sprite:  %6.2f ms\n", float64(startupProfile.LoadExplosionSprite.Microseconds())/1000.0)
		fmt.Printf("  Load Faction Sprites:   %6.2f ms\n", float64(startupProfile.LoadFactionSprites.Microseconds())/1000.0)
		fmt.Printf("  Load Font:              %6.2f ms\n", float64(startupProfile.LoadFont.Microseconds())/1000.0)
		fmt.Printf("  Create UI Dialogs:      %6.2f ms\n", float64(startupProfile.CreateUIDialogs.Microseconds())/1000.0)
		fmt.Printf("  Load Persistence:       %6.2f ms\n", float64(startupProfile.LoadPersistence.Microseconds())/1000.0)
		fmt.Printf("  Create Entity Manager:  %6.2f ms\n", float64(startupProfile.CreateEntityManager.Microseconds())/1000.0)
		fmt.Printf("  Create Audio Manager:   %6.2f ms\n", float64(startupProfile.CreateAudioManager.Microseconds())/1000.0)
		fmt.Printf("  Load Sound Effects:     %6.2f ms\n", float64(startupProfile.LoadSoundEffects.Microseconds())/1000.0)
		fmt.Printf("  Load Menu Music:        %6.2f ms\n", float64(startupProfile.LoadMenuMusic.Microseconds())/1000.0)
		fmt.Printf("  Create Chat Window:     %6.2f ms\n", float64(startupProfile.CreateChatWindow.Microseconds())/1000.0)
		fmt.Printf("  Load BGM Tracks:        %6.2f ms\n", float64(startupProfile.LoadBGMTracks.Microseconds())/1000.0)
		fmt.Printf("  Create Game Struct:     %6.2f ms\n", float64(createGameStruct.Microseconds())/1000.0)
		fmt.Printf("  Wire Up Game:           %6.2f ms\n", float64(wireUpGame.Microseconds())/1000.0)
		fmt.Printf("  Start Menu Music:       %6.2f ms\n", float64(startMenuMusic.Microseconds())/1000.0)
		fmt.Printf("  ---\n")
		fmt.Printf("  Total Measured:         %6.2f ms\n", float64(measured.Microseconds())/1000.0)
		fmt.Printf("  UNACCOUNTED TIME:       %6.2f ms (!)\n", float64(unaccounted.Microseconds())/1000.0)
		fmt.Printf("  TOTAL STARTUP TIME:     %6.2f ms\n", float64(startupProfile.Total.Microseconds())/1000.0)
		fmt.Printf("==================================================\n\n")
	}

	return game, nil
}

// ReturnToTitleScreen transitions to the title screen and resumes menu music
func (g *Game) ReturnToTitleScreen() {
	g.currentState = TitleScreen
	if err := g.audioManager.PlayMusic("menu"); err != nil {
		log.Printf("Warning: Failed to play menu music: %v", err)
	}
}

// StartGame transitions from title screen to in-game state by spawning ships
func (g *Game) StartGame(fleetConfig config.FleetConfig) error {
	totalStart := time.Now()

	// Store fleet config for quick restart
	g.currentFleetConfig = &fleetConfig

	// Clear entity manager for new game
	t := time.Now()
	g.entityManager.Clear()
	g.newGameProfile.ClearEntities = time.Since(t)

	// Initialize factions and spawn points
	t = time.Now()
	g.entityManager.InitializeFactions()
	g.newGameProfile.InitializeFactions = time.Since(t)

	// Spawn ships according to fleet configuration
	const spawnRadius = 75.0 // Radius for circular spawn pattern
	var playerShip entity.Ship

	t = time.Now()
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
	g.newGameProfile.SpawnShips = time.Since(t)

	// Ensure we found a player ship
	if playerShip == nil {
		return fmt.Errorf("no player ship spawned (faction 0 must have at least 1 ship)")
	}

	// Initialize camera to follow player ship
	t = time.Now()
	newX, newY := playerShip.GetPosition()
	g.cameraX = newX - float64(config.ScreenWidth)/2
	g.cameraY = newY - float64(config.ScreenHeight)/2
	g.newGameProfile.InitializeCamera = time.Since(t)

	// Play random in-game music
	t = time.Now()
	bgmIndex := rand.Intn(5) // We have 5 BGM tracks (bgm0 through bgm4)
	bgmName := fmt.Sprintf("bgm%d", bgmIndex)
	if err := g.audioManager.PlayMusic(bgmName); err != nil {
		log.Printf("Warning: Failed to play BGM: %v", err)
	}
	g.newGameProfile.LoadGameMusic = time.Since(t)

	// Calculate total new game time
	g.newGameProfile.Total = time.Since(totalStart)

	// Print new game profiling data if enabled
	if *profileNewGame {
		fmt.Printf("\n=== New Game Profiling (Play Game to game loaded) ===\n")
		fmt.Printf("  Clear Entities:         %6.2f ms\n", float64(g.newGameProfile.ClearEntities.Microseconds())/1000.0)
		fmt.Printf("  Initialize Factions:    %6.2f ms\n", float64(g.newGameProfile.InitializeFactions.Microseconds())/1000.0)
		fmt.Printf("  Spawn Ships:            %6.2f ms\n", float64(g.newGameProfile.SpawnShips.Microseconds())/1000.0)
		fmt.Printf("  Initialize Camera:      %6.2f ms\n", float64(g.newGameProfile.InitializeCamera.Microseconds())/1000.0)
		fmt.Printf("  Load Game Music:        %6.2f ms\n", float64(g.newGameProfile.LoadGameMusic.Microseconds())/1000.0)
		fmt.Printf("  TOTAL NEW GAME TIME:    %6.2f ms\n", float64(g.newGameProfile.Total.Microseconds())/1000.0)
		fmt.Printf("======================================================\n\n")
	}

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
			} else if buttonIndex == 1 { // "Settings" button
				g.currentState = Settings
			} else if buttonIndex == 2 { // "Instructions" button
				g.currentState = Instructions
			} else if buttonIndex == 3 { // "High Scores" button
				g.currentState = HighScores
			}
		}

	case InGame:
		// Handle pause toggle (P key)
		keyP := ebiten.IsKeyPressed(ebiten.KeyP)
		if keyP && !g.prevKeyP {
			// P key was just pressed - toggle pause
			g.paused = !g.paused
		}
		g.prevKeyP = keyP

		// Handle sound mute toggle (N key)
		keyN := ebiten.IsKeyPressed(ebiten.KeyN)
		if keyN && !g.prevKeyN {
			// N key was just pressed - toggle sound mute
			g.audioManager.ToggleSoundMute()

			// Save settings with new mute state
			g.settings.SoundMuted = g.audioManager.IsSoundMuted()
			if err := persistence.SaveSettings(g.settings); err != nil {
				log.Printf("Warning: Failed to save settings: %v", err)
			}
		}
		g.prevKeyN = keyN

		// Handle music mute toggle (M key)
		keyM := ebiten.IsKeyPressed(ebiten.KeyM)
		if keyM && !g.prevKeyM {
			// M key was just pressed - toggle music mute
			g.audioManager.ToggleMusicMute()

			// Save settings with new mute state
			g.settings.MusicMuted = g.audioManager.IsMusicMuted()
			if err := persistence.SaveSettings(g.settings); err != nil {
				log.Printf("Warning: Failed to save settings: %v", err)
			}
		}
		g.prevKeyM = keyM

		// Skip all game logic if paused
		if g.paused {
			return nil
		}

		updateStart := time.Now()

		// Update all entities (ships, projectiles, explosions, collisions)
		// Profiling is handled internally by EntityManager via Profiler interface
		g.entityManager.UpdateAll()

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

			// Create game over screen with ebitenui and button callbacks
			g.gameOverScreen = ui.NewGameOverScreen(playerName, score, kills, deaths, isHighScore, g.hudFont.Source,
				// Continue callback (saves high score and returns to title)
				func() {
					// Save high score
					playerName := g.gameOverScreen.GetPlayerName()
					score, kills, deaths := g.entityManager.GetPlayerStats()
					g.highScores.AddScore(playerName, score, kills, deaths)
					if err := persistence.SaveHighScores(g.highScores); err != nil {
						log.Printf("Warning: Failed to save high scores: %v", err)
					}

					// Return to title screen and clear game over screen
					g.gameOverScreen = nil
					g.ReturnToTitleScreen()
				},
				// Quick restart callback (restarts with same fleet config)
				func() {
					if g.currentFleetConfig != nil {
						// Clear game over screen
						g.gameOverScreen = nil

						// Restart with same fleet configuration
						if err := g.StartGame(*g.currentFleetConfig); err != nil {
							log.Printf("Error restarting game: %v", err)
							g.ReturnToTitleScreen()
						}
					}
				},
			)
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
					g.ReturnToTitleScreen()
				} else {
					// Clear next fleet config after using it
					g.nextFleetConfig = nil
				}
			} else {
				// Fallback: return to title if no config
				g.ReturnToTitleScreen()
			}
		}

	case Settings:
		// Create settings screen if not already created
		if g.settingsScreen == nil {
			g.settingsScreen = ui.NewSettingsScreen(
				g.audioManager.IsSoundMuted(),
				g.audioManager.GetSoundVolume(),
				g.audioManager.IsMusicMuted(),
				g.audioManager.GetMusicVolume(),
				g.settings.ChatEnabled,
				g.hudFont.Source,
				func(soundMuted bool, soundVolume float64, musicMuted bool, musicVolume float64, chatEnabled bool) {
					// Apply settings
					g.audioManager.SetSoundMuted(soundMuted)
					g.audioManager.SetSoundVolume(soundVolume)
					g.audioManager.SetMusicMuted(musicMuted)
					g.audioManager.SetMusicVolume(musicVolume)

					// Update settings struct
					g.settings.SoundMuted = soundMuted
					g.settings.SoundVolume = soundVolume
					g.settings.MusicMuted = musicMuted
					g.settings.MusicVolume = musicVolume
					g.settings.ChatEnabled = chatEnabled
				},
				func(soundMuted bool, soundVolume float64, musicMuted bool, musicVolume float64, chatEnabled bool) {
					// Save settings callback
					g.audioManager.SetSoundMuted(soundMuted)
					g.audioManager.SetSoundVolume(soundVolume)
					g.audioManager.SetMusicMuted(musicMuted)
					g.audioManager.SetMusicVolume(musicVolume)

					// Update settings struct
					g.settings.SoundMuted = soundMuted
					g.settings.SoundVolume = soundVolume
					g.settings.MusicMuted = musicMuted
					g.settings.MusicVolume = musicVolume
					g.settings.ChatEnabled = chatEnabled

					// Save to disk
					if err := persistence.SaveSettings(g.settings); err != nil {
						log.Printf("Warning: Failed to save settings: %v", err)
					}

					// Clear settings screen and return to title
					g.settingsScreen = nil
					g.currentState = TitleScreen
				},
			)
		}

		// Update settings screen
		g.settingsScreen.Update()
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
		minGridX := int(math.Floor(cameraX / float64(config.StarGridSize)))
		maxGridX := int(math.Floor((cameraX + float64(config.ScreenWidth)) / float64(config.StarGridSize)))
		minGridY := int(math.Floor(cameraY / float64(config.StarGridSize)))
		maxGridY := int(math.Floor((cameraY + float64(config.ScreenHeight)) / float64(config.StarGridSize)))

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
		// Use floor division to handle negative camera coordinates correctly
		minGridX := int(math.Floor(g.cameraX / float64(config.StarGridSize)))
		maxGridX := int(math.Floor((g.cameraX + float64(config.ScreenWidth)) / float64(config.StarGridSize)))
		minGridY := int(math.Floor(g.cameraY / float64(config.StarGridSize)))
		maxGridY := int(math.Floor((g.cameraY + float64(config.ScreenHeight)) / float64(config.StarGridSize)))

		// Draw stars for visible grid cells
		for gridX := minGridX; gridX <= maxGridX; gridX++ {
			for gridY := minGridY; gridY <= maxGridY; gridY++ {
				stars := systems.GenerateStarsForGrid(gridX, gridY)
				for _, star := range stars {
					// Use wrapped screen position to handle world wrapping correctly
					screenX, screenY := entity.GetWrappedScreenPosition(star.X, star.Y, g.cameraX, g.cameraY)

					// Only draw if on screen
					if screenX >= 0 && screenX < float64(config.ScreenWidth) && screenY >= 0 && screenY < float64(config.ScreenHeight) {
						vector.FillRect(screen, float32(screenX), float32(screenY), 1, 1, color.White, false)
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

		// Draw testudon beams
		t = time.Now()
		for _, ship := range g.entityManager.ships {
			// Check if this is a testudon with an active beam
			if ship.GetClass() == entity.ClassTestudon {
				// Type assert to get beam target info
				if testudon, ok := ship.(*entity.Testudon); ok && testudon.IsAlive() {
					beamTargetID := testudon.GetBeamTargetID()
					if beamTargetID >= 0 {
						target := g.entityManager.GetShip(beamTargetID)
						if target != nil && target.IsAlive() {
							// Get positions
							shipX, shipY := testudon.GetPosition()
							targetX, targetY := target.GetPosition()

							// Convert to screen coordinates using wrapped positions
							screenX1, screenY1 := entity.GetWrappedScreenPosition(shipX, shipY, g.cameraX, g.cameraY)
							screenX2, screenY2 := entity.GetWrappedScreenPosition(targetX, targetY, g.cameraX, g.cameraY)

							// Check if at least one endpoint is visible
							// This prevents drawing beams "the long way" when both endpoints
							// are off-screen on opposite sides
							testudonVisible := screenX1 >= 0 && screenX1 < float64(config.ScreenWidth) &&
								screenY1 >= 0 && screenY1 < float64(config.ScreenHeight)
							targetVisible := screenX2 >= 0 && screenX2 < float64(config.ScreenWidth) &&
								screenY2 >= 0 && screenY2 < float64(config.ScreenHeight)

							if testudonVisible || targetVisible {
								// Draw beam line using testudon's faction color
								factionID := testudon.GetFaction()
								beamColor := g.factionColors[factionID%len(g.factionColors)]
								vector.StrokeLine(screen, float32(screenX1), float32(screenY1), float32(screenX2), float32(screenY2), 2, beamColor, false)
							}
						}
					}
				}
			}
		}
		g.profileData.RenderBeams += time.Since(t)

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

		// Draw particles (afterburner exhaust, etc.)
		for _, particle := range g.entityManager.particles {
			particle.Render(screen, g.cameraX, g.cameraY)
		}

		// Draw minimap
		t = time.Now()
		g.renderMinimap(screen)
		g.profileData.RenderMinimap += time.Since(t)

		// Draw afterburner bar (left of minimap)
		g.renderAfterburnerBar(screen)

		// Draw chat window (left of minimap)
		g.renderChatWindow(screen)

		// Draw HUD
		textColor := color.White

		// Get player state from EntityManager
		playerScore, playerKills, _ := g.entityManager.GetPlayerStats()
		var playerShield int

		// Display shield of currently controlled or spectated ship
		if g.entityManager.IsSpectating() {
			spectatedShip := g.entityManager.GetSpectatedShip()
			if spectatedShip != nil {
				playerShield, _ = spectatedShip.GetHealth()
			}
		} else {
			playerShip := g.entityManager.GetPlayerShip()
			if playerShip != nil {
				playerShield, _ = playerShip.GetHealth()
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

		// Pause message (centered)
		if g.paused {
			pauseFont := &text.GoTextFace{
				Source: g.hudFont.Source,
				Size:   24,
			}

			pauseText := "PAUSED"
			pauseWidth, _ := text.Measure(pauseText, pauseFont, 0)
			pauseOp := &text.DrawOptions{}
			pauseOp.GeoM.Translate(float64(config.ScreenWidth/2)-pauseWidth/2, float64(config.ScreenHeight/2)-40)
			pauseOp.ColorScale.ScaleWithColor(color.RGBA{255, 255, 0, 255}) // Yellow
			text.Draw(screen, pauseText, pauseFont, pauseOp)

			// Instructions
			instructionsText := "Press P to resume"
			instrWidth, _ := text.Measure(instructionsText, g.hudFont, 0)
			instrOp := &text.DrawOptions{}
			instrOp.GeoM.Translate(float64(config.ScreenWidth/2)-instrWidth/2, float64(config.ScreenHeight/2))
			instrOp.ColorScale.ScaleWithColor(textColor)
			text.Draw(screen, instructionsText, g.hudFont, instrOp)

			// Control keys reference
			controlsText := "Controls: M (music) | N (sound) | P (pause)"
			controlsWidth, _ := text.Measure(controlsText, g.hudFont, 0)
			controlsOp := &text.DrawOptions{}
			controlsOp.GeoM.Translate(float64(config.ScreenWidth/2)-controlsWidth/2, float64(config.ScreenHeight/2)+30)
			controlsOp.ColorScale.ScaleWithColor(color.RGBA{180, 180, 180, 255}) // Gray
			text.Draw(screen, controlsText, g.hudFont, controlsOp)
		}

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
				fmt.Printf("  Weapons Update:    %6.2f ms (capacitor charging)\n", float64(g.profileData.WeaponsUpdate.Microseconds())/frameCount/1000.0)
				fmt.Printf("  AI Update:         %6.2f ms (targeting, rotation, firing)\n", float64(g.profileData.AIMovement.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Beam Weapons:      %6.2f ms (testudon beams)\n", float64(g.profileData.BeamWeapons.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Movement:          %6.2f ms (position + wrapping)\n", float64(g.profileData.Movement.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Missile Tracking:  %6.2f ms (homing missiles)\n", float64(g.profileData.MissileTracking.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Projectile Life:   %6.2f ms (lifetime checks)\n", float64(g.profileData.ProjectileLife.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Collisions:        %6.2f ms (spatial grid)\n", float64(g.profileData.Collisions.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Explosions:        %6.2f ms (animation)\n", float64(g.profileData.Explosions.Microseconds())/frameCount/1000.0)
				fmt.Printf("Render Systems:\n")
				fmt.Printf("  Stars:             %6.2f ms\n", float64(g.profileData.RenderStars.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Ships:             %6.2f ms\n", float64(g.profileData.RenderShips.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Beams:             %6.2f ms (testudon beams)\n", float64(g.profileData.RenderBeams.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Projectiles:       %6.2f ms\n", float64(g.profileData.RenderProjectiles.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Explosions:        %6.2f ms\n", float64(g.profileData.RenderExplosions.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Minimap:           %6.2f ms\n", float64(g.profileData.RenderMinimap.Microseconds())/frameCount/1000.0)
				fmt.Printf("Totals:\n")
				fmt.Printf("  Total Update:      %6.2f ms\n", float64(g.profileData.TotalUpdate.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Total Draw:        %6.2f ms\n", float64(g.profileData.TotalDraw.Microseconds())/frameCount/1000.0)
				totalFrame := g.profileData.TotalUpdate + g.profileData.TotalDraw
				fmt.Printf("  Total Frame:       %6.2f ms\n", float64(totalFrame.Microseconds())/frameCount/1000.0)
				fmt.Printf("  Frame Budget:      %6.2f ms (60 FPS)\n", 16.67)

				// Calculate entity counts for context
				shipCount := len(g.entityManager.ships)
				projCount := len(g.entityManager.projectiles)
				explCount := len(g.entityManager.explosions)
				fmt.Printf("Entity Counts: %d ships, %d projectiles, %d explosions\n", shipCount, projCount, explCount)
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
		minGridX := int(math.Floor(cameraX / float64(config.StarGridSize)))
		maxGridX := int(math.Floor((cameraX + float64(config.ScreenWidth)) / float64(config.StarGridSize)))
		minGridY := int(math.Floor(cameraY / float64(config.StarGridSize)))
		maxGridY := int(math.Floor((cameraY + float64(config.ScreenHeight)) / float64(config.StarGridSize)))
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
		minGridX := int(math.Floor(cameraX / float64(config.StarGridSize)))
		maxGridX := int(math.Floor((cameraX + float64(config.ScreenWidth)) / float64(config.StarGridSize)))
		minGridY := int(math.Floor(cameraY / float64(config.StarGridSize)))
		maxGridY := int(math.Floor((cameraY + float64(config.ScreenHeight)) / float64(config.StarGridSize)))

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
		minGridX := int(math.Floor(cameraX / float64(config.StarGridSize)))
		maxGridX := int(math.Floor((cameraX + float64(config.ScreenWidth)) / float64(config.StarGridSize)))
		minGridY := int(math.Floor(cameraY / float64(config.StarGridSize)))
		maxGridY := int(math.Floor((cameraY + float64(config.ScreenHeight)) / float64(config.StarGridSize)))

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
		minGridX := int(math.Floor(cameraX / float64(config.StarGridSize)))
		maxGridX := int(math.Floor((cameraX + float64(config.ScreenWidth)) / float64(config.StarGridSize)))
		minGridY := int(math.Floor(cameraY / float64(config.StarGridSize)))
		maxGridY := int(math.Floor((cameraY + float64(config.ScreenHeight)) / float64(config.StarGridSize)))

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
		minGridX := int(math.Floor(cameraX / float64(config.StarGridSize)))
		maxGridX := int(math.Floor((cameraX + float64(config.ScreenWidth)) / float64(config.StarGridSize)))
		minGridY := int(math.Floor(cameraY / float64(config.StarGridSize)))
		maxGridY := int(math.Floor((cameraY + float64(config.ScreenHeight)) / float64(config.StarGridSize)))

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

	case Settings:
		// Draw stars background
		cameraX, cameraY := 0.0, 0.0
		minGridX := int(math.Floor(cameraX / float64(config.StarGridSize)))
		maxGridX := int(math.Floor((cameraX + float64(config.ScreenWidth)) / float64(config.StarGridSize)))
		minGridY := int(math.Floor(cameraY / float64(config.StarGridSize)))
		maxGridY := int(math.Floor((cameraY + float64(config.ScreenHeight)) / float64(config.StarGridSize)))

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

		// Draw settings screen
		if g.settingsScreen != nil {
			g.settingsScreen.Draw(screen)
		}
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

	// Draw all ships as colored dots
	for _, ship := range g.entityManager.ships {
		if !ship.IsAlive() {
			continue
		}

		x, y := ship.GetPosition()
		minimapDotX := minimapX + float32(x*scaleX)
		minimapDotY := minimapY + float32(y*scaleY)

		factionID := ship.GetFaction()
		shipColor := g.factionColors[factionID%len(g.factionColors)]

		// Draw ship dot with size based on ship class
		// Fighters: 1x1, Destroyers: 2x2, Testudons: 3x3
		var dotSize float32
		switch ship.GetClass() {
		case entity.ClassFighter:
			dotSize = 1
		case entity.ClassDestroyer:
			dotSize = 2
		case entity.ClassTestudon:
			dotSize = 3
		default:
			dotSize = 1
		}

		// Center the dot on the ship position
		offset := dotSize / 2
		vector.FillRect(screen, minimapDotX-offset, minimapDotY-offset, dotSize, dotSize, shipColor, false)
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

// renderAfterburnerBar renders the afterburner charge bar to the left of the minimap
func (g *Game) renderAfterburnerBar(screen *ebiten.Image) {
	// Get player ship
	playerShip := g.entityManager.GetPlayerShip()
	if playerShip == nil || !playerShip.IsAlive() {
		return
	}

	// Only show for ships with afterburner (fighters and destroyers)
	if !playerShip.HasAfterburner() {
		return
	}

	// Bar constants (matching minimap position)
	const minimapSize = 120
	const minimapMargin = 10
	const minimapX = config.ScreenWidth - minimapSize - minimapMargin
	const minimapY = config.ScreenHeight - minimapSize - minimapMargin

	// Bar position (10 pixels to the left of minimap)
	const barWidth = 8
	const barSpacing = 10
	const barX = minimapX - barWidth - barSpacing
	const barY = minimapY
	const barHeight = minimapSize

	// Draw bar background (dark gray)
	vector.DrawFilledRect(screen, barX, barY, barWidth, barHeight, color.RGBA{40, 40, 40, 200}, false)

	// Calculate fill height based on charge (0-360)
	charge := playerShip.GetAfterburnerCharge()
	maxCharge := 360.0
	fillRatio := charge / maxCharge
	fillHeight := float32(barHeight) * float32(fillRatio)

	// Draw fill from bottom to top (orange for afterburner)
	if fillHeight > 0 {
		fillY := barY + barHeight - fillHeight
		fillColor := color.RGBA{255, 140, 0, 255} // Orange
		vector.DrawFilledRect(screen, barX, fillY, barWidth, fillHeight, fillColor, false)
	}

	// Draw border
	borderColor := color.RGBA{100, 100, 100, 255}
	vector.StrokeRect(screen, barX, barY, barWidth, barHeight, 1, borderColor, false)
}

// renderChatWindow renders the chat message window to the left of the minimap
func (g *Game) renderChatWindow(screen *ebiten.Image) {
	if g.chatWindow == nil || !g.settings.ChatEnabled {
		return
	}

	// Don't show chat window in spectate mode
	if g.entityManager.IsSpectating() {
		return
	}

	// Chat window constants (matching minimap position)
	const minimapSize = 120
	const minimapMargin = 10
	const minimapX = config.ScreenWidth - minimapSize - minimapMargin
	const minimapY = config.ScreenHeight - minimapSize - minimapMargin

	// Chat window dimensions (60% of screen width, same height as minimap)
	chatWidth := float32(config.ScreenWidth) * 0.6
	chatHeight := float32(minimapSize)
	const chatSpacing = 10
	chatX := float32(minimapX) - chatWidth - chatSpacing - 8 - chatSpacing // Account for afterburner bar
	chatY := float32(minimapY)

	// Draw semi-transparent background
	vector.DrawFilledRect(screen, chatX, chatY, chatWidth, chatHeight, color.RGBA{20, 20, 20, 180}, false)

	// Draw border
	borderColor := color.RGBA{100, 100, 100, 255}
	vector.StrokeRect(screen, chatX, chatY, chatWidth, chatHeight, 1, borderColor, false)

	// Draw messages
	messages := g.chatWindow.GetMessages()
	if len(messages) == 0 {
		return
	}

	// Font for chat messages
	const fontSize = 11
	const lineHeight = 12
	const padding = 5

	// Draw messages from bottom to top (most recent at bottom)
	y := chatY + chatHeight - padding - float32(fontSize)
	for i := len(messages) - 1; i >= 0 && y > chatY+padding; i-- {
		msg := messages[i]
		messageText := fmt.Sprintf("%s: %s", msg.SpeakerName, msg.Text)

		// Get faction color for this message
		factionColor := g.factionColors[msg.FactionID%len(g.factionColors)]

		// Create text options
		textOpts := &text.DrawOptions{}
		textOpts.GeoM.Translate(float64(chatX+padding), float64(y))
		textOpts.ColorScale.ScaleWithColor(factionColor)

		// Use smaller font for chat
		chatFont := &text.GoTextFace{
			Source: g.hudFont.Source,
			Size:   fontSize,
		}

		// Draw text
		text.Draw(screen, messageText, chatFont, textOpts)

		y -= lineHeight
	}
}

// Layout returns the game's screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

func main() {
	mainStart := time.Now()
	flag.Parse()

	if *profileStartup {
		fmt.Printf("\n=== Main() Profiling ===\n")
		fmt.Printf("  flag.Parse() completed:     %6.2f ms\n", float64(time.Since(mainStart).Microseconds())/1000.0)
	}

	t := time.Now()
	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle("Verdant Thane")
	if *profileStartup {
		fmt.Printf("  Window setup completed:     %6.2f ms (from start: %6.2f ms)\n",
			float64(time.Since(t).Microseconds())/1000.0,
			float64(time.Since(mainStart).Microseconds())/1000.0)
	}

	t = time.Now()
	game, err := NewGame()
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}
	if *profileStartup {
		fmt.Printf("  NewGame() completed:        %6.2f ms (from start: %6.2f ms)\n",
			float64(time.Since(t).Microseconds())/1000.0,
			float64(time.Since(mainStart).Microseconds())/1000.0)
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

	if *profileStartup {
		fmt.Printf("  Starting ebiten.RunGame():  (from start: %6.2f ms)\n",
			float64(time.Since(mainStart).Microseconds())/1000.0)
		fmt.Printf("========================\n\n")
	}

	t = time.Now()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
