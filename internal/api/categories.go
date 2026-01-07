package api

import (
	"fmt"
	"time"
)

// Category represents a category from the API
type Category struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// CategoryCreateRequest represents a create category request
type CategoryCreateRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// CategoryUpdateRequest represents an update category request
type CategoryUpdateRequest struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
}

// ListCategories fetches all categories for the current user
func (c *Client) ListCategories() ([]Category, error) {
	var categories []Category
	if err := c.Get("/categories", &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

// GetCategory fetches a single category by ID
func (c *Client) GetCategory(id int) (*Category, error) {
	var category Category
	if err := c.Get(fmt.Sprintf("/categories/%d", id), &category); err != nil {
		return nil, err
	}
	return &category, nil
}

// CreateCategory creates a new category
func (c *Client) CreateCategory(name, color string) (*Category, error) {
	req := CategoryCreateRequest{Name: name, Color: color}
	var category Category
	if err := c.Post("/categories", req, &category); err != nil {
		return nil, err
	}
	return &category, nil
}

// UpdateCategory updates an existing category
func (c *Client) UpdateCategory(id int, req CategoryUpdateRequest) (*Category, error) {
	var category Category
	if err := c.Patch(fmt.Sprintf("/categories/%d", id), req, &category); err != nil {
		return nil, err
	}
	return &category, nil
}

// DeleteCategory deletes a category
func (c *Client) DeleteCategory(id int) error {
	return c.Delete(fmt.Sprintf("/categories/%d", id), nil)
}
