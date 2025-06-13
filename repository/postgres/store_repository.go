package postgres

import (
	"context"
	"fmt"

	"geosearch-poc/domain"
	"geosearch-poc/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StoreRepository implements repository.StoreRepository
type StoreRepository struct {
	pool *pgxpool.Pool
}

// NewStoreRepository creates a new store repository instance
func NewStoreRepository(pool *pgxpool.Pool) repository.StoreRepository {
	return &StoreRepository{
		pool: pool,
	}
}

// Create creates a new store
func (r *StoreRepository) Create(ctx context.Context, store *domain.Store) error {
	query := `
		INSERT INTO stores (name, latitude, longitude, h3_index, delivery_radius)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := r.pool.QueryRow(ctx, query,
		store.Name,
		store.Latitude,
		store.Longitude,
		store.H3Index,
		store.DeliveryRadius,
	).Scan(&store.ID)

	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	return nil
}

// GetByID retrieves a store by its ID
func (r *StoreRepository) GetByID(ctx context.Context, id int) (*domain.Store, error) {
	query := `
		SELECT id, name, latitude, longitude, h3_index, delivery_radius, created_at, updated_at
		FROM stores
		WHERE id = $1`

	store := &domain.Store{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&store.ID,
		&store.Name,
		&store.Latitude,
		&store.Longitude,
		&store.H3Index,
		&store.DeliveryRadius,
		&store.CreatedAt,
		&store.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get store by ID: %w", err)
	}

	return store, nil
}

// GetByH3Index retrieves stores by their H3 index
func (r *StoreRepository) GetByH3Index(ctx context.Context, h3Index string) ([]*domain.Store, error) {
	query := `
		SELECT id, name, latitude, longitude, h3_index, delivery_radius, created_at, updated_at
		FROM stores
		WHERE h3_index = $1`

	rows, err := r.pool.Query(ctx, query, h3Index)
	if err != nil {
		return nil, fmt.Errorf("failed to query stores: %w", err)
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		store := &domain.Store{}
		err := rows.Scan(
			&store.ID,
			&store.Name,
			&store.Latitude,
			&store.Longitude,
			&store.H3Index,
			&store.DeliveryRadius,
			&store.CreatedAt,
			&store.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan store: %w", err)
		}
		stores = append(stores, store)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating store rows: %w", err)
	}

	return stores, nil
}

// GetByH3Indexes retrieves stores by multiple H3 indexes
func (r *StoreRepository) GetByH3Indexes(ctx context.Context, h3Indexes []string) ([]*domain.Store, error) {
	query := `
		SELECT id, name, latitude, longitude, h3_index, delivery_radius, created_at, updated_at
		FROM stores
		WHERE h3_index = ANY($1)`

	rows, err := r.pool.Query(ctx, query, h3Indexes)
	if err != nil {
		return nil, fmt.Errorf("failed to query stores by H3 indexes: %w", err)
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		store := &domain.Store{}
		err := rows.Scan(
			&store.ID,
			&store.Name,
			&store.Latitude,
			&store.Longitude,
			&store.H3Index,
			&store.DeliveryRadius,
			&store.CreatedAt,
			&store.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan store: %w", err)
		}
		stores = append(stores, store)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating store rows: %w", err)
	}

	return stores, nil
}

// Update updates an existing store
func (r *StoreRepository) Update(ctx context.Context, store *domain.Store) error {
	query := `
		UPDATE stores
		SET name = $1, latitude = $2, longitude = $3, h3_index = $4, delivery_radius = $5
		WHERE id = $6`

	_, err := r.pool.Exec(ctx, query,
		store.Name,
		store.Latitude,
		store.Longitude,
		store.H3Index,
		store.DeliveryRadius,
		store.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update store: %w", err)
	}

	return nil
}

// Delete deletes a store by its ID
func (r *StoreRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM stores WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete store: %w", err)
	}

	return nil
}

// Search searches for stores based on criteria
func (r *StoreRepository) Search(ctx context.Context, params *domain.SearchParams) ([]*domain.Store, int, error) {
	query := `
		SELECT id, name, latitude, longitude, h3_index, delivery_radius, created_at, updated_at
		FROM stores
		WHERE ST_DWithin(
			ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
			$3 * 1000
		)
		AND delivery_radius >= ST_Distance(
			ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
		)`

	args := []interface{}{params.Longitude, params.Latitude, params.Radius}

	// Add pagination
	query += " LIMIT $4 OFFSET $5"
	args = append(args, params.Limit, params.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query stores: %w", err)
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		store := &domain.Store{}
		err := rows.Scan(
			&store.ID,
			&store.Name,
			&store.Latitude,
			&store.Longitude,
			&store.H3Index,
			&store.DeliveryRadius,
			&store.CreatedAt,
			&store.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan store: %w", err)
		}
		stores = append(stores, store)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating store rows: %w", err)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM stores
		WHERE ST_DWithin(
			ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
			$3 * 1000
		)
		AND delivery_radius >= ST_Distance(
			ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
		)`

	var total int
	err = r.pool.QueryRow(ctx, countQuery, params.Longitude, params.Latitude, params.Radius).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return stores, total, nil
}

// SearchByRadius searches for stores within a given radius (in kilometers) from a point
func (r *StoreRepository) SearchByRadius(ctx context.Context, latitude, longitude float64, radius float64) ([]*domain.Store, error) {
	query := `
		SELECT id, name, latitude, longitude, h3_index, created_at, updated_at
		FROM stores
		WHERE ST_DWithin(
			ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
			$3 * 1000
		)
		ORDER BY ST_Distance(
			ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
		)
	`

	rows, err := r.pool.Query(ctx, query, longitude, latitude, radius)
	if err != nil {
		return nil, fmt.Errorf("failed to query stores by radius: %w", err)
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		var store domain.Store
		err := rows.Scan(
			&store.ID,
			&store.Name,
			&store.Latitude,
			&store.Longitude,
			&store.H3Index,
			&store.CreatedAt,
			&store.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan store: %w", err)
		}
		stores = append(stores, &store)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating store rows: %w", err)
	}

	return stores, nil
}
