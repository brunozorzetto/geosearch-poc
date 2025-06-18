package repository

import (
	"context"
	"geosearch-poc/domain"

	"github.com/google/uuid"
)

// ProductRepository defines the interface for product operations
type ProductRepository interface {
	// Create creates a new product
	Create(ctx context.Context, product *domain.Product) error

	// GetByID retrieves a product by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)

	// GetByStoreID retrieves all products for a store
	GetByStoreID(ctx context.Context, storeID uuid.UUID) ([]*domain.Product, error)

	// Update updates an existing product
	Update(ctx context.Context, product *domain.Product) error

	// Delete deletes a product by its ID
	Delete(ctx context.Context, id uuid.UUID) error

	// Search searches for products based on criteria
	Search(ctx context.Context, params *domain.ProductSearchParams) ([]*domain.Product, int64, error)
}
