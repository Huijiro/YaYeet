package ui

import (
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/Huijiro/YaYeet/internal/config"
	"github.com/Huijiro/YaYeet/internal/runner"
)

func Run(logger *slog.Logger, configuration *config.Configuration, runners []runner.Runner) {
	launcher := app.NewWithID("io.github.huijiro.YaYeet")
	window := launcher.NewWindow("YaYeet")

	configured, err := config.Exists()
	if err != nil {
		logger.Error("could not check configuration file", slog.Any("error", err))
		return
	}

	var showHome func()
	showHome = func() {
		var home fyne.CanvasObject
		home = homePage(logger, configuration, func() {
			settings := settingsPage(configuration, runners, window, showHome)
			window.SetContent(settings)
			window.Resize(settings.MinSize().Add(fyne.NewSize(48, 48)))
		})
		fyne.Do(func() {
			window.SetContent(home)
			window.Resize(fyne.NewSize(1280, 720))
		})
	}

	var content fyne.CanvasObject
	if configured {
		showHome()
		window.ShowAndRun()
		return
	} else {
		content = setupPage(configuration, runners, window, "Continue", showHome)
	}

	window.SetContent(content)
	window.Resize(content.MinSize().Add(fyne.NewSize(48, 48)))
	window.ShowAndRun()
}
