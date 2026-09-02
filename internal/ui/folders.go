package ui

import (
	"log/slog"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"fyne.io/fyne/v2"
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
	usersPath := filepath.Join(winePrefix, "drive_c", "users")
	for _, name := range wineUserCandidates(usersPath) {
		localAppData := filepath.Join(usersPath, name, "AppData", "Local")
		if info, err := os.Stat(localAppData); err == nil && info.IsDir() {
			return filepath.Join(localAppData, "VotV", "Assets")
		}
	}

	return filepath.Join(usersPath, defaultWineUser(), "AppData", "Local", "VotV", "Assets")
}

func wineUserCandidates(usersPath string) []string {
	candidates := []string{defaultWineUser(), "steamuser"}
	entries, err := os.ReadDir(usersPath)
	if err != nil {
		return candidates
	}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "Public" {
			candidates = append(candidates, entry.Name())
		}
	}
	return candidates
}

func defaultWineUser() string {
	current, err := user.Current()
	if err == nil && current.Username != "" {
		return current.Username
	}
	return "steamuser"
}
