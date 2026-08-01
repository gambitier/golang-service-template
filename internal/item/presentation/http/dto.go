package item

import "time"

// CreateItemRequest is the HTTP body for creating an item.
type CreateItemRequest struct {
	Name        string `json:"name" example:"Notebook"`
	Description string `json:"description" example:"A ruled notebook"`
}

// UpdateItemRequest is the HTTP body for updating an item.
type UpdateItemRequest struct {
	Name        string `json:"name" example:"Notebook"`
	Description string `json:"description" example:"Updated description"`
}

// ItemResponse is the HTTP representation of an item.
type ItemResponse struct {
	ID          string    `json:"id" example:"507f1f77bcf86cd799439011"`
	Name        string    `json:"name" example:"Notebook"`
	Description string    `json:"description" example:"A ruled notebook"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ListItemsResponse wraps a list of items.
type ListItemsResponse struct {
	Items []ItemResponse `json:"items"`
}
