package domain

import "context"

// Пользователь шлет сложный, длинный текст.
type AskRequest struct {
	Query string `json:"query"` // Длинный, неструктурированный вопрос пользователя
	TopK  int    `json:"top_k"` // Сколько документов искать в OpenSearch
}

// AskSource — структура одного источника информации для ИИ.
type AskSource struct {
	Score       float64 `json:"score"`
    DocumentID int64 `json:"document_id"`
    ChunkID    int64 `json:"chunk_id"` // Тоже меняем на int64 по ТЗ!
    Title      string `json:"title"`
    URL        string `json:"url"`
}

type AskResponse struct {
	Query    string      `json:"query"`     // Исходный длинный вопрос
	Answer   string      `json:"answer"`    // Сгенерированный лаконичный ответ
	Sources  []AskSource `json:"sources"`   // Источники, которые мы нашли по вычищенным ключевым словам
	LLMModel string      `json:"llm_model"` // Модель, которая отвечала
}

// Принимает длинный запрос, оркеструет вызовы ИИ и Поиска, возвращает структурированный ответ.
type AskUseCase interface {
	Ask(ctx context.Context, req AskRequest, userID int64) (*AskResponse, error)
}

type LLMClient interface {
	ExtractKeywords(ctx context.Context, rawQuery string) (string, error)
	GenerateAnswer(ctx context.Context, rawQuery string, docs []Post) (string, error)
	GetModelName() string
}