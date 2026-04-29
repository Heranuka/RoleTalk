// Package http provides the HTTP transport layer for the Backforge API.
// It contains the router, middleware, handlers, and HTTP server lifecycle
// management responsible for serving the API over HTTP.
package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"go-backend/internal/config"
)

// Server wraps http.Server and manages lifecycle.
type Server struct {
	cfg *config.Config
	log *zap.SugaredLogger

	router http.Handler
	srv    *http.Server
}

// NewServer creates a new HTTP server instance.
func NewServer(
	cfg *config.Config,
	log *zap.SugaredLogger,
	router http.Handler,
) *Server {
	return &Server{
		cfg:    cfg,
		log:    log,
		router: router,
	}
}

// Run starts the HTTP server and blocks until context cancellation or fatal error.
func (s *Server) Run(ctx context.Context) error {
	addr := normalizeAddr(s.cfg.HTTP.Port)

	s.srv = &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadTimeout:       s.cfg.HTTP.ReadTimeout,
		WriteTimeout:      s.cfg.HTTP.WriteTimeout,
		IdleTimeout:       s.cfg.HTTP.IdleTimeout,
		ReadHeaderTimeout: 5 * time.Second,

		// security hardening
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s.log.Infow("http server starting",
		"addr", addr,
		"read_timeout", s.cfg.HTTP.ReadTimeout,
		"write_timeout", s.cfg.HTTP.WriteTimeout,
		"idle_timeout", s.cfg.HTTP.IdleTimeout,
	)

	errCh := make(chan error, 1)

	go func() {
		err := s.srv.Serve(ln)
		if !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():

		s.log.Info("http server shutdown initiated")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			s.cfg.HTTP.ShutdownTimeout,
		)
		defer cancel()

		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			s.log.Errorw("http server graceful shutdown failed", "error", err)

			if closeErr := s.srv.Close(); closeErr != nil {
				s.log.Errorw("http server forced close failed", "error", closeErr)
			}

			return err
		}

		s.log.Info("http server stopped gracefully")

		return nil

	case err := <-errCh:

		if err != nil {
			return fmt.Errorf("http server crashed: %w", err)
		}

		return nil
	}
}

// Shutdown stops the server gracefully.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}

	s.log.Info("http server shutting down")

	return s.srv.Shutdown(ctx)
}

func normalizeAddr(port string) string {
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}
