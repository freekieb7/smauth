package telemetry

import (
	"context"
	"time"

	"github.com/freekieb7/smauth/internal"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
)

func NewMeterProvider(ctx context.Context, cfg internal.ServerConfig) (*sdkMetric.MeterProvider, error) {
	metricExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, err
	}

	meterProvider := sdkMetric.NewMeterProvider(
		sdkMetric.WithReader(sdkMetric.NewPeriodicReader(metricExporter, sdkMetric.WithInterval(1*time.Minute))),
		sdkMetric.WithResource(NewResourceAttributes(cfg)),
	)
	return meterProvider, nil
}

type Meter struct {
	metric.Meter
}

func NewMeter(name string, provider *sdkMetric.MeterProvider) Meter {
	return Meter{
		provider.Meter(name),
	}
}

func (m *Meter) RecordInt64Counter(ctx context.Context, name string, value int64, attrs ...metric.AddOption) {
	counter, err := m.Int64Counter(name)
	if err != nil {
		return
	}
	counter.Add(ctx, value, attrs...)
}
