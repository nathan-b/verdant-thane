package systems

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/nathan/verdant-thane/components"
)

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

// GetSpriteForShip returns the correct sprite for a ship's class and faction
func (fs *FactionSprites) GetSpriteForShip(shipClass components.ShipClass, factionID int) *ebiten.Image {
	var classSprites *ShipClassSprites

	// Select ship class sprites
	switch shipClass {
	case components.Fighter:
		classSprites = fs.Fighter
	case components.Destroyer:
		classSprites = fs.Destroyer
	case components.Testudon:
		classSprites = fs.Testudon
	default:
		// Fallback to fighter
		classSprites = fs.Fighter
	}

	// Select faction color
	switch factionID {
	case 0:
		return classSprites.Green
	case 1:
		return classSprites.Blue
	case 2:
		return classSprites.Red
	case 3:
		return classSprites.Yellow
	default:
		return classSprites.Green // Fallback to green
	}
}

// GetSpriteForFaction returns the fighter sprite for a given faction ID (legacy compatibility)
func (fs *FactionSprites) GetSpriteForFaction(factionID int) *ebiten.Image {
	return fs.GetSpriteForShip(components.Fighter, factionID)
}
