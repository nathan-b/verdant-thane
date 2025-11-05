package systems

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestApplyFactionPalette_AllFactions(t *testing.T) {
	// Create a simple test image (grayscale)
	baseSprite := ebiten.NewImage(24, 24)

	// Test all 4 factions
	for factionID := 0; factionID < 4; factionID++ {
		result := ApplyFactionPalette(baseSprite, factionID)

		if result == nil {
			t.Errorf("ApplyFactionPalette returned nil for faction %d", factionID)
		}

		// Verify dimensions are preserved
		bounds := result.Bounds()
		if bounds.Dx() != 24 || bounds.Dy() != 24 {
			t.Errorf("Faction %d sprite has wrong dimensions: %dx%d (expected 24x24)",
				factionID, bounds.Dx(), bounds.Dy())
		}
	}
}

func TestApplyFactionPalette_UnknownFaction(t *testing.T) {
	// Create a simple test image
	baseSprite := ebiten.NewImage(24, 24)

	// Test with unknown faction ID
	result := ApplyFactionPalette(baseSprite, 999)

	// Should return original sprite for unknown faction
	if result != baseSprite {
		t.Error("ApplyFactionPalette should return original sprite for unknown faction")
	}
}

func TestGetCachedFactionSprite_Caching(t *testing.T) {
	// Clear cache before test
	ClearPaletteCache()

	// Create a test sprite
	baseSprite := ebiten.NewImage(24, 24)

	// Get sprite twice for same faction
	sprite1 := GetCachedFactionSprite(baseSprite, 0)
	sprite2 := GetCachedFactionSprite(baseSprite, 0)

	// Should return the same cached instance
	if sprite1 != sprite2 {
		t.Error("GetCachedFactionSprite should return cached instance on second call")
	}

	// Clean up
	ClearPaletteCache()
}

func TestGetCachedFactionSprite_DifferentFactions(t *testing.T) {
	// Clear cache before test
	ClearPaletteCache()

	// Create a test sprite
	baseSprite := ebiten.NewImage(24, 24)

	// Get sprites for different factions
	sprite0 := GetCachedFactionSprite(baseSprite, 0)
	sprite1 := GetCachedFactionSprite(baseSprite, 1)

	// Should return different instances
	if sprite0 == sprite1 {
		t.Error("GetCachedFactionSprite should return different instances for different factions")
	}

	// Clean up
	ClearPaletteCache()
}

func TestClearPaletteCache(t *testing.T) {
	// Create a test sprite and cache it
	baseSprite := ebiten.NewImage(24, 24)
	sprite1 := GetCachedFactionSprite(baseSprite, 0)

	// Clear cache
	ClearPaletteCache()

	// Get sprite again - should be a new instance
	sprite2 := GetCachedFactionSprite(baseSprite, 0)

	if sprite1 == sprite2 {
		t.Error("ClearPaletteCache should clear the cache, causing new sprite to be generated")
	}

	// Clean up
	ClearPaletteCache()
}

func TestGetFactionColor(t *testing.T) {
	// Test all known factions
	for factionID := 0; factionID < 4; factionID++ {
		color := GetFactionColor(factionID)
		if color == nil {
			t.Errorf("GetFactionColor returned nil for faction %d", factionID)
		}
	}

	// Test unknown faction (should return white)
	unknownColor := GetFactionColor(999)
	if unknownColor == nil {
		t.Error("GetFactionColor returned nil for unknown faction")
	}
}

func TestLoadFactionSprites(t *testing.T) {
	// This is an integration test that requires assets to be present
	// When running from the systems package, assets are in ../assets
	sprites, err := LoadFactionSpritesWithBasePath("../assets")

	if err != nil {
		t.Fatalf("LoadFactionSprites failed: %v", err)
	}

	// Verify Fighter sprites were loaded
	if sprites.Fighter == nil {
		t.Fatal("Fighter sprites are nil")
	}

	// Verify Destroyer sprites were loaded
	if sprites.Destroyer == nil {
		t.Fatal("Destroyer sprites are nil")
	}

	// Verify all faction sprites were generated for Fighter
	if sprites.Fighter.Green == nil {
		t.Error("Fighter Green sprite is nil")
	}
	if sprites.Fighter.Blue == nil {
		t.Error("Fighter Blue sprite is nil")
	}
	if sprites.Fighter.Red == nil {
		t.Error("Fighter Red sprite is nil")
	}
	if sprites.Fighter.Yellow == nil {
		t.Error("Fighter Yellow sprite is nil")
	}

	// Verify base sprite was loaded
	if sprites.Fighter.baseSprite == nil {
		t.Error("Fighter base sprite is nil")
	}

	// Verify all fighter sprites have correct dimensions (24x24)
	checkDimensions := func(name string, sprite *ebiten.Image, expectedWidth, expectedHeight int) {
		if sprite == nil {
			return
		}
		bounds := sprite.Bounds()
		if bounds.Dx() != expectedWidth || bounds.Dy() != expectedHeight {
			t.Errorf("%s sprite has wrong dimensions: %dx%d (expected %dx%d)",
				name, bounds.Dx(), bounds.Dy(), expectedWidth, expectedHeight)
		}
	}

	checkDimensions("Fighter Green", sprites.Fighter.Green, 24, 24)
	checkDimensions("Fighter Blue", sprites.Fighter.Blue, 24, 24)
	checkDimensions("Fighter Red", sprites.Fighter.Red, 24, 24)
	checkDimensions("Fighter Yellow", sprites.Fighter.Yellow, 24, 24)
	checkDimensions("Fighter Base", sprites.Fighter.baseSprite, 24, 24)

	// Verify destroyer sprites have correct dimensions (40x60)
	checkDimensions("Destroyer Green", sprites.Destroyer.Green, 40, 60)
	checkDimensions("Destroyer Blue", sprites.Destroyer.Blue, 40, 60)
	checkDimensions("Destroyer Red", sprites.Destroyer.Red, 40, 60)
	checkDimensions("Destroyer Yellow", sprites.Destroyer.Yellow, 40, 60)
	checkDimensions("Destroyer Base", sprites.Destroyer.baseSprite, 40, 60)
}

func TestFactionSprites_GetSpriteForFaction(t *testing.T) {
	// Load sprites
	sprites, err := LoadFactionSpritesWithBasePath("../assets")
	if err != nil {
		t.Fatalf("LoadFactionSprites failed: %v", err)
	}

	// Test getting fighter sprite for each faction (legacy method)
	for factionID := 0; factionID < 4; factionID++ {
		sprite := sprites.GetSpriteForFaction(factionID)
		if sprite == nil {
			t.Errorf("GetSpriteForFaction returned nil for faction %d", factionID)
		}
	}

	// Test unknown faction (should fallback to green fighter)
	fallback := sprites.GetSpriteForFaction(999)
	if fallback != sprites.Fighter.Green {
		t.Error("GetSpriteForFaction should fallback to green fighter for unknown faction")
	}
}
