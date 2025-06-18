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
	"geosearch-poc/service"
	"geosearch-poc/service/vertexai"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupProductHandlerTest(t *testing.T) (*handlers.ProductHandler, *pgxpool.Pool, func()) {
	// Setup database connection
	db, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5433/geosearch_test?sslmode=disable")
	require.NoError(t, err)

	// Clean database
	_, _ = db.Exec(context.Background(), "TRUNCATE TABLE products, stores RESTART IDENTITY CASCADE")

	// Setup repositories
	productRepo := postgres.NewProductRepository(db)
	storeRepo := postgres.NewStoreRepository(db)

	// Setup H3 indexer
	h3Indexer := h3.NewIndexer(9)

	// Setup Vertex AI client (mock for testing)
	vertexAIClient := &vertexai.Client{} // Mock client

	// Setup services
	productSearchService := service.NewProductSearchService(productRepo, storeRepo, h3Indexer, vertexAIClient)

	// Setup handler
	handler := handlers.NewProductHandler(productRepo, productSearchService)

	// Cleanup function
	cleanup := func() {
		db.Close()
	}

	return handler, db, cleanup
}

func TestProductHandler_Create(t *testing.T) {
	handler, db, cleanup := setupProductHandlerTest(t)
	defer cleanup()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/products", handler.Create)

	// Create a test store first
	storeRepo := postgres.NewStoreRepository(db)
	testStore := &domain.Store{
		ID:             uuid.New(),
		Name:           "Test Store for Products",
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

	// Garante que a loja foi persistida
	persistedStore, err := storeRepo.GetByID(context.Background(), testStore.ID)
	require.NoError(t, err)
	require.NotNil(t, persistedStore)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedFields []string
	}{
		{
			name: "should create product successfully",
			payload: map[string]interface{}{
				"store_id":    testStore.ID.String(),
				"name":        "Test Product",
				"description": "A test product description",
				"price":       29.99,
				"category":    "Electronics",
				"brand":       "TestBrand",
				"sku":         "TEST-SKU-001",
				"stock":       100,
				"h3_index":    "8928308280fffff",
				"images":      []string{"https://example.com/image1.jpg"},
			},
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"id", "store_id", "name", "description", "price", "category", "brand", "sku", "stock", "h3_index", "images"},
		},
		{
			name: "should create product with minimal fields",
			payload: map[string]interface{}{
				"store_id": testStore.ID.String(),
				"name":     "Minimal Product",
				"price":    19.99,
				"sku":      "MIN-SKU-001",
				"stock":    50,
			},
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"id", "store_id", "name", "price", "sku", "stock"},
		},
		{
			name: "should return error for invalid store_id",
			payload: map[string]interface{}{
				"store_id": "invalid-uuid",
				"name":     "Test Product",
				"price":    29.99,
				"sku":      "TEST-SKU-002",
				"stock":    100,
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name: "should return error for missing required fields",
			payload: map[string]interface{}{
				"name":  "Test Product",
				"price": 29.99,
				// Missing store_id, sku, stock
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name: "should return error for non-existent store",
			payload: map[string]interface{}{
				"store_id": uuid.New().String(),
				"name":     "Test Product",
				"price":    29.99,
				"sku":      "TEST-SKU-003",
				"stock":    100,
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean database before each test to avoid interference
			_, _ = db.Exec(context.Background(), "TRUNCATE TABLE products, stores RESTART IDENTITY CASCADE")

			// Recreate the test store for this test
			err := storeRepo.Create(context.Background(), testStore)
			require.NoError(t, err)

			// Create request
			payloadBytes, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check expected fields
			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field)
			}

			// Additional assertions for successful creation
			if tt.expectedStatus == http.StatusCreated {
				assert.NotEmpty(t, response["id"])
				assert.Equal(t, tt.payload["name"], response["name"])
				assert.Equal(t, tt.payload["store_id"], response["store_id"])
				assert.NotEmpty(t, response["created_at"])
				assert.NotEmpty(t, response["updated_at"])
			}
		})
	}
}

