package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Inforberi/go-template/internal/infra/config"
	"github.com/Inforberi/go-template/internal/infra/logger"
	"github.com/Inforberi/go-template/internal/transport/router"
	"github.com/Inforberi/go-template/internal/transport/server"
)

func New() error {
	// background context
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// config
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	// logger
	log, err := logger.New(cfg)
	if err != nil {
		return fmt.Errorf("Logger init: %w", err)
	}

	// router
	r := router.New(log)

	if err := server.Run(ctx, cfg, r, log); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}
