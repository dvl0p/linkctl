package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	ctx, cancel := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGINT,
	)

	var httpPort int
	var dataDir string

	flag.IntVar(&httpPort, "port", 8080, "port to bind server")
	flag.StringVar(&dataDir, "data", "./data", "database directory")

	flag.Parse()

	status := run(ctx, cancel, httpPort, dataDir)

	cancel()
	os.Exit(status)
}

func run(ctx context.Context, cancel context.CancelFunc,
		httpPort int, dataDir string) int {
	
	logger, err := initLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing logger: %v", err)
		return 1
	}

	logger.Info("Welcome to linkctld 🗿",
		slog.Int("port", httpPort),
		slog.String("data_dir", dataDir),
	)

	server := newServer(cancel, httpPort, logger)

	var	servErr error
	go func() {
		servErr = server.start()
	}()

	<- ctx.Done()

	ctxShutdown, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.shutdown(ctxShutdown); err != nil {
		logger.Error("could not stop server",
			slog.String("error", err.Error()),
		)
		return 1
	}
	if servErr != nil {
		logger.Error("could not start server",
			slog.String("error", servErr.Error()),
		)
		return 1
	}
	return 0
}
