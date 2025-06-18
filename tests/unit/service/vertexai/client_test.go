package vertexai_test

import (
	"context"
	"testing"
	"time"

	"geosearch-poc/service/vertexai"

	retail "cloud.google.com/go/retail/apiv2/retailpb"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSearchParams(t *testing.T) {
	params := vertexai.SearchParams{
		Query:     "test query",
		Filter:    "test filter",
		PageSize:  10,
		PageToken: "test-token",
	}

	assert.Equal(t, "test query", params.Query)
	assert.Equal(t, "test filter", params.Filter)
	assert.Equal(t, 10, params.PageSize)
	assert.Equal(t, "test-token", params.PageToken)
}

func TestSearchResult(t *testing.T) {
	now := time.Now()
	result := vertexai.SearchResult{
		ID:          uuid.New().String(),
		Title:       "Test Product",
		Description: "Test product description",
		PriceInfo: vertexai.PriceInfo{
			Price:         29.99,
			Currency:      "BRL",
			OriginalPrice: 39.99,
		},
		Categories: []string{"Electronics", "Gadgets"},
		Brands:     []string{"TestBrand", "Premium"},
		Availability: vertexai.Availability{
			Stock:     100,
			Available: true,
		},
		Attributes: map[string]string{
			"store_id": uuid.New().String(),
			"h3_index": "8928308280fffff",
			"color":    "Black",
		},
		Images:    []string{"https://example.com/image1.jpg", "https://example.com/image2.jpg"},
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "Test Product", result.Title)
	assert.Equal(t, "Test product description", result.Description)
	assert.Equal(t, 29.99, result.PriceInfo.Price)
	assert.Equal(t, "BRL", result.PriceInfo.Currency)
	assert.Equal(t, 39.99, result.PriceInfo.OriginalPrice)
	assert.Equal(t, []string{"Electronics", "Gadgets"}, result.Categories)
	assert.Equal(t, []string{"TestBrand", "Premium"}, result.Brands)
	assert.Equal(t, int64(100), result.Availability.Stock)
	assert.True(t, result.Availability.Available)
	assert.Equal(t, "8928308280fffff", result.Attributes["h3_index"])
	assert.Equal(t, "Black", result.Attributes["color"])
	assert.Equal(t, []string{"https://example.com/image1.jpg", "https://example.com/image2.jpg"}, result.Images)
	assert.Equal(t, now, result.CreatedAt)
	assert.Equal(t, now, result.UpdatedAt)
}

func TestPriceInfo(t *testing.T) {
	priceInfo := vertexai.PriceInfo{
		Price:         29.99,
		Currency:      "BRL",
		OriginalPrice: 39.99,
	}

	assert.Equal(t, 29.99, priceInfo.Price)
	assert.Equal(t, "BRL", priceInfo.Currency)
	assert.Equal(t, 39.99, priceInfo.OriginalPrice)
}

func TestAvailability(t *testing.T) {
	availability := vertexai.Availability{
		Stock:     100,
		Available: true,
	}

	assert.Equal(t, int64(100), availability.Stock)
	assert.True(t, availability.Available)
}

func TestAvailability_OutOfStock(t *testing.T) {
	availability := vertexai.Availability{
		Stock:     0,
		Available: false,
	}

	assert.Equal(t, int64(0), availability.Stock)
	assert.False(t, availability.Available)
}

// MockVertexAIClient for testing
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

func TestMockVertexAIClient_Search(t *testing.T) {
	mock := &MockVertexAIClient{
		searchResults: []vertexai.SearchResult{
			{
				ID:    uuid.New().String(),
				Title: "Test Product",
			},
		},
	}

	params := vertexai.SearchParams{
		Query:    "test",
		PageSize: 10,
	}

	results, token, err := mock.Search(context.Background(), params)

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Test Product", results[0].Title)
	assert.Equal(t, "", token)
}

func TestMockVertexAIClient_SearchError(t *testing.T) {
	mock := &MockVertexAIClient{
		searchError: assert.AnError,
	}

	params := vertexai.SearchParams{
		Query:    "test",
		PageSize: 10,
	}

	results, token, err := mock.Search(context.Background(), params)

	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Equal(t, "", token)
}

func TestMockVertexAIClient_CreateProduct(t *testing.T) {
	mock := &MockVertexAIClient{}

	product := &retail.Product{
		Id:    uuid.New().String(),
		Title: "Test Product",
	}

	err := mock.CreateProduct(context.Background(), product)
	assert.NoError(t, err)
}

func TestMockVertexAIClient_CreateProductError(t *testing.T) {
	mock := &MockVertexAIClient{
		createError: assert.AnError,
	}

	product := &retail.Product{
		Id:    uuid.New().String(),
		Title: "Test Product",
	}

	err := mock.CreateProduct(context.Background(), product)
	assert.Error(t, err)
}

func TestMockVertexAIClient_UpdateProduct(t *testing.T) {
	mock := &MockVertexAIClient{}

	product := &retail.Product{
		Id:    uuid.New().String(),
		Title: "Updated Product",
	}

	err := mock.UpdateProduct(context.Background(), product)
	assert.NoError(t, err)
}

func TestMockVertexAIClient_UpdateProductError(t *testing.T) {
	mock := &MockVertexAIClient{
		updateError: assert.AnError,
	}

	product := &retail.Product{
		Id:    uuid.New().String(),
		Title: "Updated Product",
	}

	err := mock.UpdateProduct(context.Background(), product)
	assert.Error(t, err)
}

func TestMockVertexAIClient_DeleteProduct(t *testing.T) {
	mock := &MockVertexAIClient{}

	productID := uuid.New().String()
	err := mock.DeleteProduct(context.Background(), productID)
	assert.NoError(t, err)
}

func TestMockVertexAIClient_DeleteProductError(t *testing.T) {
	mock := &MockVertexAIClient{
		deleteError: assert.AnError,
	}

	productID := uuid.New().String()
	err := mock.DeleteProduct(context.Background(), productID)
	assert.Error(t, err)
}

func TestMockVertexAIClient_Close(t *testing.T) {
	mock := &MockVertexAIClient{}

	err := mock.Close()
	assert.NoError(t, err)
}
