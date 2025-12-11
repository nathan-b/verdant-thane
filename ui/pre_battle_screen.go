package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/nathan-b/verdant-thane/config"
	"github.com/nathan-b/verdant-thane/systems"
)

// ShipButton represents a ship selection button with an image
type ShipButton struct {
	Class     config.ShipClass
	X, Y      float64
	Width     float64
	Height    float64
	Available bool
}

// PreBattleScreen manages the pre-battle screen state
type PreBattleScreen struct {
	battleNumber   int
	fleetConfig    config.FleetConfig
	playerFaction  config.FactionComposition // Player's faction (faction 0) composition
	score          int
	totalKills     int
	totalDeaths    int
	battleKills    int
	battleDeaths   int
	selectedClass  config.ShipClass // Selected starting ship class
	hudFont        *text.GoTextFace
	factionSprites *systems.FactionSprites
	onStart        func(selectedClass config.ShipClass)

	// UI elements
	shipButtons  []ShipButton
	startButton  Button
	prevMouseBtn bool // Track previous mouse button state for click detection
}

// NewPreBattleScreen creates a new pre-battle screen with manual buttons
// The onStart callback will be called when the Start Battle button is clicked, passing the selected ship class
func NewPreBattleScreen(
	battleNumber int,
	fleetConfig config.FleetConfig,
	score, totalKills, totalDeaths, battleKills, battleDeaths int,
	fontSource *text.GoTextFaceSource,
	factionSprites *systems.FactionSprites,
	onStart func(selectedClass config.ShipClass),
) *PreBattleScreen {
	// Create font faces
	hudFont := &text.GoTextFace{
		Source: fontSource,
		Size:   16,
	}

	// Get player faction composition (faction 0)
	playerFaction := fleetConfig.Compositions[0]

	// Default selection: Fighter if available, otherwise Destroyer
	defaultClass := config.ClassFighter
	if playerFaction.Fighters == 0 && playerFaction.Destroyers > 0 {
		defaultClass = config.ClassDestroyer
	}

	pbs := &PreBattleScreen{
		battleNumber:   battleNumber,
		fleetConfig:    fleetConfig,
		playerFaction:  playerFaction,
		score:          score,
		totalKills:     totalKills,
		totalDeaths:    totalDeaths,
		battleKills:    battleKills,
		battleDeaths:   battleDeaths,
		selectedClass:  defaultClass,
		hudFont:        hudFont,
		factionSprites: factionSprites,
		onStart:        onStart,
	}

	// Create ship selection buttons with fixed positions
	pbs.shipButtons = make([]ShipButton, 0, 2)
	centerX := float64(config.ScreenWidth) / 2
	buttonY := 470.0 // Y position for ship buttons (moved down to avoid overlap with stats)
	buttonWidth := 140.0
	buttonHeight := 120.0
	buttonSpacing := 30.0

	// Count available ship types to calculate positions
	availableCount := 0
	if playerFaction.Fighters > 0 {
		availableCount++
	}
	if playerFaction.Destroyers > 0 {
		availableCount++
	}

	// Calculate starting X to center the buttons
	totalWidth := float64(availableCount)*buttonWidth + float64(availableCount-1)*buttonSpacing
	startX := centerX - totalWidth/2

	// Create Fighter button if available
	buttonIndex := 0
	if playerFaction.Fighters > 0 {
		x := startX + float64(buttonIndex)*(buttonWidth+buttonSpacing)
		pbs.shipButtons = append(pbs.shipButtons, ShipButton{
			Class:     config.ClassFighter,
			X:         x,
			Y:         buttonY,
			Width:     buttonWidth,
			Height:    buttonHeight,
			Available: true,
		})
		buttonIndex++
	}

	// Create Destroyer button if available
	if playerFaction.Destroyers > 0 {
		x := startX + float64(buttonIndex)*(buttonWidth+buttonSpacing)
		pbs.shipButtons = append(pbs.shipButtons, ShipButton{
			Class:     config.ClassDestroyer,
			X:         x,
			Y:         buttonY,
			Width:     buttonWidth,
			Height:    buttonHeight,
			Available: true,
		})
	}

	// Create Start Battle button
	pbs.startButton = Button{
		Label:  "START BATTLE",
		X:      centerX - 100, // Center button (200px wide)
		Y:      buttonY + buttonHeight + 30,
		Width:  200,
		Height: 50,
	}

	return pbs
}

