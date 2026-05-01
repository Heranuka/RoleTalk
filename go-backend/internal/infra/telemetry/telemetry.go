// Package telemetry provides OpenTelemetry initialization for distributed tracing and structured logging.
// It integrates with the OTel Collector via gRPC to export data to Tempo and Loki.
package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.uber.org/zap/zapcore"

	"go-backend/internal/config"
)

// Telemetry manages the lifecycle of OpenTelemetry providers for tracing and logging.
type Telemetry struct {
	tp *sdktrace.TracerProvider
	lp *log.LoggerProvider
}

// New initializes OpenTelemetry providers and returns a zapcore.Core bridge for OTLP logging.
func New(ctx context.Context, cfg *config.Config) (*Telemetry, zapcore.Core, error) {
	// 1. Define shared resources for all telemetry data
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"", // Use default schema to avoid version conflicts
			semconv.ServiceName(cfg.App.Name),
			semconv.ServiceVersion(cfg.App.Version),
			semconv.DeploymentEnvironment(cfg.Env),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 2. Setup Tracing: OTLP gRPC Exporter for Tempo
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.Observability.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.Observability.SampleRate))),
	)
	otel.SetTracerProvider(tp)

	// 3. Setup Logging: OTLP gRPC Exporter for Loki
	logExporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(cfg.Observability.OTLPEndpoint),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create log exporter: %w", err)
	}

	lp := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)
	global.SetLoggerProvider(lp)

	// 4. Configure Text Map Propagator for distributed tracing (Go -> Python)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 5. Create Zap Bridge Core to redirect Zap logs to OpenTelemetry
	otelCore := otelzap.NewCore(cfg.App.Name, otelzap.WithLoggerProvider(lp))

	return &Telemetry{tp: tp, lp: lp}, otelCore, nil
}

// Shutdown gracefully flushes and stops the telemetry providers within the given context.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t == nil {
		return nil
	}

	// Create a timeout context for shutdown operations
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if t.tp != nil {
		if err := t.tp.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("tracer provider shutdown: %w", err)
		}
	}

	if t.lp != nil {
		if err := t.lp.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("logger provider shutdown: %w", err)
		}
	}

	return nil
}