func TestProductHandler_GetByID(t *testing.T) {
	handler, db, cleanup := setupProductHandlerTest(t)
	defer cleanup()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/products/:id", handler.GetByID)

	// Create a test store and product
	storeRepo := postgres.NewStoreRepository(db)
	productRepo := postgres.NewProductRepository(db)

	tests := []struct {
		name           string
		productID      string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "should get product by valid ID",
			productID:      "valid-uuid",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"id", "store_id", "name", "description", "price", "category", "brand", "sku", "stock", "h3_index", "images"},
		},
		{
			name:           "should return error for invalid UUID",
			productID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name:           "should return error for non-existent product",
			productID:      uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			expectedFields: []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean database and recreate test data for each test
			_, _ = db.Exec(context.Background(), "TRUNCATE TABLE products, stores RESTART IDENTITY CASCADE")

			// Create new test store and product for each test
			testStore := &domain.Store{
				ID:             uuid.New(),
				Name:           "Test Store for Product Get",
				CategoryID:     uuid.Nil,
				Latitude:       -23.550520,
				Longitude:      -46.633308,
				H3Index:        "8928308280fffff",
				Address:        "Rua Teste, 123",
				DeliveryRadius: 5.0,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			testProduct := domain.NewProduct(
				testStore.ID,
				"Test Product for Get",
				"Test product description",
				29.99,
				"Electronics",
				"TestBrand",
				"TEST-SKU-GET-001",
				100,
				"8928308280fffff",
				[]string{"https://example.com/image1.jpg"},
			)

			// Recreate the test store and product
			err := storeRepo.Create(context.Background(), testStore)
			require.NoError(t, err)

			err = productRepo.Create(context.Background(), testProduct)
			require.NoError(t, err)

			// Use actual product ID for the first test case
			productID := tt.productID
			if tt.productID == "valid-uuid" {
				productID = testProduct.ID.String()
			}

			// Create request
			req := httptest.NewRequest("GET", fmt.Sprintf("/products/%s", productID), nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check expected fields
			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field)
			}

			// Additional assertions for successful retrieval
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, testProduct.ID.String(), response["id"])
				assert.Equal(t, testProduct.Name, response["name"])
				assert.Equal(t, testProduct.StoreID.String(), response["store_id"])
			}
		})
	}
}

func TestProductHandler_Update(t *testing.T) {
	handler, db, cleanup := setupProductHandlerTest(t)
	defer cleanup()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/products/:id", handler.Update)

	// Create a test store and product
	storeRepo := postgres.NewStoreRepository(db)
	productRepo := postgres.NewProductRepository(db)

	tests := []struct {
		name           string
		productID      string
		payload        map[string]interface{}
		expectedStatus int
		expectedFields []string
	}{
		{
			name:      "should update product successfully",
			productID: "valid-uuid", // Will be replaced with actual product ID
			payload: map[string]interface{}{
				"name":        "Updated Product Name",
				"description": "Updated product description",
				"price":       39.99,
				"category":    "Updated Electronics",
				"brand":       "UpdatedBrand",
				"sku":         "UPDATED-SKU-001",
				"stock":       150,
				"h3_index":    "8928308280fffff",
				"images":      []string{"https://example.com/updated-image.jpg"},
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"id", "store_id", "name", "description", "price", "category", "brand", "sku", "stock", "h3_index", "images"},
		},
		{
			name:      "should update product with partial fields",
			productID: "valid-uuid", // Will be replaced with actual product ID
			payload: map[string]interface{}{
				"name":  "Partially Updated Product",
				"price": 49.99,
				"stock": 200,
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"id", "store_id", "name", "price", "stock"},
		},
		{
			name:           "should return error for invalid UUID",
			productID:      "invalid-uuid",
			payload:        map[string]interface{}{"name": "Test"},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name:           "should return error for non-existent product",
			productID:      uuid.New().String(),
			payload:        map[string]interface{}{"name": "Test"},
			expectedStatus: http.StatusNotFound,
			expectedFields: []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean database and recreate test data for each test
			_, _ = db.Exec(context.Background(), "TRUNCATE TABLE products, stores RESTART IDENTITY CASCADE")

			// Create new test store and product for each test
			testStore := &domain.Store{
				ID:             uuid.New(),
				Name:           "Test Store for Product Update",
				CategoryID:     uuid.Nil,
				Latitude:       -23.550520,
				Longitude:      -46.633308,
				H3Index:        "8928308280fffff",
				Address:        "Rua Teste, 123",
				DeliveryRadius: 5.0,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			testProduct := domain.NewProduct(
				testStore.ID,
				"Test Product for Update",
				"Test product description",
				29.99,
				"Electronics",
				"TestBrand",
				"TEST-SKU-UPDATE-001",
				100,
				"8928308280fffff",
				[]string{"https://example.com/image1.jpg"},
			)

			// Recreate the test store and product
			err := storeRepo.Create(context.Background(), testStore)
			require.NoError(t, err)

			err = productRepo.Create(context.Background(), testProduct)
			require.NoError(t, err)

			// Use actual product ID for valid test cases
			productID := tt.productID
			if tt.productID == "valid-uuid" {
				productID = testProduct.ID.String()
			}

			// Create request
			payloadBytes, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("PUT", fmt.Sprintf("/products/%s", productID), bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check expected fields
			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field)
			}

			// Additional assertions for successful update
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, testProduct.ID.String(), response["id"])
				if name, ok := tt.payload["name"].(string); ok {
					assert.Equal(t, name, response["name"])
				}
				assert.NotEmpty(t, response["updated_at"])
			}
		})
	}
}

