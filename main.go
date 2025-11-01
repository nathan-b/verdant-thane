package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1024
	screenHeight = 768

	gameWidth = 5040
	gameHeight = 5040

	// Rotation speed in radians per tick (60 ticks per second)
	rotationSpeed = 3.0 * math.Pi / 180.0 // 3 degrees per tick

	// Movement constants (for Fighter class)
	maxSpeed     = 6.0           // pixels per tick
	acceleration = 4.0 / 60.0    // pixels per second per tick

	// Starfield constants
	starDensity = 0.0003 // stars per pixel
	starGridSize = 200   // grid size for deterministic star generation
)

// Game represents the main game state
type Game struct {
	ships    []*Ship
	player   *Player
	cameraX  float64 // Camera position (follows player)
	cameraY  float64
}

type ShipClass int
const (
	Fighter ShipClass = iota
	Destroyer
	Testudon
	Mothership
)

type Ship struct {
	sprite   *ebiten.Image
	class    ShipClass
	faction  int
	x        float64
	y        float64
	angle    float64
	speed    float64
}

type Player struct {
	ship *Ship
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
	switch class {
	case Fighter:
		spritePath = "assets/fighter.png"
	case Destroyer:
		spritePath = "assets/destroyer.png"
	case Testudon:
		spritePath = "assets/testudon.png"
	case Mothership:
		spritePath = "assets/mothership.png"
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
		sprite:  sprite,
		class:   class,
		faction: faction,
		x:       x,
		y:       y,
		angle:   angle,
		speed:   0,  // TODO Handle acceleration too
	}, nil
}

// NewGame creates and initializes a new game
func NewGame() (*Game, error) {
	// Build the player ship at the center of the game world
	startX := float64(gameWidth) / 2
	startY := float64(gameHeight) / 2
	pship , err := NewShip(Fighter, 0, startX, startY, 0)

	if err != nil {
		return nil, err
	}

	return &Game{
		ships: []*Ship{pship},
		player: &Player{ship: pship},
		cameraX: startX - float64(screenWidth)/2,
		cameraY: startY - float64(screenHeight)/2,
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
