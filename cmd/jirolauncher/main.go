package main

import (
	"log/slog"

	"github.com/huijiro/jirolauncher/internal/config"
	"github.com/huijiro/jirolauncher/internal/logging"
	"github.com/huijiro/jirolauncher/internal/runner"
	"github.com/huijiro/jirolauncher/internal/ui"
)

func main() {
	logger, logFile, err := logging.New()
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	configuration, err := config.New(logger)
	if err != nil {
		logger.Error("could not initialize configuration", slog.Any("error", err))
		return
	}

	logger.Info("configuration initialized", slog.String("game_folder", configuration.InstallationPath), slog.String("wine_prefix", configuration.WinePrefix))

	runners, err := runner.List(logger)
	if err != nil {
		logger.Error("could not discover Wine runners", slog.Any("error", err))
	} else {
		for _, discovered := range runners {
			logger.Info("Wine runner discovered",
				slog.String("name", discovered.Name),
				slog.String("kind", string(discovered.Kind)),
				slog.String("version", discovered.Version),
				slog.String("executable", discovered.Executable),
				slog.String("source", string(discovered.Source)))
		}
	}

	ui.Run(logger, configuration, runners)
}
