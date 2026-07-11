package domain

import (
	"context"
	"time"
)

type Post struct {
	ChunkID     string    `json:"chunk_id"`
	DocumentID  string    `json:"document_id"`
	SourceName  string    `json:"source_name"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
}

type SearchLog struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Query       string    `json:"query"`
	Mode        string    `json:"mode"`
	LimitValue  int       `json:"limit_value"`
	ResultCount int       `json:"result_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type SearchRepository interface {
	Search(ctx context.Context, query string, limit int) ([]Post, error)
	SaveLog(ctx context.Context, log *SearchLog) error
}

type SearchUseCase interface {
	Execute(ctx context.Context, query string, limit int) ([]Post, error)
}