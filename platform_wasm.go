//go:build js && wasm

package main

// Flag variables for WASM builds - all set to default values
// WASM builds don't support command-line flags since they run in the browser
var (
	// Command-line flags (for testing) - all disabled on WASM
	perfTest       = boolPtr(false)
	perfShips      = intPtr(800)
	perfFactions   = intPtr(4)
	showFPS        = boolPtr(false)
	showProfile    = boolPtr(false)
	profileStartup = boolPtr(false)
	profileNewGame = boolPtr(false)
	useDestroyer   = boolPtr(false)
	useTestudon    = boolPtr(false)

	// Gameplay tweaking flags - all use defaults
	firingCone      = intPtr(0)
	firingRate      = float64Ptr(0)
	turnRate        = float64Ptr(0)
	projectileLife  = float64Ptr(0)
	projectileSpeed = float64Ptr(0)
	maxSpeed        = float64Ptr(0)
	acceleration    = float64Ptr(0)
	aiFire          = float64Ptr(0)

	// Version flag - disabled on WASM
	showVersion = boolPtr(false)
)

// Helper functions to create pointers to literal values
func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

// parseFlags is a no-op for WASM builds
// All flag variables are already initialized to their default values above
func parseFlags() {
	// No-op: WASM builds don't support command-line flags
	// All configuration happens through the web interface or is hard-coded to defaults
}
