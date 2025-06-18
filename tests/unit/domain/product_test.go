package domain_test

import (
	"testing"

	"geosearch-poc/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewProduct(t *testing.T) {
	storeID := uuid.New()

	tests := []struct {
		name        string
		storeID     uuid.UUID
		productName string
		description string
		price       float64
		category    string
		brand       string
		sku         string
		stock       int
		h3Index     string
		images      []string
		checkResult func(t *testing.T, product *domain.Product)
	}{
		{
			name:        "should create product with all fields",
			storeID:     storeID,
			productName: "Test Product",
			description: "Test product description",
			price:       29.99,
			category:    "Electronics",
			brand:       "TestBrand",
			sku:         "TEST-SKU-001",
			stock:       100,
			h3Index:     "8928308280fffff",
			images:      []string{"https://example.com/image1.jpg"},
			checkResult: func(t *testing.T, product *domain.Product) {
				assert.NotEqual(t, uuid.Nil, product.ID)
				assert.Equal(t, storeID, product.StoreID)
				assert.Equal(t, "Test Product", product.Name)
				assert.Equal(t, "Test product description", product.Description)
				assert.Equal(t, 29.99, product.Price)
				assert.Equal(t, "Electronics", product.Category)
				assert.Equal(t, "TestBrand", product.Brand)
				assert.Equal(t, "TEST-SKU-001", product.SKU)
				assert.Equal(t, 100, product.Stock)
				assert.Equal(t, "8928308280fffff", product.H3Index)
				assert.Equal(t, []string{"https://example.com/image1.jpg"}, product.Images)
				assert.False(t, product.CreatedAt.IsZero())
				assert.False(t, product.UpdatedAt.IsZero())
			},
		},
		{
			name:        "should create product with minimal fields",
			storeID:     storeID,
			productName: "Minimal Product",
			description: "",
			price:       19.99,
			category:    "",
			brand:       "",
			sku:         "MIN-SKU-001",
			stock:       50,
			h3Index:     "",
			images:      []string{},
			checkResult: func(t *testing.T, product *domain.Product) {
				assert.NotEqual(t, uuid.Nil, product.ID)
				assert.Equal(t, storeID, product.StoreID)
				assert.Equal(t, "Minimal Product", product.Name)
				assert.Equal(t, "", product.Description)
				assert.Equal(t, 19.99, product.Price)
				assert.Equal(t, "", product.Category)
				assert.Equal(t, "", product.Brand)
				assert.Equal(t, "MIN-SKU-001", product.SKU)
				assert.Equal(t, 50, product.Stock)
				assert.Equal(t, "", product.H3Index)
				assert.Empty(t, product.Images)
				assert.False(t, product.CreatedAt.IsZero())
				assert.False(t, product.UpdatedAt.IsZero())
			},
		},
		{
			name:        "should create product with zero price",
			storeID:     storeID,
			productName: "Free Product",
			description: "Free product",
			price:       0.0,
			category:    "Free",
			brand:       "FreeBrand",
			sku:         "FREE-SKU-001",
			stock:       0,
			h3Index:     "8928308280fffff",
			images:      []string{},
			checkResult: func(t *testing.T, product *domain.Product) {
				assert.Equal(t, 0.0, product.Price)
				assert.Equal(t, 0, product.Stock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create product
			product := domain.NewProduct(
				tt.storeID,
				tt.productName,
				tt.description,
				tt.price,
				tt.category,
				tt.brand,
				tt.sku,
				tt.stock,
				tt.h3Index,
				tt.images,
			)

			// Run checks
			tt.checkResult(t, product)
		})
	}
}

func TestProduct_Update(t *testing.T) {
	storeID := uuid.New()
	originalProduct := domain.NewProduct(
		storeID,
		"Original Product",
		"Original description",
		29.99,
		"Electronics",
		"OriginalBrand",
		"ORIGINAL-SKU-001",
		100,
		"8928308280fffff",
		[]string{"https://example.com/original.jpg"},
	)

	originalCreatedAt := originalProduct.CreatedAt
	originalUpdatedAt := originalProduct.UpdatedAt

	tests := []struct {
		name        string
		updates     func(*domain.Product)
		checkResult func(t *testing.T, product *domain.Product)
	}{
		{
			name: "should update all fields",
			updates: func(p *domain.Product) {
				p.Update(
					"Updated Product",
					"Updated description",
					39.99,
					"Updated Electronics",
					"UpdatedBrand",
					"UPDATED-SKU-001",
					150,
					"8928308281fffff",
					[]string{"https://example.com/updated.jpg"},
				)
			},
			checkResult: func(t *testing.T, product *domain.Product) {
				assert.Equal(t, "Updated Product", product.Name)
				assert.Equal(t, "Updated description", product.Description)
				assert.Equal(t, 39.99, product.Price)
				assert.Equal(t, "Updated Electronics", product.Category)
				assert.Equal(t, "UpdatedBrand", product.Brand)
				assert.Equal(t, "UPDATED-SKU-001", product.SKU)
				assert.Equal(t, 150, product.Stock)
				assert.Equal(t, "8928308281fffff", product.H3Index)
				assert.Equal(t, []string{"https://example.com/updated.jpg"}, product.Images)

				// Check timestamps
				assert.Equal(t, originalCreatedAt, product.CreatedAt)      // Should not change
				assert.True(t, product.UpdatedAt.After(originalUpdatedAt)) // Should be updated
			},
		},
		{
			name: "should update partial fields",
			updates: func(p *domain.Product) {
				p.Update(
					"Partially Updated Product",
					p.Description, // Keep original
					p.Price,       // Keep original
					p.Category,    // Keep original
					p.Brand,       // Keep original
					p.SKU,         // Keep original
					200,           // Update stock
					p.H3Index,     // Keep original
					p.Images,      // Keep original
				)
			},
			checkResult: func(t *testing.T, product *domain.Product) {
				assert.Equal(t, "Partially Updated Product", product.Name)
				assert.Equal(t, "Original description", product.Description)
				assert.Equal(t, 29.99, product.Price)
				assert.Equal(t, "Electronics", product.Category)
				assert.Equal(t, "OriginalBrand", product.Brand)
				assert.Equal(t, "ORIGINAL-SKU-001", product.SKU)
				assert.Equal(t, 200, product.Stock)
				assert.Equal(t, "8928308280fffff", product.H3Index)
				assert.Equal(t, []string{"https://example.com/original.jpg"}, product.Images)

				// Check timestamps
				assert.Equal(t, originalCreatedAt, product.CreatedAt)      // Should not change
				assert.True(t, product.UpdatedAt.After(originalUpdatedAt)) // Should be updated
			},
		},
		{
			name: "should update with empty values",
			updates: func(p *domain.Product) {
				p.Update(
					"Empty Updated Product",
					"",
					0.0,
					"",
					"",
					"EMPTY-SKU-001",
					0,
					"",
					[]string{},
				)
			},
			checkResult: func(t *testing.T, product *domain.Product) {
				assert.Equal(t, "Empty Updated Product", product.Name)
				assert.Equal(t, "", product.Description)
				assert.Equal(t, 0.0, product.Price)
				assert.Equal(t, "", product.Category)
				assert.Equal(t, "", product.Brand)
				assert.Equal(t, "EMPTY-SKU-001", product.SKU)
				assert.Equal(t, 0, product.Stock)
				assert.Equal(t, "", product.H3Index)
				assert.Empty(t, product.Images)

				// Check timestamps
				assert.Equal(t, originalCreatedAt, product.CreatedAt)      // Should not change
				assert.True(t, product.UpdatedAt.After(originalUpdatedAt)) // Should be updated
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the original product
			product := &domain.Product{
				ID:          originalProduct.ID,
				StoreID:     originalProduct.StoreID,
				Name:        originalProduct.Name,
				Description: originalProduct.Description,
				Price:       originalProduct.Price,
				Category:    originalProduct.Category,
				Brand:       originalProduct.Brand,
				SKU:         originalProduct.SKU,
				Stock:       originalProduct.Stock,
				H3Index:     originalProduct.H3Index,
				Images:      originalProduct.Images,
				CreatedAt:   originalProduct.CreatedAt,
				UpdatedAt:   originalProduct.UpdatedAt,
			}

			// Apply updates
			tt.updates(product)

			// Run checks
			tt.checkResult(t, product)
		})
	}
}
