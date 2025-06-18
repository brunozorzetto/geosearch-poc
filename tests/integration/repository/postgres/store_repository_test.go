package postgres

import (
	"context"
	"math"
	"testing"

	"geosearch-poc/config"
	"geosearch-poc/domain"
	"geosearch-poc/pkg/h3"
	"geosearch-poc/repository/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreRepository_GetByH3Indexes(t *testing.T) {
	// Setup
	ctx := context.Background()
	dbConfig := config.NewTestDatabaseConfig()
	pool, err := dbConfig.NewPool()
	require.NoError(t, err)
	defer pool.Close()

	// Clean database
	_, _ = pool.Exec(ctx, "TRUNCATE TABLE products, stores RESTART IDENTITY CASCADE")

	repo := postgres.NewStoreRepository(pool)
	indexer := h3.NewIndexer(9)

	// Create test stores
	stores := []*domain.Store{
		domain.NewStore(
			"Store 1",
			uuid.Nil,
			-23.550520,
			-46.633308,
			func() string {
				h3Index, err := indexer.GetCellFromLatLng(-23.550520, -46.633308)
				require.NoError(t, err)
				return h3Index
			}(),
			"Rua Teste, 123",
			5000, // 5km
		),
		domain.NewStore(
			"Store 2",
			uuid.Nil,
			-23.560520,
			-46.643308,
			func() string {
				h3Index, err := indexer.GetCellFromLatLng(-23.560520, -46.643308)
				require.NoError(t, err)
				return h3Index
			}(),
			"Rua Teste, 456",
			3000, // 3km
		),
	}

	// Test cases
	tests := []struct {
		name      string
		h3Indexes []string
		want      int
		check     func(t *testing.T, stores []*domain.Store)
	}{
		{
			name:      "should return store for valid H3 index",
			h3Indexes: []string{stores[0].H3Index},
			want:      1,
			check: func(t *testing.T, stores []*domain.Store) {
				require.Len(t, stores, 1)
				assert.Equal(t, "Store 1", stores[0].Name)
				assert.Equal(t, float64(5000), stores[0].DeliveryRadius)
			},
		},
		{
			name:      "should return no stores for invalid H3 index",
			h3Indexes: []string{"invalid_h3_index"},
			want:      0,
			check: func(t *testing.T, stores []*domain.Store) {
				require.Empty(t, stores)
			},
		},
		{
			name:      "should return multiple stores for multiple H3 indexes",
			h3Indexes: []string{stores[0].H3Index, stores[1].H3Index},
			want:      2,
			check: func(t *testing.T, stores []*domain.Store) {
				require.Len(t, stores, 2)
				names := []string{stores[0].Name, stores[1].Name}
				assert.Contains(t, names, "Store 1")
				assert.Contains(t, names, "Store 2")
				// Check delivery radius
				for _, store := range stores {
					if store.Name == "Store 1" {
						assert.Equal(t, float64(5000), store.DeliveryRadius)
					} else if store.Name == "Store 2" {
						assert.Equal(t, float64(3000), store.DeliveryRadius)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean database before each test to avoid interference
			_, _ = pool.Exec(ctx, "TRUNCATE TABLE products, stores RESTART IDENTITY CASCADE")

			// Recreate test stores
			for _, store := range stores {
				err := repo.Create(ctx, store)
				require.NoError(t, err)
			}

			got, err := repo.GetByH3Indexes(ctx, tt.h3Indexes)
			require.NoError(t, err)
			assert.Len(t, got, tt.want)

			// Run additional checks if provided
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestStoreRepository_CRUD(t *testing.T) {
	// Setup
	ctx := context.Background()
	dbConfig := config.NewTestDatabaseConfig()
	pool, err := dbConfig.NewPool()
	require.NoError(t, err)
	defer pool.Close()

	// Clean database
	_, _ = pool.Exec(ctx, "TRUNCATE TABLE products, stores RESTART IDENTITY CASCADE")

	repo := postgres.NewStoreRepository(pool)
	indexer := h3.NewIndexer(9)

	// Test cases
	tests := []struct {
		name    string
		setup   func() *domain.Store
		action  func(t *testing.T, store *domain.Store)
		verify  func(t *testing.T, store *domain.Store)
		cleanup func(t *testing.T, store *domain.Store)
	}{
		{
			name: "should create and read store",
			setup: func() *domain.Store {
				return domain.NewStore(
					"Test Store",
					uuid.Nil,
					-23.550520,
					-46.633308,
					func() string {
						h3Index, err := indexer.GetCellFromLatLng(-23.550520, -46.633308)
						require.NoError(t, err)
						return h3Index
					}(),
					"Rua Teste, 123",
					5000, // 5km
				)
			},
			action: func(t *testing.T, store *domain.Store) {
				err := repo.Create(ctx, store)
				require.NoError(t, err)
				require.NotEmpty(t, store.ID)
			},
			verify: func(t *testing.T, store *domain.Store) {
				got, err := repo.GetByID(ctx, store.ID)
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, store.Name, got.Name)
				assert.Equal(t, store.Latitude, got.Latitude)
				assert.Equal(t, store.Longitude, got.Longitude)
				assert.Equal(t, store.H3Index, got.H3Index)
				assert.Equal(t, store.DeliveryRadius, got.DeliveryRadius)
			},
			cleanup: func(t *testing.T, store *domain.Store) {
				err := repo.Delete(ctx, store.ID)
				require.NoError(t, err)
			},
		},
		{
			name: "should update store",
			setup: func() *domain.Store {
				store := domain.NewStore(
					"Test Store",
					uuid.Nil,
					-23.550520,
					-46.633308,
					func() string {
						h3Index, err := indexer.GetCellFromLatLng(-23.550520, -46.633308)
						require.NoError(t, err)
						return h3Index
					}(),
					"Rua Teste, 123",
					5000, // 5km
				)
				err := repo.Create(ctx, store)
				require.NoError(t, err)
				return store
			},
			action: func(t *testing.T, store *domain.Store) {
				store.Name = "Updated Store"
				store.DeliveryRadius = 3000 // Update to 3km
				err := repo.Update(ctx, store)
				require.NoError(t, err)
			},
			verify: func(t *testing.T, store *domain.Store) {
				got, err := repo.GetByID(ctx, store.ID)
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, "Updated Store", got.Name)
				assert.Equal(t, float64(3000), got.DeliveryRadius)
			},
			cleanup: func(t *testing.T, store *domain.Store) {
				err := repo.Delete(ctx, store.ID)
				require.NoError(t, err)
			},
		},
		{
			name: "should delete store",
			setup: func() *domain.Store {
				store := domain.NewStore(
					"Test Store",
					uuid.Nil,
					-23.550520,
					-46.633308,
					func() string {
						h3Index, err := indexer.GetCellFromLatLng(-23.550520, -46.633308)
						require.NoError(t, err)
						return h3Index
					}(),
					"Rua Teste, 123",
					5000, // 5km
				)
				err := repo.Create(ctx, store)
				require.NoError(t, err)
				return store
			},
			action: func(t *testing.T, store *domain.Store) {
				err := repo.Delete(ctx, store.ID)
				require.NoError(t, err)
			},
			verify: func(t *testing.T, store *domain.Store) {
				got, err := repo.GetByID(ctx, store.ID)
				require.Error(t, err)
				assert.Nil(t, got)
			},
			cleanup: func(t *testing.T, store *domain.Store) {
				// No cleanup needed as store is already deleted
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			store := tt.setup()

			// Execute action
			tt.action(t, store)

			// Verify results
			tt.verify(t, store)

			// Cleanup
			tt.cleanup(t, store)
		})
	}
}

func TestStoreRepository_SearchByRadius(t *testing.T) {
	// Setup
	ctx := context.Background()
	dbConfig := config.NewTestDatabaseConfig()
	pool, err := dbConfig.NewPool()
	require.NoError(t, err)
	defer pool.Close()

	// Clean database
	_, _ = pool.Exec(ctx, "TRUNCATE TABLE products, stores RESTART IDENTITY CASCADE")

	repo := postgres.NewStoreRepository(pool)
	indexer := h3.NewIndexer(9)

	// Clean up stores table before test
	_, err = pool.Exec(ctx, "DELETE FROM stores")
	require.NoError(t, err)

	// Create test stores
	stores := []*domain.Store{
		domain.NewStore(
			"Store 1",
			uuid.Nil,
			-23.550520,
			-46.633308,
			func() string {
				h3Index, err := indexer.GetCellFromLatLng(-23.550520, -46.633308)
				require.NoError(t, err)
				return h3Index
			}(),
			"Rua Teste, 123",
			1000, // 1km
		),
		domain.NewStore(
			"Store 2",
			uuid.Nil,
			-23.560520,
			-46.643308,
			func() string {
				h3Index, err := indexer.GetCellFromLatLng(-23.560520, -46.643308)
				require.NoError(t, err)
				return h3Index
			}(),
			"Rua Teste, 456",
			2000, // 2km
		),
		domain.NewStore(
			"Store 3",
			uuid.Nil,
			-23.570520,
			-46.653308,
			func() string {
				h3Index, err := indexer.GetCellFromLatLng(-23.570520, -46.653308)
				require.NoError(t, err)
				return h3Index
			}(),
			"Rua Teste, 789",
			500, // 0.5km, não deve aparecer nos testes
		),
	}

	for _, store := range stores {
		err := repo.Create(ctx, store)
		require.NoError(t, err)
	}

	// Test cases
	tests := []struct {
		name      string
		latitude  float64
		longitude float64
		radius    float64
		want      int
		check     func(t *testing.T, stores []*domain.Store)
	}{
		{
			name:      "should return stores within 1km radius",
			latitude:  -23.550520,
			longitude: -46.633308,
			radius:    1.0,
			want:      1,
			check: func(t *testing.T, stores []*domain.Store) {
				require.Len(t, stores, 1)
				assert.Equal(t, "Store 1", stores[0].Name)
				distance := calculateDistance(-23.550520, -46.633308, stores[0].Latitude, stores[0].Longitude)
				assert.LessOrEqual(t, distance, 1.0)
			},
		},
		{
			name:      "should return stores within 2km radius",
			latitude:  -23.550520,
			longitude: -46.633308,
			radius:    2.0,
			want:      2,
			check: func(t *testing.T, stores []*domain.Store) {
				require.Len(t, stores, 2)
				names := []string{stores[0].Name, stores[1].Name}
				assert.Contains(t, names, "Store 1")
				assert.Contains(t, names, "Store 2")
				for _, store := range stores {
					distance := calculateDistance(-23.550520, -46.633308, store.Latitude, store.Longitude)
					assert.LessOrEqual(t, distance, 2.0)
				}
			},
		},
		{
			name:      "should return no stores for far location",
			latitude:  -23.000000,
			longitude: -46.000000,
			radius:    1.0,
			want:      0,
			check: func(t *testing.T, stores []*domain.Store) {
				require.Empty(t, stores)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := domain.NewSearchParams(tt.latitude, tt.longitude, tt.radius)
			got, total, err := repo.Search(ctx, params)
			require.NoError(t, err)
			assert.Equal(t, tt.want, total)
			assert.Len(t, got, tt.want)

			// Run additional checks if provided
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

// calculateDistance calculates the distance between two points in kilometers
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth's radius in kilometers

	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
