package ui

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

//go:embed assets/ShareTechMono-Regular.ttf
var shareTechMono []byte

//go:embed assets/background.png
var backgroundImage []byte

type launcherTheme struct {
	base fyne.Theme
	font fyne.Resource
}

type quietInteractionTheme struct {
	fyne.Theme
}

func (t *quietInteractionTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameHover || name == theme.ColorNameFocus {
		return color.Transparent
	}
	return t.Theme.Color(name, variant)
}

func withBackground(content fyne.CanvasObject) fyne.CanvasObject {
	background := canvas.NewImageFromResource(fyne.NewStaticResource("background.png", backgroundImage))
	background.FillMode = canvas.ImageFillStretch
	background.SetMinSize(fyne.NewSize(0, 0))
	return container.NewStack(background, content)
}

func withoutInteractionEffect(content fyne.CanvasObject) fyne.CanvasObject {
	return container.NewThemeOverride(content, &quietInteractionTheme{Theme: newLauncherTheme()})
}

func outlinedInput(input fyne.CanvasObject) fyne.CanvasObject {
	border := canvas.NewRectangle(color.Transparent)
	border.StrokeColor = buttonGreen
	border.StrokeWidth = theme.Size(theme.SizeNameInputBorder)
	return container.NewStack(border, input)
}

func newLauncherTheme() fyne.Theme {
	return &launcherTheme{
		base: theme.DarkTheme(),
		font: fyne.NewStaticResource("ShareTechMono-Regular.ttf", shareTechMono),
	}
}

func (t *launcherTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameInputBackground:
		return color.Transparent
	case theme.ColorNameInputBorder, theme.ColorNamePrimary, theme.ColorNameScrollBar:
		return buttonGreen
	case theme.ColorNameFocus:
		return color.NRGBA{R: buttonGreen.R, G: buttonGreen.G, B: buttonGreen.B, A: 26}
	case theme.ColorNameMenuBackground:
		return color.NRGBA{A: 204}
	default:
		return t.base.Color(name, variant)
	}
}

func (t *launcherTheme) Font(fyne.TextStyle) fyne.Resource {
	return t.font
}

func (t *launcherTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *launcherTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameButtonRadius, theme.SizeNameInputRadius, theme.SizeNameScrollBarRadius:
		return 0
	default:
		return t.base.Size(name)
	}
}
