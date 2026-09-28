// Package config загружает конфигурацию сервиса из переменных окружения.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// Config — полная конфигурация сервиса.
type Config struct {
	Log  Log
	HTTP HTTP
}

// Log — настройки корневого логгера.
type Log struct {
	Level  slog.Level
	Format string // "json" или "text"
}

// HTTP — настройки HTTP-сервера.
type HTTP struct {
	Addr              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

// Load читает конфигурацию из окружения, подставляя значения по умолчанию.
// Обо всех некорректных значениях сообщается сразу, одной ошибкой.
func Load() (Config, error) {
	var l loader

	cfg := Config{
		Log: Log{
			Level:  l.level("LOG_LEVEL", slog.LevelInfo),
			Format: l.string("LOG_FORMAT", "json"),
		},
		HTTP: HTTP{
			Addr:              l.string("HTTP_ADDR", ":8080"),
			ReadHeaderTimeout: l.duration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
			ReadTimeout:       l.duration("HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:      l.duration("HTTP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:       l.duration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout:   l.duration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
	}

	if cfg.Log.Format != "json" && cfg.Log.Format != "text" {
		l.errs = append(l.errs,
			fmt.Errorf("LOG_FORMAT: unsupported value %q", cfg.Log.Format))
	}

	if err := errors.Join(l.errs...); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// loader накапливает ошибки разбора, чтобы вернуть их все вместе.
type loader struct {
	errs []error
}

func (l *loader) string(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func (l *loader) duration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s: %w", key, err))
		return def
	}
	return d
}

func (l *loader) level(key string, def slog.Level) slog.Level {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(v)); err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s: %w", key, err))
		return def
	}
	return lvl
}
