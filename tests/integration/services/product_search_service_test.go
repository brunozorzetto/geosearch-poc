package services

import (
	"context"
	"testing"
	"time"

	"geosearch-poc/domain"
	"geosearch-poc/pkg/h3"
	"geosearch-poc/repository/postgres"
	"geosearch-poc/service"
	"geosearch-poc/service/vertexai"

	retail "cloud.google.com/go/retail/apiv2/retailpb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockVertexAIClient is a mock implementation of the Vertex AI client for testing
type MockVertexAIClient struct {
	searchResults []vertexai.SearchResult
	searchError   error
	createError   error
	updateError   error
	deleteError   error
}

func (m *MockVertexAIClient) Search(ctx context.Context, params vertexai.SearchParams) ([]vertexai.SearchResult, string, error) {
	if m.searchError != nil {
		return nil, "", m.searchError
	}
	return m.searchResults, "", nil
}

func (m *MockVertexAIClient) CreateProduct(ctx context.Context, product *retail.Product) error {
	return m.createError
}

func (m *MockVertexAIClient) UpdateProduct(ctx context.Context, product *retail.Product) error {
	return m.updateError
}

func (m *MockVertexAIClient) DeleteProduct(ctx context.Context, productID string) error {
	return m.deleteError
}

func (m *MockVertexAIClient) Close() error {
	return nil
}

func setupProductSearchServiceTest(t *testing.T) (*service.ProductSearchService, *pgxpool.Pool, *MockVertexAIClient, func()) {
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

	// Setup mock Vertex AI client
	mockVertexAI := &MockVertexAIClient{}

	// Setup service
	productSearchService := service.NewProductSearchService(productRepo, storeRepo, h3Indexer, mockVertexAI)

	// Cleanup function
	cleanup := func() {
		db.Close()
	}

	return productSearchService, db, mockVertexAI, cleanup
}

func TestProductSearchService_SearchProducts(t *testing.T) {
	service, db, mockVertexAI, cleanup := setupProductSearchServiceTest(t)
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

	err := storeRepo.Create(context.Background(), store1)
	require.NoError(t, err)
	err = storeRepo.Create(context.Background(), store2)
	require.NoError(t, err)

	// Mock search results
	mockResults := []vertexai.SearchResult{
		{
			ID:          uuid.New().String(),
			Title:       "Test Product 1",
			Description: "Test product description 1",
			PriceInfo: vertexai.PriceInfo{
				Price:    29.99,
				Currency: "BRL",
			},
			Categories: []string{"Electronics"},
			Brands:     []string{"TestBrand"},
			Availability: vertexai.Availability{
				Stock:     100,
				Available: true,
			},
			Attributes: map[string]string{
				"store_id": store1.ID.String(),
				"h3_index": store1.H3Index,
			},
			Images:    []string{"https://example.com/image1.jpg"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Title:       "Test Product 2",
			Description: "Test product description 2",
			PriceInfo: vertexai.PriceInfo{
				Price:    39.99,
				Currency: "BRL",
			},
			Categories: []string{"Electronics"},
			Brands:     []string{"TestBrand2"},
			Availability: vertexai.Availability{
				Stock:     50,
				Available: true,
			},
			Attributes: map[string]string{
				"store_id": store2.ID.String(),
				"h3_index": store2.H3Index,
			},
			Images:    []string{"https://example.com/image2.jpg"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	tests := []struct {
		name          string
		searchParams  *domain.ProductSearchParams
		expectedCount int
		expectedError bool
		setupMock     func()
	}{
		{
			name: "should search products successfully",
			searchParams: &domain.ProductSearchParams{
				Query:     "test product",
				Latitude:  -23.550520,
				Longitude: -46.633308,
				Radius:    5.0,
				PageSize:  10,
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "should search products with category filter",
			searchParams: &domain.ProductSearchParams{
				Query:     "test product",
				Latitude:  -23.550520,
				Longitude: -46.633308,
				Radius:    5.0,
				Category:  "Electronics",
				PageSize:  10,
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "should search products with brand filter",
			searchParams: &domain.ProductSearchParams{
				Query:     "test product",
				Latitude:  -23.550520,
				Longitude: -46.633308,
				Radius:    5.0,
				Brand:     "TestBrand",
				PageSize:  10,
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "should search products with price filters",
			searchParams: &domain.ProductSearchParams{
				Query:     "test product",
				Latitude:  -23.550520,
				Longitude: -46.633308,
				Radius:    5.0,
				MinPrice:  20.0,
				MaxPrice:  50.0,
				PageSize:  10,
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "should return empty results when no stores in radius",
			searchParams: &domain.ProductSearchParams{
				Query:     "test product",
				Latitude:  -23.600520, // Far from stores
				Longitude: -46.683308,
				Radius:    1.0, // Small radius
				PageSize:  10,
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "should return error when Vertex AI search fails",
			searchParams: &domain.ProductSearchParams{
				Query:     "test product",
				Latitude:  -23.550520,
				Longitude: -46.633308,
				Radius:    5.0,
				PageSize:  10,
			},
			expectedCount: 0,
			expectedError: true,
			setupMock: func() {
				mockVertexAI.searchError = assert.AnError
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock if needed
			if tt.setupMock != nil {
				tt.setupMock()
			} else {
				// Reset mock to default state
				mockVertexAI.searchError = nil
				mockVertexAI.searchResults = mockResults
			}

			// Execute search
			result, err := service.SearchProducts(context.Background(), tt.searchParams)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Products, tt.expectedCount)
				assert.Equal(t, int64(tt.expectedCount), result.TotalCount)
			}
		})
	}
}

func TestProductSearchService_GetVertexAIClient(t *testing.T) {
	service, _, mockVertexAI, cleanup := setupProductSearchServiceTest(t)
	defer cleanup()

	// Test that the service returns the correct Vertex AI client
	client := service.GetVertexAIClient()
	assert.Equal(t, mockVertexAI, client)
}
