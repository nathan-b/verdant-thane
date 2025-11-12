# Verdant Thane Development Tools

This directory contains development utilities for working with assets and resources for Verdant Thane.

_NOTE_: I had AI write this documentation because I couldn't be bothered, so it might be weird. I did take a look over it and it seems reasonable.

## Usage

All tools are invoked through the main tools command:

```bash
go run ./tools <command> [options]
```

Run `go run ./tools help` to see available commands.

## Commands

### build-sounds

Generates sound effect WAV files using procedural synthesis.

**Usage:**
```bash
go run ./tools build-sounds [--type TYPE] [--output FILE] [--duration SECONDS]
```

**Options:**
- `--type` - Type of sound to generate: `laser`, `impact`, `explosion` (default: `laser`)
- `--output` - Output WAV file path (default: `out.wav`)
- `--duration` - Sound duration in seconds (default: `0.2`)

**Examples:**
```bash
# Generate default laser sound
go run ./tools build-sounds --type laser --output laser.wav

# Generate impact sound
go run ./tools build-sounds --type impact --output impact.wav

# Generate explosion sound
go run ./tools build-sounds --type explosion --output explosion.wav --duration 0.4

# Generate a longer laser effect
go run ./tools build-sounds --type laser --duration 0.5 --output custom_laser.wav
```

---

### convert-grayscale

Converts a PNG image to grayscale using the standard luminosity formula.

**Usage:**
```bash
go run ./tools convert-grayscale <input.png> <output.png>
```

**Example:**
```bash
go run ./tools convert-grayscale assets/ship_color.png assets/ship_gray.png
```

**Details:**
- Uses standard luminosity formula: Y = 0.299×R + 0.587×G + 0.114×B
- Preserves alpha channel transparency
- Output format: PNG with same dimensions as input

**Use case:** Creating grayscale base sprites for palette swapping system.

---

### extract-palette

Extracts unique colors from an image and generates a palette strip for visualization.

**Usage:**
```bash
go run ./tools extract-palette [--output FILE] <input.png>
```

**Options:**
- `--output` - Output palette strip file (default: `palette_strip.png`)

**Example:**
```bash
go run ./tools extract-palette assets/fighter.png
go run ./tools extract-palette --output my_palette.png assets/testudon.png
```

**Details:**
- Extracts all unique non-transparent colors from source image
- Sorts colors by brightness (darkest to lightest)
- Generates horizontal strip image (20px height, 1px per color)
- Useful for analyzing sprite palettes and color schemes

---

## Development Notes

### Adding New Tools

To add a new tool:

1. Create a new `.go` file in this directory (e.g., `my_tool.go`)
2. Implement your tool as a function that returns `error`:
   ```go
   func myTool(args ...string) error {
       // Implementation
       return nil
   }
   ```
3. Add a case to the switch statement in `tools_main.go`
4. Update this README with documentation

### Code Style

- Use proper error handling (return errors, don't panic)
- Provide informative error messages with context
- Print success messages with relevant details
- Use standard Go conventions (gofmt, etc.)
