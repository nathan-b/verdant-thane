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
)

const (
	screenWidth  = 1024
	screenHeight = 768

	gameWidth  = 5040
	gameHeight = 5040

	// Rotation speed in radians per tick (60 ticks per second)
	rotationSpeed = 3.0 * math.Pi / 180.0 // 3 degrees per tick

	// Movement constants (for Fighter class)
	maxSpeed     = 6.0        // pixels per tick
	acceleration = 4.0 / 60.0 // pixels per second per tick

	// Firing constants
	capacitorChargeTime = 600.0 / 1000.0                     // 600ms in seconds
	capacitorChargeRate = 1.0 / (capacitorChargeTime * 60.0) // charge per tick (60 ticks/sec)
	firingConeAngle     = 30.0 * math.Pi / 180.0             // 30 degrees in radians
	projectileSpeed     = maxSpeed * 2.0                     // twice max ship speed
	projectileLifetime  = 180                                // ticks (3 seconds at 60 TPS)

	// Starfield constants
	starDensity  = 0.0003 // stars per pixel
	starGridSize = 200    // grid size for deterministic star generation
)

// Game represents the main game state
type Game struct {
	ships       []*Ship
	player      *Player
	projectiles []*Projectile
	laserSprite *ebiten.Image    // Shared sprite for all projectiles
	hudFont     *text.GoTextFace // Font for HUD rendering
	cameraX     float64          // Camera position (follows player)
	cameraY     float64
}

type ShipClass int

const (
	Fighter ShipClass = iota
	Destroyer
	Testudon
	Mothership
)

type Ship struct {
	sprite    *ebiten.Image
	class     ShipClass
	faction   int
	x         float64
	y         float64
	angle     float64
	speed     float64
	capacitor float64 // 0.0 to 1.0, controls firing ability
	hull      int     // hit points
}

type Player struct {
	ship  *Ship
	score int
	kills int
}

type Projectile struct {
	x        float64
	y        float64
	vx       float64 // velocity X
	vy       float64 // velocity Y
	faction  int     // which team fired it
	lifetime int     // ticks remaining
}

// modulo performs proper modulo operation (handles negatives correctly)
func modulo(a, b int) int {
	return ((a % b) + b) % b
}

