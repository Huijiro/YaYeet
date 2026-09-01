package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const logDir = "jirolauncher"

type consoleHandler struct {
	writer io.Writer
}

func (h *consoleHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *consoleHandler) Handle(_ context.Context, record slog.Record) error {
	executable := ""
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == "executable" && attr.Value.Kind() == slog.KindString {
			executable = attr.Value.String()
		}
		return true
	})

	if executable != "" {
		_, err := fmt.Fprintf(
			h.writer,
			"[%s]: %q executable=%s\n",
			strings.ToUpper(record.Level.String()),
			record.Message,
			executable,
		)
		return err
	}

	_, err := fmt.Fprintf(h.writer, "[%s]: %q\n", strings.ToUpper(record.Level.String()), record.Message)
	return err
}

func (h *consoleHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *consoleHandler) WithGroup(_ string) slog.Handler {
	return h
}

type dualHandler struct {
	console slog.Handler
	file    slog.Handler
}

func (h *dualHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.console.Enabled(ctx, level) || h.file.Enabled(ctx, level)
}

func (h *dualHandler) Handle(ctx context.Context, record slog.Record) error {
	if err := h.console.Handle(ctx, record); err != nil {
		return err
	}
	return h.file.Handle(ctx, record)
}

func (h *dualHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &dualHandler{
		console: h.console.WithAttrs(attrs),
		file:    h.file.WithAttrs(attrs),
	}
}

func (h *dualHandler) WithGroup(name string) slog.Handler {
	return &dualHandler{
		console: h.console.WithGroup(name),
		file:    h.file.WithGroup(name),
	}
}

func New() (*slog.Logger, *os.File, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, err
	}
	stateDir := filepath.Join(homeDir, ".local", "state", logDir)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return nil, nil, err
	}

	logFile, err := os.OpenFile(
		filepath.Join(stateDir, "jirolauncher.log"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0o644,
	)
	if err != nil {
		return nil, nil, err
	}

	fileHandler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(&dualHandler{
		console: &consoleHandler{writer: os.Stderr},
		file:    fileHandler,
	})

	return logger, logFile, nil
}
