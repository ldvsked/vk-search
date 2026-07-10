package mocks

import (
	"context"
	"strings"
	"time"
	"vk-search/internal/domain"
)

type SearchMockRepository struct {
	posts []domain.Post
}

func NewSearchMockRepository() domain.SearchRepository {
	return &SearchMockRepository{
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