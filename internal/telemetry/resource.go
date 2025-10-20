package telemetry

import (
	"github.com/freekieb7/smauth/internal"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func NewResourceAttributes(cfg internal.ServerConfig) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(cfg.Name),
		semconv.ServiceVersionKey.String(cfg.Version),
	)
}