// hashPosition creates a deterministic hash for a grid position
// Wraps grid coordinates to ensure consistent stars across world boundaries
func hashPosition(gridX, gridY int) int {
	// Calculate number of grid cells in the game world
	gridCountX := gameWidth / starGridSize
	gridCountY := gameHeight / starGridSize

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

// Build a ship
func NewShip(class ShipClass, faction int, x, y, angle float64) (*Ship, error) {
	// Get the sprite path based on class
	var spritePath string
	var hull int
	switch class {
	case Fighter:
		spritePath = "assets/fighter.png"
		hull = 8
	case Destroyer:
		spritePath = "assets/destroyer.png"
		hull = 32
	case Testudon:
		spritePath = "assets/testudon.png"
		hull = 50
	case Mothership:
		spritePath = "assets/mothership.png"
		hull = 200
	default:
		return nil, fmt.Errorf("unknown ship class: %v", class)
	}

	// Load the sprite
	// TODO: Cache loaded sprites to avoid reloading
	sprite, _, err := ebitenutil.NewImageFromFile(spritePath)
	if err != nil {
		return nil, err
	}

	return &Ship{
		sprite:    sprite,
		class:     class,
		faction:   faction,
		x:         x,
		y:         y,
		angle:     angle,
		speed:     0,
		capacitor: 1.0, // Start fully charged,
		hull:      hull,
	}, nil
}

// fireWeapon attempts to fire a projectile from the given ship toward the target coordinates
// Returns true if the weapon was fired, false otherwise (e.g., capacitor not charged)
func (g *Game) fireWeapon(ship *Ship, targetX, targetY float64) bool {
	// Check if capacitor is charged
	if ship.capacitor < 1.0 {
		return false
	}

	// Calculate angle to target
	dx := targetX - ship.x
	dy := targetY - ship.y
	// Adjust for sprite orientation (sprite faces up at angle 0)
	targetAngle := math.Atan2(dx, -dy)

	// Calculate angle difference from ship's facing direction
	angleDiff := targetAngle - ship.angle
	// Normalize to [-π, π]
	for angleDiff > math.Pi {
		angleDiff -= 2 * math.Pi
	}
	for angleDiff < -math.Pi {
		angleDiff += 2 * math.Pi
	}

	// Constrain to firing cone
	firingAngle := ship.angle
	halfCone := firingConeAngle / 2
	if angleDiff > halfCone {
		firingAngle += halfCone
	} else if angleDiff < -halfCone {
		firingAngle -= halfCone
	} else {
		firingAngle = targetAngle
	}

	// Create projectile
	vx := math.Sin(firingAngle) * projectileSpeed
	vy := -math.Cos(firingAngle) * projectileSpeed

	g.projectiles = append(g.projectiles, &Projectile{
		x:        ship.x,
		y:        ship.y,
		vx:       vx,
		vy:       vy,
		faction:  ship.faction,
		lifetime: projectileLifetime,
	})

	// Drain capacitor
	ship.capacitor = 0.0

	return true
}

// NewGame creates and initializes a new game
func NewGame() (*Game, error) {
	// Build the player ship at the center of the game world
	startX := float64(gameWidth) / 2
	startY := float64(gameHeight) / 2
	pship, err := NewShip(Fighter, 0, startX, startY, 0)

	if err != nil {
		return nil, err
	}

	// Load laser sprite (shared by all projectiles)
	laserSprite, _, err := ebitenutil.NewImageFromFile("assets/laser.png")
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

	return &Game{
		ships:       []*Ship{pship},
		player:      &Player{ship: pship, score: 0, kills: 0},
		laserSprite: laserSprite,
		hudFont:     hudFont,
		cameraX:     startX - float64(screenWidth)/2,
		cameraY:     startY - float64(screenHeight)/2,
	}, nil
}

// Update updates the game logic
// This is called 60 times per second
func (g *Game) Update() error {
	ship := g.player.ship

	// Handle rotation
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		// Rotate left (counter-clockwise)
		ship.angle -= rotationSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		// Rotate right (clockwise)
		ship.angle += rotationSpeed
	}

	// Handle acceleration/deceleration
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		// Accelerate
		ship.speed += acceleration
		if ship.speed > maxSpeed {
			ship.speed = maxSpeed
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		// Decelerate
		ship.speed -= acceleration
		if ship.speed < 0 {
			ship.speed = 0
		}
	}

	// Apply velocity to position
	// Sprite faces upward at angle 0, so we adjust the standard math:
	// - Upward (angle 0) means negative Y in screen coordinates
	// - Use sin() for X and -cos() for Y
	ship.x += math.Sin(ship.angle) * ship.speed
	ship.y += -math.Cos(ship.angle) * ship.speed

	// Keep ship within game bounds (wrap around if necessary)
	if ship.x < 0 {
		ship.x += float64(gameWidth)
	} else if ship.x >= float64(gameWidth) {
		ship.x -= float64(gameWidth)
	}
	if ship.y < 0 {
		ship.y += float64(gameHeight)
	} else if ship.y >= float64(gameHeight) {
		ship.y -= float64(gameHeight)
	}

	// Charge capacitor
	if ship.capacitor < 1.0 {
		ship.capacitor += capacitorChargeRate
	}

	// Handle firing
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		// Get mouse position in world coordinates
		mouseX, mouseY := ebiten.CursorPosition()
		worldMouseX := float64(mouseX) + g.cameraX
		worldMouseY := float64(mouseY) + g.cameraY

		// Attempt to fire at mouse position
		g.fireWeapon(ship, worldMouseX, worldMouseY)
	}

	// Update projectiles
	activeProjectiles := []*Projectile{}
	for _, p := range g.projectiles {
		// Update position
		p.x += p.vx
		p.y += p.vy

		// Wrap around world boundaries
		if p.x < 0 {
			p.x += float64(gameWidth)
		} else if p.x >= float64(gameWidth) {
			p.x -= float64(gameWidth)
		}
		if p.y < 0 {
			p.y += float64(gameHeight)
		} else if p.y >= float64(gameHeight) {
			p.y -= float64(gameHeight)
		}

		// Decrease lifetime
		p.lifetime--

		// Keep if still alive
		if p.lifetime > 0 {
			activeProjectiles = append(activeProjectiles, p)
		}
	}
	g.projectiles = activeProjectiles

	// Update camera to follow player (centered on player)
	g.cameraX = ship.x - float64(screenWidth)/2
	g.cameraY = ship.y - float64(screenHeight)/2

	return nil
}

