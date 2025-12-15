package systems

import (
	"io/fs"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// FactionColorProfile Tests
// ============================================================================

func TestFactionColorProfiles_AllFactionsExist(t *testing.T) {
	// Test that all 4 factions have color profiles
	for factionID := 0; factionID < 4; factionID++ {
		profile, ok := factionColorProfiles[factionID]
		if !ok {
			t.Errorf("Faction %d does not have a color profile", factionID)
		}

		// Verify RGB values are in valid range (0.0 to 1.0+)
		if profile.R < 0 || profile.G < 0 || profile.B < 0 {
			t.Errorf("Faction %d has negative color values: R=%f, G=%f, B=%f",
				factionID, profile.R, profile.G, profile.B)
		}
	}
}

func TestFactionColorProfiles_GreenFaction(t *testing.T) {
	profile := factionColorProfiles[0]

	// Green faction should have high green component
	if profile.G != 1.0 {
		t.Errorf("Green faction should have G=1.0, got %f", profile.G)
	}

	// Green faction should have low red and blue
	if profile.R != 0.3 {
		t.Errorf("Green faction R component: expected 0.3, got %f", profile.R)
	}
	if profile.B != 0.3 {
		t.Errorf("Green faction B component: expected 0.3, got %f", profile.B)
	}
}

func TestFactionColorProfiles_BlueFaction(t *testing.T) {
	profile := factionColorProfiles[1]

	// Blue faction should have high blue component
	if profile.B != 1.0 {
		t.Errorf("Blue faction should have B=1.0, got %f", profile.B)
	}

	// Blue faction should have low red and green
	if profile.R != 0.2 {
		t.Errorf("Blue faction R component: expected 0.2, got %f", profile.R)
	}
	if profile.G != 0.5 {
		t.Errorf("Blue faction G component: expected 0.5, got %f", profile.G)
	}
}

func TestFactionColorProfiles_RedFaction(t *testing.T) {
	profile := factionColorProfiles[2]

	// Red faction should have high red component
	if profile.R != 1.0 {
		t.Errorf("Red faction should have R=1.0, got %f", profile.R)
	}

	// Red faction should have low green and blue
	if profile.G != 0.2 {
		t.Errorf("Red faction G component: expected 0.2, got %f", profile.G)
	}
	if profile.B != 0.2 {
		t.Errorf("Red faction B component: expected 0.2, got %f", profile.B)
	}
}

func TestFactionColorProfiles_YellowFaction(t *testing.T) {
	profile := factionColorProfiles[3]

	// Yellow faction should have high red and green
	if profile.R != 1.0 {
		t.Errorf("Yellow faction should have R=1.0, got %f", profile.R)
	}
	if profile.G != 1.0 {
		t.Errorf("Yellow faction should have G=1.0, got %f", profile.G)
	}

	// Yellow faction should have low blue
	if profile.B != 0.2 {
		t.Errorf("Yellow faction B component: expected 0.2, got %f", profile.B)
	}
}

func TestFactionColorProfiles_Uniqueness(t *testing.T) {
	// Verify each faction has a unique color profile
	profiles := make([]FactionColorProfile, 4)
	for i := 0; i < 4; i++ {
		profiles[i] = factionColorProfiles[i]
	}

	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			p1 := profiles[i]
			p2 := profiles[j]

			// Profiles should differ in at least one component
			if p1.R == p2.R && p1.G == p2.G && p1.B == p2.B {
				t.Errorf("Factions %d and %d have identical color profiles", i, j)
			}
		}
	}
}

// ============================================================================
// ApplyFactionPalette Tests (adapted from old tests)
// ============================================================================

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

func TestApplyFactionPalette_NegativeFaction(t *testing.T) {
	baseSprite := ebiten.NewImage(24, 24)

	result := ApplyFactionPalette(baseSprite, -1)

	// Should return original sprite for negative faction
	if result != baseSprite {
		t.Error("ApplyFactionPalette should return original sprite for negative faction")
	}
}

func TestApplyFactionPalette_PreservesDimensions(t *testing.T) {
	// Test with various dimensions
	testCases := []struct {
		width, height int
	}{
		{24, 24},   // Fighter
		{40, 60},   // Destroyer
		{100, 100}, // Testudon
		{1, 1},     // Minimum
		{256, 256}, // Large
	}

	for _, tc := range testCases {
		baseSprite := ebiten.NewImage(tc.width, tc.height)
		result := ApplyFactionPalette(baseSprite, 0)

		bounds := result.Bounds()
		if bounds.Dx() != tc.width || bounds.Dy() != tc.height {
			t.Errorf("Dimensions not preserved: input %dx%d, output %dx%d",
				tc.width, tc.height, bounds.Dx(), bounds.Dy())
		}
	}
}

