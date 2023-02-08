package logging

import (
	"context"
	"go.uber.org/zap"
)

type loggerKey struct {
}

var loggerKeyVal = &loggerKey{}

func TryGetLoggerFromContext(ctx context.Context) *zap.Logger {
	value := ctx.Value(loggerKeyVal)
	if value == nil {
		return nil
	}
	return value.(*zap.Logger)
}

func L(ctx context.Context, opts ...zap.Option) *zap.Logger {
	value := ctx.Value(loggerKeyVal)
	if value == nil {
		panic("Logging from a context without a logger")
	}
	logger := value.(*zap.Logger)
	if len(opts) > 0 {
		return logger.WithOptions(opts...)
	} else {
		return logger
	}
}

func SL(ctx context.Context, opts ...zap.Option) *zap.SugaredLogger {
	logger := L(ctx, opts...)
	return logger.Sugar()
}

func ImbueContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKeyVal, logger)
}

func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	logger := L(ctx)
	return ImbueContext(ctx, logger.With(fields...))
}
