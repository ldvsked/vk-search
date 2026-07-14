package document

import (
	"context"
	"vk-search/internal/domain"
)

type usecase struct {
	repo domain.DocumentRepository
}

func NewDocumentUseCase(repo domain.DocumentRepository) domain.DocumentUseCase {
	return &usecase{
		repo: repo,
	}
}

func (u *usecase) GetDocumentByID(ctx context.Context, id int64) (*domain.Document, error) {
	return u.repo.GetByID(ctx, id)
}