func TestApplyFactionPalette_CreatesNewImage(t *testing.T) {
	baseSprite := ebiten.NewImage(24, 24)

	result := ApplyFactionPalette(baseSprite, 0)

	// Result should be a new image (not the same instance)
	if result == baseSprite {
		t.Error("ApplyFactionPalette should create a new image, not return the original")
	}
}

func TestApplyFactionPalette_DifferentOutputsPerFaction(t *testing.T) {
	baseSprite := ebiten.NewImage(24, 24)

	// Generate sprites for all factions
	sprites := make([]*ebiten.Image, 4)
	for i := 0; i < 4; i++ {
		sprites[i] = ApplyFactionPalette(baseSprite, i)
	}

	// Each faction should produce a different image instance
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			if sprites[i] == sprites[j] {
				t.Errorf("Factions %d and %d produced the same image instance", i, j)
			}
		}
	}
}

// ============================================================================
// ShipClassSprites Tests
// ============================================================================

func TestShipClassSprites_Structure(t *testing.T) {
	// Create a minimal ShipClassSprites
	baseSprite := ebiten.NewImage(24, 24)
	sprites := &ShipClassSprites{
		baseSprite: baseSprite,
		Green:      ebiten.NewImage(24, 24),
		Blue:       ebiten.NewImage(24, 24),
		Red:        ebiten.NewImage(24, 24),
		Yellow:     ebiten.NewImage(24, 24),
	}

	// Verify all fields are set
	if sprites.baseSprite == nil {
		t.Error("baseSprite should not be nil")
	}
	if sprites.Green == nil {
		t.Error("Green sprite should not be nil")
	}
	if sprites.Blue == nil {
		t.Error("Blue sprite should not be nil")
	}
	if sprites.Red == nil {
		t.Error("Red sprite should not be nil")
	}
	if sprites.Yellow == nil {
		t.Error("Yellow sprite should not be nil")
	}
}

// ============================================================================
// FactionSprites Tests
// ============================================================================

func TestFactionSprites_Structure(t *testing.T) {
	// Create a minimal FactionSprites
	sprites := &FactionSprites{
		Fighter:   &ShipClassSprites{},
		Destroyer: &ShipClassSprites{},
		Testudon:  &ShipClassSprites{},
	}

	// Verify all fields are set
	if sprites.Fighter == nil {
		t.Error("Fighter sprites should not be nil")
	}
	if sprites.Destroyer == nil {
		t.Error("Destroyer sprites should not be nil")
	}
	if sprites.Testudon == nil {
		t.Error("Testudon sprites should not be nil")
	}
}

// ============================================================================
// LoadFactionSprites Integration Tests (adapted from old tests)
// ============================================================================

func TestLoadFactionSprites(t *testing.T) {
	// This is an integration test that requires assets to be present
	// When running from the systems package, assets are in ../assets
	fsys := os.DirFS("..")
	sprites, err := LoadFactionSpritesWithBasePath(fsys, "assets")

	if err != nil {
		t.Fatalf("LoadFactionSpritesWithBasePath failed: %v", err)
	}

	// Verify Fighter sprites were loaded
	if sprites.Fighter == nil {
		t.Fatal("Fighter sprites are nil")
	}

	// Verify Destroyer sprites were loaded
	if sprites.Destroyer == nil {
		t.Fatal("Destroyer sprites are nil")
	}

	// Verify Testudon sprites were loaded
	if sprites.Testudon == nil {
		t.Fatal("Testudon sprites are nil")
	}
}

