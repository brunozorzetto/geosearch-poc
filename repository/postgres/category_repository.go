package postgres

import (
	"context"
	"geosearch-poc/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO categories (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`,
		category.ID, category.Name,
	)
	return err
}
