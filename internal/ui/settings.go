package ui

import (
	"log/slog"

	"fyne.io/fyne/v2"

	"github.com/Huijiro/YaYeet/internal/config"
	"github.com/Huijiro/YaYeet/internal/runner"
)

func settingsPage(logger *slog.Logger, configuration *config.Configuration, runners []runner.Runner, window fyne.Window, onComplete func()) fyne.CanvasObject {
	return setupPage(logger, configuration, runners, window, "Confirm", onComplete)
}
