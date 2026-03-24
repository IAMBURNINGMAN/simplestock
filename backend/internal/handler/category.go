package handler

import (
	"net/http"

	"simplestock/internal/domain"
	"simplestock/internal/repository"
)

type CategoryHandler struct {
	categoryRepo *repository.CategoryRepo
}

func NewCategoryHandler(categoryRepo *repository.CategoryRepo) *CategoryHandler {
	return &CategoryHandler{categoryRepo: categoryRepo}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	cats, err := h.categoryRepo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения категорий")
		return
	}
	if cats == nil {
		cats = []domain.Category{}
	}
	writeJSON(w, http.StatusOK, cats)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "название категории обязательно")
		return
	}

	c := &domain.Category{Name: req.Name}
	if err := h.categoryRepo.Create(r.Context(), c); err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка создания категории: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, c)
}
