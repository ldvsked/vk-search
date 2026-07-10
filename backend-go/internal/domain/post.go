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

type SearchRepository interface {
	Search(ctx context.Context, query string, limit int) ([]Post, error)
}

type SearchUseCase interface {
	Execute(ctx context.Context, query string, limit int) ([]Post, error)
}