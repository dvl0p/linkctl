package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
)

type closeLoggerFunc func() error

func initLogger() (*slog.Logger, closeLoggerFunc, error) {
	logPath := os.Getenv("LINKCTL_LOGPATH")
	if logPath == "" {
		logPath = "./log/linkctl.log"
	}

	_, isDebug := os.LookupEnv("LINKCTL_DEBUG")

	logFile, err := os.OpenFile(
		logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"could not open log file at path %v: %w", logPath, err,
		)
	}

	logFileBuff := bufio.NewWriterSize(logFile, 4096)

	CloseLogFile := func() error {
		if err := logFileBuff.Flush(); err != nil {
			return fmt.Errorf("could not flush log file buffer: %w", err)
		}
		if err := logFile.Close(); err != nil {
			return fmt.Errorf("could not close log file: %w", err)
		}
		return nil
	}

	fileHandler := slog.NewJSONHandler(logFileBuff, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	stderrLevel := slog.LevelError
	if isDebug {
		stderrLevel = slog.LevelDebug
	}

	errorHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: stderrLevel,
	})

	multiHandler := slog.NewMultiHandler(errorHandler, fileHandler)

	return slog.New(multiHandler), CloseLogFile, nil
}
