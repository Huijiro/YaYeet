package ui

import (
	"context"
	"log/slog"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Huijiro/YaYeet/internal/config"
	"github.com/Huijiro/YaYeet/internal/game"
)

type thinProgressLayout struct{}

type verticalFillLayout struct{}

func (verticalFillLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	height := size.Height / float32(len(objects))
	for index, object := range objects {
		object.Move(fyne.NewPos(0, float32(index)*height))
		object.Resize(fyne.NewSize(size.Width, height))
	}
}

func (verticalFillLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var width, height float32
	for _, object := range objects {
		minimum := object.MinSize()
		width = fyne.Max(width, minimum.Width)
		height += minimum.Height
	}
	return fyne.NewSize(width, height)
}

func (thinProgressLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, object := range objects {
		object.Resize(size)
	}
}

func (thinProgressLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, theme.Size(theme.SizeNameSeparatorThickness)*2)
}

func homePage(logger *slog.Logger, configuration *config.Configuration, openSettings func()) fyne.CanvasObject {
	byLabel := make(map[string]game.VersionOption)
	versionSelect := widget.NewSelect(nil, nil)
	versionSelect.Disable()
	progress := widget.NewProgressBar()
	progress.TextFormatter = func() string { return "" }
	progress.Hide()
	installedVersion := ""
	versionsLoaded := false
	detectionFinished := false
	install := newOutlinedButton("Play", nil)
	install.TextSizeName = theme.SizeNameHeadingText
	install.Disable()
	openGameFolder := newOutlinedButton("Open Game Folder", func() {
		openFolder(logger, configuration.InstallationPath)
	})
	openGameFolder.Disable()
	openCustomContentFolder := newOutlinedButton("Open Custom Content Folder", func() {
		openFolder(logger, customContentPath(configuration.WinePrefix))
	})
	openCustomContentFolder.Disable()
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
				openGameFolder.Enable()
				openCustomContentFolder.Enable()
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
		progress.Show()

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
				progress.Hide()
				if err != nil {
					return
				}
				progress.SetValue(1)
				installedVersion = selected.Name
				openGameFolder.Enable()
				openCustomContentFolder.Enable()
				updateAction()
			})
		}()
	}

	title := canvas.NewText("YaYeet Launcher", theme.Color(theme.ColorNameForeground))
	title.TextSize = theme.Size(theme.SizeNameHeadingText) * 2
	title.TextStyle = fyne.TextStyle{Bold: true}
	configurationButton := newOutlinedButton("Configuration", openSettings)
	bottom := container.NewGridWithColumns(
		3,
		container.NewVBox(
			container.New(thinProgressLayout{}, progress),
			outlinedInput(versionSelect),
			install,
		),
		layout.NewSpacer(),
		container.NewHBox(
			layout.NewSpacer(),
			container.New(verticalFillLayout{}, openGameFolder, openCustomContentFolder),
			configurationButton,
		),
	)

	return container.NewBorder(title, bottom, nil, nil, layout.NewSpacer())
}
