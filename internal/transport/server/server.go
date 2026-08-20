package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Inforberi/go-template/internal/infra/config"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config, handler http.Handler, log *zap.Logger) error {
	srv := http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	srvErr := make(chan error, 1)

	go func() {
		log.Info("Start server", zap.String("PORT", cfg.Port))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
			return
		}

		srvErr <- nil
	}()

	select {
	case err := <-srvErr:
		if err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	shutdownctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownctx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	log.Info("http server stopped gracefully")

	return nil
}
