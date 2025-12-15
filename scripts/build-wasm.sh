#!/bin/bash

# build-wasm.sh - Build Verdant Thane for WebAssembly deployment
# This script compiles the game to WASM and prepares the web directory for deployment

set -e  # Exit on error

echo "Building Verdant Thane for WebAssembly..."
echo

# Check if we're in the project root
if [ ! -f "go.mod" ]; then
    echo "Error: Must run from project root directory"
    exit 1
fi

# Create web directory if it doesn't exist
mkdir -p web

# Build WASM binary
echo "Compiling WASM binary..."
GOOS=js GOARCH=wasm go build -o web/verdant-thane.wasm
echo "✓ Compiled web/verdant-thane.wasm"

# Check size
WASM_SIZE=$(du -h web/verdant-thane.wasm | cut -f1)
echo "  Binary size: $WASM_SIZE"

# Copy or download wasm_exec.js
echo
echo "Updating wasm_exec.js..."

# Try to copy from local Go installation first
GOROOT=$(go env GOROOT)
WASM_EXEC_FOUND=false

# Check Go 1.25+ location first
if [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
    cp "$GOROOT/lib/wasm/wasm_exec.js" web/
    echo "✓ Copied wasm_exec.js from Go installation (lib/wasm)"
    WASM_EXEC_FOUND=true
# Fall back to older Go location
elif [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
    cp "$GOROOT/misc/wasm/wasm_exec.js" web/
    echo "✓ Copied wasm_exec.js from Go installation (misc/wasm)"
    WASM_EXEC_FOUND=true
fi

if [ "$WASM_EXEC_FOUND" = false ]; then
    # Download from Go repository if not found locally
    echo "  wasm_exec.js not found in Go installation, downloading from GitHub..."
    curl -sL "https://raw.githubusercontent.com/golang/go/release-branch.go1.23/misc/wasm/wasm_exec.js" \
        -o web/wasm_exec.js

    if [ $? -eq 0 ] && [ -f web/wasm_exec.js ] && [ -s web/wasm_exec.js ]; then
        echo "✓ Downloaded wasm_exec.js from GitHub"
    else
        echo "Error: Failed to download wasm_exec.js"
        exit 1
    fi
fi

# Verify all required files are present
echo
echo "Verifying deployment files..."

required_files=("web/index.html" "web/wasm_exec.js" "web/verdant-thane.wasm")
all_present=true

for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        echo "✓ $file"
    else
        echo "✗ Missing: $file"
        all_present=false
    fi
done

echo

if [ "$all_present" = true ]; then
    echo "Build successful! 🎮"
    echo
    echo "To test locally, run:"
    echo "  python3 -m http.server --directory web 8080"
    echo
    echo "Then open http://localhost:8080 in your browser"
else
    echo "Build incomplete - missing required files"
    exit 1
fi
