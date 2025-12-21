package ui

import (
	"fmt"
	"image/color"

	"github.com/ebitenui/ebitenui"
	ebitenui_image "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// SettingsScreen represents the settings dialog using ebitenui
type SettingsScreen struct {
	ui              *ebitenui.UI
	soundMute       *widget.Checkbox
	soundVolume     *widget.Slider
	musicMute       *widget.Checkbox
	musicVolume     *widget.Slider
	chatEnabled     *widget.Checkbox
	soundVolumeText *widget.Text
	musicVolumeText *widget.Text
	onApply         func(soundMuted bool, soundVolume float64, musicMuted bool, musicVolume float64, chatEnabled bool)
	onClose         func(soundMuted bool, soundVolume float64, musicMuted bool, musicVolume float64, chatEnabled bool)
	font            text.Face
	hudFont         *text.GoTextFace
}

// NewSettingsScreen creates a new settings screen with ebitenui widgets
func NewSettingsScreen(soundMuted bool, soundVolume float64, musicMuted bool, musicVolume float64, chatEnabled bool,
	fontSource *text.GoTextFaceSource,
	onApply func(soundMuted bool, soundVolume float64, musicMuted bool, musicVolume float64, chatEnabled bool),
	onClose func(soundMuted bool, soundVolume float64, musicMuted bool, musicVolume float64, chatEnabled bool)) *SettingsScreen {

	// Create font faces
	hudFont := &text.GoTextFace{
		Source: fontSource,
		Size:   14,
	}

	ss := &SettingsScreen{
		onApply: onApply,
		onClose: onClose,
		font:    hudFont,
		hudFont: hudFont,
	}

	// Create root container with anchor layout (for centering)
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	// Create main content container with background and vertical layout
	contentContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ebitenui_image.NewNineSliceColor(color.NRGBA{20, 20, 30, 240})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(20),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top:    50,
				Bottom: 30,
				Left:   40,
				Right:  40,
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(500, 300),
		),
	)

	// Title text
	titleText := widget.NewText(
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
		widget.TextOpts.Text("Settings", &ss.font, color.NRGBA{255, 255, 255, 255}),
	)

	// Sound Effects Section
	soundEffectsRow := ss.createSettingsRow(
		"Sound Effects",
		soundMuted,
		soundVolume,
		func(checked bool) {
			// Sound mute checkbox changed
			ss.applyCurrentSettings()
		},
		func(args *widget.SliderChangedEventArgs) {
			// Sound volume slider changed
			ss.soundVolumeText.Label = fmt.Sprintf("%d%%", args.Current)
			ss.applyCurrentSettings()
		},
	)
	ss.soundMute = soundEffectsRow.checkbox
	ss.soundVolume = soundEffectsRow.slider
	ss.soundVolumeText = soundEffectsRow.volumeText

	// Music Section
	musicRow := ss.createSettingsRow(
		"Music",
		musicMuted,
		musicVolume,
		func(checked bool) {
			// Music mute checkbox changed
			ss.applyCurrentSettings()
		},
		func(args *widget.SliderChangedEventArgs) {
			// Music volume slider changed
			ss.musicVolumeText.Label = fmt.Sprintf("%d%%", args.Current)
			ss.applyCurrentSettings()
		},
	)
	ss.musicMute = musicRow.checkbox
	ss.musicVolume = musicRow.slider
	ss.musicVolumeText = musicRow.volumeText

	// Chat Section (simple checkbox, no volume slider)
	chatRow := ss.createSimpleCheckboxRow(
		"Enable Chat",
		chatEnabled,
		func(checked bool) {
			// Chat enabled checkbox changed
			ss.applyCurrentSettings()
		},
	)
	ss.chatEnabled = chatRow.checkbox

	// Close button
	closeButton := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(120, 40),
		),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    ebitenui_image.NewNineSliceColor(color.NRGBA{R: 70, G: 130, B: 180, A: 255}),
			Hover:   ebitenui_image.NewNineSliceColor(color.NRGBA{R: 90, G: 150, B: 200, A: 255}),
			Pressed: ebitenui_image.NewNineSliceColor(color.NRGBA{R: 50, G: 110, B: 160, A: 255}),
		}),
		widget.ButtonOpts.Text("Close", &ss.font, &widget.ButtonTextColor{
			Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   30,
			Right:  30,
			Top:    10,
			Bottom: 10,
		}),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if ss.onClose != nil {
				ss.onClose(
					ss.soundMute.State() == widget.WidgetChecked,
					float64(ss.soundVolume.Current)/100.0,
					ss.musicMute.State() == widget.WidgetChecked,
					float64(ss.musicVolume.Current)/100.0,
					ss.chatEnabled.State() == widget.WidgetChecked,
				)
			}
		}),
	)

	// Add all widgets to content container
	contentContainer.AddChild(titleText)
	contentContainer.AddChild(soundEffectsRow.container)
	contentContainer.AddChild(musicRow.container)
	contentContainer.AddChild(chatRow.container)
	contentContainer.AddChild(closeButton)

	// Add content to root container
	rootContainer.AddChild(contentContainer)

	// Create UI
	ss.ui = &ebitenui.UI{
		Container: rootContainer,
	}

	return ss
}

type settingsRow struct {
	container  *widget.Container
	checkbox   *widget.Checkbox
	slider     *widget.Slider
	volumeText *widget.Text
}

