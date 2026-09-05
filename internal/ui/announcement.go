package ui

import (
	"log/slog"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

func announcementArea(logger *slog.Logger) fyne.CanvasObject {
	plushieURL := &url.URL{
		Scheme: "https",
		Host:   "www.makeship.com",
		Path:   "/products/argemia-the-ariral-plush",
	}
	plushieButton := newRainbowOutlinedButton("Buy the plushie", func() {
		if err := fyne.CurrentApp().OpenURL(plushieURL); err != nil {
			logger.Error("could not open announcement link", slog.String("url", plushieURL.String()), slog.Any("error", err))
		}
	})
	plushieButton.SizeScale = 2

	return container.NewHBox(plushieButton, layout.NewSpacer())
}
