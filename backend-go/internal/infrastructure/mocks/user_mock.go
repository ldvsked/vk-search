package mocks

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"vk-search/internal/domain"
)

type UserMockRepository struct {
	users map[string]*domain.User
}

func NewUserMockRepository() domain.UserRepository {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		panic("failed to generate mock password hash: " + err.Error())
	}

	return &UserMockRepository{
		users: map[string]*domain.User{
			"maria": {
				ID:           1,
				Username:     "maria",
				PasswordHash: string(hashedPassword),
				RoleID:       1,
			},
		},
	}
}

func (r *UserMockRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, ok := r.users[username]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}