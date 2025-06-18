package services

import (
	"context"
	"testing"

	"geosearch-poc/domain"
	"geosearch-poc/pkg/h3"
	"geosearch-poc/repository/postgres"
	"geosearch-poc/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStoreSearchServiceTest(t *testing.T) (*service.StoreSearchService, *pgxpool.Pool, func()) {
	// Setup database connection
	db, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5433/geosearch_test?sslmode=disable")
	require.NoError(t, err)

	// Clean database more thoroughly
	_, _ = db.Exec(context.Background(), "TRUNCATE TABLE products, stores RESTART IDENTITY CASCADE")
	_, _ = db.Exec(context.Background(), "DELETE FROM products")
	_, _ = db.Exec(context.Background(), "DELETE FROM stores")

	// Setup repositories
	storeRepo := postgres.NewStoreRepository(db)

	// Setup H3 indexer
	h3Indexer := h3.NewIndexer(9)

	// Setup service
	storeSearchService := service.NewStoreSearchService(storeRepo, h3Indexer)

	// Cleanup function
	cleanup := func() {
		db.Close()
	}

	return storeSearchService, db, cleanup
}

func TestStoreSearchService_Search(t *testing.T) {
	service, db, cleanup := setupStoreSearchServiceTest(t)
	defer cleanup()

	// Create test stores
	storeRepo := postgres.NewStoreRepository(db)
	h3Indexer := h3.NewIndexer(9)

	// Store 1 - São Paulo center
	store1 := domain.NewStore(
		"Store São Paulo Center",
		uuid.Nil,
		-23.550520,
		-46.633308,
		func() string {
			h3Index, err := h3Indexer.GetCellFromLatLng(-23.550520, -46.633308)
			require.NoError(t, err)
			return h3Index
		}(),
		"Av. Paulista, 1000",
		5.0,
	)

	// Store 2 - Near São Paulo center (within 1km)
	store2 := domain.NewStore(
		"Store Near Center",
		uuid.Nil,
		-23.551520,
		-46.634308,
		func() string {
			h3Index, err := h3Indexer.GetCellFromLatLng(-23.551520, -46.634308)
			require.NoError(t, err)
			return h3Index
		}(),
		"Rua Augusta, 500",
		3.0,
	)

	// Store 3 - Far from São Paulo center (more than 5km)
	store3 := domain.NewStore(
		"Store Far Away",
		uuid.Nil,
		-23.600520,
		-46.683308,
		func() string {
			h3Index, err := h3Indexer.GetCellFromLatLng(-23.600520, -46.683308)
			require.NoError(t, err)
			return h3Index
		}(),
		"Av. Brigadeiro Faria Lima, 2000",
		2.0,
	)

	err := storeRepo.Create(context.Background(), store1)
	require.NoError(t, err)
	err = storeRepo.Create(context.Background(), store2)
	require.NoError(t, err)
	err = storeRepo.Create(context.Background(), store3)
	require.NoError(t, err)

	tests := []struct {
		name           string
		params         domain.StoreSearchParams
		expectedStores int
		checkStores    func(t *testing.T, stores []domain.Store)
	}{
		{
			name: "should find stores within 1km radius",
			params: domain.StoreSearchParams{
				Latitude:  -23.550520,
				Longitude: -46.633308,
				Radius:    1.0,
			},
			expectedStores: 2, // store1 and store2
			checkStores: func(t *testing.T, stores []domain.Store) {
				assert.Len(t, stores, 2)
				if len(stores) >= 2 {
					storeNames := []string{stores[0].Name, stores[1].Name}
					assert.Contains(t, storeNames, "Store São Paulo Center")
					assert.Contains(t, storeNames, "Store Near Center")

					// Check distances are calculated
					for _, store := range stores {
						assert.GreaterOrEqual(t, store.Distance, 0.0)
						assert.LessOrEqual(t, store.Distance, 1.0)
					}
				}
			},
		},
		{
			name: "should find stores within 5km radius",
			params: domain.StoreSearchParams{
				Latitude:  -23.550520,
				Longitude: -46.633308,
				Radius:    5.0,
			},
			expectedStores: 2, // store1 and store2, store3 is at 7.54km
			checkStores: func(t *testing.T, stores []domain.Store) {
				assert.Len(t, stores, 2)
				if len(stores) >= 2 {
					storeNames := []string{stores[0].Name, stores[1].Name}
					assert.Contains(t, storeNames, "Store São Paulo Center")
					assert.Contains(t, storeNames, "Store Near Center")

					// Check distances are calculated
					for _, store := range stores {
						assert.GreaterOrEqual(t, store.Distance, 0.0)
						assert.LessOrEqual(t, store.Distance, 5.0)
					}
				}
			},
		},
		{
			name: "should find stores from different location",
			params: domain.StoreSearchParams{
				Latitude:  -23.600520,
				Longitude: -46.683308,
				Radius:    1.0,
			},
			expectedStores: 1, // only store3
			checkStores: func(t *testing.T, stores []domain.Store) {
				assert.Len(t, stores, 1)
				if len(stores) >= 1 {
					assert.Equal(t, "Store Far Away", stores[0].Name)
					assert.GreaterOrEqual(t, stores[0].Distance, 0.0)
					assert.LessOrEqual(t, stores[0].Distance, 1.0)
				}
			},
		},
		{
			name: "should return empty results for far location",
			params: domain.StoreSearchParams{
				Latitude:  -24.000000,
				Longitude: -47.000000,
				Radius:    1.0,
			},
			expectedStores: 0,
			checkStores: func(t *testing.T, stores []domain.Store) {
				assert.Empty(t, stores)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute search
			stores, err := service.Search(context.Background(), tt.params)

			// Assert no error
			assert.NoError(t, err)

			// Assert expected number of stores
			assert.Len(t, stores, tt.expectedStores)

			// Run additional checks if provided
			if tt.checkStores != nil {
				tt.checkStores(t, stores)
			}
		})
	}
}

