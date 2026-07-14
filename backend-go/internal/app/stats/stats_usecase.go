package stats

import (
	"context"
	"vk-search/internal/domain"
)

type usecase struct {
	statsRepo  domain.StatsRepository
	chunkRepo  domain.ChunkRepository
}

func NewStatsUseCase(sr domain.StatsRepository, cr domain.ChunkRepository) domain.StatsUseCase {
	return &usecase{
		statsRepo: sr,
		chunkRepo: cr,
	}
}

func (u *usecase) GetStats(ctx context.Context) (*domain.Stats, error) {
	stats, err := u.statsRepo.GetGeneralStats(ctx)
	if err != nil {
		return nil, err
	}

	chunksCount, err := u.chunkRepo.GetChunksCount(ctx)
	if err != nil {
		return nil, err
	}

	stats.ChunksCount = chunksCount

	return stats, nil
}