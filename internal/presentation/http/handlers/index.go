package handlers

import (
	appitem "github.com/gambitier/golang-service-template/internal/application/item"
	itemhandler "github.com/gambitier/golang-service-template/internal/presentation/http/handlers/item"
)

// Handlers aggregates all HTTP handlers.
type Handlers struct {
	Item *itemhandler.Handler
}

// NewHandlers constructs the handler bag.
func NewHandlers(itemSvc *appitem.Service) *Handlers {
	return &Handlers{
		Item: itemhandler.NewHandler(itemSvc),
	}
}
