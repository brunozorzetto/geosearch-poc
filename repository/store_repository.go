package repository

import (
	"context"
	"geosearch-poc/domain"
)

// StoreRepository defines the interface for store operations
type StoreRepository interface {
	// Create creates a new store
	Create(ctx context.Context, store *domain.Store) error

	// GetByID retrieves a store by its ID
	GetByID(ctx context.Context, id int) (*domain.Store, error)

	// GetByH3Index retrieves stores by their H3 index
	GetByH3Index(ctx context.Context, h3Index string) ([]*domain.Store, error)

	// GetByH3Indexes retrieves stores by multiple H3 indexes
	GetByH3Indexes(ctx context.Context, h3Indexes []string) ([]*domain.Store, error)

	// Update updates an existing store
	Update(ctx context.Context, store *domain.Store) error

	// Delete deletes a store by its ID
	Delete(ctx context.Context, id int) error

	// Search searches for stores based on criteria
	Search(ctx context.Context, params *domain.SearchParams) ([]*domain.Store, int, error)

	// SearchByRadius searches for stores within a given radius
	SearchByRadius(ctx context.Context, latitude, longitude float64, radius float64) ([]*domain.Store, error)
}
