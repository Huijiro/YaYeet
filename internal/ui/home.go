package ui

import (
	"context"
	"log/slog"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Huijiro/YaYeet/internal/config"
	"github.com/Huijiro/YaYeet/internal/game"
)

type thinProgressLayout struct{}

func (thinProgressLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, object := range objects {
		object.Resize(size)
	}
}

func (thinProgressLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, theme.Size(theme.SizeNameSeparatorThickness))
}

func homePage(logger *slog.Logger, configuration *config.Configuration, openSettings func()) fyne.CanvasObject {
	byLabel := make(map[string]game.VersionOption)
	versionSelect := widget.NewSelect(nil, nil)
	versionSelect.Disable()
	progress := widget.NewProgressBar()
	progress.TextFormatter = func() string { return "" }
	installedVersion := ""
	versionsLoaded := false
	detectionFinished := false
	install := widget.NewButton("Play", nil)
	install.Disable()
	updateAction := func() {
		if !detectionFinished {
			return
		}
		selected, ok := byLabel[versionSelect.Selected]
		if ok && installedVersion != "" && strings.HasPrefix(selected.Name, installedVersion) {
			install.SetText("Play")
		} else {
			install.SetText("Install selected version")
		}
	}
	versionSelect.OnChanged = func(string) { updateAction() }

	go func() {
		detected, err := game.Detect(context.Background(), configuration.InstallationPath)
		fyne.Do(func() {
			if err == nil {
				installedVersion = detected.Installed
			}
			detectionFinished = true
			updateAction()
			if versionsLoaded {
				install.Enable()
			}
		})
	}()

	go func() {
		versions, latest, err := game.AvailableVersions(context.Background())
		if err != nil {
			logger.Error("could not fetch game versions", slog.Any("error", err))
			return
		}

		if !configuration.ShowRevisions {
			versions = game.LatestRevisions(versions)
		}

		labels := make([]string, 0, len(versions))
		versionsByLabel := make(map[string]game.VersionOption, len(versions))
		selectedLabel := latest
		for _, version := range versions {
			if version.Unstable && !configuration.ShowUnstable {
				continue
			}
			if version.Test && !configuration.ShowTest {
				continue
			}
			label := version.Label
			if !configuration.ShowRevisions {
				label = version.RevisionlessLabel
			}
			labels = append(labels, label)
			versionsByLabel[label] = version
			if version.Label == latest {
				selectedLabel = label
			}
		}
		fyne.Do(func() {
			byLabel = versionsByLabel
			versionSelect.Options = labels
			versionSelect.SetSelected(selectedLabel)
			versionSelect.Enable()
			versionsLoaded = true
			updateAction()
			if detectionFinished {
				install.Enable()
			}
		})
	}()

	install.OnTapped = func() {
		selected, ok := byLabel[versionSelect.Selected]
		if !ok {
			return
		}
		if installedVersion != "" && strings.HasPrefix(selected.Name, installedVersion) {
			logger.Info("play requested", slog.String("version", selected.Name))
			go func() {
				if err := game.Launch(context.Background(), logger, configuration.InstallationPath, configuration.Runner.Executable, configuration.WinePrefix); err != nil {
					logger.Error("play request failed", slog.Any("error", err))
				}
			}()
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

	title := widget.NewLabel("YaYeet Launcher")
	bottom := container.NewGridWithColumns(
		3,
		container.NewVBox(
			versionSelect,
			container.New(thinProgressLayout{}, progress),
			install,
		),
		layout.NewSpacer(),
		container.NewHBox(
			layout.NewSpacer(),
			widget.NewButton("Configuration", openSettings),
		),
	)

	return container.NewBorder(title, bottom, nil, nil, layout.NewSpacer())
}