// Update processes UI updates (call this from main game loop)
func (pbs *PreBattleScreen) Update() {
	// Handle mouse clicks
	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	// Detect click (button just pressed)
	if mousePressed && !pbs.prevMouseBtn {
		mouseX, mouseY := ebiten.CursorPosition()

		// Check ship button clicks
		for i := range pbs.shipButtons {
			btn := &pbs.shipButtons[i]
			if float64(mouseX) >= btn.X && float64(mouseX) <= btn.X+btn.Width &&
				float64(mouseY) >= btn.Y && float64(mouseY) <= btn.Y+btn.Height {
				pbs.selectedClass = btn.Class
				break
			}
		}

		// Check start button click
		if float64(mouseX) >= pbs.startButton.X && float64(mouseX) <= pbs.startButton.X+pbs.startButton.Width &&
			float64(mouseY) >= pbs.startButton.Y && float64(mouseY) <= pbs.startButton.Y+pbs.startButton.Height {
			if pbs.onStart != nil {
				pbs.onStart(pbs.selectedClass)
			}
		}
	}

	pbs.prevMouseBtn = mousePressed
}

// Draw renders the pre-battle screen
func (pbs *PreBattleScreen) Draw(screen *ebiten.Image) {
	centerX := float64(config.ScreenWidth) / 2
	titleFont := &text.GoTextFace{
		Source: pbs.hudFont.Source,
		Size:   28,
	}
	sectionFont := &text.GoTextFace{
		Source: pbs.hudFont.Source,
		Size:   20,
	}

	textColor := color.RGBA{255, 255, 255, 255}
	goldColor := color.RGBA{255, 215, 0, 255}

	// Draw title
	titleText := fmt.Sprintf("BATTLE %d", pbs.battleNumber)
	titleWidth, _ := text.Measure(titleText, titleFont, 0)
	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(centerX-titleWidth/2, 80)
	titleOp.ColorScale.ScaleWithColor(goldColor)
	text.Draw(screen, titleText, titleFont, titleOp)

	// Draw battle info section
	y := 140.0

	// Player fleet breakdown section
	fleetBreakdownText := "YOUR FLEET"
	fleetBreakdownWidth, _ := text.Measure(fleetBreakdownText, sectionFont, 0)
	fleetBreakdownOp := &text.DrawOptions{}
	fleetBreakdownOp.GeoM.Translate(centerX-fleetBreakdownWidth/2, y)
	fleetBreakdownOp.ColorScale.ScaleWithColor(goldColor)
	text.Draw(screen, fleetBreakdownText, sectionFont, fleetBreakdownOp)
	y += 35

	// Fleet composition
	fleetCompositionText := fmt.Sprintf("Fighters: %d  |  Destroyers: %d  |  Testudons: %d",
		pbs.playerFaction.Fighters, pbs.playerFaction.Destroyers, pbs.playerFaction.Testudons)
	fleetCompositionWidth, _ := text.Measure(fleetCompositionText, pbs.hudFont, 0)
	fleetCompositionOp := &text.DrawOptions{}
	fleetCompositionOp.GeoM.Translate(centerX-fleetCompositionWidth/2, y)
	fleetCompositionOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, fleetCompositionText, pbs.hudFont, fleetCompositionOp)
	y += 40

	// Stats section
	statsText := "STATISTICS"
	statsWidth, _ := text.Measure(statsText, sectionFont, 0)
	statsOp := &text.DrawOptions{}
	statsOp.GeoM.Translate(centerX-statsWidth/2, y)
	statsOp.ColorScale.ScaleWithColor(goldColor)
	text.Draw(screen, statsText, sectionFont, statsOp)
	y += 35

	// Score
	scoreText := fmt.Sprintf("Score: %d", pbs.score)
	scoreWidth, _ := text.Measure(scoreText, pbs.hudFont, 0)
	scoreOp := &text.DrawOptions{}
	scoreOp.GeoM.Translate(centerX-scoreWidth/2, y)
	scoreOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, scoreText, pbs.hudFont, scoreOp)
	y += 25

	// Kills (show battle kills if this is not the first battle)
	var killsText string
	if pbs.battleNumber > 1 {
		killsText = fmt.Sprintf("Kills: %d (Last Battle: %d)", pbs.totalKills, pbs.battleKills)
	} else {
		killsText = fmt.Sprintf("Kills: %d", pbs.totalKills)
	}
	killsWidth, _ := text.Measure(killsText, pbs.hudFont, 0)
	killsOp := &text.DrawOptions{}
	killsOp.GeoM.Translate(centerX-killsWidth/2, y)
	killsOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, killsText, pbs.hudFont, killsOp)
	y += 25

	// Deaths (show battle deaths if this is not the first battle)
	var deathsText string
	if pbs.battleNumber > 1 {
		deathsText = fmt.Sprintf("Deaths: %d (Last Battle: %d)", pbs.totalDeaths, pbs.battleDeaths)
	} else {
		deathsText = fmt.Sprintf("Deaths: %d", pbs.totalDeaths)
	}
	deathsWidth, _ := text.Measure(deathsText, pbs.hudFont, 0)
	deathsOp := &text.DrawOptions{}
	deathsOp.GeoM.Translate(centerX-deathsWidth/2, y)
	deathsOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, deathsText, pbs.hudFont, deathsOp)

	// Draw label above ship buttons
	labelY := pbs.shipButtons[0].Y - 30
	labelText := "Your ship:"
	labelWidth, _ := text.Measure(labelText, pbs.hudFont, 0)
	labelOp := &text.DrawOptions{}
	labelOp.GeoM.Translate(centerX-labelWidth/2, labelY)
	labelOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, labelText, pbs.hudFont, labelOp)

	// Draw ship selection buttons
	for i := range pbs.shipButtons {
		pbs.drawShipButton(screen, &pbs.shipButtons[i])
	}

	// Draw start button
	pbs.drawStartButton(screen)
}

