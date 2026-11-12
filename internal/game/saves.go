package game

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BackupSaveGames copies SaveGames into installationPath/save-backups when switching versions.
// It is a no-op when fromVersion is empty, matches toVersion, or no save files exist.
func BackupSaveGames(winePrefix, installationPath, fromVersion, toVersion string) (string, error) {
	if fromVersion == "" || toVersion == "" || fromVersion == toVersion {
		return "", nil
	}

	label := backupVersionLabel(fromVersion)
	if label == "" {
		label = "unknown"
	}

	source := SaveGamesPath(winePrefix)
	hasSaves, err := directoryHasFiles(source)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("check save games: %w", err)
	}
	if !hasSaves {
		return "", nil
	}

	stamp := time.Now().Format("20060102-150405")
	destRoot := filepath.Join(installationPath, "save-backups", label+"-"+stamp)
	dest := filepath.Join(destRoot, "SaveGames")
	if err := os.MkdirAll(destRoot, 0o755); err != nil {
		return "", fmt.Errorf("create save backup folder: %w", err)
	}
	if err := copyDirectory(source, dest); err != nil {
		_ = os.RemoveAll(destRoot)
		return "", fmt.Errorf("copy save games: %w", err)
	}
	return destRoot, nil
}

func backupVersionLabel(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return ""
	}
	version = strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', 0:
			return '_'
		default:
			return r
		}
	}, version)
	return filepath.Base(version)
}

func directoryHasFiles(path string) (bool, error) {
	found := false
	err := filepath.WalkDir(path, func(_ string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		found = true
		return fs.SkipAll
	})
	if err != nil {
		return false, err
	}
	return found, nil
}
