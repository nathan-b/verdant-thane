package entity

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Particle is a simple visual effect entity (e.g., afterburner exhaust)
type Particle struct {
	ID       int
	X, Y     float64
	VX, VY   float64 // Velocity (for trailing effect)
	Lifetime int     // Ticks remaining
	MaxLife  int     // Total lifetime for fade calculation
	Color    color.RGBA
	Size     float64
	Alive    bool
}

// NewAfterburnerParticle creates an orange particle for afterburner exhaust
func NewAfterburnerParticle(id int, x, y, vx, vy float64) *Particle {
	lifetime := 15 // 15 ticks = 0.25 seconds at 60 TPS
	return &Particle{
		ID:       id,
		X:        x,
		Y:        y,
		VX:       vx * 0.3, // Particles move slower than ship (trailing effect)
		VY:       vy * 0.3,
		Lifetime: lifetime,
		MaxLife:  lifetime,
		Color:    color.RGBA{255, 165, 0, 255}, // Orange
		Size:     2.0,
		Alive:    true,
	}
}

// ============================================================================
// Entity Interface Implementation
// ============================================================================

// Update decrements the particle lifetime
func (p *Particle) Update(ctx GameContext) error {
	if !p.Alive {
		return nil
	}

	// Move particle
	p.X += p.VX
	p.Y += p.VY

	// Wrap position
	p.X, p.Y = WrapPosition(p.X, p.Y)

	// Decrement lifetime
	p.Lifetime--
	if p.Lifetime <= 0 {
		p.Alive = false
	}

	return nil
}

// Render draws the particle as a small colored circle with fade-out
func (p *Particle) Render(screen *ebiten.Image, cameraX, cameraY float64) {
	if !p.Alive {
		return
	}

	// Calculate alpha fade based on remaining lifetime
	alphaRatio := float64(p.Lifetime) / float64(p.MaxLife)
	alpha := uint8(float64(p.Color.A) * alphaRatio)

	fadeColor := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}

	// Get wrapped screen position (handles world wrapping)
	screenX, screenY := GetWrappedScreenPosition(p.X, p.Y, cameraX, cameraY)

	// Draw small circle
	vector.FillCircle(screen, float32(screenX), float32(screenY), float32(p.Size), fadeColor, false)
}

// GetID returns the particle's ID
func (p *Particle) GetID() int {
	return p.ID
}

// GetPosition returns the particle's position
func (p *Particle) GetPosition() (float64, float64) {
	return p.X, p.Y
}

// IsAlive returns whether the particle is still active
func (p *Particle) IsAlive() bool {
	return p.Alive
}
