package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

// Game represents the main game state
type Game struct{}

// Update updates the game logic
// This is called 60 times per second
func (g *Game) Update() error {
	return nil
}

// Draw renders the game screen
func (g *Game) Draw(screen *ebiten.Image) {
	// Fill the screen with black
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Draw "Hello World" text
	ebitenutil.DebugPrint(screen, "Hello World from Verdant Thane!")
}

// Layout returns the game's screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Verdant Thane")

	game := &Game{}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
