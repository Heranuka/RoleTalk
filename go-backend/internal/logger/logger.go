// Package logger provides a production-ready Zap logger with OpenTelemetry integration.
package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config defines the specific parameters needed to initialize the logger.
// Keeping this local prevents circular dependencies with the global config package.
type Config struct {
	Environment string
	Level       string
	OTelCore    zapcore.Core
}

// New creates a new *zap.SugaredLogger.
// It uses a "Tee" core to send logs to both stdout and OpenTelemetry simultaneously.
func New(cfg Config) (*zap.SugaredLogger, error) {
	var encoderCfg zapcore.EncoderConfig
	var encoder zapcore.Encoder

	if cfg.Environment == "production" {
		encoderCfg = zap.NewProductionEncoderConfig()
		// Standards for Loki/Tempo correlation
		encoderCfg.TimeKey = "timestamp"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	// Create the standard output core
	stdoutCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)

	// Combine stdout with OTelCore if provided
	finalCore := stdoutCore
	if cfg.OTelCore != nil {
		finalCore = zapcore.NewTee(stdoutCore, cfg.OTelCore)
	}

	// AddCaller allows us to see the file and line number in logs
	l := zap.New(finalCore, zap.AddCaller())

	return l.Sugar(), nil
}

// Sync flushes any buffered log entries.
// Should be called via defer in main.go to ensure all logs reach the collector.
func Sync(l *zap.SugaredLogger) {
	if l != nil {
		_ = l.Sync()
	}
}

// FromContext extracts RequestID and OpenTelemetry Tracing IDs from the context.
// This allows Loki logs to be directly linked to Tempo traces in Grafana.
func FromContext(ctx context.Context, l *zap.SugaredLogger) *zap.SugaredLogger {
	if l == nil {
		return zap.NewNop().Sugar()
	}

	reqID := middleware.GetReqID(ctx)
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()

	// If no metadata is found, return the logger as is
	if reqID == "" && !sc.IsValid() {
		return l
	}

	fields := make([]interface{}, 0, 6)
	if reqID != "" {
		fields = append(fields, "request_id", reqID)
	}
	if sc.IsValid() {
		fields = append(fields,
			"trace_id", sc.TraceID().String(),
			"span_id", sc.SpanID().String(),
		)
	}

	return l.With(fields...)
}
