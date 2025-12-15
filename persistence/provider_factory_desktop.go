//go:build !js && !wasm

package persistence

// NewProvider creates a desktop file-based persistence provider
func NewProvider() PersistenceProvider {
	return newDesktopProvider()
}
