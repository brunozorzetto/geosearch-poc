package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"geosearch-poc/domain"
	"geosearch-poc/handlers"
	"geosearch-poc/pkg/h3"
	"geosearch-poc/repository/postgres"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStoreHandlerTest(t *testing.T) (*handlers.StoreHandler, *pgxpool.Pool, func()) {
	// Setup database connection
	db, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5433/geosearch_test?sslmode=disable")
	require.NoError(t, err)

	// Clean database
	_, _ = db.Exec(context.Background(), "TRUNCATE TABLE products, stores RESTART IDENTITY CASCADE")

	// Setup repositories
	storeRepo := postgres.NewStoreRepository(db)

	// Setup H3 indexer
	h3Indexer := h3.NewIndexer(9)

	// Setup handler
	handler := handlers.NewStoreHandler(storeRepo, h3Indexer)

	// Cleanup function
	cleanup := func() {
		db.Close()
	}

	return handler, db, cleanup
}

func TestStoreHandler_Create(t *testing.T) {
	handler, _, cleanup := setupStoreHandlerTest(t)
	defer cleanup()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/stores", handler.Create)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedFields []string
	}{
		{
			name: "should create store successfully",
			payload: map[string]interface{}{
				"name":            "Test Store",
				"latitude":        -23.550520,
				"longitude":       -46.633308,
				"address":         "Rua Teste, 123",
				"delivery_radius": 5.0,
			},
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"id", "name", "latitude", "longitude", "address", "delivery_radius", "h3_index"},
		},
		{
			name: "should create store with category_id",
			payload: map[string]interface{}{
				"name":            "Test Store with Category",
				"category_id":     uuid.New().String(),
				"latitude":        -23.550520,
				"longitude":       -46.633308,
				"address":         "Rua Teste, 456",
				"delivery_radius": 3.0,
			},
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"id", "name", "category_id", "latitude", "longitude", "address", "delivery_radius", "h3_index"},
		},
		{
			name: "should return error for invalid request",
			payload: map[string]interface{}{
				"name": "Test Store",
				// Missing required fields
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name: "should return error for invalid category_id",
			payload: map[string]interface{}{
				"name":            "Test Store",
				"category_id":     "invalid-uuid",
				"latitude":        -23.550520,
				"longitude":       -46.633308,
				"address":         "Rua Teste, 789",
				"delivery_radius": 2.0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			payloadBytes, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/stores", bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")

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

			// Additional assertions for successful creation
			if tt.expectedStatus == http.StatusCreated {
				assert.NotEmpty(t, response["id"])
				assert.Equal(t, tt.payload["name"], response["name"])
				assert.NotEmpty(t, response["h3_index"])
				assert.NotEmpty(t, response["created_at"])
				assert.NotEmpty(t, response["updated_at"])
			}
		})
	}
}

func TestStoreHandler_GetByID(t *testing.T) {
	handler, db, cleanup := setupStoreHandlerTest(t)
	defer cleanup()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/stores/:id", handler.GetByID)

	// Create a test store first
	storeRepo := postgres.NewStoreRepository(db)
	testStore := &domain.Store{
		ID:             uuid.New(),
		Name:           "Test Store for Get",
		CategoryID:     uuid.Nil,
		Latitude:       -23.550520,
		Longitude:      -46.633308,
		H3Index:        "8928308280fffff",
		Address:        "Rua Teste, 123",
		DeliveryRadius: 5.0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := storeRepo.Create(context.Background(), testStore)
	require.NoError(t, err)

	tests := []struct {
		name           string
		storeID        string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "should get store by valid ID",
			storeID:        testStore.ID.String(),
			expectedStatus: http.StatusOK,
			expectedFields: []string{"id", "name", "latitude", "longitude", "address", "delivery_radius", "h3_index"},
		},
		{
			name:           "should return error for invalid UUID",
			storeID:        "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name:           "should return error for non-existent store",
			storeID:        uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			expectedFields: []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", fmt.Sprintf("/stores/%s", tt.storeID), nil)

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

			// Additional assertions for successful retrieval
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, testStore.ID.String(), response["id"])
				assert.Equal(t, testStore.Name, response["name"])
			}
		})
	}
}

