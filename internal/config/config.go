package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/huijiro/jirolauncher/internal/files"
	"github.com/huijiro/jirolauncher/internal/runner"
)

const (
	configDir     = "jirolauncher"
	configVersion = 1
)

type Configuration struct {
	Version          int    `json:"version"`
	InstallationPath string `json:"installation_path"`
	DownloadPath     string `json:"download_path"`
	WinePrefix       string `json:"wine_prefix"`
	Runner           Runner `json:"runner"`
}

type Runner struct {
	Kind       string `json:"kind"`
	Executable string `json:"executable"`
	Name       string `json:"name"`
	Version    string `json:"version"`
}

func New(logger *slog.Logger) (*Configuration, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err == nil {
		var configuration Configuration
		if err := json.Unmarshal(data, &configuration); err != nil {
			return nil, fmt.Errorf("decode configuration: %w", err)
		}
		if err := configuration.Validate(); err != nil {
			return nil, err
		}
		return &configuration, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read configuration: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("no user home directory", slog.Any("error", err))
		return nil, err
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	return &Configuration{
		Version:          configVersion,
		InstallationPath: files.GameFolderPath(homeDir),
		DownloadPath:     filepath.Join(cacheDir, configDir, "downloads"),
		WinePrefix:       filepath.Join(files.GameFolderPath(homeDir), "pfx"),
	}, nil
}

func Exists() (bool, error) {
	path, err := Path()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func Path() (string, error) {
	configHome, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configHome, configDir, "config.json"), nil
}

func (c *Configuration) SetRunner(discovered runner.Runner) {
	c.Runner = Runner{
		Kind:       string(discovered.Kind),
		Executable: discovered.Executable,
		Name:       discovered.Name,
		Version:    discovered.Version,
	}
}

func DefaultRunner(runners []runner.Runner) runner.Runner {
	best := runners[0]
	for _, candidate := range runners[1:] {
		if runnerRank(candidate) > runnerRank(best) {
			best = candidate
		}
	}
	return best
}

var versionPattern = regexp.MustCompile(`(\d+)(?:\.(\d+))?`)

func runnerRank(discovered runner.Runner) int {
	rank := 0
	if discovered.Kind == runner.RunnerProton {
		rank = 1_000_000
	}
	matches := versionPattern.FindStringSubmatch(discovered.Version)
	if len(matches) < 3 {
		return rank
	}
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	return rank + major*1_000 + minor
}

func (c Configuration) Validate() error {
	if c.Version != configVersion {
		return fmt.Errorf("unsupported configuration version: %d", c.Version)
	}
	if c.InstallationPath == "" || c.DownloadPath == "" || c.WinePrefix == "" {
		return errors.New("configuration contains empty required paths")
	}
	return nil
}

func (c Configuration) Save() error {
	if err := c.Validate(); err != nil {
		return err
	}
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	temporary, err := os.CreateTemp(filepath.Dir(path), ".config-*.tmp")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)

	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryName, path)
}
