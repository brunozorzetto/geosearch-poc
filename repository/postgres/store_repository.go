package postgres

import (
	"context"
	"fmt"

	"geosearch-poc/domain"
	"geosearch-poc/repository"

	"github.com/google/uuid"
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
		INSERT INTO stores (name, category_id, latitude, longitude, h3_index, address)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	categoryID, err := uuid.Parse(store.Category)
	if err != nil {
		return fmt.Errorf("invalid category ID: %w", err)
	}

	err = r.pool.QueryRow(ctx, query,
		store.Name,
		categoryID,
		store.Latitude,
		store.Longitude,
		store.H3Index,
		store.Address,
	).Scan(&store.ID, &store.CreatedAt, &store.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	return nil
}

// GetByID retrieves a store by its ID
func (r *StoreRepository) GetByID(ctx context.Context, id string) (*domain.Store, error) {
	query := `
		SELECT s.id, s.name, c.id as category_id, s.latitude, s.longitude, s.h3_index, s.address, s.created_at, s.updated_at
		FROM stores s
		JOIN categories c ON s.category_id = c.id
		WHERE s.id = $1`

	store := &domain.Store{}
	var categoryID uuid.UUID

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&store.ID,
		&store.Name,
		&categoryID,
		&store.Latitude,
		&store.Longitude,
		&store.H3Index,
		&store.Address,
		&store.CreatedAt,
		&store.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get store: %w", err)
	}

	store.Category = categoryID.String()
	return store, nil
}

// GetByH3Index retrieves stores by their H3 index
func (r *StoreRepository) GetByH3Index(ctx context.Context, h3Index string) ([]*domain.Store, error) {
	query := `
		SELECT s.id, s.name, c.id as category_id, s.latitude, s.longitude, s.h3_index, s.address, s.created_at, s.updated_at
		FROM stores s
		JOIN categories c ON s.category_id = c.id
		WHERE s.h3_index = $1`

	rows, err := r.pool.Query(ctx, query, h3Index)
	if err != nil {
		return nil, fmt.Errorf("failed to query stores: %w", err)
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		store := &domain.Store{}
		var categoryID uuid.UUID

		err := rows.Scan(
			&store.ID,
			&store.Name,
			&categoryID,
			&store.Latitude,
			&store.Longitude,
			&store.H3Index,
			&store.Address,
			&store.CreatedAt,
			&store.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan store: %w", err)
		}

		store.Category = categoryID.String()
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
		SELECT s.id, s.name, c.id as category_id, s.latitude, s.longitude, s.h3_index, s.address, s.created_at, s.updated_at
		FROM stores s
		JOIN categories c ON s.category_id = c.id
		WHERE s.h3_index = ANY($1)`

	rows, err := r.pool.Query(ctx, query, h3Indexes)
	if err != nil {
		return nil, fmt.Errorf("failed to query stores: %w", err)
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		store := &domain.Store{}
		var categoryID uuid.UUID

		err := rows.Scan(
			&store.ID,
			&store.Name,
			&categoryID,
			&store.Latitude,
			&store.Longitude,
			&store.H3Index,
			&store.Address,
			&store.CreatedAt,
			&store.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan store: %w", err)
		}

		store.Category = categoryID.String()
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
		SET name = $1, category_id = $2, latitude = $3, longitude = $4, h3_index = $5, address = $6
		WHERE id = $7
		RETURNING updated_at`

	categoryID, err := uuid.Parse(store.Category)
	if err != nil {
		return fmt.Errorf("invalid category ID: %w", err)
	}

	err = r.pool.QueryRow(ctx, query,
		store.Name,
		categoryID,
		store.Latitude,
		store.Longitude,
		store.H3Index,
		store.Address,
		store.ID,
	).Scan(&store.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update store: %w", err)
	}

	return nil
}

// Delete deletes a store by its ID
func (r *StoreRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM stores WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete store: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("store not found")
	}

	return nil
}

// Search searches for stores based on criteria
func (r *StoreRepository) Search(ctx context.Context, params *domain.SearchParams) ([]*domain.Store, int, error) {
	query := `
		SELECT s.id, s.name, c.id as category_id, s.latitude, s.longitude, s.h3_index, s.address, s.created_at, s.updated_at
		FROM stores s
		JOIN categories c ON s.category_id = c.id
		WHERE ST_DWithin(
			ST_SetSRID(ST_MakePoint(s.longitude, s.latitude), 4326)::geography,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
			$3
		)`

	args := []interface{}{params.Longitude, params.Latitude, params.Radius}
	argCount := 3

	if params.Category != "" {
		query += fmt.Sprintf(" AND c.id = $%d", argCount+1)
		args = append(args, params.Category)
		argCount++
	}

	// Add pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query stores: %w", err)
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		store := &domain.Store{}
		var categoryID uuid.UUID

		err := rows.Scan(
			&store.ID,
			&store.Name,
			&categoryID,
			&store.Latitude,
			&store.Longitude,
			&store.H3Index,
			&store.Address,
			&store.CreatedAt,
			&store.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan store: %w", err)
		}

		store.Category = categoryID.String()
		stores = append(stores, store)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating store rows: %w", err)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM stores s
		JOIN categories c ON s.category_id = c.id
		WHERE ST_DWithin(
			ST_SetSRID(ST_MakePoint(s.longitude, s.latitude), 4326)::geography,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
			$3
		)`

	countArgs := []interface{}{params.Longitude, params.Latitude, params.Radius}
	if params.Category != "" {
		countQuery += " AND c.id = $4"
		countArgs = append(countArgs, params.Category)
	}

	var total int
	err = r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return stores, total, nil
}
