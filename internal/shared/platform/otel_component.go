package platform

import (
	"context"

	"github.com/gambitier/go-pkgs/logging"
	"github.com/gambitier/go-pkgs/observability"
)

// OTelComponent manages OpenTelemetry init/shutdown.
type OTelComponent struct {
	cfg      observability.Config
	logger   logging.Logger
	shutdown func(context.Context) error
}

// NewOTelComponent builds an OTel lifecycle component.
func NewOTelComponent(cfg observability.Config, logger logging.Logger) *OTelComponent {
	return &OTelComponent{cfg: cfg, logger: logger}
}

func (c *OTelComponent) Name() string { return "otel" }

func (c *OTelComponent) Start(ctx context.Context) error {
	shutdown, err := InitObservability(ctx, c.cfg, c.logger)
	if err != nil {
		return err
	}
	c.shutdown = shutdown
	return nil
}

func (c *OTelComponent) Stop(ctx context.Context) error {
	if c.shutdown == nil {
		return nil
	}
	err := c.shutdown(ctx)
	c.shutdown = nil
	return err
}
