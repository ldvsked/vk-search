package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"vk-search/internal/domain"
)

type SearchHandler struct {
	useCase domain.SearchUseCase
}

func NewSearchHandler(useCase domain.SearchUseCase) *SearchHandler {
	return &SearchHandler{useCase: useCase}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	results, err := h.useCase.Execute(r.Context(), query, limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "search failed"}`))
		return
	}

	response := map[string]interface{}{
		"query":   query,
		"limit":   limit,
		"count":   len(results),
		"results": results,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}