package item

import (
	"github.com/gofiber/fiber/v3"
)

// Register mounts item routes on the versioned API group.
func Register(v1 fiber.Router, h *Handler) {
	v1.Post("/items", h.Create)
	v1.Get("/items", h.List)
	v1.Get("/items/:id", h.GetByID)
	v1.Put("/items/:id", h.Update)
	v1.Delete("/items/:id", h.Delete)
}
