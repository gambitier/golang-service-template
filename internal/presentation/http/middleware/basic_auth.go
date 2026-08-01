package middleware

import (
	"encoding/base64"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// BasicAuthMiddleware wraps a handler with HTTP Basic Authentication.
func BasicAuthMiddleware(handler fiber.Handler, username, password string) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
			c.Set("WWW-Authenticate", `Basic realm="Swagger"`)
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}

		encoded := strings.TrimPrefix(authHeader, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			c.Set("WWW-Authenticate", `Basic realm="Swagger"`)
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}

		credentials := strings.SplitN(string(decoded), ":", 2)
		if len(credentials) != 2 || credentials[0] != username || credentials[1] != password {
			c.Set("WWW-Authenticate", `Basic realm="Swagger"`)
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}

		return handler(c)
	}
}
