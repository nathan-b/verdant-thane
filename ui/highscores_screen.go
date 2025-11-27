package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/nathan-b/verdant-thane/config"
	"github.com/nathan-b/verdant-thane/persistence"
)

// CreateHighScoresDialog creates the high scores screen dialog
func CreateHighScoresDialog() *Dialog {
	return &Dialog{
		Title:   "High Scores",
		X:       float64(config.ScreenWidth)/2 - 350,
		Y:       80,
		Width:   700,
		Height:  500,
		Buttons: []Button{}, // No buttons, uses X close button
	}
}

// DrawHighScoresTable draws the high scores table
func DrawHighScoresTable(screen *ebiten.Image, font *text.GoTextFaceSource, highScores *persistence.HighScores) {
	hudFont := &text.GoTextFace{
		Source: font,
		Size:   16,
	}

	textColor := color.RGBA{255, 255, 255, 255}
	goldColor := color.RGBA{255, 215, 0, 255}     // Gold for rank 1
	silverColor := color.RGBA{192, 192, 192, 255} // Silver for rank 2
	bronzeColor := color.RGBA{205, 127, 50, 255}  // Bronze for rank 3

	startY := 160.0
	lineHeight := 35.0
	leftX := float64(config.ScreenWidth)/2 - 320

	// Header
	headerY := startY - 30
	rankHeader := "Rank"
	nameHeader := "Name"
	scoreHeader := "Score"
	killsHeader := "Kills"
	dateHeader := "Date"

	rankOp := &text.DrawOptions{}
	rankOp.GeoM.Translate(leftX, headerY)
	rankOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, rankHeader, hudFont, rankOp)

	nameOp := &text.DrawOptions{}
	nameOp.GeoM.Translate(leftX+80, headerY)
	nameOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, nameHeader, hudFont, nameOp)

	scoreOp := &text.DrawOptions{}
	scoreOp.GeoM.Translate(leftX+280, headerY)
	scoreOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, scoreHeader, hudFont, scoreOp)

	killsOp := &text.DrawOptions{}
	killsOp.GeoM.Translate(leftX+400, headerY)
	killsOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, killsHeader, hudFont, killsOp)

	dateOp := &text.DrawOptions{}
	dateOp.GeoM.Translate(leftX+520, headerY)
	dateOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, dateHeader, hudFont, dateOp)

	// Entries
	currentY := startY
	for i, entry := range highScores.Entries {
		if i >= 10 {
			break
		}

		// Determine color based on rank
		entryColor := textColor
		if i == 0 {
			entryColor = goldColor
		} else if i == 1 {
			entryColor = silverColor
		} else if i == 2 {
			entryColor = bronzeColor
		}

		// Rank
		rankText := fmt.Sprintf("%d", i+1)
		rankEntryOp := &text.DrawOptions{}
		rankEntryOp.GeoM.Translate(leftX+10, currentY)
		rankEntryOp.ColorScale.ScaleWithColor(entryColor)
		text.Draw(screen, rankText, hudFont, rankEntryOp)

		// Name (truncate if too long)
		nameText := entry.Name
		if len(nameText) > 15 {
			nameText = nameText[:15] + "..."
		}
		nameEntryOp := &text.DrawOptions{}
		nameEntryOp.GeoM.Translate(leftX+80, currentY)
		nameEntryOp.ColorScale.ScaleWithColor(entryColor)
		text.Draw(screen, nameText, hudFont, nameEntryOp)

		// Score
		scoreText := fmt.Sprintf("%d", entry.Score)
		scoreEntryOp := &text.DrawOptions{}
		scoreEntryOp.GeoM.Translate(leftX+280, currentY)
		scoreEntryOp.ColorScale.ScaleWithColor(entryColor)
		text.Draw(screen, scoreText, hudFont, scoreEntryOp)

		// Kills
		killsText := fmt.Sprintf("%d", entry.Kills)
		killsEntryOp := &text.DrawOptions{}
		killsEntryOp.GeoM.Translate(leftX+420, currentY)
		killsEntryOp.ColorScale.ScaleWithColor(entryColor)
		text.Draw(screen, killsText, hudFont, killsEntryOp)

		// Date (format as MM/DD/YYYY)
		dateText := entry.Date.Format("01/02/2006")
		dateEntryOp := &text.DrawOptions{}
		dateEntryOp.GeoM.Translate(leftX+520, currentY)
		dateEntryOp.ColorScale.ScaleWithColor(entryColor)
		text.Draw(screen, dateText, hudFont, dateEntryOp)

		currentY += lineHeight
	}

	// If no scores, display message
	if len(highScores.Entries) == 0 {
		noScoresText := "No high scores yet. Play a game to set one!"
		noScoresWidth, _ := text.Measure(noScoresText, hudFont, 0)
		noScoresOp := &text.DrawOptions{}
		noScoresOp.GeoM.Translate(float64(config.ScreenWidth)/2-noScoresWidth/2, startY+100)
		noScoresOp.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, noScoresText, hudFont, noScoresOp)
	}
}