func TestProductHandler_Delete(t *testing.T) {
	handler, db, cleanup := setupProductHandlerTest(t)
	defer cleanup()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/products/:id", handler.Delete)

	// Create a test store and product
	storeRepo := postgres.NewStoreRepository(db)
	productRepo := postgres.NewProductRepository(db)

	tests := []struct {
		name           string
		productID      string
		expectedStatus int
	}{
		{
			name:           "should delete product successfully",
			productID:      "valid-uuid", // Will be replaced with actual product ID
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "should return error for invalid UUID",
			productID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should return error for non-existent product",
			productID:      uuid.New().String(),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean database and recreate test data for each test
			_, _ = db.Exec(context.Background(), "TRUNCATE TABLE products, stores RESTART IDENTITY CASCADE")

			// Create new test store and product for each test
			testStore := &domain.Store{
				ID:             uuid.New(),
				Name:           "Test Store for Product Delete",
				CategoryID:     uuid.Nil,
				Latitude:       -23.550520,
				Longitude:      -46.633308,
				H3Index:        "8928308280fffff",
				Address:        "Rua Teste, 123",
				DeliveryRadius: 5.0,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			testProduct := domain.NewProduct(
				testStore.ID,
				"Test Product for Delete",
				"Test product description",
				29.99,
				"Electronics",
				"TestBrand",
				"TEST-SKU-DELETE-001",
				100,
				"8928308280fffff",
				[]string{"https://example.com/image1.jpg"},
			)

			// Recreate the test store and product
			err := storeRepo.Create(context.Background(), testStore)
			require.NoError(t, err)

			err = productRepo.Create(context.Background(), testProduct)
			require.NoError(t, err)

			// Use actual product ID for valid test case
			productID := tt.productID
			if tt.productID == "valid-uuid" {
				productID = testProduct.ID.String()
			}

			// Create request
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/products/%s", productID), nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// For successful deletion, verify product is actually deleted
			if tt.expectedStatus == http.StatusNoContent {
				_, err := productRepo.GetByID(context.Background(), testProduct.ID)
				assert.Error(t, err) // Should return error as product is deleted
			}
		})
	}
}

func TestProductHandler_Search(t *testing.T) {
	handler, _, cleanup := setupProductHandlerTest(t)
	defer cleanup()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/products/search", handler.Search)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "should return error for missing query parameter",
			queryParams:    "lat=-23.550520&lng=-46.633308&radius=5.0",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name:           "should return error for missing latitude",
			queryParams:    "q=test&lng=-46.633308&radius=5.0",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name:           "should return error for missing longitude",
			queryParams:    "q=test&lat=-23.550520&radius=5.0",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name:           "should return error for missing radius",
			queryParams:    "q=test&lat=-23.550520&lng=-46.633308",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name:           "should return error for invalid latitude",
			queryParams:    "q=test&lat=invalid&lng=-46.633308&radius=5.0",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name:           "should return error for invalid longitude",
			queryParams:    "q=test&lat=-23.550520&lng=invalid&radius=5.0",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name:           "should return error for invalid radius",
			queryParams:    "q=test&lat=-23.550520&lng=-46.633308&radius=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", fmt.Sprintf("/products/search?%s", tt.queryParams), nil)

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
		})
	}
}
