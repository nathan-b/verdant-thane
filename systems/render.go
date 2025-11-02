package systems

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
)

// RenderShips draws all ships with sprites, rotation, and position
func RenderShips(w donburi.World, screen *ebiten.Image, cameraX, cameraY float64) {
	query := donburi.NewQuery(
		filter.Contains(
			components.IsShip,
			components.Position,
			components.Rotation,
			components.Sprite,
		),
	)

	for entry := range query.Iter(w) {
		pos := components.Position.Get(entry)
		rot := components.Rotation.Get(entry)
		sprite := components.Sprite.Get(entry)

		op := &ebiten.DrawImageOptions{}

		bounds := sprite.Image.Bounds()
		width := float64(bounds.Dx())
		height := float64(bounds.Dy())

		// Apply transformations:
		// 1. Translate to center around origin (for rotation)
		op.GeoM.Translate(-width/2, -height/2)
		// 2. Rotate around origin
		op.GeoM.Rotate(rot.Angle)
		// 3. Translate to position in world
		op.GeoM.Translate(pos.X, pos.Y)
		// 4. Apply camera offset
		op.GeoM.Translate(-cameraX, -cameraY)

		screen.DrawImage(sprite.Image, op)
	}
}

// RenderProjectiles draws all projectiles (no rotation needed for square sprites)
func RenderProjectiles(w donburi.World, screen *ebiten.Image, cameraX, cameraY float64) {
	query := donburi.NewQuery(
		filter.Contains(
			components.IsProjectile,
			components.Position,
			components.Sprite,
		),
	)

	for entry := range query.Iter(w) {
		pos := components.Position.Get(entry)
		sprite := components.Sprite.Get(entry)

		op := &ebiten.DrawImageOptions{}

		bounds := sprite.Image.Bounds()
		width := float64(bounds.Dx())
		height := float64(bounds.Dy())

		// Simple centered positioning (no rotation needed)
		op.GeoM.Translate(pos.X-width/2, pos.Y-height/2)
		op.GeoM.Translate(-cameraX, -cameraY)

		screen.DrawImage(sprite.Image, op)
	}
}
