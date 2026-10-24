package ui

import (
	"context"
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/Huijiro/YaYeet/internal/runner"
)

func offerProtonGEInstall(logger *slog.Logger, window fyne.Window, onInstalled func([]runner.Runner)) {
	fyne.Do(func() {
		message := "No Wine or Proton runners were found. Install the latest ProtonGE? It is currently the most compatible option for Voices of the Void."
		dialog.ShowConfirm("Install ProtonGE", message, func(install bool) {
			if !install {
				return
			}
			progress := dialog.NewProgress("Installing ProtonGE", "Downloading ProtonGE...", window)
			progress.Show()
			go func() {
				tag, err := runner.InstallLatestProtonGE(context.Background(), logger, func(update runner.ProtonGEProgress) {
					if update.Phase != "download" || update.Total <= 0 {
						return
					}
					value := float64(update.Current) / float64(update.Total)
					fyne.Do(func() { progress.SetValue(value) })
				})
				var runners []runner.Runner
				if err == nil {
					runners, err = runner.List(logger)
				}
				fyne.Do(func() {
					progress.Hide()
					if err != nil {
						logger.Error("could not install ProtonGE", slog.Any("error", err))
						dialog.ShowError(err, window)
						return
					}
					onInstalled(runners)
					dialog.ShowInformation("ProtonGE installed", fmt.Sprintf("%s was installed.", tag), window)
				})
			}()
		}, window)
	})
}
