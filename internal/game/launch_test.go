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

	if err := Launch(context.Background(), slog.Default(), root, runnerPath, filepath.Join(root, "prefix")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(markerPath); err != nil {
		t.Fatalf("game process had not finished when Launch returned: %v", err)
	}
}
