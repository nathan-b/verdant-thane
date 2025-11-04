package systems

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// FactionColorProfile defines the color transformation for a faction
type FactionColorProfile struct {
	// ColorM matrix for transforming grayscale to faction colors
	// These values are tuned to match the extracted palettes
	R, G, B float64 // RGB multipliers for grayscale intensity
}

var factionColorProfiles = map[int]FactionColorProfile{
	0: {R: 0.3, G: 1.0, B: 0.3},   // Green (faction 0)
	1: {R: 0.2, G: 0.5, B: 1.0},   // Blue (faction 1)
	2: {R: 1.0, G: 0.2, B: 0.2},   // Red (faction 2)
	3: {R: 1.0, G: 1.0, B: 0.2},   // Yellow (faction 3)
}

// ApplyFactionPalette applies a faction-specific color transformation to a grayscale sprite
// The input sprite should be grayscale (or will be treated as such)
// Returns a new image with the faction's colors applied
func ApplyFactionPalette(baseSprite *ebiten.Image, factionID int) *ebiten.Image {
	profile, ok := factionColorProfiles[factionID]
	if !ok {
		// Unknown faction, return original sprite
		return baseSprite
	}

	// Get sprite dimensions
	bounds := baseSprite.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a new image for the result
	result := ebiten.NewImage(width, height)

	// Create ColorM for palette transformation
	var colorMatrix ebiten.ColorM

	// For a grayscale image, R=G=B, so we can use any channel as the intensity
	// We'll scale each output channel by the faction profile
	// The ColorM works by: output = colorMatrix * input
	// For grayscale to color, we want:
	//   R_out = gray * profile.R
	//   G_out = gray * profile.G
	//   B_out = gray * profile.B
	//   A_out = alpha (unchanged)

	// ColorM is applied as:
	// [R']   [m00 m01 m02 m03 m04]   [R]
	// [G'] = [m10 m11 m12 m13 m14] * [G]
	// [B']   [m20 m21 m22 m23 m24]   [B]
	// [A']   [m30 m31 m32 m33 m34]   [A]
	//                                 [1]

	// For grayscale, R=G=B, so we take the average and apply faction colors
	// We'll use the red channel as the grayscale intensity source
	colorMatrix.Scale(profile.R, profile.G, profile.B, 1.0)

	// Draw the base sprite with color transformation
	opts := &ebiten.DrawImageOptions{}
	opts.ColorM = colorMatrix
	result.DrawImage(baseSprite, opts)

	return result
}

// paletteCache caches generated faction sprites to avoid regenerating
var paletteCache = make(map[int]*ebiten.Image)

// GetCachedFactionSprite returns a cached faction sprite or generates it if not cached
func GetCachedFactionSprite(baseSprite *ebiten.Image, factionID int) *ebiten.Image {
	// Check cache
	if cached, ok := paletteCache[factionID]; ok {
		return cached
	}

	// Generate and cache
	sprite := ApplyFactionPalette(baseSprite, factionID)
	paletteCache[factionID] = sprite
	return sprite
}

// ClearPaletteCache clears the palette cache (useful for testing)
func ClearPaletteCache() {
	paletteCache = make(map[int]*ebiten.Image)
}

// GetFactionColor returns the primary color for a faction (for UI elements, etc.)
func GetFactionColor(factionID int) color.Color {
	switch factionID {
	case 0:
		return color.RGBA{0, 255, 0, 255} // Green
	case 1:
		return color.RGBA{0, 128, 255, 255} // Blue
	case 2:
		return color.RGBA{255, 0, 0, 255} // Red
	case 3:
		return color.RGBA{255, 255, 0, 255} // Yellow
	default:
		return color.White
	}
}
