package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	serverPort string
	jwtSecret  string
	dbHost     string
	dbPort     string
	dbUser     string
	dbPass     string
	dbName     string

	// LLM настройки
	llmEnabled          bool
	llmBaseURL          string
	llmAPIKey           string
	llmModel            string
	llmTimeoutSeconds   int
	llmMaxContextChars  int
}

func Load() (*Config, error) {
	// Твоя проверенная загрузка .env
	_ = godotenv.Load()

	// Парсим булевое значение для LLM_ENABLED с дефолтом true
	llmEnabled := true
	if enabledStr := os.Getenv("LLM_ENABLED"); enabledStr != "" {
		if val, err := strconv.ParseBool(enabledStr); err == nil {
			llmEnabled = val
		}
	}

	// Парсим таймаут с дефолтом 40 секунд
	llmTimeout := 40
	if timeoutStr := os.Getenv("LLM_TIMEOUT_SECONDS"); timeoutStr != "" {
		if val, err := strconv.Atoi(timeoutStr); err == nil {
			llmTimeout = val
		}
	}

	// Парсим лимит символов контекста с дефолтом 12000
	llmMaxChars := 12000
	if charsStr := os.Getenv("LLM_MAX_CONTEXT_CHARS"); charsStr != "" {
		if val, err := strconv.Atoi(charsStr); err == nil {
			llmMaxChars = val
		}
	}

	cfg := &Config{
		serverPort:         os.Getenv("SERVER_PORT"),
		jwtSecret:          os.Getenv("JWT_SECRET"),
		dbHost:             os.Getenv("POSTGRES_HOST"),
		dbPort:             os.Getenv("POSTGRES_PORT"),
		dbUser:             os.Getenv("POSTGRES_USER"),
		dbPass:             os.Getenv("POSTGRES_PASSWORD"),
		dbName:             os.Getenv("POSTGRES_DB"),

		// Инициализация LLM полей
		llmEnabled:         llmEnabled,
		llmBaseURL:         os.Getenv("LLM_BASE_URL"),
		llmAPIKey:          os.Getenv("LLM_API_KEY"),
		llmModel:           os.Getenv("LLM_MODEL"),
		llmTimeoutSeconds:  llmTimeout,
		llmMaxContextChars: llmMaxChars,
	}

	return cfg, nil
}

// --- Существующие геттеры ---

func (c *Config) GetJWTSecret() string {
	return c.jwtSecret
}

func (c *Config) GetHTTPPort() string {
	if c.serverPort == "" {
		return "8080"
	}
	return c.serverPort
}

func (c *Config) GetDBDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.dbUser,
		c.dbPass,
		c.dbHost,
		c.dbPort,
		c.dbName,
	)
}

// --- Новые геттеры для LLM ---

func (c *Config) IsLLMEnabled() bool {
	return c.llmEnabled
}

func (c *Config) GetLLMBaseURL() string {
	if c.llmBaseURL == "" {
		return "https://openrouter.ai/api/v1"
	}
	return c.llmBaseURL
}

func (c *Config) GetLLMAPIKey() string {
	return c.llmAPIKey
}

func (c *Config) GetLLMModel() string {
	if c.llmModel == "" {
		return "meta-llama/llama-3-8b-instruct:free"
	}
	return c.llmModel
}

func (c *Config) GetLLMTimeout() int {
	return c.llmTimeoutSeconds
}

func (c *Config) GetLLMMaxContextChars() int {
	return c.llmMaxContextChars
}