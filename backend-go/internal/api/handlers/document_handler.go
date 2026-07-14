package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"vk-search/internal/domain"
)

type DocumentHandler struct {
	usecase domain.DocumentUseCase
}

func NewDocumentHandler(u domain.DocumentUseCase) *DocumentHandler {
	return &DocumentHandler{usecase: u}
}

func (h *DocumentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.writeError(w, "invalid document id", http.StatusBadRequest)
		return
	}

	doc, err := h.usecase.GetDocumentByID(r.Context(), id)
	if err != nil {
		h.writeError(w, "failed to retrieve document", http.StatusInternalServerError)
		return
	}

	if doc == nil {
		h.writeError(w, "document not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(doc); err != nil {
		h.writeError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *DocumentHandler) writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}