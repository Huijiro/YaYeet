package runner

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type RunnerKind string

const (
	RunnerWine   RunnerKind = "wine"
	RunnerProton RunnerKind = "proton"
)

type SourceKind string

const (
	SourcePATH   SourceKind = "PATH"
	SourceLutris SourceKind = "Lutris"
	SourceSteam  SourceKind = "Steam"
)

type Runner struct {
	Name       string
	Kind       RunnerKind
	Executable string
	Version    string
	Source     SourceKind
}

var WINE_RELIABLES = [4]string{"wine", "wine64", "wine-staging", "wine-development"}
var LUTRIS_FOLDERS = [2]string{
	filepath.Join(".local", "share", "lutris", "runners", "wine"),
	filepath.Join(".var", "app", "net.lutris.Lutris", "data", "lutris", "runners", "wine"),
}

var STEAM_FOLDERS = [8]string{
	filepath.Join(".steam", "root", "compatibilitytools.d"),
	filepath.Join(".steam", "steam", "compatibilitytools.d"),
	filepath.Join(".local", "share", "Steam", "compatibilitytools.d"),
	filepath.Join(".steam", "root", "steamapps", "common"),
	filepath.Join(".steam", "steam", "steamapps", "common"),
	filepath.Join(".local", "share", "Steam", "steamapps", "common"),
	filepath.Join(".var", "app", "com.valvesoftware.Steam", "data", "Steam", "compatibilitytools.d"),
	filepath.Join(".var", "app", "com.valvesoftware.Steam", "data", "Steam", "steamapps", "common"),
}

func lutrisRunners(logger *slog.Logger) ([]Runner, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var lutrisRunners []Runner
	for _, lutrisFolder := range LUTRIS_FOLDERS {
		lutrisRunnerPath := filepath.Join(homeDir, lutrisFolder)
		pathInfo, err := os.Stat(lutrisRunnerPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			logger.Warn("cannot inspect Lutris runner directory", slog.String("path", lutrisRunnerPath), slog.Any("error", err))
			continue
		}
		if !pathInfo.IsDir() {
			logger.Warn("Lutris runner path is not a directory", slog.String("path", lutrisRunnerPath))
			continue
		}

		runnerDirs, err := os.ReadDir(lutrisRunnerPath)
		if err != nil {
			logger.Warn("cannot read Lutris runner directory", slog.String("path", lutrisRunnerPath), slog.Any("error", err))
			continue
		}

		for _, runnerDir := range runnerDirs {
			if !runnerDir.IsDir() {
				continue
			}

			runnerPath := filepath.Join(lutrisRunnerPath, runnerDir.Name(), "bin", "wine")
			runnerInfo, err := os.Stat(runnerPath)
			if err != nil || runnerInfo.IsDir() {
				continue
			}

			version, err := runnerVersion(runnerPath)
			if err != nil {
				logger.Warn("cannot query Lutris runner version", slog.String("path", runnerPath), slog.Any("error", err))
				continue
			}

			lutrisRunners = append(lutrisRunners, Runner{
				Name:       runnerDir.Name(),
				Kind:       RunnerWine,
				Executable: runnerPath,
				Version:    version,
				Source:     SourceLutris,
			})
		}
	}

	return lutrisRunners, nil
}

func SteamRunners(logger *slog.Logger) ([]Runner, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var steamRunners []Runner
	for _, steamFolder := range STEAM_FOLDERS {
		steamRunnerPath := filepath.Join(homeDir, steamFolder)
		entries, err := os.ReadDir(steamRunnerPath)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				logger.Warn("cannot read Steam runner directory", slog.String("path", steamRunnerPath), slog.Any("error", err))
			}
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			runnerPath := filepath.Join(steamRunnerPath, entry.Name(), "files", "bin", "wine")
			runnerInfo, err := os.Stat(runnerPath)
			if err != nil || runnerInfo.IsDir() {
				continue
			}

			version, err := runnerVersion(runnerPath)
			if err != nil {
				logger.Warn("cannot query Proton runner version", slog.String("path", runnerPath), slog.Any("error", err))
				continue
			}

			steamRunners = append(steamRunners, Runner{
				Name:       entry.Name(),
				Kind:       RunnerProton,
				Executable: runnerPath,
				Version:    version,
				Source:     SourceSteam,
			})
		}
	}

	return steamRunners, nil
}

func runnerVersion(path string) (string, error) {
	output, err := exec.Command(path, "--version").Output()
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(output))
	if version == "" {
		return "", errors.New("runner returned an empty version")
	}
	return version, nil
}

func List(logger *slog.Logger) ([]Runner, error) {
	var runners []Runner
	seen := make(map[string]struct{})

	addRunner := func(runner Runner) {
		canonicalPath, err := filepath.EvalSymlinks(runner.Executable)
		if err == nil {
			runner.Executable = canonicalPath
		}

		if _, exists := seen[runner.Executable]; exists {
			return
		}
		seen[runner.Executable] = struct{}{}
		runners = append(runners, runner)
	}

	for _, name := range WINE_RELIABLES {
		path, err := exec.LookPath(name)
		if err != nil {
			logger.Debug("runner not found in PATH", slog.String("name", name))
			continue
		}

		version, err := runnerVersion(path)
		if err != nil {
			logger.Warn("cannot query PATH runner version", slog.String("path", path), slog.Any("error", err))
			continue
		}

		addRunner(Runner{
			Name:       name,
			Kind:       RunnerWine,
			Executable: path,
			Version:    version,
			Source:     SourcePATH,
		})
	}

	lutrisRunners, err := lutrisRunners(logger)
	if err != nil {
		logger.Warn("cannot scan Lutris runners", slog.Any("error", err))
	} else {
		for _, runner := range lutrisRunners {
			addRunner(runner)
		}
	}

	protonRunners, err := SteamRunners(logger)
	if err != nil {
		logger.Warn("cannot scan Steam runners", slog.Any("error", err))
	} else {
		for _, runner := range protonRunners {
			addRunner(runner)
		}
	}

	if len(runners) == 0 {
		return nil, fmt.Errorf("no Wine runners found")
	}
	return runners, nil
}
