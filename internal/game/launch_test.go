package game

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestLaunchWaitsForGameProcess(t *testing.T) {
	root := t.TempDir()
	gameDir := filepath.Join(root, "WindowsNoEditor")
	if err := os.MkdirAll(gameDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gameDir, "VotV.exe"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	markerPath := filepath.Join(root, "finished")
	runnerPath := filepath.Join(root, "runner")
	runner := "#!/bin/sh\nsleep 0.1\nprintf finished > \"$YAYEET_TEST_MARKER\"\n"
	if err := os.WriteFile(runnerPath, []byte(runner), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("YAYEET_TEST_MARKER", markerPath)

	if err := Launch(context.Background(), slog.Default(), root, runnerPath, filepath.Join(root, "prefix"), ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(markerPath); err != nil {
		t.Fatalf("game process had not finished when Launch returned: %v", err)
	}
}

func TestLaunchPassesCustomArgumentsAfterExecutable(t *testing.T) {
	root := t.TempDir()
	gameDir := filepath.Join(root, "WindowsNoEditor")
	if err := os.MkdirAll(gameDir, 0o755); err != nil {
		t.Fatal(err)
	}
	executablePath := filepath.Join(gameDir, "VotV.exe")
	if err := os.WriteFile(executablePath, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	argumentsPath := filepath.Join(root, "arguments")
	runnerPath := filepath.Join(root, "runner")
	runner := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$YAYEET_TEST_ARGUMENTS\"\n"
	if err := os.WriteFile(runnerPath, []byte(runner), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("YAYEET_TEST_ARGUMENTS", argumentsPath)

	if err := Launch(context.Background(), slog.Default(), root, runnerPath, filepath.Join(root, "prefix"), "-cefdisablegpu -windowed"); err != nil {
		t.Fatal(err)
	}
	arguments, err := os.ReadFile(argumentsPath)
	if err != nil {
		t.Fatal(err)
	}
	want := executablePath + "\n-cefdisablegpu\n-windowed\n"
	if string(arguments) != want {
		t.Fatalf("arguments = %q, want %q", arguments, want)
	}
}

func TestEnsureSteamCompatDataPathUsesParentWhenNamedPfx(t *testing.T) {
	root := t.TempDir()
	prefix := filepath.Join(root, "pfx")
	compatPath, err := ensureSteamCompatDataPath(prefix)
	if err != nil {
		t.Fatal(err)
	}
	if compatPath != root {
		t.Fatalf("compat path = %q, want %q", compatPath, root)
	}
}

func TestEnsureSteamCompatDataPathSymlinksCustomPrefix(t *testing.T) {
	root := t.TempDir()
	prefix := filepath.Join(root, "pfx-test")
	if err := os.MkdirAll(prefix, 0o755); err != nil {
		t.Fatal(err)
	}

	compatPath, err := ensureSteamCompatDataPath(prefix)
	if err != nil {
		t.Fatal(err)
	}
	wantCompat := prefix + ".steamcompat"
	if compatPath != wantCompat {
		t.Fatalf("compat path = %q, want %q", compatPath, wantCompat)
	}

	link := filepath.Join(compatPath, "pfx")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	wantTarget, err := filepath.Abs(prefix)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Clean(target) != filepath.Clean(wantTarget) {
		t.Fatalf("pfx link = %q, want %q", target, wantTarget)
	}

	// Second call should be idempotent and leave the existing symlink alone.
	compatPathAgain, err := ensureSteamCompatDataPath(prefix)
	if err != nil {
		t.Fatal(err)
	}
	if compatPathAgain != compatPath {
		t.Fatalf("second compat path = %q, want %q", compatPathAgain, compatPath)
	}
}

func TestEnsureSteamCompatDataPathRejectsExistingNonSymlink(t *testing.T) {
	root := t.TempDir()
	prefix := filepath.Join(root, "pfx-test")
	compatDir := prefix + ".steamcompat"
	if err := os.MkdirAll(filepath.Join(compatDir, "pfx"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(prefix, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := ensureSteamCompatDataPath(prefix); err == nil {
		t.Fatal("expected error for existing non-symlink pfx path")
	}
}
