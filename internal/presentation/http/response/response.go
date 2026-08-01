package response

import (
	"net/http"

	"github.com/gambitier/go-pkgs/errors/httpresp"
	"github.com/gofiber/fiber/v3"
)

const responseErrorLocalKey = "response_error"

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

// Write maps an error to an HTTP envelope via go-pkgs/errors/httpresp.
func Write(c fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	reqID := requestIDFromCtx(c)
	c.Locals(responseErrorLocalKey, err)
	built := httpresp.BuildError(err, reqID)
	return c.Status(built.Status).JSON(fiber.Map{
		"success":   0,
		"message":   built.Msg,
		"data":      nil,
		"requestId": reqID,
		"code":      string(built.Code),
		"fields":    built.Fields,
	})
}

// OK writes a 200 success envelope.
func OK[T any](c fiber.Ctx, data T) error {
	reqID := requestIDFromCtx(c)
	env := httpresp.BuildOKEnvelope(data, reqID)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"success":   1,
		"message":   "ok",
		"data":      env.Data,
		"requestId": reqID,
	})
}

// Created writes a 201 success envelope.
func Created[T any](c fiber.Ctx, data T) error {
	reqID := requestIDFromCtx(c)
	env := httpresp.BuildCreatedEnvelope(data, reqID)
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"success":   1,
		"message":   "created",
		"data":      env.Data,
		"requestId": reqID,
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
