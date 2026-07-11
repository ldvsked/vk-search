package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"vk-search/internal/domain"
)

type authUseCase struct {
	userRepo  domain.UserRepository
	jwtSecret []byte
}

type TokenConfig interface {
	GetJWTSecret() string
}

func NewAuthUseCase(userRepo domain.UserRepository, cfg TokenConfig) domain.AuthUseCase {
	return &authUseCase{
		userRepo:  userRepo,
		jwtSecret: []byte(cfg.GetJWTSecret()),
	}
}

func (uc *authUseCase) Login(ctx context.Context, username, password string) (string, error) {
	user, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	
	if user == nil {
		return "", errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role_id":  user.RoleID,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(uc.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
