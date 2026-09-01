package game

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

func Launch(ctx context.Context, logger *slog.Logger, installationPath, wineExecutable, winePrefix string) error {
	executablePath := filepath.Join(installationPath, "WindowsNoEditor", "VotV.exe")
	if _, err := os.Stat(executablePath); err != nil {
		return fmt.Errorf("find VotV.exe: %w", err)
	}

	runnerExecutable := wineExecutable
	arguments := []string{executablePath}
	environment := append(os.Environ(), "WINEPREFIX="+winePrefix)
	if protonExecutable, steamRoot, ok := protonScript(wineExecutable); ok {
		runnerExecutable = protonExecutable
		arguments = []string{"run", executablePath}
		environment = append(
			environment,
			"STEAM_COMPAT_DATA_PATH="+filepath.Dir(winePrefix),
			"STEAM_COMPAT_CLIENT_INSTALL_PATH="+steamRoot,
		)
	}

	command := exec.CommandContext(ctx, runnerExecutable, arguments...)
	command.Dir = filepath.Dir(executablePath)
	command.Env = environment
	processLog := slog.NewLogLogger(logger.Handler(), slog.LevelInfo)
	command.Stdout = processLog.Writer()
	command.Stderr = processLog.Writer()

	logger.Info("starting game process", slog.String("runner", runnerExecutable), slog.Any("arguments", arguments), slog.String("prefix", winePrefix), slog.String("executable", executablePath))
	if err := command.Start(); err != nil {
		logger.Error("could not start game process", slog.Any("error", err))
		return fmt.Errorf("launch VotV.exe: %w", err)
	}

	go func() {
		err := command.Wait()
		if err != nil {
			logger.Error("game process exited with an error", slog.Any("error", err))
			return
		}
		logger.Info("game process exited")
	}()
	return nil
}

func protonScript(wineExecutable string) (string, string, bool) {
	winePath := filepath.Clean(wineExecutable)
	if filepath.Base(filepath.Dir(winePath)) != "bin" || filepath.Base(filepath.Dir(filepath.Dir(winePath))) != "files" {
		return "", "", false
	}

	runnerRoot := filepath.Dir(filepath.Dir(filepath.Dir(winePath)))
	candidate := filepath.Join(runnerRoot, "proton")
	if _, err := os.Stat(candidate); err != nil {
		return "", "", false
	}

	containerDir := filepath.Dir(runnerRoot)
	if filepath.Base(containerDir) == "compatibilitytools.d" {
		return candidate, filepath.Dir(containerDir), true
	}
	if filepath.Base(containerDir) == "common" && filepath.Base(filepath.Dir(containerDir)) == "steamapps" {
		return candidate, filepath.Dir(filepath.Dir(containerDir)), true
	}
	return "", "", false
}
