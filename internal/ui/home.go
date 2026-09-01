package ui

import (
	"context"
	"log/slog"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/huijiro/jirolauncher/internal/config"
	"github.com/huijiro/jirolauncher/internal/game"
)

func homePage(logger *slog.Logger, configuration *config.Configuration, versions []game.VersionOption, latest string, openSettings func()) fyne.CanvasObject {
	labels := make([]string, 0, len(versions))
	byLabel := make(map[string]game.VersionOption, len(versions))
	for _, version := range versions {
		labels = append(labels, version.Label)
		byLabel[version.Label] = version
	}

	versionSelect := widget.NewSelect(labels, nil)
	versionSelect.SetSelected(latest)
	progress := widget.NewProgressBar()
	installedVersion := ""
	if detected, err := game.Detect(context.Background(), configuration.InstallationPath); err == nil {
		installedVersion = detected.Installed
	}
	install := widget.NewButton("Install selected version", nil)
	updateAction := func() {
		selected, ok := byLabel[versionSelect.Selected]
		if ok && installedVersion != "" && strings.HasPrefix(selected.Name, installedVersion) {
			install.SetText("Play")
		} else {
			install.SetText("Install selected version")
		}
	}
	versionSelect.OnChanged = func(string) { updateAction() }
	updateAction()

	install.OnTapped = func() {
		selected, ok := byLabel[versionSelect.Selected]
		if !ok {
			return
		}
		if installedVersion != "" && strings.HasPrefix(selected.Name, installedVersion) {
			logger.Info("play requested", slog.String("version", selected.Name))
			if err := game.Launch(context.Background(), logger, configuration.InstallationPath, configuration.Runner.Executable, configuration.WinePrefix); err != nil {
				logger.Error("play request failed", slog.Any("error", err))
			}
			return
		}

		ctx := context.Background()
		install.Disable()
		progress.SetValue(0)

		go func() {
			err := game.Install(ctx, selected.URL, configuration.InstallationPath, configuration.DownloadPath, func(update game.InstallProgress) {
				fyne.Do(func() {
					if update.Total > 0 {
						progress.SetValue(float64(update.Current) / float64(update.Total))
					}
				})
			})
			fyne.Do(func() {
				install.Enable()
				if err != nil {
					return
				}
				progress.SetValue(1)
				installedVersion = selected.Name
				updateAction()
			})
		}()
	}

	return container.NewCenter(container.NewVBox(
		widget.NewLabel("JiroLauncher is ready."),
		widget.NewLabel("Select a game version:"),
		versionSelect,
		progress,
		install,
		widget.NewButton("Configuration", openSettings),
	))
}
