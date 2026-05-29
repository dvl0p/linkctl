package main

import (
	"log/slog"
	"os"
)

func initLogger() (*slog.Logger, error) {

	_, isDebug := os.LookupEnv("LINKCTL_DEBUG")

	level := slog.LevelInfo
	if isDebug {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler), nil
}
