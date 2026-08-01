package middleware

import (
	"net/url"
	"strings"

	"github.com/gambitier/go-pkgs/logging"
	"github.com/gambitier/golang-service-template/internal/config"
	"github.com/gambitier/golang-service-template/internal/shared/platform"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// Deps are inputs for the global HTTP middleware stack.
type Deps struct {
	Logger logging.Logger
	CORS   config.CORSConfig
}

// Register applies the global middleware stack to app.
// Order: OTel → recover → request scope → CORS.
func Register(app *fiber.App, deps Deps) {
	app.Use(platform.FiberMiddleware())
	app.Use(RecoverMiddleware(deps.Logger))
	app.Use(HttpRequestMiddleware(deps.Logger))
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: buildOriginMatcher(deps.CORS.AllowOrigins, deps.CORS.AllowOriginSuffixes),
		AllowMethods:     splitCSV(deps.CORS.AllowMethods),
		AllowHeaders:     splitCSV(deps.CORS.AllowHeaders),
		ExposeHeaders:    splitCSV(deps.CORS.ExposeHeaders),
		AllowCredentials: deps.CORS.AllowCredentials,
	}))
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
