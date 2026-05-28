package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Endpoint func(http.ResponseWriter, *http.Request) (int, error)

type server struct {
	httpServer	*http.Server
	logger		*slog.Logger
	cancel		context.CancelFunc
}

func newServer(cancel context.CancelFunc,
		httpPort int, logger *slog.Logger) *server {

	serveMux := http.NewServeMux()

	httpServer := &http.Server{
		Addr: "127.0.0.1:" + strconv.Itoa(httpPort),
		Handler: serveMux,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 120 * time.Second,
	}

	s := &server{
		httpServer: httpServer,
		logger: logger,
		cancel: cancel,
	}

	adaptor := loggerAdaptor(logger)

	serveMux.Handle("GET /healthz", adaptor(s.handlerHealth))

	return s
}

func (s *server) start() error {
	s.logger.Debug(
		"opening listening socket",
		slog.String("addr", s.httpServer.Addr),
	)
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("could not bind tcp socket: %w", err)
	}

	s.logger.Debug(
		"initializing server",
		slog.String("addr", s.httpServer.Addr),
	)
	if err := s.httpServer.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("could not serve on tcp socket: %w", err)
	}
	return nil
}

func (s *server) shutdown(ctx context.Context) error {
	s.logger.Debug(
		"shutting down server", slog.String("addr", s.httpServer.Addr),
	)
	return s.httpServer.Shutdown(ctx)
}

func loggerAdaptor(logger *slog.Logger) func(Endpoint) http.Handler {
	return func(endPoint Endpoint) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			status, err := endPoint(w, r)
			if err != nil {
				http.Error(w, err.Error(), status)
			}
			
			logger.Info("http request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", status),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

func (s *server) handlerHealth(w http.ResponseWriter, 
		r *http.Request) (int, error) {
	code := http.StatusOK
	w.WriteHeader(code)
	return code, nil
}
