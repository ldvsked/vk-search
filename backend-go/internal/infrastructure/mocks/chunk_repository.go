package mocks

import (
	"context"
	"vk-search/internal/domain"
)

type chunkRepository struct{}

func NewChunkRepository() domain.ChunkRepository {
	return &chunkRepository{}
}

func (r *chunkRepository) GetChunksCount(ctx context.Context) (int, error) {
	return 17309, nil
}