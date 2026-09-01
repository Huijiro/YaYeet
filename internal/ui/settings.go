package ui

import (
	"fyne.io/fyne/v2"

	"github.com/Huijiro/YaYeet/internal/config"
	"github.com/Huijiro/YaYeet/internal/runner"
)

func settingsPage(configuration *config.Configuration, runners []runner.Runner, window fyne.Window, onComplete func()) fyne.CanvasObject {
	return setupPage(configuration, runners, window, "Confirm", onComplete)
}
