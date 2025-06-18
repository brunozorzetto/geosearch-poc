package service

import (
	"context"
	"fmt"

	"geosearch-poc/domain"
	"geosearch-poc/repository"
)

type StoreSearchService struct {
	storeRepo repository.StoreRepository
	h3Indexer H3Indexer
}

func NewStoreSearchService(storeRepo repository.StoreRepository, h3Indexer H3Indexer) *StoreSearchService {
	return &StoreSearchService{
		storeRepo: storeRepo,
		h3Indexer: h3Indexer,
	}
}

// Search performs a store search using H3 for geographic filtering
func (s *StoreSearchService) Search(ctx context.Context, params domain.StoreSearchParams) ([]domain.Store, error) {
	// Get H3 cells within radius
	cells, err := s.h3Indexer.GetCellsInRadius(params.Latitude, params.Longitude, params.Radius)
	if err != nil {
		return nil, fmt.Errorf("failed to get H3 cells: %w", err)
	}

	// Get stores in those cells
	stores, err := s.storeRepo.GetByH3Cells(ctx, cells)
	if err != nil {
		return nil, fmt.Errorf("failed to get stores: %w", err)
	}

	// Calculate distance for each store and filter by delivery radius
	var filteredStores []domain.Store
	for _, store := range stores {
		distance := s.h3Indexer.GetDistance(
			params.Latitude,
			params.Longitude,
			store.Latitude,
			store.Longitude,
		)

		// Only include stores that can deliver to the search location
		if distance <= store.DeliveryRadius {
			store.Distance = distance
			filteredStores = append(filteredStores, store)
		}
	}

	return filteredStores, nil
}
