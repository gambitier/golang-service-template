package routes

import (
	"github.com/gambitier/go-pkgs/logging"
	"github.com/gambitier/golang-service-template/internal/config"
	"github.com/gambitier/golang-service-template/internal/presentation/http/handlers"
	"github.com/gambitier/golang-service-template/internal/presentation/http/middleware"
	"github.com/gambitier/golang-service-template/internal/presentation/http/response"
	swaggo "github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
)

type probeStatus struct {
	Status string `json:"status"`
}

// RegisterRoutes registers probes, swagger, and item API routes.
func RegisterRoutes(app *fiber.App, cfg *config.Config, logger logging.Logger, h *handlers.Handlers) {
	app.Get("/livez", func(c fiber.Ctx) error {
		return response.OK(c, probeStatus{Status: "ok"})
	})

	app.Get("/healthz", func(c fiber.Ctx) error {
		return response.OK(c, probeStatus{Status: "ok"})
	})

	items := app.Group("/items")

	if cfg.Swagger.Enabled && cfg.Server.Env.IsDevelopment() {
		logger.Info("swagger enabled", logging.Fields{
			"enabled": cfg.Swagger.Enabled,
			"env":     cfg.Server.Env.String(),
		})
		swaggerHandler := swaggo.HandlerDefault
		if cfg.Swagger.Username != "" && cfg.Swagger.Password != "" {
			logger.Info("applying swagger basic auth", logging.Fields{"username": cfg.Swagger.Username})
			swaggerHandler = middleware.BasicAuthMiddleware(swaggerHandler, cfg.Swagger.Username, cfg.Swagger.Password)
		} else {
			logger.Warn("swagger enabled without auth", nil)
		}
		items.All("/swagger/*", swaggerHandler)
		logger.Info("swagger docs available", logging.Fields{"path": "/items/swagger/index.html"})
	} else {
		logger.Debug("swagger disabled", logging.Fields{
			"enabled": cfg.Swagger.Enabled,
			"env":     cfg.Server.Env.String(),
		})
	}

	v1 := items.Group("/api/v1")

	v1.Get("/livez", func(c fiber.Ctx) error {
		return response.OK(c, probeStatus{Status: "ok"})
	})
	v1.Get("/healthz", func(c fiber.Ctx) error {
		return response.OK(c, probeStatus{Status: "ok"})
	})

	v1.Post("/items", h.Item.Create)
	v1.Get("/items", h.Item.List)
	v1.Get("/items/:id", h.Item.GetByID)
	v1.Put("/items/:id", h.Item.Update)
	v1.Delete("/items/:id", h.Item.Delete)

	logger.Info("routes registered", nil)
}
