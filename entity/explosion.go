package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nathan-b/verdant-thane/config"
)

// Explosion is an animation entity that plays a destruction effect
type Explosion struct {
	ID           int
	X, Y         float64
	CurrentFrame int
	FrameTimer   int
	Sprite       *ebiten.Image // 4-frame explosion sprite sheet
	Alive        bool
}

// NewExplosion creates a new explosion animation
func NewExplosion(id int, x, y float64, sprite *ebiten.Image) *Explosion {
	return &Explosion{
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

// Update advances the explosion animation
func (e *Explosion) Update(ctx GameContext) error {
	if !e.Alive {
		return nil
	}

	// Advance frame timer
	e.FrameTimer++

	// Change frame every 5 ticks
	if e.FrameTimer >= 5 {
		e.FrameTimer = 0
		e.CurrentFrame++

		// Check if animation is complete
		if e.CurrentFrame >= config.ExplosionFrameCount {
			e.Alive = false
		}
	}

	return nil
}

// Render draws the current explosion frame to the screen
func (e *Explosion) Render(screen *ebiten.Image, cameraX, cameraY float64) {
	if !e.Alive || e.Sprite == nil {
		return
	}

	// Calculate frame position in sprite sheet
	// Assuming horizontal sprite sheet layout
	frameWidth := e.Sprite.Bounds().Dx() / config.ExplosionFrameCount
	frameHeight := e.Sprite.Bounds().Dy()

	// Get the current frame sub-image
	frameX := e.CurrentFrame * frameWidth
	frameRect := e.Sprite.Bounds()
	frameRect.Min.X = frameX
	frameRect.Max.X = frameX + frameWidth
	currentFrameImage := e.Sprite.SubImage(frameRect).(*ebiten.Image)

	// Get wrapped screen position (handles world wrapping)
	screenX, screenY := GetWrappedScreenPosition(e.X, e.Y, cameraX, cameraY)

	opts := &ebiten.DrawImageOptions{}

	// Center sprite
	opts.GeoM.Translate(-float64(frameWidth)/2, -float64(frameHeight)/2)

	// Position (relative to camera, accounting for world wrapping)
	opts.GeoM.Translate(screenX, screenY)

	screen.DrawImage(currentFrameImage, opts)
}

// GetID returns the explosion's ID
func (e *Explosion) GetID() int {
	return e.ID
}

// GetPosition returns the explosion's position
func (e *Explosion) GetPosition() (float64, float64) {
	return e.X, e.Y
}

// IsAlive returns whether the explosion animation is still playing
func (e *Explosion) IsAlive() bool {
	return e.Alive
}
