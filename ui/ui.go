package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Button represents a clickable UI button
type Button struct {
	Label   string
	X, Y    float64 // Position
	Width   float64
	Height  float64
	OnClick func() // Callback when clicked
}

// Dialog represents a UI dialog box with buttons
type Dialog struct {
	Title   string
	X, Y    float64 // Position (top-left corner)
	Width   float64
	Height  float64
	Buttons []Button
}

// RenderButton draws a dialog button with border and centered text
func RenderButton(screen *ebiten.Image, parent *Dialog, button *Button, font *text.GoTextFace) {
	// Set coordinates relative to dialog
	buttonX := parent.X + button.X
	buttonY := parent.Y + button.Y

	// Button background
	vector.FillRect(screen,
		float32(buttonX), float32(buttonY),
		float32(button.Width), float32(button.Height),
		color.RGBA{40, 40, 40, 255}, false)

	// Button border
	vector.StrokeRect(screen,
		float32(buttonX), float32(buttonY),
		float32(button.Width), float32(button.Height),
		2, color.RGBA{150, 150, 150, 255}, false)

	// Button text (centered)
	textWidth, textHeight := text.Measure(button.Label, font, 0)
	textX := buttonX + (button.Width-textWidth)/2
	// Center the text: place baseline so the visible glyphs are centered
	// Baseline is typically at ~75% down from top of text bounding box
	textY := buttonY + (button.Height-textHeight)/2 + textHeight*0.75

	textOp := &text.DrawOptions{}
	textOp.GeoM.Translate(textX, textY)
	textOp.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, button.Label, font, textOp)
}

// RenderDialog draws a dialog background and all its buttons
func RenderDialog(screen *ebiten.Image, dialog *Dialog, font *text.GoTextFace) {
	// Dialog background
	vector.FillRect(screen,
		float32(dialog.X), float32(dialog.Y),
		float32(dialog.Width), float32(dialog.Height),
		color.RGBA{20, 20, 20, 230}, false)

	// Dialog border
	vector.StrokeRect(screen,
		float32(dialog.X), float32(dialog.Y),
		float32(dialog.Width), float32(dialog.Height),
		3, color.RGBA{100, 100, 100, 255}, false)

	// Render all buttons
	for i := range dialog.Buttons {
		RenderButton(screen, dialog, &dialog.Buttons[i], font)
	}
}

// CheckButtonClick checks if a mouse click at (mouseX, mouseY) hit any button in the dialog
// Returns the index of the clicked button, or -1 if no button was clicked
func CheckButtonClick(dialog *Dialog, mouseX, mouseY int) int {
	mx := float64(mouseX)
	my := float64(mouseY)

	for i, button := range dialog.Buttons {
		// Convert button's relative coordinates to absolute screen coordinates
		buttonX := dialog.X + button.X
		buttonY := dialog.Y + button.Y

		// Check if click is within button bounds
		if mx >= buttonX && mx <= buttonX+button.Width &&
			my >= buttonY && my <= buttonY+button.Height {
			return i
		}
	}

	return -1
}

// CheckCloseButtonClick checks if a mouse click hit the X close button in the dialog's upper-right corner
// Returns true if the close button was clicked
func CheckCloseButtonClick(dialog *Dialog, mouseX, mouseY int) bool {
	closeButtonSize := 24.0
	closeButtonX := dialog.X + dialog.Width - closeButtonSize - 8
	closeButtonY := dialog.Y + 8

	mx := float64(mouseX)
	my := float64(mouseY)

	return mx >= closeButtonX && mx <= closeButtonX+closeButtonSize &&
		my >= closeButtonY && my <= closeButtonY+closeButtonSize
}

// DrawCloseButton draws an X close button in the upper-right corner of the dialog
func DrawCloseButton(screen *ebiten.Image, dialog *Dialog) {
	closeButtonSize := 24.0
	closeButtonX := dialog.X + dialog.Width - closeButtonSize - 8
	closeButtonY := dialog.Y + 8

	// Draw button background
	vector.FillRect(screen, float32(closeButtonX), float32(closeButtonY), float32(closeButtonSize), float32(closeButtonSize), color.RGBA{60, 60, 60, 255}, false)
	// Draw button border
	vector.StrokeRect(screen, float32(closeButtonX), float32(closeButtonY), float32(closeButtonSize), float32(closeButtonSize), 1, color.RGBA{100, 100, 100, 255}, false)

	// Draw X
	padding := 6.0
	x1 := float32(closeButtonX + padding)
	y1 := float32(closeButtonY + padding)
	x2 := float32(closeButtonX + closeButtonSize - padding)
	y2 := float32(closeButtonY + closeButtonSize - padding)

	// Draw two diagonal lines to form X
	vector.StrokeLine(screen, x1, y1, x2, y2, 2, color.White, false)
	vector.StrokeLine(screen, x2, y1, x1, y2, 2, color.White, false)
}
