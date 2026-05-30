package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dvl0p/linkctl/internal/api"
	"github.com/dvl0p/linkctl/internal/daemon"
	"github.com/dvl0p/linkctl/internal/store"
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

	daemon := daemon.New(store, logger)

	daemonCtx, daemonCancel := context.WithCancel(context.Background())
	defer daemonCancel()

	returnCode := 0
	var wg sync.WaitGroup
	daemonErr := make(chan error, 1)
	wg.Go(func() {
		daemonErr <- daemon.Start(daemonCtx)
	})

	server := api.NewServer(cancel, store, daemon, httpPort, logger)

	serverErr := make(chan error, 1)
	wg.Go(func() {
		serverErr <- server.Start()
	})

	select {
	case err := <-daemonErr:
		if err != nil {
			logger.Error("could not start daemon",
				slog.String("error", err.Error()),
			)
			returnCode = 1
		}
	case err := <-serverErr:
		if err != nil {
			logger.Error("could not run server",
				slog.String("error", err.Error()),
			)
			returnCode = 1
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	ctxShutdown, cancelShutdown := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancelShutdown()

	err = server.Shutdown(ctxShutdown)
	if err != nil {
		logger.Error("could not stop server",
			slog.String("error", err.Error()),
		)
		returnCode = 1
	}

	daemon.CloseQueue()
	wg.Wait()

	select {
	case err := <-serverErr:
		if err != nil {
			logger.Error("server exited with error",
				slog.String("error", err.Error()),
			)
			returnCode = 1
		}
	default:
	}

	select {
	case err := <-daemonErr:
		if err != nil {
			logger.Error("daemon exited with error",
				slog.String("error", err.Error()),
			)
			returnCode = 1
		}
	default:
	}

	return returnCode
}

func getDBPath(dbArg string, logger *slog.Logger) string {
	dbEnv, set := os.LookupEnv("LINKCTL_DBPATH")
	if set {
		logger.Debug("getting db path from env",
			slog.String("db_path", dbEnv),
		)
		return dbEnv
	}

	logger.Debug("setting db path from flag",
		slog.String("db_arg", dbArg),
	)
	if strings.HasSuffix(dbArg, ".db") {
		return dbArg
	}
	return strings.TrimSuffix(dbArg, "/") + "/linkctl.db"
}
