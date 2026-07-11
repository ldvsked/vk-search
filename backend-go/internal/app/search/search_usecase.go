package search

import (
	"context"
	"log"
	"vk-search/internal/api/middleware"
	"vk-search/internal/domain"
)

type SearchUseCase struct {
	repo domain.SearchRepository
}

func NewSearchUseCase(repo domain.SearchRepository) domain.SearchUseCase {
	return &SearchUseCase{repo: repo}
}

func (uc *SearchUseCase) Execute(ctx context.Context, query string, limit int) ([]domain.Post, error) {
	results, err := uc.repo.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	userID, _ := ctx.Value(middleware.UserIDKey).(int64)

	searchLog := &domain.SearchLog{
		UserID:      userID,
		Query:       query,
		Mode:        "search",
		LimitValue:  limit,
		ResultCount: len(results),
	}

	// Сохраняем лог в базу (асинхронно, чтобы не задерживать ответ)
	go func(l *domain.SearchLog) {
		if err := uc.repo.SaveLog(context.Background(), l); err != nil {
			log.Printf("failed to save search log: %v", err)
		}
	}(searchLog)

	return results, nil
}