// drawShipButton draws a single ship selection button with sprite and label
func (pbs *PreBattleScreen) drawShipButton(screen *ebiten.Image, btn *ShipButton) {
	// Draw button background
	vector.FillRect(screen,
		float32(btn.X), float32(btn.Y),
		float32(btn.Width), float32(btn.Height),
		color.RGBA{40, 40, 50, 255}, false)

	// Draw selection border if this button is selected
	if pbs.selectedClass == btn.Class {
		borderWidth := float32(3.0)
		goldColor := color.RGBA{255, 215, 0, 255}
		vector.StrokeRect(screen,
			float32(btn.X), float32(btn.Y),
			float32(btn.Width), float32(btn.Height),
			borderWidth, goldColor, false)
	} else {
		// Draw normal border
		vector.StrokeRect(screen,
			float32(btn.X), float32(btn.Y),
			float32(btn.Width), float32(btn.Height),
			2, color.RGBA{100, 100, 100, 255}, false)
	}

	// Draw ship sprite
	var sprite *ebiten.Image
	switch btn.Class {
	case config.ClassFighter:
		sprite = pbs.factionSprites.Fighter.Green
	case config.ClassDestroyer:
		sprite = pbs.factionSprites.Destroyer.Green
	}

	if sprite != nil {
		bounds := sprite.Bounds()
		spriteWidth := float64(bounds.Dx())

		// Center sprite horizontally, position near top vertically
		spriteX := btn.X + (btn.Width-spriteWidth)/2
		spriteY := btn.Y + 10 // 10px margin from top

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(spriteX, spriteY)
		screen.DrawImage(sprite, op)
	}

	// Draw label text below sprite
	count := 0
	switch btn.Class {
	case config.ClassFighter:
		count = pbs.playerFaction.Fighters
	case config.ClassDestroyer:
		count = pbs.playerFaction.Destroyers
	}

	// Draw text centered at bottom of button
	lines := []string{"Fighter", fmt.Sprintf("(%d available)", count)}
	if btn.Class == config.ClassDestroyer {
		lines[0] = "Destroyer"
	}

	textY := btn.Y + 70 // Start text 70px from top (below sprite)
	for _, line := range lines {
		textWidth, _ := text.Measure(line, pbs.hudFont, 0)
		textX := btn.X + (btn.Width-textWidth)/2

		textOp := &text.DrawOptions{}
		textOp.GeoM.Translate(textX, textY)
		textOp.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, line, pbs.hudFont, textOp)
		textY += 20
	}
}

// drawStartButton draws the start battle button
func (pbs *PreBattleScreen) drawStartButton(screen *ebiten.Image) {
	// Button background (green)
	vector.FillRect(screen,
		float32(pbs.startButton.X), float32(pbs.startButton.Y),
		float32(pbs.startButton.Width), float32(pbs.startButton.Height),
		color.RGBA{70, 180, 70, 255}, false)

	// Button border
	vector.StrokeRect(screen,
		float32(pbs.startButton.X), float32(pbs.startButton.Y),
		float32(pbs.startButton.Width), float32(pbs.startButton.Height),
		2, color.RGBA{50, 130, 50, 255}, false)

	// Button text (centered)
	textWidth, _ := text.Measure(pbs.startButton.Label, pbs.hudFont, 0)
	textX := pbs.startButton.X + (pbs.startButton.Width-textWidth)/2
	textY := pbs.startButton.Y + pbs.startButton.Height/2 + 6 // Vertically centered

	textOp := &text.DrawOptions{}
	textOp.GeoM.Translate(textX, textY)
	textOp.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, pbs.startButton.Label, pbs.hudFont, textOp)
}
