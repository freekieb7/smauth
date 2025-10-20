package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/freekieb7/smauth/internal"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdkLog "go.opentelemetry.io/otel/sdk/log"
)

type Logger struct {
	Name string
	*slog.Logger
}

func NewLogger(name string, provider *sdkLog.LoggerProvider) Logger {
	return Logger{
		Name: name,
		Logger: slog.New(&FanoutHandler{handlers: []slog.Handler{
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			otelslog.NewHandler(name, otelslog.WithLoggerProvider(provider)),
		}}),
	}
}

func NewLoggerProvider(ctx context.Context, cfg internal.ServerConfig) (*sdkLog.LoggerProvider, error) {
	logExporter, err := otlploggrpc.New(ctx)
	if err != nil {
		panic(err)
	}

	loggerProvider := sdkLog.NewLoggerProvider(
		sdkLog.WithProcessor(sdkLog.NewBatchProcessor(logExporter)),
		sdkLog.WithResource(NewResourceAttributes(cfg)),
	)

	return loggerProvider, nil
}

// Ensures the FanoutHandler implements slog.Handler interface
var _ slog.Handler = (*FanoutHandler)(nil)

type FanoutHandler struct {
	handlers []slog.Handler
}

func Fanout(handlers ...slog.Handler) slog.Handler {
	return &FanoutHandler{
		handlers: handlers,
	}
}

// Enabled implements slog.Handler.
func (h *FanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, level) {
			return true
		}
	}

	return false
}

// Handle implements slog.Handler.
func (h *FanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, r.Level) {
			var err error
			defer func() {
				if r := recover(); r != nil {
					if e, ok := r.(error); ok {
						err = e
					} else {
						err = fmt.Errorf("unexpected error: %+v", r)
					}
				}
			}()

			err = h.handlers[i].Handle(ctx, r.Clone())

			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	// If errs is empty, or contains only nil errors, this returns nil
	return errors.Join(errs...)
}

// WithAttrs implements slog.Handler.
func (h *FanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	for i := range h.handlers {
		h.handlers[i].WithAttrs(slices.Clone(attrs))
	}
	return h
}

// WithGroup implements slog.Handler.
func (h *FanoutHandler) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	for i := range h.handlers {
		h.handlers[i].WithGroup(name)
	}
	return h
}
