package api

import (
	"net/http"
	"vk-search/internal/api/handlers"
	"vk-search/internal/api/middleware"
	"vk-search/internal/infrastructure/config"
)

func NewRouter(authHandler *handlers.AuthHandler,
	searchHandler *handlers.SearchHandler, 
	healthHandler *handlers.HealthHandler,
	cfg *config.Config,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("GET /api/v1/health", healthHandler.Check)

	jwtSecret := []byte(cfg.GetJWTSecret())
	authMiddleware := middleware.AuthMiddleware(jwtSecret)

	mux.Handle("GET /api/v1/search", authMiddleware(http.HandlerFunc(searchHandler.Search)))

	return mux
}