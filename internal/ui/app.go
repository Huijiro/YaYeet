package ui

import (
	"context"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Huijiro/YaYeet/internal/config"
	"github.com/Huijiro/YaYeet/internal/game"
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
		versions, latest, err := game.AvailableVersions(context.Background())
		if err != nil {
			logger.Error("could not fetch game versions", slog.Any("error", err))
			return
		}
		var home fyne.CanvasObject
		home = homePage(logger, configuration, versions, latest, func() {
			settings := settingsPage(configuration, runners, window, func() {
				loading := container.NewCenter(widget.NewLabel("Loading game versions..."))
				window.SetContent(loading)
				window.Resize(loading.MinSize().Add(fyne.NewSize(48, 48)))
				go showHome()
			})
			window.SetContent(settings)
			window.Resize(settings.MinSize().Add(fyne.NewSize(48, 48)))
		})
		fyne.Do(func() {
			window.SetContent(home)
			window.Resize(home.MinSize().Add(fyne.NewSize(48, 48)))
		})
	}

	var content fyne.CanvasObject
	if configured {
		content = container.NewCenter(widget.NewLabel("Loading game versions..."))
		go showHome()
	} else {
		content = setupPage(configuration, runners, window, "Continue", func() { go showHome() })
	}

	window.SetContent(content)
	window.Resize(content.MinSize().Add(fyne.NewSize(48, 48)))
	window.ShowAndRun()
}
