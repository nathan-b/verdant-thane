package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/nathan/verdant-thane/config"
)

// GameOverScreen manages the game over screen state
type GameOverScreen struct {
	playerName   string
	score        int
	kills        int
	deaths       int
	isHighScore  bool
	nameSelected bool // Whether the name field is selected for input
	dialog       *Dialog
}

// NewGameOverScreen creates a new game over screen
func NewGameOverScreen(playerName string, score, kills, deaths int, isHighScore bool) *GameOverScreen {
	// Create dialog with Continue button
	buttons := []Button{
		{
			Label:   "Continue",
			X:       float64(config.ScreenWidth)/2 - 50,
			Y:       float64(config.ScreenHeight)/2 + 150,
			Width:   100,
			Height:  40,
			OnClick: nil, // Will be handled by main game loop
		},
	}

	dialog := &Dialog{
		Title:   "GAME OVER",
		X:       float64(config.ScreenWidth)/2 - 250,
		Y:       float64(config.ScreenHeight)/2 - 200,
		Width:   500,
		Height:  400,
		Buttons: buttons,
	}

	return &GameOverScreen{
		playerName:   playerName,
		score:        score,
		kills:        kills,
		deaths:       deaths,
		isHighScore:  isHighScore,
		nameSelected: isHighScore, // Auto-select if it's a high score
		dialog:       dialog,
	}
}

// GetPlayerName returns the current player name
func (gos *GameOverScreen) GetPlayerName() string {
	return gos.playerName
}

// HandleInput processes keyboard input for name entry
func (gos *GameOverScreen) HandleInput() {
	if !gos.nameSelected {
		return
	}

	// Handle text input
	runes := ebiten.AppendInputChars(nil)
	for _, r := range runes {
		// Only allow printable characters
		if r >= 32 && r <= 126 {
			// Limit name length to 20 characters
			if len(gos.playerName) < 20 {
				gos.playerName += string(r)
			}
		}
	}

	// Handle backspace
	if ebiten.IsKeyPressed(ebiten.KeyBackspace) {
		if len(gos.playerName) > 0 {
			gos.playerName = gos.playerName[:len(gos.playerName)-1]
		}
	}
}

// Draw renders the game over screen
func (gos *GameOverScreen) Draw(screen *ebiten.Image, font *text.GoTextFaceSource) {
	hudFont := &text.GoTextFace{
		Source: font,
		Size:   16,
	}

	textColor := color.RGBA{255, 255, 255, 255}

	// Draw dialog background
	RenderDialog(screen, gos.dialog, hudFont)

	// Calculate center X for text alignment
	centerX := float64(config.ScreenWidth) / 2

	// Draw stats
	statsY := float64(config.ScreenHeight)/2 - 120

	// Score
	scoreText := fmt.Sprintf("Final Score: %d", gos.score)
	scoreWidth, _ := text.Measure(scoreText, hudFont, 0)
	scoreOp := &text.DrawOptions{}
	scoreOp.GeoM.Translate(centerX-scoreWidth/2, statsY)
	scoreOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, scoreText, hudFont, scoreOp)

	// Kills
	killsText := fmt.Sprintf("Kills: %d", gos.kills)
	killsWidth, _ := text.Measure(killsText, hudFont, 0)
	killsOp := &text.DrawOptions{}
	killsOp.GeoM.Translate(centerX-killsWidth/2, statsY+30)
	killsOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, killsText, hudFont, killsOp)

	// Deaths
	deathsText := fmt.Sprintf("Deaths: %d", gos.deaths)
	deathsWidth, _ := text.Measure(deathsText, hudFont, 0)
	deathsOp := &text.DrawOptions{}
	deathsOp.GeoM.Translate(centerX-deathsWidth/2, statsY+60)
	deathsOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, deathsText, hudFont, deathsOp)

	// High score message (if applicable)
	if gos.isHighScore {
		highScoreText := "NEW HIGH SCORE!"
		highScoreWidth, _ := text.Measure(highScoreText, hudFont, 0)
		highScoreOp := &text.DrawOptions{}
		highScoreOp.GeoM.Translate(centerX-highScoreWidth/2, statsY+100)
		highScoreOp.ColorScale.ScaleWithColor(color.RGBA{255, 215, 0, 255}) // Gold color
		text.Draw(screen, highScoreText, hudFont, highScoreOp)
	}

	// Name entry field
	nameY := statsY + 140
	namePrompt := "Enter your name:"
	namePromptWidth, _ := text.Measure(namePrompt, hudFont, 0)
	namePromptOp := &text.DrawOptions{}
	namePromptOp.GeoM.Translate(centerX-namePromptWidth/2, nameY)
	namePromptOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, namePrompt, hudFont, namePromptOp)

	// Name input box
	nameBoxY := nameY + 30
	nameBoxWidth := 300.0
	nameBoxHeight := 40.0
	nameBoxX := centerX - nameBoxWidth/2

	// Draw input box background
	vector.FillRect(screen, float32(nameBoxX), float32(nameBoxY), float32(nameBoxWidth), float32(nameBoxHeight), color.RGBA{40, 40, 40, 255}, false)
	// Draw input box border
	vector.StrokeRect(screen, float32(nameBoxX), float32(nameBoxY), float32(nameBoxWidth), float32(nameBoxHeight), 2, color.RGBA{100, 100, 100, 255}, false)

	// Draw name text
	nameText := gos.playerName
	if gos.nameSelected {
		// Add cursor blink (using tick count)
		if (int(ebiten.ActualTPS()*10) / 30 % 2) == 0 {
			nameText += "_"
		}
	}
	nameTextOp := &text.DrawOptions{}
	nameTextOp.GeoM.Translate(nameBoxX+10, nameBoxY+12)
	nameTextOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, nameText, hudFont, nameTextOp)
}

// HandleClick processes mouse clicks on the dialog
func (gos *GameOverScreen) HandleClick(mouseX, mouseY int) int {
	return CheckButtonClick(gos.dialog, mouseX, mouseY)
}
