package ui

import (
	"fmt"
	"image/color"

	"github.com/ebitenui/ebitenui"
	ebitenui_image "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/nathan/verdant-thane/config"
)

// GameOverScreen manages the game over screen state
type GameOverScreen struct {
	score          int
	kills          int
	deaths         int
	isHighScore    bool
	ui             *ebitenui.UI
	nameInput      *widget.TextInput
	continueButton *widget.Button
	restartButton  *widget.Button
	newGameButton  *widget.Button
	font           text.Face // Store as interface for ebitenui
	hudFont        *text.GoTextFace
}

// NewGameOverScreen creates a new game over screen with ebitenui widgets
// The onContinue callback will be called when the Continue button is clicked
// The onRestart callback will be called when the Restart current battle button is clicked
// The onNewGame callback will be called when the New game button is clicked
func NewGameOverScreen(playerName string, score, kills, deaths int, isHighScore bool, fontSource *text.GoTextFaceSource, onContinue func(), onRestart func(), onNewGame func()) *GameOverScreen {
	// Create font faces
	hudFont := &text.GoTextFace{
		Source: fontSource,
		Size:   16,
	}

	gos := &GameOverScreen{
		score:       score,
		kills:       kills,
		deaths:      deaths,
		isHighScore: isHighScore,
		font:        hudFont, // Store as text.Face interface
		hudFont:     hudFont,
	}

	// Create UI container positioned at center of screen (transparent - no background)
	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	// Content container for vertical layout with background
	// Positioned below stats text (stats end around Y=270-290, so start at Y=310)
	contentContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ebitenui_image.NewNineSliceColor(color.NRGBA{20, 20, 30, 240})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(15),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top:    30,
				Bottom: 30,
				Left:   40,
				Right:  40,
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				StretchVertical:    false,
				Padding: &widget.Insets{
					Top: 310, // Position below stats
				},
			}),
			widget.WidgetOpts.MinSize(400, 0), // Auto height
		),
	)

	// Add "Enter your name:" label
	nameLabel := widget.NewText(
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
		widget.TextOpts.Text("Enter your name:", &gos.font, color.NRGBA{255, 255, 255, 255}),
	)

	// Create text input for player name
	gos.nameInput = widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
			widget.WidgetOpts.MinSize(300, 40),
		),
		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle:     ebitenui_image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
			Disabled: ebitenui_image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
		}),
		widget.TextInputOpts.Face(&gos.font),
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:          color.NRGBA{254, 255, 255, 255},
			Disabled:      color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			Caret:         color.NRGBA{254, 255, 255, 255},
			DisabledCaret: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		}),
		widget.TextInputOpts.Padding(widget.NewInsetsSimple(10)),
		widget.TextInputOpts.Placeholder("Enter your name"),
		widget.TextInputOpts.Validation(func(newInputText string) (bool, *string) {
			// Limit to 20 characters
			if len(newInputText) > 20 {
				return false, nil
			}
			return true, nil
		}),
	)

	// Set initial player name
	gos.nameInput.SetText(playerName)

	// Create Continue button
	gos.continueButton = widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  false,
			}),
			widget.WidgetOpts.MinSize(120, 40),
		),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    ebitenui_image.NewNineSliceColor(color.NRGBA{R: 70, G: 130, B: 180, A: 255}),
			Hover:   ebitenui_image.NewNineSliceColor(color.NRGBA{R: 90, G: 150, B: 200, A: 255}),
			Pressed: ebitenui_image.NewNineSliceColor(color.NRGBA{R: 50, G: 110, B: 160, A: 255}),
		}),
		widget.ButtonOpts.Text("Continue", &gos.font, &widget.ButtonTextColor{
			Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   30,
			Right:  30,
			Top:    10,
			Bottom: 10,
		}),
		// Click handler callback
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if onContinue != nil {
				onContinue()
			}
		}),
	)

	// Create Restart current battle button
	gos.restartButton = widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  false,
			}),
			widget.WidgetOpts.MinSize(200, 40),
		),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    ebitenui_image.NewNineSliceColor(color.NRGBA{R: 70, G: 180, B: 130, A: 255}), // Green color
			Hover:   ebitenui_image.NewNineSliceColor(color.NRGBA{R: 90, G: 200, B: 150, A: 255}),
			Pressed: ebitenui_image.NewNineSliceColor(color.NRGBA{R: 50, G: 160, B: 110, A: 255}),
		}),
		widget.ButtonOpts.Text("Retry current battle", &gos.font, &widget.ButtonTextColor{
			Idle: color.NRGBA{0xdf, 0xff, 0xf4, 0xff},
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   20,
			Right:  20,
			Top:    10,
			Bottom: 10,
		}),
		// Click handler callback
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if onRestart != nil {
				onRestart()
			}
		}),
	)

	// Create New game button
	gos.newGameButton = widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  false,
			}),
			widget.WidgetOpts.MinSize(120, 40),
		),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    ebitenui_image.NewNineSliceColor(color.NRGBA{R: 180, G: 130, B: 70, A: 255}), // Orange color
			Hover:   ebitenui_image.NewNineSliceColor(color.NRGBA{R: 200, G: 150, B: 90, A: 255}),
			Pressed: ebitenui_image.NewNineSliceColor(color.NRGBA{R: 160, G: 110, B: 50, A: 255}),
		}),
		widget.ButtonOpts.Text("New game", &gos.font, &widget.ButtonTextColor{
			Idle: color.NRGBA{0xff, 0xf4, 0xdf, 0xff},
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   30,
			Right:  30,
			Top:    10,
			Bottom: 10,
		}),
		// Click handler callback
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if onNewGame != nil {
				onNewGame()
			}
		}),
	)

	// Add widgets to content container
	contentContainer.AddChild(nameLabel)
	contentContainer.AddChild(gos.nameInput)
	contentContainer.AddChild(gos.continueButton)
	contentContainer.AddChild(gos.restartButton)
	contentContainer.AddChild(gos.newGameButton)

	// Add content to main container
	container.AddChild(contentContainer)

	// Create UI
	gos.ui = &ebitenui.UI{
		Container: container,
	}

	return gos
}

