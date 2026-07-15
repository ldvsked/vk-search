package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"vk-search/internal/domain"
	"vk-search/internal/infrastructure/config"
)

type openRouterClient struct {
	cfg *config.Config
}

func NewLLMClient(cfg *config.Config) domain.LLMClient {
	return &openRouterClient{
		cfg: cfg,
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func (c *openRouterClient) GetModelName() string {
	return c.cfg.GetLLMModel()
}

func (c *openRouterClient) readPromptFile(filename string) (string, error) {
	path := filepath.Join("internal", "infrastructure", "llm", "prompts", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		fallbackPath := filepath.Join("..", "..", "internal", "infrastructure", "llm", "prompts", filename)
		data, err = os.ReadFile(fallbackPath)
		if err != nil {
			return "", fmt.Errorf("failed to read prompt file %s: %w", filename, err)
		}
	}
	return string(data), nil
}

func (c *openRouterClient) ExtractKeywords(ctx context.Context, rawQuery string) (string, error) {
	template, err := c.readPromptFile("keywords.txt")
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(template, rawQuery)
	return c.sendRequest(ctx, prompt)
}

func (c *openRouterClient) GenerateAnswer(ctx context.Context, rawQuery string, docs []domain.Post) (string, error) {
	template, err := c.readPromptFile("answer.txt")
	if err != nil {
		return "", err
	}

	var docsContext string
	for i, d := range docs {
		content := d.Content
		if len(content) > 2000 {
			content = content[:2000] + "..."
		}
		docsContext += fmt.Sprintf("[%d]\nЗаголовок: %s\nИсточник: %s\nСсылка: %s\nТекст: %s\n\n", 
            i+1, 
            d.Title, 
            d.SourceName, 
            d.URL, 
            content,
        )
	}

	prompt := fmt.Sprintf(template, docsContext, rawQuery)
	return c.sendRequest(ctx, prompt)
}

func (c *openRouterClient) sendRequest(ctx context.Context, prompt string) (string, error) {
	if !c.cfg.IsLLMEnabled() {
		return "", fmt.Errorf("LLM disabled")
	}

	apiKey := c.cfg.GetLLMAPIKey()
	if apiKey == "" {
		return "", fmt.Errorf("missing LLM API key")
	}

	reqBody := chatRequest{
		Model: c.cfg.GetLLMModel(),
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: time.Duration(c.cfg.GetLLMTimeout()) * time.Second,
	}

	apiURL := c.cfg.GetLLMBaseURL()
	if !strings.HasSuffix(apiURL, "/chat/completions") {
		apiURL = fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(apiURL, "/"))
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/misis-student/vk-search")
	req.Header.Set("X-Title", "VK Search Academic Project")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var responseData chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return "", err
	}

	if len(responseData.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}

	return strings.TrimSpace(responseData.Choices[0].Message.Content), nil
}