package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/gambitier/golang-service-template/internal/presentation/http/response"
	"github.com/gofiber/fiber/v3"
)

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("150405.000000000")))
	}
	return hex.EncodeToString(b[:])
}

// HttpRequestMiddleware adds a request ID and logs request completion.
func HttpRequestMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		reqID := c.Get("X-Request-Id")
		if reqID == "" {
			reqID = c.Get("X-Request-ID")
		}
		if reqID == "" {
			reqID = newRequestID()
		}
		c.Locals("request_id", reqID)
		c.Set("X-Request-Id", reqID)

		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		status := c.Response().StatusCode()
		attrs := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"request_id", reqID,
		}

		if respErr := response.GetResponseError(c); respErr != nil {
			attrs = append(attrs, "error", respErr.Error())
			logger.Error("request completed with error", attrs...)
			return err
		}
		if err != nil {
			attrs = append(attrs, "error", err.Error())
			logger.Error("request failed", attrs...)
			return err
		}

		logger.Info("request completed", attrs...)
		return nil
	}
}

// RecoverMiddleware recovers from panics and returns 500.
func RecoverMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered", "panic", r, "path", c.Path())
				_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": 0,
					"message": "internal error",
					"data":    nil,
				})
			}
		}()
		return c.Next()
	}
}
