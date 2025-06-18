package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"geosearch-poc/domain"
	"geosearch-poc/handlers"
	"geosearch-poc/pkg/h3"
	"geosearch-poc/repository/postgres"
	"geosearch-poc/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSearchHandlerTest(t *testing.T) (*handlers.SearchHandler, *pgxpool.Pool, func()) {
	// Setup database connection
	db, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5433/geosearch_test?sslmode=disable")
	require.NoError(t, err)

	// Clean database
	_, _ = db.Exec(context.Background(), "TRUNCATE TABLE products, stores RESTART IDENTITY CASCADE")

	// Setup repositories
	storeRepo := postgres.NewStoreRepository(db)

	// Setup H3 indexer
	h3Indexer := h3.NewIndexer(9)

	// Setup services
	storeSearchService := service.NewStoreSearchService(storeRepo, h3Indexer)

	// Setup handler
	handler := handlers.NewSearchHandler(storeSearchService)

	// Cleanup function
	cleanup := func() {
		db.Close()
	}

	return handler, db, cleanup
}

func TestSearchHandler_Search(t *testing.T) {
	handler, db, cleanup := setupSearchHandlerTest(t)
	defer cleanup()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/search", handler.Search)

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

	// Verify stores were created
	createdStore1, err := storeRepo.GetByID(context.Background(), store1.ID)
	require.NoError(t, err)
	require.NotNil(t, createdStore1)

	createdStore2, err := storeRepo.GetByID(context.Background(), store2.ID)
	require.NoError(t, err)
	require.NotNil(t, createdStore2)

	createdStore3, err := storeRepo.GetByID(context.Background(), store3.ID)
	require.NoError(t, err)
	require.NotNil(t, createdStore3)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedFields []string
		expectedStores int
	}{
		{
			name:           "should search stores within 1km radius",
			queryParams:    "lat=-23.550520&lng=-46.633308&radius=1.0",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"stores", "total"},
			expectedStores: 2, // store1 and store2 should be within 1km
		},
		{
			name:           "should search stores within 5km radius",
			queryParams:    "lat=-23.550520&lng=-46.633308&radius=5.0",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"stores", "total"},
			expectedStores: 2, // store1 and store2 should be within 5km, store3 is at 7.54km
		},
		{
			name:           "should search stores from different location",
			queryParams:    "lat=-23.600520&lng=-46.683308&radius=1.0",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"stores", "total"},
			expectedStores: 1, // only store3 should be within 1km
		},
		{
			name:           "should return error for missing latitude",
			queryParams:    "lng=-46.633308&radius=1.0",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
			expectedStores: 0,
		},
		{
			name:           "should return error for missing longitude",
			queryParams:    "lat=-23.550520&radius=1.0",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
			expectedStores: 0,
		},
		{
			name:           "should return error for missing radius",
			queryParams:    "lat=-23.550520&lng=-46.633308",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
			expectedStores: 0,
		},
		{
			name:           "should return error for invalid latitude",
			queryParams:    "lat=invalid&lng=-46.633308&radius=1.0",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
			expectedStores: 0,
		},
		{
			name:           "should return error for invalid longitude",
			queryParams:    "lat=-23.550520&lng=invalid&radius=1.0",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
			expectedStores: 0,
		},
		{
			name:           "should return error for invalid radius",
			queryParams:    "lat=-23.550520&lng=-46.633308&radius=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
			expectedStores: 0,
		},
		{
			name:           "should return empty results for far location",
			queryParams:    "lat=-24.000000&lng=-47.000000&radius=1.0",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"stores", "total"},
			expectedStores: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/search?"+tt.queryParams, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check expected fields
			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field)
			}

			// Additional assertions for successful search
			if tt.expectedStatus == http.StatusOK {
				stores, ok := response["stores"].([]interface{})
				if !ok {
					// Handle case where stores is null
					stores = []interface{}{}
				}
				assert.Len(t, stores, tt.expectedStores)

				total, ok := response["total"].(float64)
				assert.True(t, ok)
				assert.Equal(t, float64(tt.expectedStores), total)

				// Check store structure for successful searches
				if tt.expectedStores > 0 {
					// Check that all stores have the required fields
					for i, store := range stores {
						storeMap := store.(map[string]interface{})
						expectedStoreFields := []string{"id", "name", "latitude", "longitude", "address", "delivery_radius", "h3_index"}
						for _, field := range expectedStoreFields {
							assert.Contains(t, storeMap, field)
						}

						// Check that distance is calculated for all stores
						distance, ok := storeMap["distance"].(float64)
						assert.True(t, ok, "Store %d should have distance field", i)
						assert.GreaterOrEqual(t, distance, 0.0)
					}
				}
			}
		})
	}
}
