package domain

import (
	"time"
)

// Category represents a store category
type Category struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewCategory creates a new category instance
func NewCategory(name, description string) *Category {
	now := time.Now()
	return &Category{
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Update updates the category information
func (c *Category) Update(name, description string) {
	c.Name = name
	c.Description = description
	c.UpdatedAt = time.Now()
}
