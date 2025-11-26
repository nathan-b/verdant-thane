package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nathan/verdant-thane/config"
)

// Impact is an animation entity that plays when a ship is hit by a projectile
type Impact struct {
	ID           int
	X, Y         float64
	CurrentFrame int
	FrameTimer   int
	Sprite       *ebiten.Image // 5-frame impact sprite sheet
	Alive        bool
}

// NewImpact creates a new impact animation
func NewImpact(id int, x, y float64, sprite *ebiten.Image) *Impact {
	return &Impact{
		ID:           id,
		X:            x,
		Y:            y,
		CurrentFrame: 0,
		FrameTimer:   0,
		Sprite:       sprite,
		Alive:        true,
	}
}

// ============================================================================
// Entity Interface Implementation
// ============================================================================

// Update advances the impact animation
func (i *Impact) Update(ctx GameContext) error {
	if !i.Alive {
		return nil
	}

	// Advance frame timer
	i.FrameTimer++

	// Change frame every 3 ticks (faster than explosion for quick hit feedback)
	if i.FrameTimer >= 3 {
		i.FrameTimer = 0
		i.CurrentFrame++

		// Check if animation is complete
		if i.CurrentFrame >= config.ImpactFrameCount {
			i.Alive = false
		}
	}

	return nil
}

// Render draws the current impact frame to the screen
func (i *Impact) Render(screen *ebiten.Image, cameraX, cameraY float64) {
	if !i.Alive || i.Sprite == nil {
		return
	}

	// Calculate frame position in sprite sheet
	// Assuming horizontal sprite sheet layout
	frameWidth := i.Sprite.Bounds().Dx() / config.ImpactFrameCount

	// Get the current frame sub-image
	frameX := i.CurrentFrame * frameWidth
	frameRect := i.Sprite.Bounds()
	frameRect.Min.X = frameX
	frameRect.Max.X = frameX + frameWidth
	currentFrameImage := i.Sprite.SubImage(frameRect).(*ebiten.Image)

	// Get wrapped screen position (handles world wrapping)
	screenX, screenY := GetWrappedScreenPosition(i.X, i.Y, cameraX, cameraY)

	opts := &ebiten.DrawImageOptions{}

	// Scale down to reasonable size (impact frames are 400x400, scale to ~40 pixels)
	const targetSize = 40.0
	scale := targetSize / float64(frameWidth)
	opts.GeoM.Scale(scale, scale)

	// Center sprite (after scaling)
	opts.GeoM.Translate(-targetSize/2, -targetSize/2)

	// Position (relative to camera, accounting for world wrapping)
	opts.GeoM.Translate(screenX, screenY)

	screen.DrawImage(currentFrameImage, opts)
}

// GetID returns the impact's ID
func (i *Impact) GetID() int {
	return i.ID
}

// GetPosition returns the impact's position
func (i *Impact) GetPosition() (float64, float64) {
	return i.X, i.Y
}

// IsAlive returns whether the impact animation is still playing
func (i *Impact) IsAlive() bool {
	return i.Alive
}
