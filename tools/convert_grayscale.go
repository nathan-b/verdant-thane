package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// convertGrayscale converts a PNG image to grayscale
func convertGrayscale(inputPath, outputPath string) error {
	// Open input image
	inFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inFile.Close()

	srcImg, err := png.Decode(inFile)
	if err != nil {
		return fmt.Errorf("failed to decode PNG: %w", err)
	}

	// Create grayscale image
	bounds := srcImg.Bounds()
	grayImg := image.NewNRGBA(bounds)

	// Convert each pixel to grayscale using luminosity method
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldColor := srcImg.At(x, y)
			r, g, b, a := oldColor.RGBA()

			// Convert to grayscale using standard luminosity formula
			// Y = 0.299*R + 0.587*G + 0.114*B
			// Values are in 16-bit format, so convert to 8-bit
			gray := uint32((299*r + 587*g + 114*b) / 1000)
			gray8 := uint8(gray >> 8)

			// Create gray color preserving original alpha
			newColor := color.NRGBA{
				R: gray8,
				G: gray8,
				B: gray8,
				A: uint8(a >> 8),
			}

			grayImg.Set(x, y, newColor)
		}
	}

	// Save output image
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, grayImg); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	fmt.Printf("Converted %s to grayscale → %s\n", inputPath, outputPath)
	return nil
}
