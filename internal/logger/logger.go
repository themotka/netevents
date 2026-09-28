// Package logger строит логгер сервиса поверх log/slog и позволяет
// передавать атрибуты уровня запроса через context.Context.
package logger

import (
	"context"
	"io"
	"log/slog"
)

// New возвращает логгер, пишущий в w в заданном формате ("json" или
// "text"). Атрибуты, положенные в контекст через WithAttrs, добавляются
// в каждую запись, залогированную с этим контекстом (InfoContext, ErrorContext, ...).
func New(w io.Writer, level slog.Level, format string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level}

	var h slog.Handler
	switch format {
	case "text":
		h = slog.NewTextHandler(w, opts)
	default:
		h = slog.NewJSONHandler(w, opts)
	}

	return slog.New(contextHandler{Handler: h})
}

// Err — сокращение для стандартного атрибута ошибки.
func Err(err error) slog.Attr {
	return slog.Any("error", err)
}

type ctxKey struct{}

// WithAttrs возвращает копию ctx, содержащую attrs в дополнение
// к уже сохранённым в нём атрибутам.
func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	existing, _ := ctx.Value(ctxKey{}).([]slog.Attr)

	merged := make([]slog.Attr, 0, len(existing)+len(attrs))
	merged = append(merged, existing...)
	merged = append(merged, attrs...)

	return context.WithValue(ctx, ctxKey{}, merged)
}

var _ slog.Handler = contextHandler{}

// contextHandler добавляет атрибуты из контекста в каждую запись.
type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(ctxKey{}).([]slog.Attr); ok {
		r.AddAttrs(attrs...)
	}
	return h.Handler.Handle(ctx, r)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{Handler: h.Handler.WithGroup(name)}
}
