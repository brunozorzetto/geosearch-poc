package postgres

import (
	"context"
	"geosearch-poc/internal/domain"
	"geosearch-poc/internal/repository/postgres"
	"geosearch-poc/tests/integration/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestProductRepository(t *testing.T) (*postgres.ProductRepository, func()) {
	// Create a test database configuration
	dbConfig := config.NewTestDatabaseConfig()

	// Create a test database connection
	pool, err := dbConfig.NewPool()
	require.NoError(t, err)

	// Create repository
	repo := postgres.NewProductRepository(pool)

	// Cleanup function
	cleanup := func() {
		pool.Close()
	}

	return repo, cleanup
}

func TestProductRepository_Create(t *testing.T) {
	repo, cleanup := setupTestProductRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test product
	product := &domain.Product{
		StoreID:     "00000000-0000-0000-0000-000000000001",
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
	}

	err := repo.Create(ctx, product)
	require.NoError(t, err)
	assert.NotEmpty(t, product.ID)
	assert.NotZero(t, product.CreatedAt)
	assert.NotZero(t, product.UpdatedAt)
}

func TestProductRepository_GetByID(t *testing.T) {
	repo, cleanup := setupTestProductRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test product
	product := &domain.Product{
		StoreID:     "00000000-0000-0000-0000-000000000001",
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
	}

	err := repo.Create(ctx, product)
	require.NoError(t, err)

	// Retrieve the product
	retrieved, err := repo.GetByID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, product.ID, retrieved.ID)
	assert.Equal(t, product.StoreID, retrieved.StoreID)
	assert.Equal(t, product.Name, retrieved.Name)
	assert.Equal(t, product.Description, retrieved.Description)
	assert.Equal(t, product.Price, retrieved.Price)
}

func TestProductRepository_GetByStoreID(t *testing.T) {
	repo, cleanup := setupTestProductRepository(t)
	defer cleanup()

	ctx := context.Background()
	storeID := "00000000-0000-0000-0000-000000000001"

	// Create multiple products for the same store
	products := []*domain.Product{
		{
			StoreID:     storeID,
			Name:        "Product 1",
			Description: "Description 1",
			Price:       99.99,
		},
		{
			StoreID:     storeID,
			Name:        "Product 2",
			Description: "Description 2",
			Price:       149.99,
		},
	}

	for _, p := range products {
		err := repo.Create(ctx, p)
		require.NoError(t, err)
	}

	// Retrieve products with pagination
	retrieved, err := repo.GetByStoreID(ctx, storeID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, retrieved, 2)
}

func TestProductRepository_Update(t *testing.T) {
	repo, cleanup := setupTestProductRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test product
	product := &domain.Product{
		StoreID:     "00000000-0000-0000-0000-000000000001",
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
	}

	err := repo.Create(ctx, product)
	require.NoError(t, err)

	// Update the product
	product.Name = "Updated Product"
	product.Description = "Updated Description"
	product.Price = 149.99

	err = repo.Update(ctx, product)
	require.NoError(t, err)

	// Verify the update
	retrieved, err := repo.GetByID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Product", retrieved.Name)
	assert.Equal(t, "Updated Description", retrieved.Description)
	assert.Equal(t, 149.99, retrieved.Price)
}

func TestProductRepository_Delete(t *testing.T) {
	repo, cleanup := setupTestProductRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test product
	product := &domain.Product{
		StoreID:     "00000000-0000-0000-0000-000000000001",
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
	}

	err := repo.Create(ctx, product)
	require.NoError(t, err)

	// Delete the product
	err = repo.Delete(ctx, product.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, product.ID)
	assert.Error(t, err)
}

func TestProductRepository_GetByStoreIDs(t *testing.T) {
	repo, cleanup := setupTestProductRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create products for multiple stores
	stores := []string{
		"00000000-0000-0000-0000-000000000001",
		"00000000-0000-0000-0000-000000000002",
	}

	for _, storeID := range stores {
		for i := 0; i < 3; i++ {
			product := &domain.Product{
				StoreID:     storeID,
				Name:        "Product " + storeID,
				Description: "Description " + storeID,
				Price:       99.99,
			}
			err := repo.Create(ctx, product)
			require.NoError(t, err)
		}
	}

	// Retrieve products with limit per store
	productsByStore, err := repo.GetByStoreIDs(ctx, stores, 2)
	require.NoError(t, err)

	// Verify results
	for _, storeID := range stores {
		products, exists := productsByStore[storeID]
		assert.True(t, exists)
		assert.LessOrEqual(t, len(products), 2)
	}
}