func TestStoreHandler_Update(t *testing.T) {
	handler, db, cleanup := setupStoreHandlerTest(t)
	defer cleanup()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/stores/:id", handler.Update)

	// Create a test store first
	storeRepo := postgres.NewStoreRepository(db)
	testStore := &domain.Store{
		ID:             uuid.New(),
		Name:           "Test Store for Update",
		CategoryID:     uuid.Nil,
		Latitude:       -23.550520,
		Longitude:      -46.633308,
		H3Index:        "8928308280fffff",
		Address:        "Rua Teste, 123",
		DeliveryRadius: 5.0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := storeRepo.Create(context.Background(), testStore)
	require.NoError(t, err)

	tests := []struct {
		name           string
		storeID        string
		payload        map[string]interface{}
		expectedStatus int
		expectedFields []string
	}{
		{
			name:    "should update store successfully",
			storeID: testStore.ID.String(),
			payload: map[string]interface{}{
				"name":            "Updated Store Name",
				"latitude":        -23.560520,
				"longitude":       -46.643308,
				"address":         "Rua Atualizada, 456",
				"delivery_radius": 7.0,
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"id", "name", "latitude", "longitude", "address", "delivery_radius", "h3_index"},
		},
		{
			name:    "should update store with category_id",
			storeID: testStore.ID.String(),
			payload: map[string]interface{}{
				"name":            "Updated Store with Category",
				"category_id":     uuid.New().String(),
				"latitude":        -23.550520,
				"longitude":       -46.633308,
				"address":         "Rua Teste, 789",
				"delivery_radius": 4.0,
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"id", "name", "category_id", "latitude", "longitude", "address", "delivery_radius", "h3_index"},
		},
		{
			name:           "should return error for invalid UUID",
			storeID:        "invalid-uuid",
			payload:        map[string]interface{}{"name": "Test"},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name:           "should return error for non-existent store",
			storeID:        uuid.New().String(),
			payload:        map[string]interface{}{"name": "Test"},
			expectedStatus: http.StatusNotFound,
			expectedFields: []string{"error"},
		},
		{
			name:    "should return error for invalid category_id",
			storeID: testStore.ID.String(),
			payload: map[string]interface{}{
				"name":        "Test Store",
				"category_id": "invalid-uuid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			payloadBytes, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("PUT", fmt.Sprintf("/stores/%s", tt.storeID), bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")

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

			// Additional assertions for successful update
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, testStore.ID.String(), response["id"])
				if name, ok := tt.payload["name"].(string); ok {
					assert.Equal(t, name, response["name"])
				}
				assert.NotEmpty(t, response["updated_at"])
			}
		})
	}
}

func TestStoreHandler_Delete(t *testing.T) {
	handler, db, cleanup := setupStoreHandlerTest(t)
	defer cleanup()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/stores/:id", handler.Delete)

	// Create a test store first
	storeRepo := postgres.NewStoreRepository(db)
	testStore := &domain.Store{
		ID:             uuid.New(),
		Name:           "Test Store for Delete",
		CategoryID:     uuid.Nil,
		Latitude:       -23.550520,
		Longitude:      -46.633308,
		H3Index:        "8928308280fffff",
		Address:        "Rua Teste, 123",
		DeliveryRadius: 5.0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := storeRepo.Create(context.Background(), testStore)
	require.NoError(t, err)

	tests := []struct {
		name           string
		storeID        string
		expectedStatus int
	}{
		{
			name:           "should delete store successfully",
			storeID:        testStore.ID.String(),
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "should return error for invalid UUID",
			storeID:        "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should return error for non-existent store",
			storeID:        uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/stores/%s", tt.storeID), nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// For successful deletion, verify store is actually deleted
			if tt.expectedStatus == http.StatusNoContent {
				_, err := storeRepo.GetByID(context.Background(), testStore.ID)
				assert.Error(t, err) // Should return error as store is deleted
			}
		})
	}
}
