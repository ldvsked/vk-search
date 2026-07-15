package domain

import (
	"context"
	"time"
)

type Post struct {
	Score       float64 `json:"score"`
    ChunkID     int64     `json:"chunk_id"`    // Теперь это тоже int64!
    DocumentID  int64     `json:"document_id"` // Уже исправили на int64
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
	Search(ctx context.Context, query string, limit int, source string, dateFrom string, dateTo string) ([]Post, error)
	SaveLog(ctx context.Context, log *SearchLog) error
}

type SearchUseCase interface {
    Execute(ctx context.Context, query string, limit int, source string, dateFrom string, dateTo string) ([]Post, error)
}