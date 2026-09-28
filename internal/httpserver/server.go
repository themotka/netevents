// Package httpserver — обёртка над net/http с корректной (graceful) остановкой.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/themotka/netevents/internal/config"
)

// Server — HTTP-сервер, который корректно останавливается по завершении контекста.
type Server struct {
	srv             *http.Server
	log             *slog.Logger
	shutdownTimeout time.Duration
}

// New создаёт сервер для обработчика h с настройками cfg.
func New(cfg config.HTTP, log *slog.Logger, h http.Handler) *Server {
	return &Server{
		srv: &http.Server{
			Addr:              cfg.Addr,
			Handler:           h,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
			ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
		},
		log:             log,
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

// Run обслуживает запросы до завершения ctx, затем останавливается,
// давая запросам в обработке завершиться в пределах shutdown-таймаута.
func (s *Server) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.srv.Addr, err)
	}
	s.log.Info("http server started", slog.String("addr", ln.Addr().String()))

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- s.srv.Serve(ln)
	}()

	select {
	case err := <-serveErr:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}

	s.log.Info("shutting down http server")

	shutdownCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx), s.shutdownTimeout)
	defer cancel()

	if err = s.srv.Shutdown(shutdownCtx); err != nil {
		err = s.srv.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("shutdown: %w", err)
	}
	if err = <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}
