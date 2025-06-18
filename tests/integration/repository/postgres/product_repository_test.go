package postgres

import (
	"context"
	"testing"
	"time"

	"geosearch-poc/domain"
	"geosearch-poc/repository/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupProductRepositoryTest(t *testing.T) (*postgres.ProductRepository, *pgxpool.Pool, func()) {
	// Setup database connection
	db, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5433/geosearch_test?sslmode=disable")
	require.NoError(t, err)

	// Clean database more thoroughly
	_, _ = db.Exec(context.Background(), "TRUNCATE TABLE products, stores RESTART IDENTITY CASCADE")
	_, _ = db.Exec(context.Background(), "DELETE FROM products")
	_, _ = db.Exec(context.Background(), "DELETE FROM stores")

	// Setup repository
	repo := postgres.NewProductRepository(db)

	// Cleanup function
	cleanup := func() {
		db.Close()
	}

	return repo.(*postgres.ProductRepository), db, cleanup
}

func TestProductRepository_Create(t *testing.T) {
	repo, db, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

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

	tests := []struct {
		name        string
		product     *domain.Product
		expectError bool
	}{
		{
			name: "should create product successfully",
			product: domain.NewProduct(
				testStore.ID,
				"Test Product",
				"Test product description",
				29.99,
				"Electronics",
				"TestBrand",
				"TEST-SKU-001",
				100,
				"8928308280fffff",
				[]string{"https://example.com/image1.jpg"},
			),
			expectError: false,
		},
		{
			name: "should create product with minimal fields",
			product: domain.NewProduct(
				testStore.ID,
				"Minimal Product",
				"",
				19.99,
				"",
				"",
				"MIN-SKU-001",
				50,
				"8928308280fffff",
				[]string{},
			),
			expectError: false,
		},
		{
			name: "should fail to create product with non-existent store",
			product: domain.NewProduct(
				uuid.New(), // Non-existent store ID
				"Test Product",
				"Test product description",
				29.99,
				"Electronics",
				"TestBrand",
				"TEST-SKU-002",
				100,
				"8928308280fffff",
				[]string{"https://example.com/image1.jpg"},
			),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute create
			err := repo.Create(context.Background(), tt.product)

			// Assert result
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify product was created
				createdProduct, err := repo.GetByID(context.Background(), tt.product.ID)
				assert.NoError(t, err)
				assert.NotNil(t, createdProduct)
				if createdProduct != nil {
					assert.Equal(t, tt.product.Name, createdProduct.Name)
					assert.Equal(t, tt.product.StoreID, createdProduct.StoreID)
					assert.Equal(t, tt.product.Price, createdProduct.Price)
					assert.Equal(t, tt.product.SKU, createdProduct.SKU)
					assert.Equal(t, tt.product.Stock, createdProduct.Stock)
				}
			}
		})
	}
}

func TestProductRepository_GetByID(t *testing.T) {
	repo, db, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	// Create a test store and product
	storeRepo := postgres.NewStoreRepository(db)
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

	err := storeRepo.Create(context.Background(), testStore)
	require.NoError(t, err)

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

	err = repo.Create(context.Background(), testProduct)
	require.NoError(t, err)

	tests := []struct {
		name        string
		productID   uuid.UUID
		expectError bool
		checkResult func(t *testing.T, product *domain.Product)
	}{
		{
			name:        "should get product by valid ID",
			productID:   testProduct.ID,
			expectError: false,
			checkResult: func(t *testing.T, product *domain.Product) {
				if product != nil {
					assert.Equal(t, testProduct.ID, product.ID)
					assert.Equal(t, testProduct.Name, product.Name)
					assert.Equal(t, testProduct.StoreID, product.StoreID)
					assert.Equal(t, testProduct.Price, product.Price)
					assert.Equal(t, testProduct.SKU, product.SKU)
					assert.Equal(t, testProduct.Stock, product.Stock)
					assert.Equal(t, testProduct.Category, product.Category)
					assert.Equal(t, testProduct.Brand, product.Brand)
					assert.Equal(t, testProduct.H3Index, product.H3Index)
					assert.Equal(t, testProduct.Images, product.Images)
				}
			},
		},
		{
			name:        "should return error for non-existent product",
			productID:   uuid.New(),
			expectError: true,
			checkResult: func(t *testing.T, product *domain.Product) {
				assert.Nil(t, product)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute get
			product, err := repo.GetByID(context.Background(), tt.productID)

			// Assert result
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Run additional checks if provided
			if tt.checkResult != nil {
				tt.checkResult(t, product)
			}
		})
	}
}

