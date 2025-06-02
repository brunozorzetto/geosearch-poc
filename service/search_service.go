package service

import (
	"context"

	"geosearch-poc/domain"
	"geosearch-poc/pkg/distance"
	"geosearch-poc/pkg/h3"
	"geosearch-poc/repository/postgres"
)

// SearchService handles store search operations
type SearchService struct {
	storeRepo    *postgres.StoreRepository
	productRepo  *postgres.ProductRepository
	indexer      *h3.Indexer
	productLimit int
}

// NewSearchService creates a new search service instance
func NewSearchService(
	storeRepo *postgres.StoreRepository,
	productRepo *postgres.ProductRepository,
	indexer *h3.Indexer,
	productLimit int,
) *SearchService {
	return &SearchService{
		storeRepo:    storeRepo,
		productRepo:  productRepo,
		indexer:      indexer,
		productLimit: productLimit,
	}
}

// SearchStores searches for stores based on the provided parameters
func (s *SearchService) SearchStores(ctx context.Context, params *domain.SearchParams) (*domain.SearchResponse, error) {
	// Get H3 cells within the search radius
	h3Cells, err := s.indexer.GetCellsInRadius(params.Latitude, params.Longitude, params.Radius)
	if err != nil {
		return nil, err
	}

	// Search stores in the cells
	stores, err := s.storeRepo.GetByH3Indexes(ctx, h3Cells)
	if err != nil {
		return nil, err
	}

	// Get store IDs for product lookup
	storeIDs := make([]string, len(stores))
	for i, store := range stores {
		storeIDs[i] = store.ID
	}

	// Get products for the stores
	productsByStore, err := s.productRepo.GetByStoreIDs(ctx, storeIDs, s.productLimit)
	if err != nil {
		return nil, err
	}

	// Build the response
	response := &domain.SearchResponse{
		Stores: make([]domain.StoreWithProducts, 0, len(stores)),
		Total:  len(stores),
	}

	for _, store := range stores {
		// Calculate distance
		dist := distance.CalculateDistance(
			params.Latitude,
			params.Longitude,
			store.Latitude,
			store.Longitude,
		)

		// Only include stores within the radius
		if dist <= params.Radius {
			// Convert []*Product to []Product
			products := make([]domain.Product, len(productsByStore[store.ID]))
			for i, p := range productsByStore[store.ID] {
				products[i] = *p
			}

			storeWithProducts := domain.StoreWithProducts{
				Store:    *store,
				Products: products,
				Distance: dist,
			}
			response.Stores = append(response.Stores, storeWithProducts)
		}
	}

	return response, nil
}
