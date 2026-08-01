package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	domainerr "github.com/gambitier/go-pkgs/errors"
	"github.com/gambitier/go-pkgs/logging"
	"github.com/gambitier/go-pkgs/logging/correlation"
	"github.com/gambitier/golang-service-template/internal/platform"
	"github.com/gambitier/golang-service-template/internal/presentation/http/response"
	"github.com/gofiber/fiber/v3"
)

func newCorrelationID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("150405.000000000")))
	}
	return hex.EncodeToString(b[:])
}

// HttpRequestMiddleware propagates X-Correlation-ID and logs request completion.
func HttpRequestMiddleware(logger logging.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		incoming := c.Get(correlation.HeaderName)
		cid := incoming
		if cid == "" {
			cid = newCorrelationID()
		}
		c.Set(correlation.HeaderName, cid)
		c.Locals("correlation_id", cid)

		ctx, _ := correlation.EnsureCorrelationID(c.Context(), cid)
		c.SetContext(ctx)
		reqLogger := logger.WithCorrelationID(cid)

		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		status := c.Response().StatusCode()
		fields := logging.Fields{
			"method":         c.Method(),
			"path":           c.Path(),
			"status":         status,
			"latency_ms":     latency.Milliseconds(),
			"correlation_id": cid,
		}

		if respErr := response.GetResponseError(c); respErr != nil {
			for k, v := range platform.DomainErrFields(respErr) {
				fields[k] = v
			}
			reqLogger.Error("request completed with error", respErr, fields)
			return err
		}
		if err != nil {
			reqLogger.Error("request failed", err, fields)
			return err
		}

		reqLogger.Info("request completed", fields)
		return nil
	}
}

// RecoverMiddleware recovers from panics and returns an RFC 9457 problem.
func RecoverMiddleware(logger logging.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered", nil, logging.Fields{
					"panic": r,
					"path":  c.Path(),
				})
				_ = response.WriteError(c, domainerr.Internal("internal error", nil, map[string]any{
					"panic": r,
				}))
			}
		}()
		return c.Next()
	}
}
