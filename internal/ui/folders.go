package ui

import (
	"log/slog"
	"net/url"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"

	"github.com/Huijiro/YaYeet/internal/game"
)

func openFolder(logger *slog.Logger, path string) {
	go func() {
		if err := os.MkdirAll(path, 0o755); err != nil {
			logger.Error("could not create folder", slog.String("path", path), slog.Any("error", err))
			return
		}

		absolutePath, err := filepath.Abs(path)
		if err != nil {
			logger.Error("could not resolve folder path", slog.String("path", path), slog.Any("error", err))
			return
		}
		if err := fyne.CurrentApp().OpenURL(&url.URL{Scheme: "file", Path: absolutePath}); err != nil {
			logger.Error("could not open folder", slog.String("path", absolutePath), slog.Any("error", err))
		}
	}()
}

func customContentPath(winePrefix string) string {
	return game.CustomContentPath(winePrefix)
}
