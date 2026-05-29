package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/dvl0p/linkctl/internal/store"
)

type Endpoint func(http.ResponseWriter, *http.Request) int

type Server struct {
	httpServer	*http.Server
	logger		*slog.Logger
	cancel		context.CancelFunc
	store		*store.Store
}

func NewServer(cancel context.CancelFunc, store *store.Store,
		httpPort int, logger *slog.Logger) *Server {

	serveMux := http.NewServeMux()

	httpServer := &http.Server{
		Addr: "127.0.0.1:" + strconv.Itoa(httpPort),
		Handler: serveMux,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 120 * time.Second,
	}

	s := &Server{
		httpServer: httpServer,
		logger: logger,
		cancel: cancel,
		store: store,
	}

	adaptor := loggerAdaptor(logger)

	serveMux.Handle("GET /healthz", adaptor(handlerHealth))
	/*
	serveMux.Handle("POST /v1/links", adaptor(s.handlerCreateLink))
	serveMux.Handle("GET /v1/links", adaptor(s.handlerListLinks))
	serveMux.Handle("GET /v1/links/{linkID}", adaptor(s.handlerGetLink))
	serveMux.Handle("GET /v1/links/url/{linkURL}", adaptor(s.handlerGetLinkFromURL))
	serveMux.Handle("PATCH /v1/links", adaptor(s.handlerUpdateLinks))
	serveMux.Handle("PUT /v1/links", adaptor(s.handlerUpdateLinks))
	*/

	return s
}

func (s *Server) Start() error {
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

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Debug(
		"shutting down server", slog.String("addr", s.httpServer.Addr),
	)
	return s.httpServer.Shutdown(ctx)
}

func loggerAdaptor(logger *slog.Logger) func(Endpoint) http.Handler {
	return func(endPoint Endpoint) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			status := endPoint(w, r)
			
			logger.Info("http request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", status),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

func handlerHealth(w http.ResponseWriter, 
		r *http.Request) int {
	code := http.StatusOK
	w.WriteHeader(code)
	return code
}
