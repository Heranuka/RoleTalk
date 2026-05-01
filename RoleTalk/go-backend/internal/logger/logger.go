// Package logger provides a production-ready Zap logger for internal use.
//
// It supports structured logging, environment-specific configs, and optional
// contextual fields for request tracing.
package logger

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config defines the logger configuration.
type Config struct {
	Environment string   // "production" or "development"
	Level       string   // debug, info, warn, error
	OutputPaths []string // optional
	ErrorOutput []string // optional
}

// New creates a new *zap.SugaredLogger based on the provided config.
//
// It does NOT set any global variables. Designed for DI usage in handlers/services.
func New(cfg Config) (*zap.SugaredLogger, error) {
	var zapCfg zap.Config

	switch cfg.Environment {
	case "production":
		zapCfg = zap.NewProductionConfig()
		zapCfg.Encoding = "json"
		if len(cfg.OutputPaths) > 0 {
			zapCfg.OutputPaths = cfg.OutputPaths
		} else {
			zapCfg.OutputPaths = []string{"stdout"}
		}
		if len(cfg.ErrorOutput) > 0 {
			zapCfg.ErrorOutputPaths = cfg.ErrorOutput
		} else {
			zapCfg.ErrorOutputPaths = []string{"stderr"}
		}
		zapCfg.EncoderConfig.TimeKey = "timestamp"
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	default:
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.Encoding = "console"
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapCfg.OutputPaths = []string{"stdout"}
		zapCfg.ErrorOutputPaths = []string{"stderr"}
	}

	if cfg.Level != "" {
		if lvl, err := zapcore.ParseLevel(cfg.Level); err == nil {
			zapCfg.Level = zap.NewAtomicLevelAt(lvl)
		} else {
			return nil, fmt.Errorf("invalid log level: %w", err)
		}
	}

	zapLogger, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	return zapLogger.Sugar(), nil
}

// Sync flushes any buffered logs. Call before application exit.
func Sync(log *zap.SugaredLogger) {
	if log != nil {
		_ = log.Sync() //nolint:errcheck
	}
}

// FromContext returns a logger enriched with request_id, trace_id, and span_id.
func FromContext(ctx context.Context, l *zap.SugaredLogger) *zap.SugaredLogger {
	if l == nil {
		return zap.NewNop().Sugar()
	}

	newLog := l

	// 1. Extract Request ID
	if reqID := middleware.GetReqID(ctx); reqID != "" {
		newLog = newLog.With("request_id", reqID)
	}

	// 2. Extract Tracing metadata (TraceID and SpanID)
	span := trace.SpanFromContext(ctx)
	if sc := span.SpanContext(); sc.IsValid() {
		newLog = newLog.With(
			"trace_id", sc.TraceID().String(),
			"span_id", sc.SpanID().String(),
		)
	}

	return newLog
}
