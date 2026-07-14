package domain

import (
	"context"
	"time"
)

type Document struct {
	ID          int64      `json:"id"`
	SourceName  string     `json:"source_name"` 
	Title       string     `json:"title"`
	Text        string     `json:"text"`        
	URL         string     `json:"url"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

type DocumentRepository interface {
	GetByID(ctx context.Context, id int64) (*Document, error)
}

type DocumentUseCase interface {
	GetDocumentByID(ctx context.Context, id int64) (*Document, error)
}