func TestStoreSearchService_SearchWithDeliveryRadius(t *testing.T) {
	service, db, cleanup := setupStoreSearchServiceTest(t)
	defer cleanup()

	// Create test stores with different delivery radii
	storeRepo := postgres.NewStoreRepository(db)
	h3Indexer := h3.NewIndexer(9)

	// Store with large delivery radius
	store1 := domain.NewStore(
		"Store with Large Delivery",
		uuid.Nil,
		-23.550520,
		-46.633308,
		func() string {
			h3Index, err := h3Indexer.GetCellFromLatLng(-23.550520, -46.633308)
			require.NoError(t, err)
			return h3Index
		}(),
		"Av. Paulista, 1000",
		10.0, // 10km delivery radius
	)

	// Store with small delivery radius
	store2 := domain.NewStore(
		"Store with Small Delivery",
		uuid.Nil,
		-23.551520,
		-46.634308,
		func() string {
			h3Index, err := h3Indexer.GetCellFromLatLng(-23.551520, -46.634308)
			require.NoError(t, err)
			return h3Index
		}(),
		"Rua Augusta, 500",
		1.0, // 1km delivery radius
	)

	err := storeRepo.Create(context.Background(), store1)
	require.NoError(t, err)
	err = storeRepo.Create(context.Background(), store2)
	require.NoError(t, err)

	tests := []struct {
		name           string
		params         domain.StoreSearchParams
		expectedStores int
		checkStores    func(t *testing.T, stores []domain.Store)
	}{
		{
			name: "should find stores within search radius and delivery radius",
			params: domain.StoreSearchParams{
				Latitude:  -23.550520,
				Longitude: -46.633308,
				Radius:    5.0, // 5km search radius
			},
			expectedStores: 2, // both stores within 5km
			checkStores: func(t *testing.T, stores []domain.Store) {
				assert.Len(t, stores, 2)
				if len(stores) >= 2 {
					storeNames := []string{stores[0].Name, stores[1].Name}
					assert.Contains(t, storeNames, "Store with Large Delivery")
					assert.Contains(t, storeNames, "Store with Small Delivery")
				}
			},
		},
		{
			name: "should filter by delivery radius when search is far",
			params: domain.StoreSearchParams{
				Latitude:  -23.560520, // 1.1km away
				Longitude: -46.643308,
				Radius:    5.0, // 5km search radius
			},
			expectedStores: 1, // only store1 with large delivery radius
			checkStores: func(t *testing.T, stores []domain.Store) {
				assert.Len(t, stores, 1)
				if len(stores) >= 1 {
					assert.Equal(t, "Store with Large Delivery", stores[0].Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute search
			stores, err := service.Search(context.Background(), tt.params)

			// Assert no error
			assert.NoError(t, err)

			// Assert expected number of stores
			assert.Len(t, stores, tt.expectedStores)

			// Run additional checks if provided
			if tt.checkStores != nil {
				tt.checkStores(t, stores)
			}
		})
	}
}