// createSettingsRow creates a row with label, mute checkbox, and volume slider
func (ss *SettingsScreen) createSettingsRow(label string, muted bool, volume float64,
	onCheckboxChanged func(bool), onSliderChanged func(*widget.SliderChangedEventArgs)) settingsRow {

	// Row container
	rowContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(8),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
	)

	// Mute checkbox row (label + checkbox)
	muteRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(10),
		)),
	)

	muteLabel := widget.NewText(
		widget.TextOpts.Text(fmt.Sprintf("Mute %s", label), &ss.font, color.NRGBA{255, 255, 255, 255}),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)

	checkbox := widget.NewCheckbox(
		widget.CheckboxOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(20, 20),
		),
		widget.CheckboxOpts.Image(&widget.CheckboxImage{
			Unchecked: ebitenui_image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
			Checked:   ebitenui_image.NewNineSliceColor(color.NRGBA{100, 255, 100, 255}),
		}),
		widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
			if onCheckboxChanged != nil {
				onCheckboxChanged(args.State == widget.WidgetChecked)
			}
		}),
	)

	// Set initial state
	if muted {
		checkbox.SetState(widget.WidgetChecked)
	}

	muteRow.AddChild(muteLabel)
	muteRow.AddChild(checkbox)

	// Volume slider row (label + slider + percentage)
	volumeRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(10),
		)),
	)

	volumeLabel := widget.NewText(
		widget.TextOpts.Text(fmt.Sprintf("%s Volume", label), &ss.font, color.NRGBA{255, 255, 255, 255}),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
			widget.WidgetOpts.MinSize(120, 0),
		),
	)

	slider := widget.NewSlider(
		widget.SliderOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
			widget.WidgetOpts.MinSize(200, 20),
		),
		widget.SliderOpts.MinMax(0, 100),
		widget.SliderOpts.Images(
			&widget.SliderTrackImage{
				Idle:  ebitenui_image.NewNineSliceColor(color.NRGBA{40, 40, 40, 255}),
				Hover: ebitenui_image.NewNineSliceColor(color.NRGBA{50, 50, 50, 255}),
			},
			&widget.ButtonImage{
				Idle:    ebitenui_image.NewNineSliceColor(color.NRGBA{200, 200, 200, 255}),
				Hover:   ebitenui_image.NewNineSliceColor(color.NRGBA{220, 220, 220, 255}),
				Pressed: ebitenui_image.NewNineSliceColor(color.NRGBA{180, 180, 180, 255}),
			},
		),
		widget.SliderOpts.ChangedHandler(onSliderChanged),
	)

	// Set initial volume
	slider.Current = int(volume * 100)

	volumeText := widget.NewText(
		widget.TextOpts.Text(fmt.Sprintf("%d%%", int(volume*100)), &ss.font, color.NRGBA{180, 180, 180, 255}),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionEnd,
			}),
			widget.WidgetOpts.MinSize(50, 0),
		),
	)

	volumeRow.AddChild(volumeLabel)
	volumeRow.AddChild(slider)
	volumeRow.AddChild(volumeText)

	// Add both rows to container
	rowContainer.AddChild(muteRow)
	rowContainer.AddChild(volumeRow)

	return settingsRow{
		container:  rowContainer,
		checkbox:   checkbox,
		slider:     slider,
		volumeText: volumeText,
	}
}

type simpleCheckboxRow struct {
	container *widget.Container
	checkbox  *widget.Checkbox
}

// createSimpleCheckboxRow creates a row with just a label and checkbox (no volume slider)
func (ss *SettingsScreen) createSimpleCheckboxRow(label string, checked bool,
	onCheckboxChanged func(bool)) simpleCheckboxRow {

	// Row container
	rowContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(15),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
	)

	// Label text
	labelText := widget.NewText(
		widget.TextOpts.Text(label, &ss.font, color.NRGBA{255, 255, 255, 255}),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)

	// Checkbox
	checkbox := widget.NewCheckbox(
		widget.CheckboxOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(20, 20),
		),
		widget.CheckboxOpts.Image(&widget.CheckboxImage{
			Unchecked: ebitenui_image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
			Checked:   ebitenui_image.NewNineSliceColor(color.NRGBA{100, 255, 100, 255}),
		}),
		widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
			onCheckboxChanged(args.State == widget.WidgetChecked)
		}),
	)

	// Set initial state
	if checked {
		checkbox.SetState(widget.WidgetChecked)
	}

	rowContainer.AddChild(labelText)
	rowContainer.AddChild(checkbox)

	return simpleCheckboxRow{
		container: rowContainer,
		checkbox:  checkbox,
	}
}

// applyCurrentSettings calls the onApply callback with current widget states
// This helper reduces duplication of the same callback logic across multiple handlers
func (ss *SettingsScreen) applyCurrentSettings() {
	if ss.onApply != nil {
		ss.onApply(
			ss.soundMute.State() == widget.WidgetChecked,
			float64(ss.soundVolume.Current)/100.0,
			ss.musicMute.State() == widget.WidgetChecked,
			float64(ss.musicVolume.Current)/100.0,
			ss.chatEnabled.State() == widget.WidgetChecked,
		)
	}
}

// Update processes UI updates
func (ss *SettingsScreen) Update() {
	ss.ui.Update()
}

// Draw renders the settings screen
func (ss *SettingsScreen) Draw(screen *ebiten.Image) {
	ss.ui.Draw(screen)
}
