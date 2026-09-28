// Package main is the entry point for the netevents service.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/themotka/netevents/internal/app"
	"github.com/themotka/netevents/internal/config"
	"github.com/themotka/netevents/internal/logger"
)

// Задаются при сборке через -ldflags "-X main.version=... -X main.commit=...".
var (
	version = "dev"
	commit  = "none"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "netevents: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(os.Stdout, cfg.Log.Level, cfg.Log.Format).With(
		slog.String("service", "netevents"),
		slog.String("version", version),
	)
	// Библиотеки, пишущие через дефолтный slog-логгер, получат тот же
	// формат и уровень. Наш собственный код получает логгер явно.
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info("starting service", slog.String("commit", commit))

	if err := app.New(cfg, log).Run(ctx); err != nil {
		log.Error("service stopped with error", logger.Err(err))
		return err
	}

	log.Info("service stopped")
	return nil
}
