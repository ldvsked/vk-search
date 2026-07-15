package ask

import (
	"context"
	"fmt"
	"log"

	"vk-search/internal/domain"
)

type askUseCase struct {
	searchUC  domain.SearchUseCase
	llmClient domain.LLMClient
}

func NewAskUseCase(searchUC domain.SearchUseCase, llmClient domain.LLMClient) domain.AskUseCase {
	return &askUseCase{
		searchUC:  searchUC,
		llmClient: llmClient,
	}
}

func (uc *askUseCase) Ask(ctx context.Context, req domain.AskRequest, userID int64) (*domain.AskResponse, error) {
	if req.TopK <= 0 {
		req.TopK = 3
	}

	cleanedQuery, err := uc.llmClient.ExtractKeywords(ctx, req.Query)
	log.Printf("[DEBUG] Original: %q, Cleaned: %q", req.Query, cleanedQuery)
	if err != nil {
		cleanedQuery = req.Query
	}

	askCtx := context.WithValue(ctx, domain.ModeKey, "ask")

	posts, err := uc.searchUC.Execute(askCtx, cleanedQuery, req.TopK, "", "", "")
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	sources := make([]domain.AskSource, 0, len(posts))
	for _, p := range posts {
		sources = append(sources, domain.AskSource{
			Score:      p.Score,
			DocumentID: p.DocumentID, // int64 -> int64
			ChunkID:    p.ChunkID,    // int64 -> int64
			Title:      p.Title,
			URL:        p.URL,
		})
	}

	answer, err := uc.llmClient.GenerateAnswer(ctx, req.Query, posts)
	log.Printf("[DEBUG] LLM generation failed: %v", err)
	if err != nil {
		answer = "Извините, не удалось получить ответ от ИИ. Источники найдены, но генератор ответов временно недоступен."
	}

	return &domain.AskResponse{
		Query:    req.Query,
		Answer:   answer,
		Sources:  sources,
		LLMModel: uc.llmClient.GetModelName(),
	}, nil
}