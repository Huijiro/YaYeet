package game

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Launch(ctx context.Context, logger *slog.Logger, installationPath, wineExecutable, winePrefix, customArguments string) error {
	executablePath := filepath.Join(installationPath, "WindowsNoEditor", "VotV.exe")
	if _, err := os.Stat(executablePath); err != nil {
		return fmt.Errorf("find VotV.exe: %w", err)
	}

	runnerExecutable := wineExecutable
	gameArguments := strings.Fields(customArguments)
	arguments := append([]string{executablePath}, gameArguments...)
	environment := append(os.Environ(), "WINEPREFIX="+winePrefix)
	if protonExecutable, steamRoot, ok := protonScript(wineExecutable); ok {
		runnerExecutable = protonExecutable
		arguments = append([]string{"run", executablePath}, gameArguments...)
		compatDataPath, err := ensureSteamCompatDataPath(winePrefix)
		if err != nil {
			return fmt.Errorf("prepare proton prefix: %w", err)
		}
		environment = append(
			environment,
			"STEAM_COMPAT_DATA_PATH="+compatDataPath,
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

	if err := command.Wait(); err != nil {
		logger.Error("game process exited with an error", slog.Any("error", err))
		return fmt.Errorf("wait for game process: %w", err)
	}
	logger.Info("game process exited")
	return nil
}

// ensureSteamCompatDataPath returns the STEAM_COMPAT_DATA_PATH for winePrefix.
// Proton always uses $STEAM_COMPAT_DATA_PATH/pfx, so a configured prefix not named
// "pfx" gets a sibling .steamcompat directory with a pfx symlink to the real prefix.
func ensureSteamCompatDataPath(winePrefix string) (string, error) {
	winePrefix = filepath.Clean(winePrefix)
	if filepath.Base(winePrefix) == "pfx" {
		return filepath.Dir(winePrefix), nil
	}

	compatDir := winePrefix + ".steamcompat"
	pfxLink := filepath.Join(compatDir, "pfx")
	if err := os.MkdirAll(compatDir, 0o755); err != nil {
		return "", fmt.Errorf("create steam compat data path: %w", err)
	}

	target, err := filepath.Abs(winePrefix)
	if err != nil {
		return "", fmt.Errorf("resolve wine prefix: %w", err)
	}
	if err := ensureSymlink(pfxLink, target); err != nil {
		return "", err
	}
	return compatDir, nil
}

func ensureSymlink(link, target string) error {
	info, err := os.Lstat(link)
	if err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("steam compat path %s exists and is not a symlink to %s", link, target)
		}
		current, readErr := os.Readlink(link)
		if readErr != nil {
			return fmt.Errorf("read steam compat pfx link: %w", readErr)
		}
		if filepath.Clean(current) == filepath.Clean(target) {
			return nil
		}
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("replace steam compat pfx link: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat steam compat pfx link: %w", err)
	}

	if err := os.Symlink(target, link); err != nil {
		return fmt.Errorf("link steam compat pfx: %w", err)
	}
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
