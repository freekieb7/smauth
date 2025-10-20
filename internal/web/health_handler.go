package web

import (
	"context"
	"time"

	"github.com/freekieb7/smauth/internal/http"
	"github.com/freekieb7/smauth/internal/telemetry"
	"go.opentelemetry.io/otel/attribute"
)

type HealthHandler struct {
	Logger *telemetry.Logger
	Tracer *telemetry.Tracer
	Meter  *telemetry.Meter
}

func NewHealthHandler(logger *telemetry.Logger, tracer *telemetry.Tracer, meter *telemetry.Meter) HealthHandler {
	return HealthHandler{
		Logger: logger,
		Tracer: tracer,
		Meter:  meter,
	}
}

func (h *HealthHandler) RegisterRoutes(router *http.Router) {
	router.GET("/health", h.HealthCheck, NoCacheMiddleware())
}

func (h *HealthHandler) HealthCheck(ctx context.Context, req *http.Request, res *http.Response) error {
	ctx, span := h.Tracer.Start(ctx, "health_handler.go")
	defer span.End()

	h.Logger.InfoContext(ctx, "Health check endpoint called")

	timestamp := int(time.Now().Unix())
	timevalueAttr := attribute.Int("health.time", timestamp)
	span.SetAttributes(timevalueAttr)

	h.Meter.RecordInt64Counter(ctx, "health_check_requests_total", 1)

	return res.SendJSON(http.StatusOK, map[string]string{"status": "ok"})
}
