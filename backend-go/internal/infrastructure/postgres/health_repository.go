package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"vk-search/internal/domain"
)

type healthRepository struct {
	pool *pgxpool.Pool
}

func NewHealthRepository(pool *pgxpool.Pool) domain.HealthRepository {
	return &healthRepository{pool: pool}
}

func (r *healthRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}