.PHONY: all build build-wasm test test-all clean install-hooks help

# Default target
all: test build build-wasm

# Build desktop version
build:
	@echo "Building desktop version..."
	go build -o verdant-thane .
	@echo "✓ Desktop build complete: ./verdant-thane"

# Build WASM version
build-wasm:
	@echo "Building WASM version..."
	@./scripts/build-wasm.sh

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "✓ Coverage report: coverage.html"

# Test all platforms (desktop + WASM compilation)
test-all: test
	@echo "Testing desktop build..."
	@go build -o /tmp/verdant-test . && rm /tmp/verdant-test
	@echo "✓ Desktop build OK"
	@echo "Testing WASM build..."
	@GOOS=js GOARCH=wasm go build -o /tmp/verdant-wasm-test . && rm /tmp/verdant-wasm-test
	@echo "✓ WASM build OK"

# Run the game
run:
	@go run .

# Run with performance testing
run-perf:
	@go run . -perf -ships 800 -factions 4 -fps

# Serve WASM build locally
serve-wasm: build-wasm
	@echo "Starting local server at http://localhost:8080"
	@echo "Press Ctrl+C to stop"
	@python3 -m http.server --directory web 8080

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

# Install pre-commit hooks
install-hooks:
	@echo "Installing pre-commit hooks..."
	@cp scripts/pre-commit-hook .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✓ Pre-commit hooks installed"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f verdant-thane
	@rm -rf build/
	@rm -f coverage.txt coverage.html
	@echo "✓ Clean complete"

# Check for WASM compatibility issues
check-wasm:
	@echo "Checking for WASM compatibility issues..."
	@echo "→ Checking for forbidden imports in WASM files..."
	@! grep -r "\"os/user\"" --include="*_wasm.go" . 2>/dev/null || (echo "Error: WASM files should not import os/user" && exit 1)
	@echo "→ Checking WASM build compiles..."
	@GOOS=js GOARCH=wasm go build -o /tmp/verdant-wasm-check . && rm /tmp/verdant-wasm-check
	@echo "✓ WASM compatibility OK"

# Help
help:
	@echo "Verdant Thane - Build Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all            - Run tests and build both desktop and WASM versions (default)"
	@echo "  build          - Build desktop version"
	@echo "  build-wasm     - Build WASM version"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-all       - Run tests and verify both builds compile"
	@echo "  run            - Run the game (desktop)"
	@echo "  run-perf       - Run with performance testing mode"
	@echo "  serve-wasm     - Build and serve WASM version locally"
	@echo "  fmt            - Format all Go code"
	@echo "  install-hooks  - Install pre-commit hooks"
	@echo "  check-wasm     - Check for WASM compatibility issues"
	@echo "  clean          - Remove build artifacts"
	@echo "  help           - Show this help message"
