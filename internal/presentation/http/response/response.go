package response

import (
	"net/http"

	"github.com/gambitier/golang-service-template/internal/domain/domainerr"
	"github.com/gofiber/fiber/v3"
)

const responseErrorLocalKey = "response_error"

// Envelope is the standard API response shape.
type Envelope struct {
	Success   int    `json:"success"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	RequestID string `json:"requestId,omitempty"`
	Code      string `json:"code,omitempty"`
	Fields    any    `json:"fields,omitempty"`
}

func requestIDFromCtx(c fiber.Ctx) string {
	if v := c.Get("X-Request-Id"); v != "" {
		return v
	}
	if v := c.Get("X-Request-ID"); v != "" {
		return v
	}
	if v := c.Locals("request_id"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Write maps an error to an HTTP envelope.
func Write(c fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	reqID := requestIDFromCtx(c)
	c.Locals(responseErrorLocalKey, err)

	if de, ok := domainerr.As(err); ok {
		return c.Status(de.HTTPStatus()).JSON(Envelope{
			Success:   0,
			Message:   de.Message,
			Data:      nil,
			RequestID: reqID,
			Code:      string(de.Code),
			Fields:    de.Fields,
		})
	}

	return c.Status(http.StatusInternalServerError).JSON(Envelope{
		Success:   0,
		Message:   "internal error",
		Data:      nil,
		RequestID: reqID,
		Code:      string(domainerr.CodeInternal),
	})
}

// OK writes a 200 success envelope.
func OK[T any](c fiber.Ctx, data T) error {
	return c.Status(http.StatusOK).JSON(Envelope{
		Success:   1,
		Message:   "ok",
		Data:      data,
		RequestID: requestIDFromCtx(c),
	})
}

// Created writes a 201 success envelope.
func Created[T any](c fiber.Ctx, data T) error {
	return c.Status(http.StatusCreated).JSON(Envelope{
		Success:   1,
		Message:   "created",
		Data:      data,
		RequestID: requestIDFromCtx(c),
	})
}

// NoContent writes a 204 with empty body.
func NoContent(c fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}

// GetResponseError returns the error stored by Write, if any.
func GetResponseError(c fiber.Ctx) error {
	if v := c.Locals(responseErrorLocalKey); v != nil {
		if err, ok := v.(error); ok {
			return err
		}
	}
	return nil
}
