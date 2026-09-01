package ui

import (
	"fyne.io/fyne/v2"

	"github.com/huijiro/jirolauncher/internal/config"
	"github.com/huijiro/jirolauncher/internal/runner"
)

func settingsPage(configuration *config.Configuration, runners []runner.Runner, window fyne.Window, onComplete func()) fyne.CanvasObject {
	return setupPage(configuration, runners, window, "Confirm", onComplete)
}
