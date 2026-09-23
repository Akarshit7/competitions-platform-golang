package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func New(env string) *slog.Logger {
	level := slog.LevelInfo
	if strings.EqualFold(env, "development") || strings.EqualFold(env, "debug") {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: false,
	}
	var handler slog.Handler
	if strings.EqualFold(env, "development") {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(RequestIDKey).(string); ok {
		return v
	}
	return ""
}
