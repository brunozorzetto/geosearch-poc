package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

// Product represents a product in the system
type Product struct {
	ID          uuid.UUID `json:"id"`
	StoreID     uuid.UUID `json:"store_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	Brand       string    `json:"brand"`
	SKU         string    `json:"sku"`
	Stock       int       `json:"stock"`
	H3Index     string    `json:"h3_index"` // H3 index of the store's location
	Images      []string  `json:"images"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductSearchParams represents the parameters for a product search
type ProductSearchParams struct {
	Query     string  `json:"query"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    float64 `json:"radius"` // in kilometers
	Category  string  `json:"category,omitempty"`
	Brand     string  `json:"brand,omitempty"`
	MinPrice  float64 `json:"min_price,omitempty"`
	MaxPrice  float64 `json:"max_price,omitempty"`
	PageSize  int     `json:"page_size,omitempty"`
	PageToken string  `json:"page_token,omitempty"`
}

// ProductSearchResponse represents the response of a product search
type ProductSearchResponse struct {
	Products      []Product `json:"products"`
	TotalCount    int64     `json:"total_count"`
	NextPageToken string    `json:"next_page_token,omitempty"`
}

// NewProduct creates a new product instance
func NewProduct(storeID uuid.UUID, name, description string, price float64, category, brand, sku string, stock int, h3Index string, images []string) *Product {
	now := time.Now()
	return &Product{
		ID:          uuid.New(),
		StoreID:     storeID,
		Name:        name,
		Description: description,
		Price:       price,
		Category:    category,
		Brand:       brand,
		SKU:         sku,
		Stock:       stock,
		H3Index:     h3Index,
		Images:      images,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Update updates the product information
func (p *Product) Update(name, description string, price float64, category, brand, sku string, stock int, h3Index string, images []string) {
	p.Name = name
	p.Description = description
	p.Price = price
	p.Category = category
	p.Brand = brand
	p.SKU = sku
	p.Stock = stock
	p.H3Index = h3Index
	p.Images = images
	p.UpdatedAt = time.Now()
}
