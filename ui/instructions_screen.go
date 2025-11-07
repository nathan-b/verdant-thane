package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/nathan/verdant-thane/config"
)

// CreateInstructionsDialog creates the instructions screen dialog
func CreateInstructionsDialog() *Dialog {
	return &Dialog{
		Title:   "Instructions",
		X:       float64(config.ScreenWidth)/2 - 300,
		Y:       80,
		Width:   600,
		Height:  350,
		Buttons: []Button{}, // No buttons, uses X close button
	}
}

// DrawInstructionsText draws the instructions text content
func DrawInstructionsText(screen *ebiten.Image, font *text.GoTextFaceSource) {
	hudFont := &text.GoTextFace{
		Source: font,
		Size:   16,
	}

	textColor := color.RGBA{255, 255, 255, 255}
	startY := 140.0
	lineHeight := 25.0
	centerX := float64(config.ScreenWidth) / 2

	// Instructions text from PLAN.md
	instructions := []string{
		"Use A / D to turn and W / S to control your speed. Mouse fires.",
		"",
		"There is no friendly fire or collisions. Shoot and fly freely!",
		"",
		"The map loops infinitely (like Asteroids).",
		"",
		"When your ship is destroyed, you'll spectate allied ships.",
		"Use A / D to cycle through ships, SPACE to take control.",
		"(You cannot take control of Testudons)",
	}

	currentY := startY
	for _, line := range instructions {
		if line == "" {
			currentY += lineHeight / 2
			continue
		}

		lineWidth, _ := text.Measure(line, hudFont, 0)
		lineOp := &text.DrawOptions{}
		lineOp.GeoM.Translate(centerX-lineWidth/2, currentY)
		lineOp.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, line, hudFont, lineOp)

		currentY += lineHeight
	}
}
