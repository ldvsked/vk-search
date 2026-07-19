package stats

import (
	"context"
	"vk-search/internal/domain"
)

type usecase struct {
	statsRepo domain.StatsRepository
}

func NewStatsUseCase(sr domain.StatsRepository) domain.StatsUseCase {
	return &usecase{
		statsRepo: sr,
	}
}

func (u *usecase) GetStats(ctx context.Context) (*domain.Stats, error) {
	return u.statsRepo.GetGeneralStats(ctx)
}