package systems

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// FactionSprites holds the sprite images for each faction
type FactionSprites struct {
	baseSprite *ebiten.Image // Grayscale base sprite
	Green      *ebiten.Image // Faction 0 - Player
	Blue       *ebiten.Image // Faction 1
	Red        *ebiten.Image // Faction 2
	Yellow     *ebiten.Image // Faction 3
}

// LoadFactionSprites loads the grayscale base sprite and generates faction-colored sprites
// via palette swapping
func LoadFactionSprites() (*FactionSprites, error) {
	return LoadFactionSpritesFromPath("assets/fighter.png")
}

// LoadFactionSpritesFromPath loads faction sprites from a specific path
// This is useful for testing with different base paths
func LoadFactionSpritesFromPath(basePath string) (*FactionSprites, error) {
	// Load the grayscale base sprite
	baseSprite, _, err := ebitenutil.NewImageFromFile(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load base fighter sprite: %w", err)
	}

	// Generate faction sprites using palette swapping
	green := ApplyFactionPalette(baseSprite, 0)
	blue := ApplyFactionPalette(baseSprite, 1)
	red := ApplyFactionPalette(baseSprite, 2)
	yellow := ApplyFactionPalette(baseSprite, 3)

	return &FactionSprites{
		baseSprite: baseSprite,
		Green:      green,
		Blue:       blue,
		Red:        red,
		Yellow:     yellow,
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
