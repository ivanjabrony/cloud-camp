package logger

import (
	"fmt"
	"log/slog"
	"os"
)

const (
	LogFormatText = "text"
	LogFormatJson = "json"
)

const (
	EnvTest = "test"
	EnvDev  = "dev"
	EnvProd = "prod"
)

type MyLogger struct {
	*slog.Logger
}

func New(env string, format string) *MyLogger {
	var loggerLevel slog.Level

	switch env {
	case EnvTest:
		loggerLevel = slog.LevelDebug
	case EnvDev:
		loggerLevel = slog.LevelDebug
	case EnvProd:
		loggerLevel = slog.LevelInfo
	}

	var logger *slog.Logger
	switch format {
	case LogFormatJson:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel}))
	case LogFormatText:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel}))
	default:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel}))
		logger.Warn(fmt.Sprintf("unsupported logging format %s, using default format instead", format))
	}

	return &MyLogger{logger}
}
