package telemetry

import (
	"context"
	"time"

	"github.com/freekieb7/smauth/internal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func NewTracerProvider(ctx context.Context, cfg internal.ServerConfig) (*sdkTrace.TracerProvider, error) {
	// Set up propagator.
	// This is necessary for proper context propagation in distributed systems.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(traceExporter, sdkTrace.WithBatchTimeout(5*time.Second)),
		sdkTrace.WithResource(NewResourceAttributes(cfg)),
	)
	return tracerProvider, nil
}

type Tracer struct {
	trace.Tracer
}

func NewTracer(name string, provider *sdkTrace.TracerProvider) Tracer {
	return Tracer{
		provider.Tracer(name),
	}
}
