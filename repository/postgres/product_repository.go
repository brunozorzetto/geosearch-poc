package postgres

import (
	"context"
	"fmt"

	"geosearch-poc/domain"
	"geosearch-poc/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProductRepository implements repository.ProductRepository
type ProductRepository struct {
	pool *pgxpool.Pool
}

// NewProductRepository creates a new product repository instance
func NewProductRepository(pool *pgxpool.Pool) repository.ProductRepository {
	return &ProductRepository{
		pool: pool,
	}
}

// Create creates a new product
func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (store_id, name, description, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	storeID, err := uuid.Parse(product.StoreID)
	if err != nil {
		return fmt.Errorf("invalid store ID: %w", err)
	}

	err = r.pool.QueryRow(ctx, query,
		storeID,
		product.Name,
		product.Description,
		product.Price,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetByID retrieves a product by its ID
func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	query := `
		SELECT id, store_id, name, description, price, created_at, updated_at
		FROM products
		WHERE id = $1`

	product := &domain.Product{}
	var storeID uuid.UUID

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&storeID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	product.StoreID = storeID.String()
	return product, nil
}

// GetByStoreID retrieves products by store ID
func (r *ProductRepository) GetByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*domain.Product, error) {
	query := `
		SELECT id, store_id, name, description, price, created_at, updated_at
		FROM products
		WHERE store_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	storeUUID, err := uuid.Parse(storeID)
	if err != nil {
		return nil, fmt.Errorf("invalid store ID: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, storeUUID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		product := &domain.Product{}
		var storeID uuid.UUID

		err := rows.Scan(
			&product.ID,
			&storeID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		product.StoreID = storeID.String()
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product rows: %w", err)
	}

	return products, nil
}

// Update updates an existing product
func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3
		WHERE id = $4
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.ID,
	).Scan(&product.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// Delete deletes a product by its ID
func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// GetByStoreIDs retrieves products for multiple stores
func (r *ProductRepository) GetByStoreIDs(ctx context.Context, storeIDs []string, limit int) (map[string][]*domain.Product, error) {
	// Convert string IDs to UUIDs
	storeUUIDs := make([]uuid.UUID, len(storeIDs))
	for i, id := range storeIDs {
		uuid, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("invalid store ID: %w", err)
		}
		storeUUIDs[i] = uuid
	}

	query := `
		WITH ranked_products AS (
			SELECT 
				id, store_id, name, description, price, created_at, updated_at,
				ROW_NUMBER() OVER (PARTITION BY store_id ORDER BY created_at DESC) as rn
			FROM products
			WHERE store_id = ANY($1)
		)
		SELECT id, store_id, name, description, price, created_at, updated_at
		FROM ranked_products
		WHERE rn <= $2`

	rows, err := r.pool.Query(ctx, query, storeUUIDs, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	productsByStore := make(map[string][]*domain.Product)
	for rows.Next() {
		product := &domain.Product{}
		var storeID uuid.UUID

		err := rows.Scan(
			&product.ID,
			&storeID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		product.StoreID = storeID.String()
		productsByStore[product.StoreID] = append(productsByStore[product.StoreID], product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product rows: %w", err)
	}

	return productsByStore, nil
}
