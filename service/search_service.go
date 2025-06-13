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
	storeRepo *postgres.StoreRepository
	indexer   *h3.Indexer
}

// NewSearchService creates a new search service instance
func NewSearchService(
	storeRepo *postgres.StoreRepository,
	indexer *h3.Indexer,
) *SearchService {
	return &SearchService{
		storeRepo: storeRepo,
		indexer:   indexer,
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

	// Build the response
	response := &domain.SearchResponse{
		Stores: make([]domain.StoreWithDistance, 0, len(stores)),
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

		// Only include stores within the search radius AND within their delivery radius
		if dist <= params.Radius && dist <= float64(store.DeliveryRadius)/1000 { // Convert delivery_radius from meters to kilometers
			storeWithDistance := domain.StoreWithDistance{
				Store:    *store,
				Distance: dist,
			}
			response.Stores = append(response.Stores, storeWithDistance)
		}
	}

	return response, nil
}
