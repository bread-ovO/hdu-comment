package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hdu-dp/backend/internal/config"
)

// HTTPServer wraps the application's HTTP listener and graceful shutdown policy.
type HTTPServer struct {
	server          *http.Server
	logger          *slog.Logger
	shutdownTimeout time.Duration
}

// New creates a configured HTTP server.
func New(cfg *config.Config, engine *gin.Engine, logger *slog.Logger) *HTTPServer {
	return &HTTPServer{
		server: &http.Server{
			Addr:              ":" + cfg.Server.Port,
			Handler:           engine,
			ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
			ReadTimeout:       cfg.Server.ReadTimeout,
			WriteTimeout:      cfg.Server.WriteTimeout,
			IdleTimeout:       cfg.Server.IdleTimeout,
		},
		logger:          logger,
		shutdownTimeout: cfg.Server.ShutdownTimeout,
	}
}

// Run starts the HTTP server and blocks until shutdown completes or an error occurs.
func (s *HTTPServer) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.logger.Info("http server listening", slog.String("addr", s.server.Addr))
		err := s.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen and serve: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.logger.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeoutDuration())
		defer cancel()

		if err := s.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}

		if err := <-errCh; err != nil {
			return err
		}

		s.logger.Info("http server stopped")
		return nil
	}
}

func (s *HTTPServer) shutdownTimeoutDuration() time.Duration {
	return s.shutdownTimeout
}
