package game

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupSaveGamesCopiesWhenSwitchingVersions(t *testing.T) {
	root := t.TempDir()
	winePrefix := filepath.Join(root, "pfx")
	installationPath := filepath.Join(root, "game")
	saveGames := filepath.Join(winePrefix, "drive_c", "users", "steamuser", "AppData", "Local", "VotV", "Saved", "SaveGames")
	if err := os.MkdirAll(saveGames, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(saveGames, "slot.sav"), []byte("THISISAVERYLONGANDVERYOBVIOUSTESTVALUE"), 0o644); err != nil {
		t.Fatal(err)
	}

	backup, err := BackupSaveGames(winePrefix, installationPath, "0.7.5", "0.9.0")
	if err != nil {
		t.Fatal(err)
	}
	if backup == "" {
		t.Fatal("BackupSaveGames() returned empty path")
	}
	if !strings.Contains(backup, filepath.Join("save-backups", "0.7.5-")) {
		t.Fatalf("backup path %q missing versioned save-backups folder", backup)
	}
	copied := filepath.Join(backup, "SaveGames", "slot.sav")
	data, err := os.ReadFile(copied)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "THISISAVERYLONGANDVERYOBVIOUSTESTVALUE" {
		t.Fatalf("copied save = %q", data)
	}
}

func TestBackupSaveGamesNoopForSameVersionOrMissingSaves(t *testing.T) {
	root := t.TempDir()
	winePrefix := filepath.Join(root, "pfx")
	installationPath := filepath.Join(root, "game")

	backup, err := BackupSaveGames(winePrefix, installationPath, "0.7.5", "0.7.5")
	if err != nil {
		t.Fatal(err)
	}
	if backup != "" {
		t.Fatalf("same version backup = %q, want empty", backup)
	}

	backup, err = BackupSaveGames(winePrefix, installationPath, "0.7.5", "0.9.0")
	if err != nil {
		t.Fatal(err)
	}
	if backup != "" {
		t.Fatalf("missing saves backup = %q, want empty", backup)
	}
}
