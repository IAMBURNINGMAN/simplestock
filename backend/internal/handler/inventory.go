package handler

import (
	"net/http"

	"simplestock/internal/domain"
	"simplestock/internal/dto"
	"simplestock/internal/middleware"
	"simplestock/internal/repository"
)

type InventoryHandler struct {
	invRepo *repository.InventoryRepo
}

func NewInventoryHandler(invRepo *repository.InventoryRepo) *InventoryHandler {
	return &InventoryHandler{invRepo: invRepo}
}

func (h *InventoryHandler) List(w http.ResponseWriter, r *http.Request) {
	invs, err := h.invRepo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения инвентаризаций")
		return
	}
	if invs == nil {
		invs = []domain.Inventory{}
	}
	writeJSON(w, http.StatusOK, invs)
}

func (h *InventoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "неверный ID")
		return
	}
	inv, err := h.invRepo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "инвентаризация не найдена")
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

func (h *InventoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	inv := &domain.Inventory{
		UserID: middleware.GetUserID(r.Context()),
	}

	if err := h.invRepo.Create(r.Context(), inv); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, inv)
}

func (h *InventoryHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	invID, err := urlParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "неверный ID")
		return
	}

	var req dto.AddInventoryItemRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	item := &domain.InventoryItem{
		InventoryID:    invID,
		ProductID:      req.ProductID,
		ActualQuantity: req.ActualQuantity,
	}

	if err := h.invRepo.AddItem(r.Context(), invID, item); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *InventoryHandler) Complete(w http.ResponseWriter, r *http.Request) {
	invID, err := urlParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "неверный ID")
		return
	}

	inv, err := h.invRepo.Complete(r.Context(), invID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, inv)
}
