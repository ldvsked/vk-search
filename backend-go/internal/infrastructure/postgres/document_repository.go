package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vk-search/internal/domain"
)

type documentRepository struct {
	pool *pgxpool.Pool
}

func NewDocumentRepository(pool *pgxpool.Pool) domain.DocumentRepository {
	return &documentRepository{
		pool: pool,
	}
}

func (r *documentRepository) GetByID(ctx context.Context, id int64) (*domain.Document, error) {
	query := `
		SELECT 
			d.id, 
			s.name as source_name, 
			d.title, 
			d.text, 
			d.url, 
			d.published_at 
		FROM documents d
		JOIN sources s ON d.source_id = s.id
		WHERE d.id = $1
		LIMIT 1;
	`

	var doc domain.Document

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID,
		&doc.SourceName,
		&doc.Title,
		&doc.Text,
		&doc.URL,
		&doc.PublishedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Если документа нет, возвращаем nil без ошибки (это обработает UseCase/Handler)
		}
		return nil, fmt.Errorf("failed to fetch document from db: %w", err)
	}

	return &doc, nil
}