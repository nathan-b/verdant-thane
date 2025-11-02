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

		bounds := sprite.Image.Bounds()
		width := float64(bounds.Dx())
		height := float64(bounds.Dy())

		// Helper function to draw ship at a specific world position
		drawShipAtPosition := func(worldX, worldY float64) {
			// Only draw if potentially visible on screen
			screenX := worldX - cameraX
			screenY := worldY - cameraY

			// Check if ship might be visible (with some margin for sprite size)
			margin := width
			if height > margin {
				margin = height
			}
			if screenX < -margin || screenX > float64(ScreenWidth)+margin ||
				screenY < -margin || screenY > float64(ScreenHeight)+margin {
				return
			}

			op := &ebiten.DrawImageOptions{}

			// Apply transformations:
			// 1. Translate to center around origin (for rotation)
			op.GeoM.Translate(-width/2, -height/2)
			// 2. Rotate around origin
			op.GeoM.Rotate(rot.Angle)
			// 3. Translate to position in world
			op.GeoM.Translate(worldX, worldY)
			// 4. Apply camera offset
			op.GeoM.Translate(-cameraX, -cameraY)

			screen.DrawImage(sprite.Image, op)
		}

		// Draw ship at its primary position
		drawShipAtPosition(pos.X, pos.Y)

		// Handle world wrapping: draw at wrapped positions if near edges
		if pos.X < cameraX {
			// Ship is to the left of camera, try drawing wrapped to the right
			drawShipAtPosition(pos.X+float64(GameWidth), pos.Y)
		}
		if pos.X > cameraX+float64(ScreenWidth) {
			// Ship is to the right of camera, try drawing wrapped to the left
			drawShipAtPosition(pos.X-float64(GameWidth), pos.Y)
		}
		if pos.Y < cameraY {
			// Ship is above camera, try drawing wrapped below
			drawShipAtPosition(pos.X, pos.Y+float64(GameHeight))
		}
		if pos.Y > cameraY+float64(ScreenHeight) {
			// Ship is below camera, try drawing wrapped above
			drawShipAtPosition(pos.X, pos.Y-float64(GameHeight))
		}
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

		bounds := sprite.Image.Bounds()
		width := float64(bounds.Dx())
		height := float64(bounds.Dy())

		// Helper function to draw projectile at a specific world position
		drawProjectileAtPosition := func(worldX, worldY float64) {
			// Only draw if visible on screen
			screenX := worldX - cameraX
			screenY := worldY - cameraY

			// Check if projectile is visible (with small margin for sprite size)
			margin := width
			if height > margin {
				margin = height
			}
			if screenX < -margin || screenX > float64(ScreenWidth)+margin ||
				screenY < -margin || screenY > float64(ScreenHeight)+margin {
				return
			}

			op := &ebiten.DrawImageOptions{}

			// Simple centered positioning (no rotation needed)
			op.GeoM.Translate(worldX-width/2, worldY-height/2)
			op.GeoM.Translate(-cameraX, -cameraY)

			screen.DrawImage(sprite.Image, op)
		}

		// Draw projectile at its primary position
		drawProjectileAtPosition(pos.X, pos.Y)

		// Handle world wrapping: draw at wrapped positions if near edges
		if pos.X < cameraX {
			// Projectile is to the left of camera, try drawing wrapped to the right
			drawProjectileAtPosition(pos.X+float64(GameWidth), pos.Y)
		}
		if pos.X > cameraX+float64(ScreenWidth) {
			// Projectile is to the right of camera, try drawing wrapped to the left
			drawProjectileAtPosition(pos.X-float64(GameWidth), pos.Y)
		}
		if pos.Y < cameraY {
			// Projectile is above camera, try drawing wrapped below
			drawProjectileAtPosition(pos.X, pos.Y+float64(GameHeight))
		}
		if pos.Y > cameraY+float64(ScreenHeight) {
			// Projectile is below camera, try drawing wrapped above
			drawProjectileAtPosition(pos.X, pos.Y-float64(GameHeight))
		}
	}
}
