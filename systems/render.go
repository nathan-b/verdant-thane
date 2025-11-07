package systems

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/nathan/verdant-thane/components"
	"github.com/nathan/verdant-thane/config"
)

// Faction colors for minimap
var factionColors = map[int]color.Color{
	0: color.RGBA{0, 255, 0, 255},   // Green
	1: color.RGBA{0, 128, 255, 255}, // Blue
	2: color.RGBA{255, 0, 0, 255},   // Red
	3: color.RGBA{255, 255, 0, 255}, // Yellow
}

var playerMarkerColor = color.RGBA{128, 255, 128, 255} // Light green for player marker

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

		// Early visibility check - skip if ship is far from visible area
		// Rough estimate with generous margin for world wrapping
		const roughMargin = 100.0
		screenX := pos.X - cameraX
		screenY := pos.Y - cameraY

		// Quick rejection test before loading sprite data
		farFromScreen := (screenX < -config.GameWidth/2-roughMargin || screenX > config.GameWidth/2+roughMargin) &&
			(screenY < -config.GameHeight/2-roughMargin || screenY > config.GameHeight/2+roughMargin)

		if farFromScreen {
			continue // Skip this ship entirely
		}

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
			if screenX < -margin || screenX > float64(config.ScreenWidth)+margin ||
				screenY < -margin || screenY > float64(config.ScreenHeight)+margin {
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
			drawShipAtPosition(pos.X+float64(config.GameWidth), pos.Y)
		}
		if pos.X > cameraX+float64(config.ScreenWidth) {
			// Ship is to the right of camera, try drawing wrapped to the left
			drawShipAtPosition(pos.X-float64(config.GameWidth), pos.Y)
		}
		if pos.Y < cameraY {
			// Ship is above camera, try drawing wrapped below
			drawShipAtPosition(pos.X, pos.Y+float64(config.GameHeight))
		}
		if pos.Y > cameraY+float64(config.ScreenHeight) {
			// Ship is below camera, try drawing wrapped above
			drawShipAtPosition(pos.X, pos.Y-float64(config.GameHeight))
		}
	}
}

// RenderProjectiles draws all projectiles (missiles have rotation, lasers do not)
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

		// Check if this projectile has rotation (e.g., missiles)
		var rotation float64
		hasRotation := entry.HasComponent(components.Rotation)
		if hasRotation {
			rot := components.Rotation.Get(entry)
			rotation = rot.Angle
		}

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
			if screenX < -margin || screenX > float64(config.ScreenWidth)+margin ||
				screenY < -margin || screenY > float64(config.ScreenHeight)+margin {
				return
			}

			op := &ebiten.DrawImageOptions{}

			// Apply rotation if this projectile has rotation component
			if hasRotation {
				// Center sprite, rotate, then translate to world position
				op.GeoM.Translate(-width/2, -height/2)
				op.GeoM.Rotate(rotation)
				op.GeoM.Translate(worldX, worldY)
				op.GeoM.Translate(-cameraX, -cameraY)
			} else {
				// Simple centered positioning (no rotation)
				op.GeoM.Translate(worldX-width/2, worldY-height/2)
				op.GeoM.Translate(-cameraX, -cameraY)
			}

			screen.DrawImage(sprite.Image, op)
		}

		// Draw projectile at its primary position
		drawProjectileAtPosition(pos.X, pos.Y)

		// Handle world wrapping: draw at wrapped positions if near edges
		if pos.X < cameraX {
			// Projectile is to the left of camera, try drawing wrapped to the right
			drawProjectileAtPosition(pos.X+float64(config.GameWidth), pos.Y)
		}
		if pos.X > cameraX+float64(config.ScreenWidth) {
			// Projectile is to the right of camera, try drawing wrapped to the left
			drawProjectileAtPosition(pos.X-float64(config.GameWidth), pos.Y)
		}
		if pos.Y < cameraY {
			// Projectile is above camera, try drawing wrapped below
			drawProjectileAtPosition(pos.X, pos.Y+float64(config.GameHeight))
		}
		if pos.Y > cameraY+float64(config.ScreenHeight) {
			// Projectile is below camera, try drawing wrapped above
			drawProjectileAtPosition(pos.X, pos.Y-float64(config.GameHeight))
		}
	}
}

// RenderExplosions draws explosion animations
func RenderExplosions(w donburi.World, screen *ebiten.Image, cameraX, cameraY float64) {
	query := donburi.NewQuery(filter.Contains(
		components.IsExplosion,
		components.Position,
		components.Explosion,
		components.Sprite,
	))

	for entry := range query.Iter(w) {
		pos := components.Position.Get(entry)
		explosion := components.Explosion.Get(entry)
		sprite := components.Sprite.Get(entry)

		// Each frame is 100x70 pixels, sprite sheet is horizontal
		frameWidth := 100.0
		frameHeight := 70.0
		srcX := float64(explosion.CurrentFrame) * frameWidth

		op := &ebiten.DrawImageOptions{}

		// Center the explosion on the position
		op.GeoM.Translate(-frameWidth/2, -frameHeight/2)
		op.GeoM.Translate(pos.X, pos.Y)
		op.GeoM.Translate(-cameraX, -cameraY)

		// Draw the current frame from sprite sheet
		screen.DrawImage(
			sprite.Image.SubImage(image.Rect(
				int(srcX), 0,
				int(srcX+frameWidth), int(frameHeight),
			)).(*ebiten.Image),
			op,
		)
	}
}

