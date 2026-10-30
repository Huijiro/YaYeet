package game

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestVersionHintFromArchiveName(t *testing.T) {
	tests := map[string]string{
		"/tmp/0100_0009.7z":   "0100_0009",
		"0100_0009.7z":        "0100_0009",
		"game.zip":            "game",
		"/tmp/weird/name":     "name",
		"/tmp/nested/path.7z": "path",
	}
	for input, expected := range tests {
		if actual := VersionHintFromArchiveName(input); actual != expected {
			t.Errorf("VersionHintFromArchiveName(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestValidateVersionNameRejectsPaths(t *testing.T) {
	if err := ValidateVersionName("0100_0009"); err != nil {
		t.Fatalf("ValidateVersionName(valid) = %v", err)
	}
	for _, version := range []string{"", "../0100_0009", "foo/bar", `foo\bar`} {
		if err := ValidateVersionName(version); err == nil {
			t.Fatalf("ValidateVersionName(%q) = nil, want error", version)
		}
	}
}

func TestFindWindowsNoEditor(t *testing.T) {
	root := t.TempDir()
	gameRoot := filepath.Join(root, "0100_0009", "WindowsNoEditor")
	if err := os.MkdirAll(gameRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gameRoot, "VotV.exe"), []byte("exe"), 0o644); err != nil {
		t.Fatal(err)
	}

	found, err := findWindowsNoEditor(root)
	if err != nil {
		t.Fatal(err)
	}
	if found != gameRoot {
		t.Fatalf("findWindowsNoEditor() = %q, want %q", found, gameRoot)
	}
}

func TestInstallFromArchivePromotesExtractedGame(t *testing.T) {
	previous := extractSeven
	defer func() { extractSeven = previous }()

	extractSeven = func(ctx context.Context, archivePath, dest string, report ProgressFunc) error {
		gameRoot := filepath.Join(dest, "0100_0009", "WindowsNoEditor")
		if err := os.MkdirAll(filepath.Join(gameRoot, "VotV", "Content", "Paks"), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(gameRoot, "VotV.exe"), []byte("exe"), 0o644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(gameRoot, "VotV", "Content", "Paks", "VotV-WindowsNoEditor.pak"), []byte("pak"), 0o644)
	}

	installDir := t.TempDir()
	oldGame := filepath.Join(installDir, "WindowsNoEditor")
	if err := os.MkdirAll(oldGame, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldGame, "stale.txt"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := InstallFromArchive(context.Background(), filepath.Join(t.TempDir(), "0100_0009.7z"), installDir, "0100_0009", nil)
	if err == nil {
		t.Fatal("InstallFromArchive() missing archive should fail before extract override path")
	}

	archive := filepath.Join(t.TempDir(), "0100_0009.7z")
	if err := os.WriteFile(archive, []byte("archive"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := InstallFromArchive(context.Background(), archive, installDir, "0100_0009", nil); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(installDir, "WindowsNoEditor", "VotV.exe")); err != nil {
		t.Fatalf("installed executable missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(installDir, "WindowsNoEditor", "stale.txt")); !os.IsNotExist(err) {
		t.Fatalf("stale game files were not replaced: %v", err)
	}
	marked, err := readInstalledVersion(installDir)
	if err != nil {
		t.Fatal(err)
	}
	if marked != "0100_0009" {
		t.Fatalf("readInstalledVersion() = %q, want 0100_0009", marked)
	}
}

func TestParseSevenZipPercent(t *testing.T) {
	percent, ok := parseSevenZipPercent(" 42%")
	if !ok || percent != 42 {
		t.Fatalf("parseSevenZipPercent() = %d, %v", percent, ok)
	}
	if _, ok := parseSevenZipPercent("Extracting"); ok {
		t.Fatal("parseSevenZipPercent() accepted non-percent line")
	}
}
