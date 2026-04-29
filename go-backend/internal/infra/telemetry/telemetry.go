// Package telemetry provides OpenTelemetry setup for distributed tracing.
package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation" // Добавлено
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.uber.org/zap"

	"go-backend/internal/config"
)

// Telemetry manages the lifecycle of OpenTelemetry TracerProvider.
type Telemetry struct {
	tp *sdktrace.TracerProvider
}

// New initializes OpenTelemetry Tracing and returns a Telemetry instance.
func New(ctx context.Context, cfg *config.Config, logger *zap.SugaredLogger) (*Telemetry, error) {
	// 1. Create OTLP Exporter for Tempo
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.Observability.OTLPEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}

	// 2. Define Resource metadata
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.App.Name),
			semconv.ServiceVersion(cfg.App.Version),
			semconv.DeploymentEnvironment(cfg.Env),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create telemetry resource: %w", err)
	}

	// 3. Create TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		// Sample everything in dev, or use cfg.Observability.SampleRate
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// 4. Set Global Providers
	otel.SetTracerProvider(tp)

	// CRITICAL: Set the propagator to allow cross-service tracing (Go -> Python)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Infow("telemetry initialized", "endpoint", cfg.Observability.OTLPEndpoint)
	return &Telemetry{tp: tp}, nil
}

// Shutdown gracefully flushes traces.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t == nil || t.tp == nil {
		return nil
	}
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := t.tp.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("tracer shutdown: %w", err)
	}
	return nil
}
