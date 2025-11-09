package systems

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// FactionColorProfile defines the color transformation for a faction
type FactionColorProfile struct {
	R, G, B float64 // RGB multipliers for grayscale intensity
}

var factionColorProfiles = map[int]FactionColorProfile{
	0: {R: 0.3, G: 1.0, B: 0.3}, // Green (faction 0)
	1: {R: 0.2, G: 0.5, B: 1.0}, // Blue (faction 1)
	2: {R: 1.0, G: 0.2, B: 0.2}, // Red (faction 2)
	3: {R: 1.0, G: 1.0, B: 0.2}, // Yellow (faction 3)
}

// ApplyFactionPalette applies a faction-specific color transformation to a grayscale sprite
func ApplyFactionPalette(baseSprite *ebiten.Image, factionID int) *ebiten.Image {
	profile, ok := factionColorProfiles[factionID]
	if !ok {
		return baseSprite
	}

	bounds := baseSprite.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	result := ebiten.NewImage(width, height)

	var colorMatrix ebiten.ColorM
	colorMatrix.Scale(profile.R, profile.G, profile.B, 1.0)

	opts := &ebiten.DrawImageOptions{}
	opts.ColorM = colorMatrix
	result.DrawImage(baseSprite, opts)

	return result
}

// ShipClassSprites holds faction-colored sprites for a single ship class
type ShipClassSprites struct {
	baseSprite *ebiten.Image // Grayscale base sprite
	Green      *ebiten.Image // Faction 0 - Player
	Blue       *ebiten.Image // Faction 1
	Red        *ebiten.Image // Faction 2
	Yellow     *ebiten.Image // Faction 3
}

// FactionSprites holds sprites for all ship classes
type FactionSprites struct {
	Fighter   *ShipClassSprites
	Destroyer *ShipClassSprites
	Testudon  *ShipClassSprites
}

// loadShipClassSprites loads and generates faction sprites for a ship class
func loadShipClassSprites(basePath string) (*ShipClassSprites, error) {
	// Load the grayscale base sprite
	baseSprite, _, err := ebitenutil.NewImageFromFile(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load base sprite from %s: %w", basePath, err)
	}

	// Generate faction sprites using palette swapping
	green := ApplyFactionPalette(baseSprite, 0)
	blue := ApplyFactionPalette(baseSprite, 1)
	red := ApplyFactionPalette(baseSprite, 2)
	yellow := ApplyFactionPalette(baseSprite, 3)

	return &ShipClassSprites{
		baseSprite: baseSprite,
		Green:      green,
		Blue:       blue,
		Red:        red,
		Yellow:     yellow,
	}, nil
}

// LoadFactionSprites loads sprites for all ship classes
func LoadFactionSprites() (*FactionSprites, error) {
	return LoadFactionSpritesWithBasePath("assets")
}

// LoadFactionSpritesWithBasePath loads sprites with a custom base path (useful for testing)
func LoadFactionSpritesWithBasePath(basePath string) (*FactionSprites, error) {
	// Load fighter sprites
	fighterSprites, err := loadShipClassSprites(basePath + "/fighter.png")
	if err != nil {
		return nil, fmt.Errorf("failed to load fighter sprites: %w", err)
	}

	// Load destroyer sprites
	destroyerSprites, err := loadShipClassSprites(basePath + "/destroyer.png")
	if err != nil {
		return nil, fmt.Errorf("failed to load destroyer sprites: %w", err)
	}

	// Load testudon sprites
	testudonSprites, err := loadShipClassSprites(basePath + "/testudon.png")
	if err != nil {
		return nil, fmt.Errorf("failed to load testudon sprites: %w", err)
	}

	return &FactionSprites{
		Fighter:   fighterSprites,
		Destroyer: destroyerSprites,
		Testudon:  testudonSprites,
	}, nil
}
