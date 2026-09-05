package ui

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/Huijiro/YaYeet/internal/buildinfo"
	launcherupdate "github.com/Huijiro/YaYeet/internal/update"
)

func checkForLauncherUpdate(logger *slog.Logger, window fyne.Window) {
	if !buildinfo.UpdateChecksEnabled() {
		logger.Debug("launcher update check disabled", slog.String("version", buildinfo.Version))
		return
	}

	release, available, err := launcherupdate.Check(context.Background(), buildinfo.Version)
	if err != nil {
		logger.Warn("could not check for launcher update", slog.Any("error", err))
		return
	}
	if !available {
		return
	}

	fyne.Do(func() {
		if buildinfo.InstallMethod == "appimage" && os.Getenv("APPIMAGE") != "" && release.AppImageURL != "" && release.AppImageChecksumURL != "" {
			showAppImageUpdate(logger, window, release)
			return
		}
		showPackageUpdate(logger, window, release)
	})
}

func showAppImageUpdate(logger *slog.Logger, window fyne.Window, release launcherupdate.Release) {
	message := fmt.Sprintf("YaYeet %s is available. Download and install it now?", release.Version)
	dialog.ShowConfirm("Launcher update available", message, func(install bool) {
		if !install {
			return
		}
		go func() {
			err := launcherupdate.InstallAppImage(context.Background(), release, os.Getenv("APPIMAGE"))
			fyne.Do(func() {
				if err != nil {
					logger.Error("could not install launcher update", slog.Any("error", err))
					dialog.ShowError(err, window)
					return
				}
				dialog.ShowInformation(
					"Launcher updated",
					"The update was installed. Restart YaYeet to use the new version.",
					window,
				)
			})
		}()
	}, window)
}

func showPackageUpdate(logger *slog.Logger, window fyne.Window, release launcherupdate.Release) {
	message := fmt.Sprintf("YaYeet %s is available. Open the release page?", release.Version)
	dialog.ShowConfirm("Launcher update available", message, func(open bool) {
		if !open {
			return
		}
		releaseURL, err := url.Parse(release.PageURL)
		if err != nil {
			logger.Error("could not parse launcher release URL", slog.Any("error", err))
			dialog.ShowError(err, window)
			return
		}
		if err := fyne.CurrentApp().OpenURL(releaseURL); err != nil {
			logger.Error("could not open launcher release", slog.String("url", release.PageURL), slog.Any("error", err))
			dialog.ShowError(err, window)
		}
	}, window)
}
