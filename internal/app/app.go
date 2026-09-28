// Package app связывает компоненты сервиса воедино.
package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/themotka/netevents/internal/config"
	"github.com/themotka/netevents/internal/httpserver"
)

// App владеет компонентами сервиса и управляет их жизненным циклом.
type App struct {
	http *httpserver.Server
}

// New собирает приложение по конфигурации cfg.
func New(cfg config.Config, log *slog.Logger) *App {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleLiveness)
	mux.HandleFunc("GET /readyz", handleReadiness)

	httpLog := log.With(slog.String("component", "http"))
	handler := httpserver.Chain(mux,
		httpserver.RequestID,
		httpserver.AccessLog(httpLog),
		httpserver.Recover(httpLog),
	)

	return &App{
		http: httpserver.New(cfg.HTTP, httpLog, handler),
	}
}

// Run блокируется, пока не завершится ctx или не упадёт один из компонентов.
func (a *App) Run(ctx context.Context) error {
	return a.http.Run(ctx)
}

// handleLiveness сообщает, что процесс жив.
func handleLiveness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// handleReadiness сообщает, что сервис готов принимать трафик. Проверки
// зависимостей (Postgres, Kafka, ...) добавятся сюда на следующих этапах.
func handleReadiness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
