package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"geosearch-poc/domain"
	"geosearch-poc/repository"
)

// ProductRepository implements repository.ProductRepository
type ProductRepository struct {
	db *pgxpool.Pool
}

// NewProductRepository creates a new product repository instance
func NewProductRepository(db *pgxpool.Pool) repository.ProductRepository {
	return &ProductRepository{db: db}
}

// Create creates a new product
func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (id, store_id, h3_index, name, description, price, category, brand, sku, stock, images, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.Exec(ctx, query,
		product.ID,
		product.StoreID,
		product.H3Index,
		product.Name,
		product.Description,
		product.Price,
		product.Category,
		product.Brand,
		product.SKU,
		product.Stock,
		product.Images,
		product.CreatedAt,
		product.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	query := `
		SELECT id, store_id, h3_index, name, description, price, category, brand, sku, stock, images, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product domain.Product
	err := r.db.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.StoreID,
		&product.H3Index,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Category,
		&product.Brand,
		&product.SKU,
		&product.Stock,
		&product.Images,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// GetByStoreID retrieves all products for a store
func (r *ProductRepository) GetByStoreID(ctx context.Context, storeID uuid.UUID) ([]*domain.Product, error) {
	query := `
		SELECT id, store_id, h3_index, name, description, price, category, brand, sku, stock, images, created_at, updated_at
		FROM products
		WHERE store_id = $1
	`

	rows, err := r.db.Query(ctx, query, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var product domain.Product
		err := rows.Scan(
			&product.ID,
			&product.StoreID,
			&product.H3Index,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Category,
			&product.Brand,
			&product.SKU,
			&product.Stock,
			&product.Images,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

// Update updates a product
func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products
		SET store_id = $1, h3_index = $2, name = $3, description = $4, price = $5,
			category = $6, brand = $7, sku = $8, stock = $9, images = $10, updated_at = $11
		WHERE id = $12
	`

	_, err := r.db.Exec(ctx, query,
		product.StoreID,
		product.H3Index,
		product.Name,
		product.Description,
		product.Price,
		product.Category,
		product.Brand,
		product.SKU,
		product.Stock,
		product.Images,
		product.UpdatedAt,
		product.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// Delete deletes a product
func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// Search searches for products based on criteria
func (r *ProductRepository) Search(ctx context.Context, params *domain.ProductSearchParams) ([]*domain.Product, int64, error) {
	// This method is not used directly as we're using Vertex AI Search
	// It's kept for compatibility with the interface
	return nil, 0, fmt.Errorf("search is not implemented, use Vertex AI Search instead")
}

func (r *ProductRepository) GetByH3Index(ctx context.Context, h3Index string) ([]*domain.Product, error) {
	query := `
		SELECT id, store_id, name, description, price, category, 
			brand, sku, stock, h3_index, images, created_at, updated_at
		FROM products 
		WHERE h3_index = $1`

	rows, err := r.db.Query(ctx, query, h3Index)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		product := &domain.Product{}
		err := rows.Scan(
			&product.ID,
			&product.StoreID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Category,
			&product.Brand,
			&product.SKU,
			&product.Stock,
			&product.H3Index,
			&product.Images,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *ProductRepository) GetByH3Cells(ctx context.Context, h3Cells []string) ([]*domain.Product, error) {
	query := `
		SELECT id, store_id, name, description, price, category, 
			brand, sku, stock, h3_index, images, created_at, updated_at
		FROM products 
		WHERE h3_index = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, h3Cells)
	if err != nil {
		return nil, fmt.Errorf("failed to query products by H3 cells: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		product := &domain.Product{}
		err := rows.Scan(
			&product.ID,
			&product.StoreID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Category,
			&product.Brand,
			&product.SKU,
			&product.Stock,
			&product.H3Index,
			&product.Images,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product rows: %w", err)
	}

	return products, nil
}