// RenderBeams draws laser beams from Testudons to their targets
func RenderBeams(w donburi.World, screen *ebiten.Image, cameraX, cameraY float64) {
	query := donburi.NewQuery(filter.Contains(
		components.BeamWeapon,
		components.Position,
		components.Faction,
	))

	for entry := range query.Iter(w) {
		beamWeapon := components.BeamWeapon.Get(entry)

		// Only draw beam if there's an active firing target
		// FiringAtEntity tracks the target currently being fired at (may differ from AI navigation target)
		if !w.Valid(beamWeapon.FiringAtEntity) {
			continue
		}

		// Get attacker and target positions
		attackerPos := components.Position.Get(entry)
		targetEntry := w.Entry(beamWeapon.FiringAtEntity)

		if !targetEntry.HasComponent(components.Position) {
			continue
		}

		targetPos := components.Position.Get(targetEntry)
		faction := components.Faction.Get(entry)

		// Get beam color based on faction
		beamColor, ok := factionColors[faction.ID]
		if !ok {
			beamColor = color.White
		}

		// Helper function to draw beam line at specific world positions
		drawBeamLine := func(fromX, fromY, toX, toY float64) {
			// Convert world coordinates to screen coordinates
			screenFromX := float32(fromX - cameraX)
			screenFromY := float32(fromY - cameraY)
			screenToX := float32(toX - cameraX)
			screenToY := float32(toY - cameraY)

			// Only draw if either endpoint is visible on screen
			margin := float32(100.0)
			onScreen := (screenFromX > -margin && screenFromX < float32(config.ScreenWidth)+margin && screenFromY > -margin && screenFromY < float32(config.ScreenHeight)+margin) ||
				(screenToX > -margin && screenToX < float32(config.ScreenWidth)+margin && screenToY > -margin && screenToY < float32(config.ScreenHeight)+margin)

			if !onScreen {
				return
			}

			// Draw the beam as a thick line
			vector.StrokeLine(screen, screenFromX, screenFromY, screenToX, screenToY, 2.0, beamColor, false)
		}

		// Draw beam at primary positions
		drawBeamLine(attackerPos.X, attackerPos.Y, targetPos.X, targetPos.Y)

		// Handle world wrapping: draw beam accounting for toroidal topology
		// Calculate shortest distance (might wrap)
		dx := targetPos.X - attackerPos.X
		dy := targetPos.Y - attackerPos.Y

		// Wrap X if needed
		wrappedTargetX := targetPos.X
		if dx > float64(config.GameWidth)/2 {
			wrappedTargetX = targetPos.X - float64(config.GameWidth)
		} else if dx < -float64(config.GameWidth)/2 {
			wrappedTargetX = targetPos.X + float64(config.GameWidth)
		}

		// Wrap Y if needed
		wrappedTargetY := targetPos.Y
		if dy > float64(config.GameHeight)/2 {
			wrappedTargetY = targetPos.Y - float64(config.GameHeight)
		} else if dy < -float64(config.GameHeight)/2 {
			wrappedTargetY = targetPos.Y + float64(config.GameHeight)
		}

		// Draw wrapped beam if coordinates changed
		if wrappedTargetX != targetPos.X || wrappedTargetY != targetPos.Y {
			drawBeamLine(attackerPos.X, attackerPos.Y, wrappedTargetX, wrappedTargetY)
		}
	}
}

// RenderMinimap draws the minimap showing all ships in the game world
func RenderMinimap(w donburi.World, screen *ebiten.Image, playerEntity donburi.Entity) {
	// Draw minimap background (dark semi-transparent box)
	vector.FillRect(screen,
		float32(config.MinimapX), float32(config.MinimapY),
		float32(config.MinimapSize), float32(config.MinimapSize),
		color.RGBA{0, 0, 0, 180}, true)

	// Draw minimap border
	vector.StrokeRect(screen,
		float32(config.MinimapX), float32(config.MinimapY),
		float32(config.MinimapSize), float32(config.MinimapSize),
		2, color.RGBA{100, 100, 100, 255}, false)

	// Scale factor: minimap pixels per world units
	scale := float64(config.MinimapSize) / float64(config.GameWidth)

	// Helper function to convert world coordinates to minimap screen coordinates
	worldToMinimap := func(worldX, worldY float64) (float32, float32) {
		minimapLocalX := worldX * scale
		minimapLocalY := worldY * scale
		screenX := float32(config.MinimapX) + float32(minimapLocalX)
		screenY := float32(config.MinimapY) + float32(minimapLocalY)
		return screenX, screenY
	}

	// Draw all ships on minimap
	query := donburi.NewQuery(filter.Contains(components.IsShip, components.Position, components.Faction, components.Ship))
	for entry := range query.Iter(w) {
		pos := components.Position.Get(entry)
		faction := components.Faction.Get(entry)
		ship := components.Ship.Get(entry)

		screenX, screenY := worldToMinimap(pos.X, pos.Y)

		// Check if this is the player ship
		isPlayer := entry.Entity() == playerEntity

		// Get faction color
		c, ok := factionColors[faction.ID]
		if !ok {
			c = color.White // Fallback color
		}

		if isPlayer {
			// Draw player as a light green plus sign (5 pixels high × 5 wide)
			// Vertical line (5 pixels high)
			vector.FillRect(screen, screenX, screenY-2, 1, 5, playerMarkerColor, false)
			// Horizontal line (5 pixels wide)
			vector.FillRect(screen, screenX-2, screenY, 5, 1, playerMarkerColor, false)
		} else if ship.Class == components.Testudon {
			// Draw Testudons as 3x3 squares (centered on position)
			vector.FillRect(screen, screenX-1, screenY-1, 3, 3, c, false)
		} else {
			// Draw regular ships (Fighters, Destroyers) as a single pixel
			vector.FillRect(screen, screenX, screenY, 1, 1, c, false)
		}
	}
}
