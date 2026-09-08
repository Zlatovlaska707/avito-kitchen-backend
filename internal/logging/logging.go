package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type ctxKey struct{}

type Fields map[string]any

func New(level, format string) *slog.Logger {
	lvl := parseLevel(level)
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: lvl, AddSource: lvl == slog.LevelDebug}
	switch strings.ToLower(format) {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

func NewWithWriter(w io.Writer, level, format string) *slog.Logger {
	lvl := parseLevel(level)
	opts := &slog.HandlerOptions{Level: lvl}
	var handler slog.Handler
	switch strings.ToLower(format) {
	case "text":
		handler = slog.NewTextHandler(w, opts)
	default:
		handler = slog.NewJSONHandler(w, opts)
	}
	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func WithContext(ctx context.Context, fields Fields) context.Context {
	existing, _ := ctx.Value(ctxKey{}).(Fields)
	merged := make(Fields, len(existing)+len(fields))
	for k, v := range existing {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return context.WithValue(ctx, ctxKey{}, merged)
}

func FromContext(ctx context.Context) *slog.Logger {
	logger := slog.Default()
	fields, ok := ctx.Value(ctxKey{}).(Fields)
	if !ok || len(fields) == 0 {
		return logger
	}
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return logger.With(args...)
}

func WrapError(err error, msg string, fields ...any) error {
	if err == nil {
		return nil
	}
	args := append([]any{"error", err}, fields...)
	FromContext(context.Background()).Error(msg, args...)
	return err
}

func MaskPhone(phone string) string {
	if len(phone) < 6 {
		return "***"
	}
	return phone[:3] + strings.Repeat("*", len(phone)-6) + phone[len(phone)-3:]
}