// Draw renders the game screen
func (g *Game) Draw(screen *ebiten.Image) {
	// Fill the screen with black
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Draw stars
	// Determine which grid cells are visible
	minGridX := int(g.cameraX) / starGridSize
	maxGridX := int(g.cameraX+float64(screenWidth)) / starGridSize
	minGridY := int(g.cameraY) / starGridSize
	maxGridY := int(g.cameraY+float64(screenHeight)) / starGridSize

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
					if screenX >= 0 && screenX < float64(screenWidth) && screenY >= 0 && screenY < float64(screenHeight) {
						vector.DrawFilledRect(screen, float32(screenX), float32(screenY), 1, 1, color.White, false)
					}
				}

				// Draw star at its primary position
				drawStarAtPosition(star.x, star.y)

				// Also check if we should draw the star at wrapped positions
				// This handles the case where the camera is near world boundaries
				if star.x < g.cameraX {
					// Star is to the left of camera, try drawing wrapped to the right
					drawStarAtPosition(star.x+float64(gameWidth), star.y)
				}
				if star.x > g.cameraX+float64(screenWidth) {
					// Star is to the right of camera, try drawing wrapped to the left
					drawStarAtPosition(star.x-float64(gameWidth), star.y)
				}
				if star.y < g.cameraY {
					// Star is above camera, try drawing wrapped below
					drawStarAtPosition(star.x, star.y+float64(gameHeight))
				}
				if star.y > g.cameraY+float64(screenHeight) {
					// Star is below camera, try drawing wrapped above
					drawStarAtPosition(star.x, star.y-float64(gameHeight))
				}
			}
		}
	}

	// Draw each ship
	for _, ship := range g.ships {
		op := &ebiten.DrawImageOptions{}

		// Get ship dimensions
		bounds := ship.sprite.Bounds()
		shipWidth := float64(bounds.Dx())
		shipHeight := float64(bounds.Dy())

		// Apply transformations in order:
		// 1. Translate ship to center it around origin (for rotation)
		op.GeoM.Translate(-shipWidth/2, -shipHeight/2)
		// 2. Rotate around origin
		op.GeoM.Rotate(ship.angle)
		// 3. Translate to ship position in world
		op.GeoM.Translate(ship.x, ship.y)
		// 4. Translate by camera offset to convert to screen coordinates
		op.GeoM.Translate(-g.cameraX, -g.cameraY)

		screen.DrawImage(ship.sprite, op)
	}

	// Draw projectiles
	for _, p := range g.projectiles {
		op := &ebiten.DrawImageOptions{}

		// Get sprite dimensions for centering
		bounds := g.laserSprite.Bounds()
		spriteWidth := float64(bounds.Dx())
		spriteHeight := float64(bounds.Dy())

		// Apply transformations (no rotation needed for square sprite):
		// 1. Translate to projectile position (centered)
		op.GeoM.Translate(p.x-spriteWidth/2, p.y-spriteHeight/2)
		// 2. Apply camera offset
		op.GeoM.Translate(-g.cameraX, -g.cameraY)

		screen.DrawImage(g.laserSprite, op)
	}

	// Draw HUD
	textColor := color.White

	// Upper left: Score
	scoreText := fmt.Sprintf("Score: %d", g.player.score)
	scoreOp := &text.DrawOptions{}
	scoreOp.GeoM.Translate(10, 10)
	scoreOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, scoreText, g.hudFont, scoreOp)

	// Upper right: Shield
	shieldText := fmt.Sprintf("Shield: %d", g.player.ship.hull)
	shieldWidth, _ := text.Measure(shieldText, g.hudFont, 0)
	shieldOp := &text.DrawOptions{}
	shieldOp.GeoM.Translate(float64(screenWidth)-shieldWidth-10, 10)
	shieldOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, shieldText, g.hudFont, shieldOp)

	// Upper right: Kills
	killsText := fmt.Sprintf("Kills: %d", g.player.kills)
	killsWidth, _ := text.Measure(killsText, g.hudFont, 0)
	killsOp := &text.DrawOptions{}
	killsOp.GeoM.Translate(float64(screenWidth)-killsWidth-10, 27)
	killsOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, killsText, g.hudFont, killsOp)
}

// Layout returns the game's screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Verdant Thane")

	game, err := NewGame()
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
