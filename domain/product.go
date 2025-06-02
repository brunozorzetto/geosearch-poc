package domain

import (
	"time"
)

// Product represents a product in a store
type Product struct {
	ID          string    `json:"id"`
	StoreID     string    `json:"store_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewProduct creates a new product instance
func NewProduct(storeID, name, description string, price float64) *Product {
	now := time.Now()
	return &Product{
		StoreID:     storeID,
		Name:        name,
		Description: description,
		Price:       price,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Update updates the product information
func (p *Product) Update(name, description string, price float64) {
	p.Name = name
	p.Description = description
	p.Price = price
	p.UpdatedAt = time.Now()
}
