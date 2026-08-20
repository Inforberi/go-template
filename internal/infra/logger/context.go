package logger

import (
	"context"

	"go.uber.org/zap"
)

type contextKey struct{}

func WithContext(ctx context.Context, log *zap.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, log)
}

func FromContext(ctx context.Context) *zap.Logger {
	log, ok := ctx.Value(contextKey{}).(*zap.Logger)
	if !ok {
		return zap.NewNop()
	}

	return log
}