// GetPlayerName returns the current player name from the text input widget
func (gos *GameOverScreen) GetPlayerName() string {
	return gos.nameInput.GetText()
}

// Update processes UI updates (call this from main game loop)
func (gos *GameOverScreen) Update() {
	gos.ui.Update()
}

// Draw renders the game over screen with ebitenui and custom text
func (gos *GameOverScreen) Draw(screen *ebiten.Image) {
	titleFont := &text.GoTextFace{
		Source: gos.hudFont.Source,
		Size:   32,
	}

	textColor := color.RGBA{255, 255, 255, 255}
	centerX := float64(config.ScreenWidth) / 2

	// Draw "GAME OVER" title
	titleText := "GAME OVER"
	titleWidth, _ := text.Measure(titleText, titleFont, 0)
	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(centerX-titleWidth/2, 100)
	titleOp.ColorScale.ScaleWithColor(color.RGBA{255, 100, 100, 255}) // Red color
	text.Draw(screen, titleText, titleFont, titleOp)

	// Draw stats
	statsY := 170.0

	// Score
	scoreText := fmt.Sprintf("Final Score: %d", gos.score)
	scoreWidth, _ := text.Measure(scoreText, gos.hudFont, 0)
	scoreOp := &text.DrawOptions{}
	scoreOp.GeoM.Translate(centerX-scoreWidth/2, statsY)
	scoreOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, scoreText, gos.hudFont, scoreOp)

	// Kills
	killsText := fmt.Sprintf("Kills: %d", gos.kills)
	killsWidth, _ := text.Measure(killsText, gos.hudFont, 0)
	killsOp := &text.DrawOptions{}
	killsOp.GeoM.Translate(centerX-killsWidth/2, statsY+30)
	killsOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, killsText, gos.hudFont, killsOp)

	// Deaths
	deathsText := fmt.Sprintf("Deaths: %d", gos.deaths)
	deathsWidth, _ := text.Measure(deathsText, gos.hudFont, 0)
	deathsOp := &text.DrawOptions{}
	deathsOp.GeoM.Translate(centerX-deathsWidth/2, statsY+60)
	deathsOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, deathsText, gos.hudFont, deathsOp)

	// High score message (if applicable)
	if gos.isHighScore {
		highScoreText := "NEW HIGH SCORE!"
		highScoreWidth, _ := text.Measure(highScoreText, gos.hudFont, 0)
		highScoreOp := &text.DrawOptions{}
		highScoreOp.GeoM.Translate(centerX-highScoreWidth/2, statsY+100)
		highScoreOp.ColorScale.ScaleWithColor(color.RGBA{255, 215, 0, 255}) // Gold color
		text.Draw(screen, highScoreText, gos.hudFont, highScoreOp)
	}

	// Draw ebitenui widgets (text input and button)
	gos.ui.Draw(screen)
}
