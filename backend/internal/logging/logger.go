package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/hdu-dp/backend/internal/config"
)

// New builds the application logger from config.
func New(cfg *config.Config) (*slog.Logger, error) {
	level, err := parseLevel(cfg.Log.Level)
	if err != nil {
		return nil, err
	}

	options := &slog.HandlerOptions{Level: level}
	format := strings.ToLower(strings.TrimSpace(cfg.Log.Format))

	switch format {
	case "", "text":
		return slog.New(slog.NewTextHandler(os.Stdout, options)), nil
	case "json":
		return slog.New(slog.NewJSONHandler(os.Stdout, options)), nil
	default:
		return nil, fmt.Errorf("unsupported log format: %s", cfg.Log.Format)
	}
}

func parseLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unsupported log level: %s", value)
	}
}
