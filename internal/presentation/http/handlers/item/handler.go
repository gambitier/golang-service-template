package item

import (
	"strconv"

	domainerr "github.com/gambitier/go-pkgs/errors"
	appitem "github.com/gambitier/golang-service-template/internal/application/item"
	domainitem "github.com/gambitier/golang-service-template/internal/domain/item"
	"github.com/gambitier/golang-service-template/internal/presentation/http/response"
	"github.com/gofiber/fiber/v3"
)

// Handler exposes item HTTP endpoints.
type Handler struct {
	svc *appitem.Service
}

// NewHandler constructs an item handler.
func NewHandler(svc *appitem.Service) *Handler {
	return &Handler{svc: svc}
}

func toResponse(it *domainitem.Item) ItemResponse {
	return ItemResponse{
		ID:          it.ID.String(),
		Name:        it.Name,
		Description: it.Description,
		CreatedAt:   it.CreatedAt,
		UpdatedAt:   it.UpdatedAt,
	}
}

// Create godoc
// @Summary     Create item
// @Description Create a new item
// @Tags        items
// @Accept      json
// @Produce     json
// @Param       body body CreateItemRequest true "Item payload"
// @Success     201 {object} response.Problem
// @Failure     400 {object} response.Problem
// @Router      /items [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req CreateItemRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.WriteError(c, domainerr.InvalidArgumentWithFields("invalid request body", map[string]any{"error": err.Error()}))
	}

	it, err := h.svc.Create(c.Context(), appitem.CreateInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return response.WriteError(c, err)
	}
	return response.Created(c, toResponse(it))
}

// List godoc
// @Summary     List items
// @Description List items with optional pagination
// @Tags        items
// @Produce     json
// @Param       limit  query int false "Page size" default(50)
// @Param       offset query int false "Offset" default(0)
// @Success     200 {object} response.Problem
// @Failure     400 {object} response.Problem
// @Router      /items [get]
func (h *Handler) List(c fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	items, err := h.svc.List(c.Context(), limit, offset)
	if err != nil {
		return response.WriteError(c, err)
	}

	out := make([]ItemResponse, 0, len(items))
	for _, it := range items {
		out = append(out, toResponse(it))
	}
	return response.OK(c, ListItemsResponse{Items: out})
}

// GetByID godoc
// @Summary     Get item by ID
// @Description Fetch a single item
// @Tags        items
// @Produce     json
// @Param       id path string true "Item ID"
// @Success     200 {object} response.Problem
// @Failure     404 {object} response.Problem
// @Router      /items/{id} [get]
func (h *Handler) GetByID(c fiber.Ctx) error {
	id := domainitem.ID(c.Params("id"))
	it, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		return response.WriteError(c, err)
	}
	return response.OK(c, toResponse(it))
}

// Update godoc
// @Summary     Update item
// @Description Update an existing item
// @Tags        items
// @Accept      json
// @Produce     json
// @Param       id   path string            true "Item ID"
// @Param       body body UpdateItemRequest true "Item payload"
// @Success     200 {object} response.Problem
// @Failure     400 {object} response.Problem
// @Failure     404 {object} response.Problem
// @Router      /items/{id} [put]
func (h *Handler) Update(c fiber.Ctx) error {
	id := domainitem.ID(c.Params("id"))
	var req UpdateItemRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.WriteError(c, domainerr.InvalidArgumentWithFields("invalid request body", map[string]any{"error": err.Error()}))
	}

	it, err := h.svc.Update(c.Context(), appitem.UpdateInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return response.WriteError(c, err)
	}
	return response.OK(c, toResponse(it))
}

// Delete godoc
// @Summary     Delete item
// @Description Delete an item by ID
// @Tags        items
// @Param       id path string true "Item ID"
// @Success     204 "No Content"
// @Failure     404 {object} response.Problem
// @Router      /items/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id := domainitem.ID(c.Params("id"))
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return response.WriteError(c, err)
	}
	return response.NoContent(c)
}
