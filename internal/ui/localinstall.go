package ui

import (
	"context"
	"log/slog"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/Huijiro/YaYeet/internal/game"
)

func promptLocalGameArchiveInstall(logger *slog.Logger, installationPath string, window fyne.Window) {
	installationPath = strings.TrimSpace(installationPath)
	if installationPath == "" {
		dialog.ShowError(errString("Set the game installation path before installing a local archive."), window)
		return
	}

	chooser := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if reader == nil {
			return
		}
		archivePath := reader.URI().Path()
		_ = reader.Close()
		if archivePath == "" {
			dialog.ShowError(errString("Could not resolve the selected archive path."), window)
			return
		}
		askLocalGameArchiveVersion(logger, archivePath, installationPath, window)
	}, window)
	chooser.SetFilter(storage.NewExtensionFileFilter([]string{".7z"}))
	chooser.SetConfirmText("Use archive")
	chooser.Show()
}

func askLocalGameArchiveVersion(logger *slog.Logger, archivePath, installationPath string, window fyne.Window) {
	versionEntry := widget.NewEntry()
	versionEntry.SetText(game.VersionHintFromArchiveName(archivePath))
	versionEntry.SetPlaceHolder("0100_0009")
	versionEntry.Validator = func(value string) error {
		return game.ValidateVersionName(strings.TrimSpace(value))
	}

	dialog.ShowForm(
		"Local game version",
		"Install",
		"Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Version", versionEntry),
		},
		func(confirmed bool) {
			if !confirmed {
				return
			}
			installLocalGameArchive(logger, archivePath, installationPath, strings.TrimSpace(versionEntry.Text), window)
		},
		window,
	)
}

func installLocalGameArchive(logger *slog.Logger, archivePath, installationPath, version string, window fyne.Window) {
	progress := dialog.NewProgress("Installing local game archive", "Extracting "+version+"...", window)
	progress.Show()

	go func() {
		err := game.InstallFromArchive(context.Background(), archivePath, installationPath, version, func(update game.InstallProgress) {
			fyne.Do(func() {
				if update.Total > 0 {
					progress.SetValue(float64(update.Current) / float64(update.Total))
				}
			})
		})
		fyne.Do(func() {
			progress.Hide()
			if err != nil {
				logger.Error("local archive install failed", slog.String("archive", archivePath), slog.String("version", version), slog.Any("error", err))
				dialog.ShowError(err, window)
				return
			}
			logger.Info("local archive installed", slog.String("archive", archivePath), slog.String("version", version), slog.String("path", installationPath))
			dialog.ShowInformation(
				"Local game installed",
				version+" was installed into the game folder. Return to the home screen to play it.",
				window,
			)
		})
	}()
}

type stringError string

func (err stringError) Error() string { return string(err) }

func errString(message string) error { return stringError(message) }
