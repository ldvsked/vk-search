package search

import (
	"context"
	"log"

	"vk-search/internal/domain"
)

type SearchUseCase struct {
	repo domain.SearchRepository
}

func NewSearchUseCase(repo domain.SearchRepository) domain.SearchUseCase {
	return &SearchUseCase{repo: repo}
}

func (uc *SearchUseCase) Execute(ctx context.Context, query string, limit int, source string, dateFrom string, dateTo string) ([]domain.Post, error) {
	results, err := uc.repo.Search(ctx, query, limit, source, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	userID, _ := ctx.Value(domain.UserIDKey).(int64)

	mode := "search"
	if m, ok := ctx.Value(domain.ModeKey).(string); ok && m != "" {
		mode = m
	}

	searchLog := &domain.SearchLog{
		UserID:      userID,
		Query:       query,
		Mode:        mode,
		LimitValue:  limit,
		ResultCount: len(results),
	}

	go func(l *domain.SearchLog) {
		if err := uc.repo.SaveLog(context.Background(), l); err != nil {
			log.Printf("failed to save search log: %v", err)
		}
	}(searchLog)

	return results, nil
}