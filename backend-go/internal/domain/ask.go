package domain

import "context"

type AskRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k"` 
}

type AskSource struct {
	Score       float64 `json:"score"`
    DocumentID int64 `json:"document_id"`
    ChunkID    int64 `json:"chunk_id"` 
    Title      string `json:"title"`
    URL        string `json:"url"`
}

type AskResponse struct {
	Query    string      `json:"query"`     
	Answer   string      `json:"answer"`    
	Sources  []AskSource `json:"sources"`  
	LLMModel string      `json:"llm_model"` 
}

type AskUseCase interface {
	Ask(ctx context.Context, req AskRequest, userID int64) (*AskResponse, error)
}

type LLMClient interface {
	ExtractKeywords(ctx context.Context, rawQuery string) (string, error)
	GenerateAnswer(ctx context.Context, rawQuery string, docs []Post) (string, error)
	GetModelName() string
}