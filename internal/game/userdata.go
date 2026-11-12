package game

import (
	"os"
	"os/user"
	"path/filepath"
)

// SaveGamesPath returns the VotV SaveGames directory inside a Wine/Proton prefix.
func SaveGamesPath(winePrefix string) string {
	return filepath.Join(localVotVPath(winePrefix), "Saved", "SaveGames")
}

// CustomContentPath returns the VotV custom Assets directory inside a Wine/Proton prefix.
func CustomContentPath(winePrefix string) string {
	return filepath.Join(localVotVPath(winePrefix), "Assets")
}

func localVotVPath(winePrefix string) string {
	usersPath := filepath.Join(winePrefix, "drive_c", "users")
	for _, name := range wineUserCandidates(usersPath) {
		localAppData := filepath.Join(usersPath, name, "AppData", "Local")
		if info, err := os.Stat(localAppData); err == nil && info.IsDir() {
			return filepath.Join(localAppData, "VotV")
		}
	}
	return filepath.Join(usersPath, defaultWineUser(), "AppData", "Local", "VotV")
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
