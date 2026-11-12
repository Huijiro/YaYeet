package ui

import (
	"context"
	"log/slog"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Huijiro/YaYeet/internal/config"
	"github.com/Huijiro/YaYeet/internal/game"
	patchnotes "github.com/Huijiro/YaYeet/patch-notes"
)

type thinProgressLayout struct{}

type verticalFillLayout struct{}

type patchNotesLayout struct{}

func (patchNotesLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 2 {
		return
	}
	gap := theme.Size(theme.SizeNamePadding)
	notesWidth := (size.Width - gap) / 3
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(size.Width-notesWidth-gap, size.Height))
	objects[1].Move(fyne.NewPos(size.Width-notesWidth, 0))
	objects[1].Resize(fyne.NewSize(notesWidth, size.Height))
}

func (patchNotesLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) != 2 {
		return fyne.NewSize(0, 0)
	}
	left := objects[0].MinSize()
	right := objects[1].MinSize()
	return fyne.NewSize(left.Width+right.Width+theme.Size(theme.SizeNamePadding), fyne.Max(left.Height, right.Height))
}

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

func homePage(logger *slog.Logger, configuration *config.Configuration, window fyne.Window, session *installSession, openSettings func()) fyne.CanvasObject {
	byLabel := make(map[string]game.VersionOption)
	patchNotes := widget.NewRichTextFromMarkdown("# Patch notes\n\nLoading...")
	patchNotes.Wrapping = fyne.TextWrapWord
	patchNotesScroll := container.NewVScroll(patchNotes)
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
	configurationButton := newOutlinedButton("Configuration", openSettings)
	canEnableInstall := func() bool {
		return detectionFinished && versionsLoaded && !session.snapshot().Running
	}
	updateAction := func() {
		if session.snapshot().Running {
			install.SetText("Installing...")
			return
		}
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
	applyInstallProgress := func(value float64) {
		progress.SetValue(value)
		progress.Show()
		install.SetText("Installing...")
		install.Disable()
		versionSelect.Disable()
	}
	finishInstallUI := func(version string, err error) {
		if versionsLoaded {
			versionSelect.Enable()
		}
		progress.Hide()
		if err != nil {
			logger.Error("install failed", slog.String("version", version), slog.Any("error", err))
			updateAction()
			if canEnableInstall() {
				install.Enable()
			}
			return
		}
		progress.SetValue(1)
		installedVersion = version
		openGameFolder.Enable()
		openCustomContentFolder.Enable()
		updateAction()
		if canEnableInstall() {
			install.Enable()
		}
	}
	session.bind(applyInstallProgress, finishInstallUI)
	if snapshot := session.snapshot(); snapshot.Running {
		applyInstallProgress(snapshot.Progress)
	} else if snapshot.CompletedVersion != "" {
		installedVersion = snapshot.CompletedVersion
		openGameFolder.Enable()
		openCustomContentFolder.Enable()
	}
	versionSelect.OnChanged = func(label string) {
		updateAction()
		selected, ok := byLabel[label]
		if !ok {
			return
		}
		notes, found := patchnotes.ForVersion(selected.Name)
		if !found {
			previousVersion, previousNotes, previousFound := patchnotes.PreviousForVersion(selected.Name)
			notes = "# Patch notes\n\nNo patches for this version."
			if previousFound {
				notes += " Previous patch notes:\n\n## " + previousVersion + "\n\n" + previousNotes
			}
		}
		patchNotes.ParseMarkdown(notes)
		patchNotesScroll.ScrollToTop()
	}

	go func() {
		detected, err := game.Detect(context.Background(), configuration.InstallationPath)
		fyne.Do(func() {
			detectionFinished = true
			// Ignore partial-tree Detect results while an install is still writing files.
			if !session.snapshot().Running {
				if err == nil {
					installedVersion = detected.Installed
					openGameFolder.Enable()
					openCustomContentFolder.Enable()
				}
				if canEnableInstall() {
					install.Enable()
				}
			}
			updateAction()
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
			if !configuration.ShowUnstable {
				label = strings.TrimSuffix(label, " (stable)")
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
			versionsLoaded = true
			if !session.snapshot().Running {
				versionSelect.Enable()
			}
			updateAction()
			if canEnableInstall() {
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
			install.SetText("Launching the game")
			install.Disable()
			hideLauncher := configuration.HideWhileGameRunning
			if hideLauncher {
				window.Hide()
			}
			go func() {
				err := game.Launch(context.Background(), logger, configuration.InstallationPath, configuration.Runner.Executable, configuration.WinePrefix, configuration.LaunchArguments)
				fyne.Do(func() {
					if hideLauncher {
						window.Show()
					}
					updateAction()
					install.Enable()
				})
				if err != nil {
					logger.Error("play request failed", slog.Any("error", err))
				}
			}()
			return
		}

		if !session.tryStart(selected.Name) {
			return
		}
		installationPath := configuration.InstallationPath
		downloadPath := configuration.DownloadPath
		applyInstallProgress(0)

		go func() {
			err := game.Install(context.Background(), selected.URL, installationPath, downloadPath, func(update game.InstallProgress) {
				if update.Total > 0 {
					session.setProgress(float64(update.Current) / float64(update.Total))
				}
			})
			session.finish(selected.Name, err)
		}()
	}

	title := canvas.NewText("YaYeet Launcher", theme.Color(theme.ColorNameForeground))
	title.TextSize = theme.Size(theme.SizeNameHeadingText) * 2
	title.TextStyle = fyne.TextStyle{Bold: true}
	openSocial := func(destination *url.URL) func() {
		return func() {
			if err := fyne.CurrentApp().OpenURL(destination); err != nil {
				logger.Error("could not open social link", slog.String("url", destination.String()), slog.Any("error", err))
			}
		}
	}
	header := container.NewVBox(title)
	socials := container.NewHBox(
		newOutlinedButton("Discord", openSocial(&url.URL{Scheme: "https", Host: "discord.gg", Path: "/eternitydevgames"})),
		newOutlinedButton("Patreon", openSocial(&url.URL{Scheme: "https", Host: "www.patreon.com", Path: "/eternitydev/"})),
		newOutlinedButton("Boosty", openSocial(&url.URL{Scheme: "https", Host: "boosty.to", Path: "/mrdrnose"})),
	)
	bottom := container.NewGridWithColumns(
		3,
		container.NewVBox(
			container.New(thinProgressLayout{}, progress),
			outlinedInput(versionSelect),
			install,
		),
		container.NewCenter(socials),
		container.NewHBox(
			layout.NewSpacer(),
			container.New(verticalFillLayout{}, openGameFolder, openCustomContentFolder),
			configurationButton,
		),
	)

	mainContent := container.New(
		patchNotesLayout{},
		container.NewBorder(header, nil, nil, nil, layout.NewSpacer()),
		patchNotesScroll,
	)
	return container.NewBorder(nil, bottom, nil, nil, mainContent)
}
