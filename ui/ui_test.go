package ui

import (
	"testing"
)

func TestCheckButtonClick_HitButton(t *testing.T) {
	dialog := &Dialog{
		Title:  "Test Dialog",
		X:      100,
		Y:      100,
		Width:  200,
		Height: 200,
		Buttons: []Button{
			{
				Label:  "Button 1",
				X:      10, // Relative to dialog
				Y:      20, // Relative to dialog
				Width:  180,
				Height: 40,
			},
		},
	}
	// Button absolute position: (100+10, 100+20) = (110, 120)
	// Button absolute bounds: (110, 120) to (290, 160)

	// Click in the center of the button
	buttonIndex := CheckButtonClick(dialog, 200, 140)
	if buttonIndex != 0 {
		t.Errorf("Expected button index 0, got %d", buttonIndex)
	}

	// Click on the left edge of the button
	buttonIndex = CheckButtonClick(dialog, 110, 140)
	if buttonIndex != 0 {
		t.Errorf("Expected button index 0 for left edge, got %d", buttonIndex)
	}

	// Click on the right edge of the button (just inside)
	buttonIndex = CheckButtonClick(dialog, 289, 140)
	if buttonIndex != 0 {
		t.Errorf("Expected button index 0 for right edge, got %d", buttonIndex)
	}

	// Click on the top edge of the button
	buttonIndex = CheckButtonClick(dialog, 200, 120)
	if buttonIndex != 0 {
		t.Errorf("Expected button index 0 for top edge, got %d", buttonIndex)
	}

	// Click on the bottom edge of the button (just inside)
	buttonIndex = CheckButtonClick(dialog, 200, 159)
	if buttonIndex != 0 {
		t.Errorf("Expected button index 0 for bottom edge, got %d", buttonIndex)
	}
}

func TestCheckButtonClick_MissButton(t *testing.T) {
	dialog := &Dialog{
		Title:  "Test Dialog",
		X:      100,
		Y:      100,
		Width:  200,
		Height: 200,
		Buttons: []Button{
			{
				Label:  "Button 1",
				X:      10, // Relative to dialog
				Y:      20, // Relative to dialog
				Width:  180,
				Height: 40,
			},
		},
	}
	// Button absolute position: (110, 120) to (290, 160)

	// Click above the button
	buttonIndex := CheckButtonClick(dialog, 200, 100)
	if buttonIndex != -1 {
		t.Errorf("Expected button index -1 for miss, got %d", buttonIndex)
	}

	// Click below the button
	buttonIndex = CheckButtonClick(dialog, 200, 180)
	if buttonIndex != -1 {
		t.Errorf("Expected button index -1 for miss, got %d", buttonIndex)
	}

	// Click to the left of the button
	buttonIndex = CheckButtonClick(dialog, 90, 140)
	if buttonIndex != -1 {
		t.Errorf("Expected button index -1 for miss, got %d", buttonIndex)
	}

	// Click to the right of the button
	buttonIndex = CheckButtonClick(dialog, 300, 140)
	if buttonIndex != -1 {
		t.Errorf("Expected button index -1 for miss, got %d", buttonIndex)
	}

	// Click just outside the right edge (beyond X + Width)
	buttonIndex = CheckButtonClick(dialog, 291, 140)
	if buttonIndex != -1 {
		t.Errorf("Expected button index -1 for just outside right edge, got %d", buttonIndex)
	}

	// Click just outside the bottom edge (beyond Y + Height)
	buttonIndex = CheckButtonClick(dialog, 200, 161)
	if buttonIndex != -1 {
		t.Errorf("Expected button index -1 for just outside bottom edge, got %d", buttonIndex)
	}
}

func TestCheckButtonClick_MultipleButtons(t *testing.T) {
	dialog := &Dialog{
		Title:  "Test Dialog",
		X:      100,
		Y:      100,
		Width:  200,
		Height: 300,
		Buttons: []Button{
			{
				Label:  "Button 1",
				X:      10, // Relative to dialog
				Y:      20, // Relative to dialog
				Width:  180,
				Height: 40,
			},
			{
				Label:  "Button 2",
				X:      10, // Relative to dialog
				Y:      70, // Relative to dialog
				Width:  180,
				Height: 40,
			},
			{
				Label:  "Button 3",
				X:      10,  // Relative to dialog
				Y:      120, // Relative to dialog
				Width:  180,
				Height: 40,
			},
		},
	}
	// Button 1 absolute: (110, 120) to (290, 160)
	// Button 2 absolute: (110, 170) to (290, 210)
	// Button 3 absolute: (110, 220) to (290, 260)

	// Click on first button
	buttonIndex := CheckButtonClick(dialog, 200, 140)
	if buttonIndex != 0 {
		t.Errorf("Expected button index 0, got %d", buttonIndex)
	}

	// Click on second button
	buttonIndex = CheckButtonClick(dialog, 200, 190)
	if buttonIndex != 1 {
		t.Errorf("Expected button index 1, got %d", buttonIndex)
	}

	// Click on third button
	buttonIndex = CheckButtonClick(dialog, 200, 240)
	if buttonIndex != 2 {
		t.Errorf("Expected button index 2, got %d", buttonIndex)
	}

	// Click between buttons (in the gap)
	buttonIndex = CheckButtonClick(dialog, 200, 165)
	if buttonIndex != -1 {
		t.Errorf("Expected button index -1 for gap between buttons, got %d", buttonIndex)
	}
}

