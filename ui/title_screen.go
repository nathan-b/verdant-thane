package ui

import "github.com/nathan/verdant-thane/config"

// CreateTitleScreen creates and returns a Dialog for the title screen
func CreateTitleScreen() *Dialog {
	// Create buttons centered within dialog
	buttonX := (config.DialogWidth - config.ButtonWidth) / 2
	buttonStartY := config.InternalMargin // Start buttons with top margin

	buttons := []Button{
		{
			Label:  "Play Game",
			X:      buttonX,
			Y:      buttonStartY,
			Width:  config.ButtonWidth,
			Height: config.ButtonHeight,
		},
		{
			Label:  "Settings",
			X:      buttonX,
			Y:      buttonStartY + config.ButtonHeight + config.ButtonSpacing,
			Width:  config.ButtonWidth,
			Height: config.ButtonHeight,
		},
		{
			Label:  "Instructions",
			X:      buttonX,
			Y:      buttonStartY + 2*(config.ButtonHeight+config.ButtonSpacing),
			Width:  config.ButtonWidth,
			Height: config.ButtonHeight,
		},
		{
			Label:  "High Scores",
			X:      buttonX,
			Y:      buttonStartY + 3*(config.ButtonHeight+config.ButtonSpacing),
			Width:  config.ButtonWidth,
			Height: config.ButtonHeight,
		},
	}

	// Calculate dialog height based on buttons
	totalButtonHeight := float64(len(buttons)) * config.ButtonHeight
	totalSpacing := float64(len(buttons)-1)*config.ButtonSpacing + (2 * config.InternalMargin)
	dialogHeight := totalButtonHeight + totalSpacing

	// Center the dialog
	dialogX := (float64(config.ScreenWidth) - config.DialogWidth) / 2
	dialogY := (float64(config.ScreenHeight) - dialogHeight) / 2

	return &Dialog{
		Title:   "Verdant Thane",
		X:       dialogX,
		Y:       dialogY,
		Width:   config.DialogWidth,
		Height:  dialogHeight,
		Buttons: buttons,
	}
}

// GetTitleY returns the Y position where the title should be drawn
func GetTitleY() float64 {
	return config.TitleY
}
