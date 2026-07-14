package handlers

import (
	"encoding/json"
	"net/http"
	"vk-search/internal/domain"
)

type StatsHandler struct {
	usecase domain.StatsUseCase
}

func NewStatsHandler(u domain.StatsUseCase) *StatsHandler {
	return &StatsHandler{usecase: u}
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.usecase.GetStats(r.Context())
	if err != nil {
		h.writeError(w, "failed to fetch statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.writeError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *StatsHandler) writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}