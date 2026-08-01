package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/gambitier/go-pkgs/apiresponse"
	"github.com/gambitier/go-pkgs/logging"
	"github.com/gambitier/golang-service-template/internal/config"
	itemhttp "github.com/gambitier/golang-service-template/internal/item/presentation/http"
	"github.com/gambitier/golang-service-template/internal/shared/presentation/http/middleware"
	"github.com/gambitier/golang-service-template/internal/shared/presentation/http/response"
	"github.com/gambitier/golang-service-template/internal/shared/presentation/http/routes"
	"github.com/gofiber/fiber/v3"
)

// HTTP is the Fiber HTTP server as a lifecycle component.
type HTTP struct {
	app    *fiber.App
	config *config.Config
	logger logging.Logger
	errCh  chan error
}

// NewHTTP wires middleware and routes. Call Start via lifecycle.App.
func NewHTTP(cfg *config.Config, logger logging.Logger, itemHandler *itemhttp.Handler) (*HTTP, error) {
	app := fiber.New(fiber.Config{
		AppName:       "golang-service-template",
		ReadTimeout:   cfg.Server.HTTP.ReadTimeout,
		WriteTimeout:  cfg.Server.HTTP.WriteTimeout,
		IdleTimeout:   cfg.Server.HTTP.IdleTimeout,
		StrictRouting: true,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			mappedErr := err
			var fe *fiber.Error
			if errors.As(err, &fe) {
				mappedErr = apiresponse.ToDomainError(fe.Code, fe.Message, fe)
			}
			return response.WriteError(c, mappedErr)
		},
	})

	middleware.Register(app, middleware.Deps{
		Logger: logger,
		CORS:   cfg.Server.CORS,
	})
	routes.RegisterRoutes(app, cfg, logger, itemHandler)

	return &HTTP{
		app:    app,
		config: cfg,
		logger: logger,
		errCh:  make(chan error, 1),
	}, nil
}

func (s *HTTP) Name() string { return "http" }

// Start begins listening in the background.
func (s *HTTP) Start(_ context.Context) error {
	addr := fmt.Sprintf(":%d", s.config.Server.HTTP.Port)
	s.logger.Info("starting HTTP server", logging.Fields{"address": addr})

	go func() {
		if err := s.app.Listen(addr); err != nil {
			select {
			case s.errCh <- err:
			default:
			}
		}
	}()
	return nil
}

func (s *HTTP) Stop(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server", nil)
	return s.app.ShutdownWithContext(ctx)
}
