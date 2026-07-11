package config

import (
	"fmt"
	"os"

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
}

func Load() (*Config, error) {
	err := godotenv.Load("../.env")
	if err != nil {
		return nil, fmt.Errorf("failed to load .env from project root: %w", err)
	}

	cfg := &Config{
		serverPort: os.Getenv("SERVER_PORT"),
		jwtSecret:  os.Getenv("JWT_SECRET"),
		dbHost:     os.Getenv("POSTGRES_HOST"),
		dbPort:     os.Getenv("POSTGRES_PORT"),
		dbUser:     os.Getenv("POSTGRES_USER"),
		dbPass:     os.Getenv("POSTGRES_PASSWORD"),
		dbName:     os.Getenv("POSTGRES_DB"),
	}

	return cfg, nil
}

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