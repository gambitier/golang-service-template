package platform

import (
	"strings"

	commonobservability "github.com/gambitier/go-pkgs/observability"
	fiberotel "github.com/gofiber/contrib/v3/otel"
	"github.com/gofiber/fiber/v3"
)

var healthCheckPaths = map[string]struct{}{
	"/health":  {},
	"/livez":   {},
	"/healthz": {},
	"/swagger": {},
}

// FiberMiddleware records server spans via gofiber/contrib/v3/otel when
// observability is enabled. Health and swagger paths are skipped.
func FiberMiddleware() fiber.Handler {
	if !commonobservability.IsEnabled() {
		return func(c fiber.Ctx) error { return c.Next() }
	}
	return fiberotel.Middleware(
		fiberotel.WithNext(func(c fiber.Ctx) bool {
			return isHealthCheckPath(c.Path())
		}),
	)
}

func isHealthCheckPath(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	if _, ok := healthCheckPaths[path]; ok {
		return true
	}
	return strings.HasPrefix(path, "/swagger/")
}
