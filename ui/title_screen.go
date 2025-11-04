package ui

const (
	// Screen dimensions (should match systems.ScreenWidth/ScreenHeight)
	screenWidth  = 1024
	screenHeight = 768

	// Title screen layout
	titleY         = 150.0 // Y position of title text
	dialogWidth    = 300.0
	buttonWidth    = 250.0
	buttonHeight   = 40.0
	buttonSpacing  = 15.0              // Vertical spacing between buttons
	internalMargin = 2 * buttonSpacing // Distance from top/bottom of dialog to first/last button
)

// CreateTitleScreen creates and returns a Dialog for the title screen
func CreateTitleScreen() *Dialog {
	// Create buttons centered within dialog
	buttonX := (dialogWidth - buttonWidth) / 2
	buttonStartY := internalMargin // Start buttons with top margin

	buttons := []Button{
		{
			Label:  "Play Game",
			X:      buttonX,
			Y:      buttonStartY,
			Width:  buttonWidth,
			Height: buttonHeight,
		},
		{
			Label:  "Settings",
			X:      buttonX,
			Y:      buttonStartY + buttonHeight + buttonSpacing,
			Width:  buttonWidth,
			Height: buttonHeight,
		},
		{
			Label:  "Instructions",
			X:      buttonX,
			Y:      buttonStartY + 2 * (buttonHeight + buttonSpacing),
			Width:  buttonWidth,
			Height: buttonHeight,
		},
		{
			Label:  "High Scores",
			X:      buttonX,
			Y:      buttonStartY + 3 * (buttonHeight + buttonSpacing),
			Width:  buttonWidth,
			Height: buttonHeight,
		},
	}

	// Calculate dialog height based on buttons
	totalButtonHeight := float64(len(buttons)) * buttonHeight
	totalSpacing := float64(len(buttons) - 1) * buttonSpacing + (2 * internalMargin)
	dialogHeight := totalButtonHeight + totalSpacing

	// Center the dialog
	dialogX := (float64(screenWidth) - dialogWidth) / 2
	dialogY := (float64(screenHeight) - dialogHeight) / 2

	return &Dialog{
		Title:   "Verdant Thane",
		X:       dialogX,
		Y:       dialogY,
		Width:   dialogWidth,
		Height:  dialogHeight,
		Buttons: buttons,
	}
}

// GetTitleY returns the Y position where the title should be drawn
func GetTitleY() float64 {
	return titleY
}
