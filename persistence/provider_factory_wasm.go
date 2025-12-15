//go:build js && wasm

package persistence

// NewProvider creates a WASM localStorage-based persistence provider
func NewProvider() PersistenceProvider {
	return newWasmProvider()
}