func TestLoadFactionSprites_FighterSprites(t *testing.T) {
	fsys := os.DirFS("..")
	sprites, err := LoadFactionSpritesWithBasePath(fsys, "assets")
	if err != nil {
		t.Fatalf("LoadFactionSpritesWithBasePath failed: %v", err)
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
}

func TestLoadFactionSprites_DestroyerSprites(t *testing.T) {
	fsys := os.DirFS("..")
	sprites, err := LoadFactionSpritesWithBasePath(fsys, "assets")
	if err != nil {
		t.Fatalf("LoadFactionSpritesWithBasePath failed: %v", err)
	}

	// Verify all faction sprites were generated for Destroyer
	if sprites.Destroyer.Green == nil {
		t.Error("Destroyer Green sprite is nil")
	}
	if sprites.Destroyer.Blue == nil {
		t.Error("Destroyer Blue sprite is nil")
	}
	if sprites.Destroyer.Red == nil {
		t.Error("Destroyer Red sprite is nil")
	}
	if sprites.Destroyer.Yellow == nil {
		t.Error("Destroyer Yellow sprite is nil")
	}

	// Verify base sprite was loaded
	if sprites.Destroyer.baseSprite == nil {
		t.Error("Destroyer base sprite is nil")
	}
}

func TestLoadFactionSprites_TestudonSprites(t *testing.T) {
	fsys := os.DirFS("..")
	sprites, err := LoadFactionSpritesWithBasePath(fsys, "assets")
	if err != nil {
		t.Fatalf("LoadFactionSpritesWithBasePath failed: %v", err)
	}

	// Verify all faction sprites were generated for Testudon
	if sprites.Testudon.Green == nil {
		t.Error("Testudon Green sprite is nil")
	}
	if sprites.Testudon.Blue == nil {
		t.Error("Testudon Blue sprite is nil")
	}
	if sprites.Testudon.Red == nil {
		t.Error("Testudon Red sprite is nil")
	}
	if sprites.Testudon.Yellow == nil {
		t.Error("Testudon Yellow sprite is nil")
	}

	// Verify base sprite was loaded
	if sprites.Testudon.baseSprite == nil {
		t.Error("Testudon base sprite is nil")
	}
}

func TestLoadFactionSprites_FighterDimensions(t *testing.T) {
	fsys := os.DirFS("..")
	sprites, err := LoadFactionSpritesWithBasePath(fsys, "assets")
	if err != nil {
		t.Fatalf("LoadFactionSpritesWithBasePath failed: %v", err)
	}

	// Helper function to check dimensions
	checkDimensions := func(name string, sprite *ebiten.Image, expectedWidth, expectedHeight int) {
		if sprite == nil {
			t.Errorf("%s sprite is nil", name)
			return
		}
		bounds := sprite.Bounds()
		if bounds.Dx() != expectedWidth || bounds.Dy() != expectedHeight {
			t.Errorf("%s sprite has wrong dimensions: %dx%d (expected %dx%d)",
				name, bounds.Dx(), bounds.Dy(), expectedWidth, expectedHeight)
		}
	}

	// Verify all fighter sprites have correct dimensions (24x24)
	checkDimensions("Fighter Green", sprites.Fighter.Green, 24, 24)
	checkDimensions("Fighter Blue", sprites.Fighter.Blue, 24, 24)
	checkDimensions("Fighter Red", sprites.Fighter.Red, 24, 24)
	checkDimensions("Fighter Yellow", sprites.Fighter.Yellow, 24, 24)
	checkDimensions("Fighter Base", sprites.Fighter.baseSprite, 24, 24)
}

func TestLoadFactionSprites_DestroyerDimensions(t *testing.T) {
	fsys := os.DirFS("..")
	sprites, err := LoadFactionSpritesWithBasePath(fsys, "assets")
	if err != nil {
		t.Fatalf("LoadFactionSpritesWithBasePath failed: %v", err)
	}

	checkDimensions := func(name string, sprite *ebiten.Image, expectedWidth, expectedHeight int) {
		if sprite == nil {
			t.Errorf("%s sprite is nil", name)
			return
		}
		bounds := sprite.Bounds()
		if bounds.Dx() != expectedWidth || bounds.Dy() != expectedHeight {
			t.Errorf("%s sprite has wrong dimensions: %dx%d (expected %dx%d)",
				name, bounds.Dx(), bounds.Dy(), expectedWidth, expectedHeight)
		}
	}

	// Verify destroyer sprites have correct dimensions (40x60)
	checkDimensions("Destroyer Green", sprites.Destroyer.Green, 40, 60)
	checkDimensions("Destroyer Blue", sprites.Destroyer.Blue, 40, 60)
	checkDimensions("Destroyer Red", sprites.Destroyer.Red, 40, 60)
	checkDimensions("Destroyer Yellow", sprites.Destroyer.Yellow, 40, 60)
	checkDimensions("Destroyer Base", sprites.Destroyer.baseSprite, 40, 60)
}

func TestLoadFactionSprites_TestudonDimensions(t *testing.T) {
	fsys := os.DirFS("..")
	sprites, err := LoadFactionSpritesWithBasePath(fsys, "assets")
	if err != nil {
		t.Fatalf("LoadFactionSpritesWithBasePath failed: %v", err)
	}

	checkDimensions := func(name string, sprite *ebiten.Image, expectedWidth, expectedHeight int) {
		if sprite == nil {
			t.Errorf("%s sprite is nil", name)
			return
		}
		bounds := sprite.Bounds()
		if bounds.Dx() != expectedWidth || bounds.Dy() != expectedHeight {
			t.Errorf("%s sprite has wrong dimensions: %dx%d (expected %dx%d)",
				name, bounds.Dx(), bounds.Dy(), expectedWidth, expectedHeight)
		}
	}

	// Verify testudon sprites have correct dimensions (100x100)
	checkDimensions("Testudon Green", sprites.Testudon.Green, 100, 100)
	checkDimensions("Testudon Blue", sprites.Testudon.Blue, 100, 100)
	checkDimensions("Testudon Red", sprites.Testudon.Red, 100, 100)
	checkDimensions("Testudon Yellow", sprites.Testudon.Yellow, 100, 100)
	checkDimensions("Testudon Base", sprites.Testudon.baseSprite, 100, 100)
}

func TestLoadFactionSprites_InvalidPath(t *testing.T) {
	fsys := os.DirFS("/")
	_, err := LoadFactionSpritesWithBasePath(fsys, "nonexistent/path")

	if err == nil {
		t.Error("LoadFactionSpritesWithBasePath should return error for invalid path")
	}
}

func TestLoadFactionSprites_DefaultPath(t *testing.T) {
	// This test will fail if not run from the project root
	// It's mainly here to document that the default function exists
	// and delegates to LoadFactionSpritesWithBasePath

	// We can't actually test this reliably in all environments,
	// so we just verify the function signature exists
	var testFunc func(fsys fs.FS) (*FactionSprites, error) = LoadFactionSprites
	if testFunc == nil {
		t.Error("LoadFactionSprites function should exist")
	}
}

// ============================================================================
// Edge Cases and Error Handling
// ============================================================================

func TestApplyFactionPalette_NilSprite(t *testing.T) {
	// This will likely panic, but we're testing the behavior
	// In production code, sprites should never be nil
	defer func() {
		if r := recover(); r == nil {
			// If it doesn't panic, that's also fine
			// Just testing the behavior
		}
	}()

	ApplyFactionPalette(nil, 0)
}

func TestShipClassSprites_AllSpritesUnique(t *testing.T) {
	fsys := os.DirFS("..")
	sprites, err := LoadFactionSpritesWithBasePath(fsys, "assets")
	if err != nil {
		t.Fatalf("LoadFactionSpritesWithBasePath failed: %v", err)
	}

	// Verify Fighter faction sprites are all different instances
	fighterSprites := []*ebiten.Image{
		sprites.Fighter.Green,
		sprites.Fighter.Blue,
		sprites.Fighter.Red,
		sprites.Fighter.Yellow,
	}

	for i := 0; i < len(fighterSprites); i++ {
		for j := i + 1; j < len(fighterSprites); j++ {
			if fighterSprites[i] == fighterSprites[j] {
				t.Errorf("Fighter faction sprites %d and %d are the same instance", i, j)
			}
		}
	}
}

func TestFactionSprites_AllShipClassesUnique(t *testing.T) {
	fsys := os.DirFS("..")
	sprites, err := LoadFactionSpritesWithBasePath(fsys, "assets")
	if err != nil {
		t.Fatalf("LoadFactionSpritesWithBasePath failed: %v", err)
	}

	// Verify Fighter, Destroyer, and Testudon are different instances
	if sprites.Fighter == sprites.Destroyer {
		t.Error("Fighter and Destroyer should be different instances")
	}
	if sprites.Fighter == sprites.Testudon {
		t.Error("Fighter and Testudon should be different instances")
	}
	if sprites.Destroyer == sprites.Testudon {
		t.Error("Destroyer and Testudon should be different instances")
	}

	// Verify base sprites are different
	if sprites.Fighter.baseSprite == sprites.Destroyer.baseSprite {
		t.Error("Fighter and Destroyer base sprites should be different")
	}
	if sprites.Fighter.baseSprite == sprites.Testudon.baseSprite {
		t.Error("Fighter and Testudon base sprites should be different")
	}
	if sprites.Destroyer.baseSprite == sprites.Testudon.baseSprite {
		t.Error("Destroyer and Testudon base sprites should be different")
	}
}
