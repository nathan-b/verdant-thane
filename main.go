package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 1024
	screenHeight = 768

	// Rotation speed in radians per tick (60 ticks per second)
	rotationSpeed = 3.0 * math.Pi / 180.0 // 3 degrees per tick
)

// Game represents the main game state
type Game struct {
	ships    []*Ship
	player   *Player
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
	speed    int
}

type Player struct {
	ship *Ship
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
	// Build the player ship
	pship , err := NewShip(Fighter, 0, float64(screenWidth)/2, float64(screenHeight)/2, 0)

	if err != nil {
		return nil, err
	}

	return &Game{
		ships: []*Ship{pship},
		player: &Player{ship: pship},
	}, nil
}

// Update updates the game logic
// This is called 60 times per second
func (g *Game) Update() error {
	// Handle rotation
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		// Rotate left (counter-clockwise)
		g.player.ship.angle -= rotationSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		// Rotate right (clockwise)
		g.player.ship.angle += rotationSpeed
	}

	return nil
}

// Draw renders the game screen
func (g *Game) Draw(screen *ebiten.Image) {
	// Fill the screen with black
	screen.Fill(color.RGBA{0, 0, 0, 255})

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
		// 3. Translate to ship position
		op.GeoM.Translate(ship.x, ship.y)

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
