package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sort"
)

// brightness calculates the perceived brightness of a color using the luminance formula
// I have no idea how this works, I just looked it up, see https://www.w3.org/TR/AERT/#color-contrast
// it seems to do what I want
func brightness(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	// Convert from 16-bit to 8-bit values and calculate weighted luminance
	// Standard weights for perceived brightness: 0.299*R + 0.587*G + 0.114*B
	return 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
}

// extractPalette extracts unique colors from an image and creates a palette strip
func extractPalette(inputPath, outputPath string) error {
	// Open source image
	srcFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer srcFile.Close()

	img, _, err := image.Decode(srcFile)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Collect unique colors
	colorMap := make(map[color.Color]bool)
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			// Ignore fully transparent pixels
			if _, _, _, a := c.RGBA(); a != 0 {
				colorMap[c] = true
			}
		}
	}

	// Convert map to slice
	colorsList := make([]color.Color, 0, len(colorMap))
	for c := range colorMap {
		colorsList = append(colorsList, c)
	}

	if len(colorsList) == 0 {
		return fmt.Errorf("no non-transparent colors found in image")
	}

	// Sort colors by brightness (dark to light)
	sort.Slice(colorsList, func(i, j int) bool {
		return brightness(colorsList[i]) < brightness(colorsList[j])
	})

	// Create a horizontal palette strip image
	const stripHeight = 20 // Arbitrary because a single pixel height is annonying to view
	stripWidth := len(colorsList)
	outImg := image.NewRGBA(image.Rect(0, 0, stripWidth, stripHeight))

	// Draw each color as a vertical bar
	for i, c := range colorsList {
		rect := image.Rect(i, 0, i+1, stripHeight)
		draw.Draw(outImg, rect, &image.Uniform{c}, image.Point{}, draw.Src)
	}

	// Save as PNG
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, outImg); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	fmt.Printf("Extracted %d colors from %s → %s\n", len(colorsList), inputPath, outputPath)
	return nil
}
