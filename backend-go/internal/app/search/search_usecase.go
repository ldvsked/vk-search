package search

import (
	"context"
	"vk-search/internal/domain"
)

type SearchUseCase struct {
	repo domain.SearchRepository
}

func NewSearchUseCase(repo domain.SearchRepository) domain.SearchUseCase {
	return &SearchUseCase{repo: repo}
}

func (uc *SearchUseCase) Execute(ctx context.Context, query string, limit int) ([]domain.Post, error) {
	return uc.repo.Search(ctx, query, limit)
}