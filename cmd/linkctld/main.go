package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dvl0p/linkctl/internal/store"
	"github.com/dvl0p/linkctl/internal/api"
)

func main() {

	ctx, cancel := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM,
	)

	var httpPort int
	var dbArg string

	flag.IntVar(&httpPort, "port", 8080, "port to bind server")
	flag.StringVar(&dbArg, "data", "./data/linkctl.db", 
		"database file or directory (env overwrites flag)",
	)

	flag.Parse()

	status := run(ctx, cancel, httpPort, dbArg)

	cancel()
	os.Exit(status)
}

func run(ctx context.Context, cancel context.CancelFunc,
		httpPort int, dbArg string) int {
	
	logger, err := initLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing logger: %v", err)
		return 1
	}

	dbPath := getDBPath(dbArg, logger)

	logger.Info("welcome to linkctld 🗿",
		slog.Int("port", httpPort),
		slog.String("db_path", dbPath),
	)

	logger.Debug("initializing store")
	store, err := store.New(dbPath)
	if err != nil {
		logger.Error("could not initialize db",
			slog.String("error", err.Error()),
		)
		return 1
	}
	defer store.Close()

	server := api.NewServer(cancel, store, httpPort, logger)

	servErr := make(chan error, 1)
	go func() {
		servErr <- server.Start()
	}()

	select {
	case err := <- servErr:
		logger.Error("could not run server",
			slog.String("error", err.Error()),
		)
		return 1
	case <- ctx.Done():
		logger.Info("server shutdown signal received")
	}

	ctxShutdown, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		logger.Error("could not stop server",
			slog.String("error", err.Error()),
		)
		return 1
	}
	return 0
}

func getDBPath(dbArg string, logger *slog.Logger) string {
	dbEnv, set := os.LookupEnv("LINKCTL_DBPATH")
	if set {
		logger.Debug("getting db path from env", slog.String("db_path", dbEnv))
		return dbEnv
	}

	logger.Debug("setting db path from flag", slog.String("db_arg", dbArg))
	if strings.HasSuffix(dbArg, ".db") {
		return dbArg
	}
	return strings.TrimSuffix(dbArg, "/") + "/linkctl.db"
}
