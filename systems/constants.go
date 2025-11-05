package systems

// Screen dimensions
const (
	ScreenWidth  = 1024
	ScreenHeight = 768
)

// Game world dimensions
const (
	GameWidth  = 5040
	GameHeight = 5040
)

// Minimap dimensions and position
const (
	MinimapSize = 205                            // 20% of screen width (1024 * 0.2 ≈ 205)
	MinimapX    = ScreenWidth - MinimapSize - 10 // Bottom right corner with 10px margin
	MinimapY    = ScreenHeight - MinimapSize - 5 // Bottom with small margin
)
