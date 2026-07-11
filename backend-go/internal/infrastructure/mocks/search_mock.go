package mocks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"vk-search/internal/domain"
)

type SearchMockRepository struct {
	posts []domain.Post
	pool  *pgxpool.Pool
}

func NewSearchMockRepository(pool *pgxpool.Pool) domain.SearchRepository {
	return &SearchMockRepository{
		pool: pool,
		posts: []domain.Post{
			{
				ChunkID:     "chunk_1",
				DocumentID:  "doc_101",
				SourceName:  "history_woman",
				Title:       "Ада Лавлейс: Первая в истории программистка",
				Content:     "Ада Лавлейс, дочь поэта Джорджа Байрона, стала известна благодаря описанию вычислительной машины Чарльза Бэббиджа. Она составила первую в мире программу для этой машины, заложив основы алгоритмизации и концепции циклов.",
				URL:         "https://vk.com/wall-history_woman_101",
				PublishedAt: time.Now().Add(-24 * time.Hour),
			},
			{
				ChunkID:     "chunk_2",
				DocumentID:  "doc_102",
				SourceName:  "history_woman",
				Title:       "Мария Кюри и открытие радия",
				Content:     "Мария Склодовская-Кюри — первая женщина-лауреат Нобелевской премии и единственный учёный, получивший её в двух разных науках: физике и химии. Совместно с Пьером Кюри она открыла новые элементы — полоний и радий.",
				URL:         "https://vk.com/wall-history_woman_102",
				PublishedAt: time.Now().Add(-48 * time.Hour),
			},
		},
	}
}

func (r *SearchMockRepository) Search(ctx context.Context, query string, limit int) ([]domain.Post, error) {
	var filtered []domain.Post
	lowerQuery := strings.ToLower(query)

	for _, p := range r.posts {
		if query == "" || strings.Contains(strings.ToLower(p.Title), lowerQuery) ||
			strings.Contains(strings.ToLower(p.Content), lowerQuery) {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

func (r *SearchMockRepository) SaveLog(ctx context.Context, log *domain.SearchLog) error {
	query := `
		INSERT INTO search_logs (user_id, query, mode, limit_value, result_count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at;
	`
	err := r.pool.QueryRow(ctx, query,
		log.UserID,
		log.Query,
		log.Mode,
		log.LimitValue,
		log.ResultCount,
	).Scan(&log.ID, &log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert log in mock repo: %w", err)
	}
	return nil
}