func TestCreateTitleScreen_ButtonCount(t *testing.T) {
	titleDialog := CreateTitleScreen()

	expectedButtonCount := 4
	if len(titleDialog.Buttons) != expectedButtonCount {
		t.Errorf("Expected %d buttons, got %d", expectedButtonCount, len(titleDialog.Buttons))
	}
}

func TestCreateTitleScreen_ButtonLabels(t *testing.T) {
	titleDialog := CreateTitleScreen()

	expectedLabels := []string{"Play Game", "Settings", "Instructions", "High Scores"}

	if len(titleDialog.Buttons) != len(expectedLabels) {
		t.Fatalf("Expected %d buttons, got %d", len(expectedLabels), len(titleDialog.Buttons))
	}

	for i, expectedLabel := range expectedLabels {
		if titleDialog.Buttons[i].Label != expectedLabel {
			t.Errorf("Button %d: expected label '%s', got '%s'", i, expectedLabel, titleDialog.Buttons[i].Label)
		}
	}
}

func TestCreateTitleScreen_DialogProperties(t *testing.T) {
	titleDialog := CreateTitleScreen()

	if titleDialog.Title != "Verdant Thane" {
		t.Errorf("Expected title 'Verdant Thane', got '%s'", titleDialog.Title)
	}

	if titleDialog.Width != dialogWidth {
		t.Errorf("Expected width %f, got %f", dialogWidth, titleDialog.Width)
	}

	// Calculate expected dialog height based on button layout
	numButtons := len(titleDialog.Buttons)
	expectedHeight := float64(numButtons)*buttonHeight + float64(numButtons-1)*buttonSpacing + 2*internalMargin
	if titleDialog.Height != expectedHeight {
		t.Errorf("Expected height %f, got %f", expectedHeight, titleDialog.Height)
	}

	// Dialog should be centered
	expectedX := (float64(screenWidth) - dialogWidth) / 2
	expectedY := (float64(screenHeight) - titleDialog.Height) / 2

	if titleDialog.X != expectedX {
		t.Errorf("Expected X %f (centered), got %f", expectedX, titleDialog.X)
	}

	if titleDialog.Y != expectedY {
		t.Errorf("Expected Y %f (centered), got %f", expectedY, titleDialog.Y)
	}
}

func TestCreateTitleScreen_ButtonLayout(t *testing.T) {
	titleDialog := CreateTitleScreen()

	// All buttons should be centered within the dialog (using relative coordinates)
	dialogCenterX := titleDialog.Width / 2 // Relative center within dialog

	for i, button := range titleDialog.Buttons {
		buttonCenterX := button.X + button.Width/2 // Button X is relative to dialog

		// Check that button is centered horizontally within dialog
		if buttonCenterX != dialogCenterX {
			t.Errorf("Button %d not horizontally centered: expected center X %f, got %f", i, dialogCenterX, buttonCenterX)
		}

		// Check that button has expected dimensions
		if button.Width != buttonWidth {
			t.Errorf("Button %d: expected width %f, got %f", i, buttonWidth, button.Width)
		}

		if button.Height != buttonHeight {
			t.Errorf("Button %d: expected height %f, got %f", i, buttonHeight, button.Height)
		}

		// Check vertical spacing (except for first button which has its own offset)
		if i > 0 {
			prevButton := titleDialog.Buttons[i-1]
			expectedY := prevButton.Y + prevButton.Height + buttonSpacing
			if button.Y != expectedY {
				t.Errorf("Button %d: expected Y %f (with spacing), got %f", i, expectedY, button.Y)
			}
		}
	}
}

func TestGetTitleY(t *testing.T) {
	y := GetTitleY()
	if y != titleY {
		t.Errorf("Expected title Y %f, got %f", titleY, y)
	}
}