func TestProductRepository_GetByStoreID(t *testing.T) {
	repo, db, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	// Create test stores and products
	storeRepo := postgres.NewStoreRepository(db)

	store1 := &domain.Store{
		ID:             uuid.New(),
		Name:           "Store 1",
		CategoryID:     uuid.Nil,
		Latitude:       -23.550520,
		Longitude:      -46.633308,
		H3Index:        "8928308280fffff",
		Address:        "Rua Teste, 123",
		DeliveryRadius: 5.0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	store2 := &domain.Store{
		ID:             uuid.New(),
		Name:           "Store 2",
		CategoryID:     uuid.Nil,
		Latitude:       -23.551520,
		Longitude:      -46.634308,
		H3Index:        "8928308280fffff",
		Address:        "Rua Teste, 456",
		DeliveryRadius: 3.0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := storeRepo.Create(context.Background(), store1)
	require.NoError(t, err)
	err = storeRepo.Create(context.Background(), store2)
	require.NoError(t, err)

	// Create products for store1
	product1 := domain.NewProduct(
		store1.ID,
		"Product 1 from Store 1",
		"Description 1",
		29.99,
		"Electronics",
		"Brand1",
		"SKU-1-001",
		100,
		"8928308280fffff",
		[]string{"https://example.com/image1.jpg"},
	)

	product2 := domain.NewProduct(
		store1.ID,
		"Product 2 from Store 1",
		"Description 2",
		39.99,
		"Electronics",
		"Brand1",
		"SKU-1-002",
		50,
		"8928308280fffff",
		[]string{"https://example.com/image2.jpg"},
	)

	// Create product for store2
	product3 := domain.NewProduct(
		store2.ID,
		"Product 1 from Store 2",
		"Description 3",
		49.99,
		"Electronics",
		"Brand2",
		"SKU-2-001",
		75,
		"8928308280fffff",
		[]string{"https://example.com/image3.jpg"},
	)

	err = repo.Create(context.Background(), product1)
	require.NoError(t, err)
	err = repo.Create(context.Background(), product2)
	require.NoError(t, err)
	err = repo.Create(context.Background(), product3)
	require.NoError(t, err)

	tests := []struct {
		name          string
		storeID       uuid.UUID
		expectedCount int
		checkProducts func(t *testing.T, products []*domain.Product)
	}{
		{
			name:          "should get products for store 1",
			storeID:       store1.ID,
			expectedCount: 2,
			checkProducts: func(t *testing.T, products []*domain.Product) {
				assert.Len(t, products, 2)
				productNames := []string{products[0].Name, products[1].Name}
				assert.Contains(t, productNames, "Product 1 from Store 1")
				assert.Contains(t, productNames, "Product 2 from Store 1")

				// Verify all products belong to store1
				for _, product := range products {
					assert.Equal(t, store1.ID, product.StoreID)
				}
			},
		},
		{
			name:          "should get products for store 2",
			storeID:       store2.ID,
			expectedCount: 1,
			checkProducts: func(t *testing.T, products []*domain.Product) {
				assert.Len(t, products, 1)
				assert.Equal(t, "Product 1 from Store 2", products[0].Name)
				assert.Equal(t, store2.ID, products[0].StoreID)
			},
		},
		{
			name:          "should return empty list for non-existent store",
			storeID:       uuid.New(),
			expectedCount: 0,
			checkProducts: func(t *testing.T, products []*domain.Product) {
				assert.Empty(t, products)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute get by store ID
			products, err := repo.GetByStoreID(context.Background(), tt.storeID)

			// Assert result
			assert.NoError(t, err)
			assert.Len(t, products, tt.expectedCount)

			// Run additional checks if provided
			if tt.checkProducts != nil {
				tt.checkProducts(t, products)
			}
		})
	}
}

func TestProductRepository_Update(t *testing.T) {
	repo, db, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	// Create a test store and product
	storeRepo := postgres.NewStoreRepository(db)
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

	err := storeRepo.Create(context.Background(), testStore)
	require.NoError(t, err)

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

	err = repo.Create(context.Background(), testProduct)
	require.NoError(t, err)

	tests := []struct {
		name        string
		updates     func(*domain.Product)
		expectError bool
		checkResult func(t *testing.T, product *domain.Product)
	}{
		{
			name: "should update product successfully",
			updates: func(p *domain.Product) {
				p.Update(
					"Updated Product Name",
					"Updated product description",
					39.99,
					"Updated Electronics",
					"UpdatedBrand",
					"UPDATED-SKU-001",
					150,
					"8928308280fffff",
					[]string{"https://example.com/updated-image.jpg"},
				)
			},
			expectError: false,
			checkResult: func(t *testing.T, product *domain.Product) {
				assert.Equal(t, "Updated Product Name", product.Name)
				assert.Equal(t, "Updated product description", product.Description)
				assert.Equal(t, 39.99, product.Price)
				assert.Equal(t, "Updated Electronics", product.Category)
				assert.Equal(t, "UpdatedBrand", product.Brand)
				assert.Equal(t, "UPDATED-SKU-001", product.SKU)
				assert.Equal(t, 150, product.Stock)
				assert.Equal(t, []string{"https://example.com/updated-image.jpg"}, product.Images)
			},
		},
		{
			name: "should update product with partial changes",
			updates: func(p *domain.Product) {
				p.Update(
					"Partially Updated Product",
					p.Description,
					p.Price,
					p.Category,
					p.Brand,
					p.SKU,
					200,
					p.H3Index,
					p.Images,
				)
			},
			expectError: false,
			checkResult: func(t *testing.T, product *domain.Product) {
				assert.Equal(t, "Partially Updated Product", product.Name)
				assert.Equal(t, 200, product.Stock)
				// Other fields should remain unchanged
				assert.Equal(t, testProduct.Description, product.Description)
				assert.Equal(t, testProduct.Price, product.Price)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply updates
			tt.updates(testProduct)

			// Execute update
			err := repo.Update(context.Background(), testProduct)

			// Assert result
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify product was updated
				updatedProduct, err := repo.GetByID(context.Background(), testProduct.ID)
				assert.NoError(t, err)

				// Run additional checks if provided
				if tt.checkResult != nil {
					tt.checkResult(t, updatedProduct)
				}
			}
		})
	}
}

func TestProductRepository_Delete(t *testing.T) {
	repo, db, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	// Create a test store and product
	storeRepo := postgres.NewStoreRepository(db)
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

	err := storeRepo.Create(context.Background(), testStore)
	require.NoError(t, err)

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

	err = repo.Create(context.Background(), testProduct)
	require.NoError(t, err)

	tests := []struct {
		name        string
		productID   uuid.UUID
		expectError bool
	}{
		{
			name:        "should delete product successfully",
			productID:   testProduct.ID,
			expectError: false,
		},
		{
			name:        "should return error for non-existent product",
			productID:   uuid.New(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute delete
			err := repo.Delete(context.Background(), tt.productID)

			// Assert result
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify product was deleted
				_, err := repo.GetByID(context.Background(), tt.productID)
				assert.Error(t, err) // Should return error as product is deleted
			}
		})
	}
}
