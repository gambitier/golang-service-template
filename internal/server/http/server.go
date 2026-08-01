package http

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gambitier/go-pkgs/errors/domainerr"
	"github.com/gambitier/go-pkgs/logging"
	commonobservability "github.com/gambitier/go-pkgs/observability"
	"github.com/gambitier/golang-service-template/internal/config"
	"github.com/gambitier/golang-service-template/internal/presentation/http/handlers"
	presentationMiddleware "github.com/gambitier/golang-service-template/internal/presentation/http/middleware"
	presentationResponse "github.com/gambitier/golang-service-template/internal/presentation/http/response"
	"github.com/gambitier/golang-service-template/internal/presentation/http/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// Server is the Fiber HTTP server.
type Server struct {
	app    *fiber.App
	config *config.Config
	logger logging.Logger
}

// New wires Fiber middleware and routes.
func New(cfg *config.Config, logger logging.Logger, h *handlers.Handlers) (*Server, error) {
	app := fiber.New(fiber.Config{
		AppName:      "golang-service-template",
		ReadTimeout:  cfg.Server.HTTP.ReadTimeout,
		WriteTimeout: cfg.Server.HTTP.WriteTimeout,
		IdleTimeout:  cfg.Server.HTTP.IdleTimeout,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			mappedErr := err
			var fe *fiber.Error
			if errors.As(err, &fe) {
				mappedErr = domainerr.InvalidArgumentWithFields(fe.Message, map[string]any{"status": fe.Code})
				if fe.Code >= 500 {
					mappedErr = domainerr.Internal(fe.Message, fe, nil)
				}
			}
			return presentationResponse.Write(c, mappedErr)
		},
	})

	app.Use(commonobservability.FiberMiddleware("golang-service-template-http"))
	app.Use(presentationMiddleware.RecoverMiddleware(logger))
	app.Use(presentationMiddleware.HttpRequestMiddleware(logger))

	originAllowed := buildOriginMatcher(cfg.Server.CORS.AllowOrigins, cfg.Server.CORS.AllowOriginSuffixes)
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: originAllowed,
		AllowMethods:     splitCSV(cfg.Server.CORS.AllowMethods),
		AllowHeaders:     splitCSV(cfg.Server.CORS.AllowHeaders),
		ExposeHeaders:    splitCSV(cfg.Server.CORS.ExposeHeaders),
		AllowCredentials: cfg.Server.CORS.AllowCredentials,
	}))

	routes.RegisterRoutes(app, cfg, logger, h)

	return &Server{
		app:    app,
		config: cfg,
		logger: logger,
	}, nil
}

// Start listens until ctx is cancelled, then shuts down gracefully.
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.config.Server.HTTP.Port)
	s.logger.Info("starting HTTP server", logging.Fields{"address": addr})

	errChan := make(chan error, 1)
	go func() {
		if err := s.app.Listen(addr); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("shutting down HTTP server", nil)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.app.ShutdownWithContext(shutdownCtx)
	case err := <-errChan:
		return err
	}
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func buildOriginMatcher(allowOriginsCSV, allowOriginSuffixesCSV string) func(string) bool {
	allowAll := strings.TrimSpace(allowOriginsCSV) == "*"
	exactOrigins := make(map[string]struct{})
	for _, origin := range strings.Split(allowOriginsCSV, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" || trimmed == "*" {
			continue
		}
		exactOrigins[trimmed] = struct{}{}
	}

	var hostSuffixes []string
	for _, suffix := range strings.Split(allowOriginSuffixesCSV, ",") {
		trimmed := strings.TrimSpace(suffix)
		if trimmed == "" {
			continue
		}
		hostSuffixes = append(hostSuffixes, strings.ToLower(trimmed))
	}

	return func(origin string) bool {
		if allowAll {
			return true
		}
		if _, ok := exactOrigins[origin]; ok {
			return true
		}
		if len(hostSuffixes) == 0 {
			return false
		}

		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		host := strings.ToLower(u.Hostname())
		if host == "" {
			return false
		}

		for _, suffix := range hostSuffixes {
			if strings.HasPrefix(suffix, ".") {
				if strings.HasSuffix(host, suffix) && len(host) > len(suffix) {
					return true
				}
				continue
			}
			if host == suffix || strings.HasSuffix(host, "."+suffix) {
				return true
			}
		}
		return false
	}
}
