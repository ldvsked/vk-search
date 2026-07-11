package handlers

import (
	"encoding/json"
	"net/http"
	"vk-search/internal/domain"
)

type HealthHandler struct {
	useCase domain.HealthUseCase
}

func NewHealthHandler(uc domain.HealthUseCase) *HealthHandler {
	return &HealthHandler{useCase: uc}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
    status := "ok"
    dbStatus := "up"

    if err := h.useCase.CheckHealth(r.Context()); err != nil {
        status = "degraded" // Приложение живо, но зависимость упала
        dbStatus = "down"
    }

    w.Header().Set("Content-Type", "application/json")
    
    if dbStatus == "down" {
        w.WriteHeader(http.StatusServiceUnavailable)
    } else {
        w.WriteHeader(http.StatusOK)
    }

    json.NewEncoder(w).Encode(map[string]string{
        "status":    status, 
        "database":  dbStatus,
        "backend":   "up",  
    })
}