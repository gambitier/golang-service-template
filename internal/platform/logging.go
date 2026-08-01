package platform

import (
	"context"
	"fmt"

	"github.com/gambitier/go-pkgs/logging"
	commonobservability "github.com/gambitier/go-pkgs/observability"
	"github.com/sirupsen/logrus"
)

// NewLogger builds the service logger from config.
func NewLogger(cfg logging.Config) (logging.Logger, error) {
	if cfg.ServiceName == "" {
		cfg.ServiceName = "golang-service-template"
	}
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Format == "" {
		cfg.Format = "json"
	}
	if len(cfg.Sinks) == 0 {
		cfg.Sinks = []logging.SinkConfig{{Type: "console", Enabled: true}}
	}
	return logging.New(cfg)
}

// otelLogger adapts logging.Logger to observability.Logger without coupling the packages.
type otelLogger struct {
	inner logging.Logger
}

func (l otelLogger) AddHook(hook logrus.Hook) {
	l.inner.AddHook(hook)
}

func (l otelLogger) Warn(message string, fields map[string]any) {
	l.inner.Warn(message, logging.Fields(fields))
}

// InitObservability starts OpenTelemetry when enabled and attaches the logrus OTLP bridge.
func InitObservability(ctx context.Context, cfg commonobservability.Config, logger logging.Logger) (func(context.Context) error, error) {
	var sink commonobservability.Logger
	if logger != nil {
		sink = otelLogger{inner: logger}
	}
	shutdown, err := commonobservability.Init(ctx, cfg, sink)
	if err != nil {
		return nil, fmt.Errorf("observability init: %w", err)
	}
	return shutdown, nil
}
