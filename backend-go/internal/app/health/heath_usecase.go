package health

import (
	"context"
	"vk-search/internal/domain"
)

type healthUseCase struct {
	repo domain.HealthRepository
}

func NewHealthUseCase(repo domain.HealthRepository) domain.HealthUseCase {
	return &healthUseCase{repo: repo}
}

func (uc *healthUseCase) CheckHealth(ctx context.Context) error {
	return uc.repo.Ping(ctx)
}