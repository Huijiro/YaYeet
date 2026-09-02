package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var buttonGreen = color.NRGBA{R: 4, G: 170, B: 109, A: 255}

type outlinedButton struct {
	widget.DisableableWidget
	Text         string
	TextSizeName fyne.ThemeSizeName
	OnTapped     func()

	hovered bool
	active  bool
	timer   *time.Timer
}

func newOutlinedButton(text string, tapped func()) *outlinedButton {
	button := &outlinedButton{Text: text, TextSizeName: theme.SizeNameText, OnTapped: tapped}
	button.ExtendBaseWidget(button)
	return button
}

func (b *outlinedButton) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(color.Transparent)
	background.StrokeColor = buttonGreen
	background.StrokeWidth = b.Theme().Size(theme.SizeNameSeparatorThickness)

	label := canvas.NewText(b.Text, b.Theme().Color(theme.ColorNameForeground, fyne.CurrentApp().Settings().ThemeVariant()))
	label.Alignment = fyne.TextAlignCenter
	label.TextSize = b.Theme().Size(b.TextSizeName)
	label.TextStyle = fyne.TextStyle{Bold: true}

	renderer := &outlinedButtonRenderer{button: b, background: background, label: label}
	renderer.refresh()
	return renderer
}

func (b *outlinedButton) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (b *outlinedButton) MouseIn(*desktop.MouseEvent) {
	b.hovered = true
	b.Refresh()
}

func (b *outlinedButton) MouseMoved(*desktop.MouseEvent) {}

func (b *outlinedButton) MouseOut() {
	b.hovered = false
	b.Refresh()
}

func (b *outlinedButton) SetText(text string) {
	b.Text = text
	b.Refresh()
}

func (b *outlinedButton) Tapped(*fyne.PointEvent) {
	if b.Disabled() {
		return
	}

	b.active = true
	b.Refresh()
	if b.timer != nil {
		b.timer.Stop()
	}
	b.timer = time.AfterFunc(2*time.Second, func() {
		fyne.Do(func() {
			b.active = false
			b.Refresh()
		})
	})
	if b.OnTapped != nil {
		b.OnTapped()
	}
}

type outlinedButtonRenderer struct {
	button     *outlinedButton
	background *canvas.Rectangle
	label      *canvas.Text
}

func (r *outlinedButtonRenderer) Destroy() {}

func (r *outlinedButtonRenderer) Layout(size fyne.Size) {
	borderSize := r.button.Theme().Size(theme.SizeNameInputBorder)
	r.background.Move(fyne.NewSquareOffsetPos(borderSize / 2))
	r.background.Resize(fyne.NewSize(size.Width-borderSize-.5, size.Height-borderSize-.5))
	labelSize := r.label.MinSize()
	r.label.Resize(labelSize)
	r.label.Move(fyne.NewPos((size.Width-labelSize.Width)/2, (size.Height-labelSize.Height)/2))
}

func (r *outlinedButtonRenderer) MinSize() fyne.Size {
	padding := r.button.Theme().Size(theme.SizeNameInnerPadding) * 2
	return r.label.MinSize().Add(fyne.NewSquareSize(padding))
}

func (r *outlinedButtonRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.label}
}

func (r *outlinedButtonRenderer) Refresh() {
	r.refresh()
	canvas.Refresh(r.button)
}

func (r *outlinedButtonRenderer) refresh() {
	variant := fyne.CurrentApp().Settings().ThemeVariant()
	r.label.Text = r.button.Text
	r.label.TextSize = r.button.Theme().Size(r.button.TextSizeName)
	r.label.Color = r.button.Theme().Color(theme.ColorNameForeground, variant)
	r.background.FillColor = color.Transparent
	if r.button.Disabled() {
		r.label.Color = r.button.Theme().Color(theme.ColorNameDisabled, variant)
	} else if r.button.active {
		r.background.FillColor = buttonGreen
	} else if r.button.hovered {
		r.background.FillColor = color.NRGBA{R: buttonGreen.R, G: buttonGreen.G, B: buttonGreen.B, A: 204}
	}
	r.background.Refresh()
	r.label.Refresh()
}
