package api

import (
	"net/http"
	"vk-search/internal/api/handlers"
	"vk-search/internal/api/middleware"
	"vk-search/internal/infrastructure/config"
)

func NewRouter(authHandler *handlers.AuthHandler, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	jwtSecret := []byte(cfg.GetJWTSecret())
	authMiddleware := middleware.AuthMiddleware(jwtSecret)

	testSearchHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, _ := r.Context().Value(middleware.UsernameKey).(string)
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Привет, ` + username + `! Доступ к поиску открыт."}`))
	})

	mux.Handle("GET /api/v1/search", authMiddleware(testSearchHandler))

	return mux
}
