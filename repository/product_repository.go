package repository

import (
	"context"
	"geosearch-poc/domain"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	// Create creates a new product
	Create(ctx context.Context, product *domain.Product) error

	// GetByID retrieves a product by its ID
	GetByID(ctx context.Context, id string) (*domain.Product, error)

	// GetByStoreID retrieves products by store ID
	GetByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*domain.Product, error)

	// Update updates an existing product
	Update(ctx context.Context, product *domain.Product) error

	// Delete deletes a product by its ID
	Delete(ctx context.Context, id string) error

	// GetByStoreIDs retrieves products for multiple stores
	GetByStoreIDs(ctx context.Context, storeIDs []string, limit int) (map[string][]*domain.Product, error)
}
