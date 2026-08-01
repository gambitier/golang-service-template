package response

import (
	"net/http"

	"github.com/gambitier/go-pkgs/apiresponse"
	"github.com/gofiber/fiber/v3"
)

const responseErrorLocalKey = "response_error"

// Problem is a Swagger-friendly mirror of apiresponse.Problem (RFC 9457).
type Problem struct {
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Status   int            `json:"status"`
	Detail   string         `json:"detail,omitempty"`
	Instance string         `json:"instance,omitempty"`
	Code     string         `json:"code,omitempty"`
	Fields   map[string]any `json:"fields,omitempty"`
}

// WriteError maps an error to an RFC 9457 problem details response.
func WriteError(c fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	c.Locals(responseErrorLocalKey, err)
	built := apiresponse.BuildProblem(err, apiresponse.BuildOptions{Instance: c.Path()})
	c.Set(fiber.HeaderContentType, apiresponse.ContentTypeProblemJSON)
	return c.Status(built.Status).JSON(built.Problem)
}

// OK writes a 200 response with the resource body (no envelope).
func OK[T any](c fiber.Ctx, data T) error {
	return c.Status(http.StatusOK).JSON(data)
}

// Created writes a 201 response with the resource body (no envelope).
func Created[T any](c fiber.Ctx, data T) error {
	return c.Status(http.StatusCreated).JSON(data)
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
