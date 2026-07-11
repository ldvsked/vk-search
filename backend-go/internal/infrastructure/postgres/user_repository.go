package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vk-search/internal/domain"
)

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) domain.UserRepository {
	return &userRepository{
		pool: pool,
	}
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, password_hash, role_id 
		FROM users 
		WHERE username = $1 
		LIMIT 1;
	`

	var user domain.User

	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.RoleID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}
