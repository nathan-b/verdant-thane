package systems

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// FactionSprites holds the sprite images for each faction
type FactionSprites struct {
	Green  *ebiten.Image // Faction 0 - Player
	Blue   *ebiten.Image // Faction 1
	Red    *ebiten.Image // Faction 2
	Yellow *ebiten.Image // Faction 3
}

// LoadFactionSprites loads all faction-colored ship sprites from the assets directory
func LoadFactionSprites() (*FactionSprites, error) {
	green, _, err := ebitenutil.NewImageFromFile("assets/ship_green.png")
	if err != nil {
		return nil, fmt.Errorf("failed to load green ship sprite: %w", err)
	}

	blue, _, err := ebitenutil.NewImageFromFile("assets/ship_blue.png")
	if err != nil {
		return nil, fmt.Errorf("failed to load blue ship sprite: %w", err)
	}

	red, _, err := ebitenutil.NewImageFromFile("assets/ship_red.png")
	if err != nil {
		return nil, fmt.Errorf("failed to load red ship sprite: %w", err)
	}

	yellow, _, err := ebitenutil.NewImageFromFile("assets/ship_yellow.png")
	if err != nil {
		return nil, fmt.Errorf("failed to load yellow ship sprite: %w", err)
	}

	return &FactionSprites{
		Green:  green,
		Blue:   blue,
		Red:    red,
		Yellow: yellow,
	}, nil
}

// GetSpriteForFaction returns the correct sprite for a given faction ID
func (fs *FactionSprites) GetSpriteForFaction(factionID int) *ebiten.Image {
	switch factionID {
	case 0:
		return fs.Green
	case 1:
		return fs.Blue
	case 2:
		return fs.Red
	case 3:
		return fs.Yellow
	default:
		return fs.Green // Fallback to green
	}
}
