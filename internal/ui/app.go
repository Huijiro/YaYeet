package ui

import (
	"log/slog"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/Huijiro/YaYeet/internal/config"
	"github.com/Huijiro/YaYeet/internal/runner"
)

func Run(logger *slog.Logger, configuration *config.Configuration, runners []runner.Runner) {
	launcher := app.NewWithID("io.github.huijiro.YaYeet")
	launcher.Settings().SetTheme(newLauncherTheme())
	window := launcher.NewWindow("YaYeet")

	configured, err := config.Exists()
	if err != nil {
		logger.Error("could not check configuration file", slog.Any("error", err))
		return
	}

	var updateCheck sync.Once
	var showHome func()
	showHome = func() {
		var home fyne.CanvasObject
		home = withBackground(homePage(logger, configuration, window, func() {
			settings := withBackground(settingsPage(configuration, runners, window, showHome))
			window.SetContent(settings)
			window.Resize(settings.MinSize().Add(fyne.NewSize(48, 48)))
		}))
		fyne.Do(func() {
			window.SetContent(home)
			window.Resize(fyne.NewSize(1280, 720))
		})
		updateCheck.Do(func() {
			go checkForLauncherUpdate(logger, window)
		})
	}

	var content fyne.CanvasObject
	if configured {
		showHome()
		window.ShowAndRun()
		return
	} else {
		content = withBackground(setupPage(configuration, runners, window, "Continue", showHome))
	}

	window.SetContent(content)
	window.Resize(content.MinSize().Add(fyne.NewSize(48, 48)))
	window.ShowAndRun()
}
