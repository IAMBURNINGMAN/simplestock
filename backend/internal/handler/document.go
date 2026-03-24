package handler

import (
	"net/http"
	"time"

	"simplestock/internal/domain"
	"simplestock/internal/dto"
	"simplestock/internal/middleware"
	"simplestock/internal/repository"
)

type DocumentHandler struct {
	docRepo *repository.DocumentRepo
}

func NewDocumentHandler(docRepo *repository.DocumentRepo) *DocumentHandler {
	return &DocumentHandler{docRepo: docRepo}
}

func (h *DocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	params := dto.DocumentListParams{
		DocType:  r.URL.Query().Get("doc_type"),
		Status:   r.URL.Query().Get("status"),
		Page:     queryInt(r, "page", 1),
		PageSize: queryInt(r, "page_size", 20),
	}

	docs, total, err := h.docRepo.List(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения документов")
		return
	}
	if docs == nil {
		docs = []domain.Document{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": docs, "total": total, "page": params.Page})
}

func (h *DocumentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "неверный ID")
		return
	}
	doc, err := h.docRepo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "документ не найден")
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateDocumentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if req.DocType != "incoming" && req.DocType != "outgoing" {
		writeError(w, http.StatusBadRequest, "тип документа должен быть incoming или outgoing")
		return
	}
	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "документ должен содержать хотя бы одну позицию")
		return
	}

	docDate := time.Now()
	if req.DocDate != "" {
		parsed, err := time.Parse("2006-01-02", req.DocDate)
		if err == nil {
			docDate = parsed
		}
	}

	docNumber, err := h.docRepo.NextDocNumber(r.Context(), req.DocType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка генерации номера")
		return
	}

	doc := &domain.Document{
		DocType:      req.DocType,
		DocNumber:    docNumber,
		Counterparty: req.Counterparty,
		ExpenseType:  req.ExpenseType,
		Status:       "draft",
		UserID:       middleware.GetUserID(r.Context()),
		DocDate:      docDate,
	}

	var items []domain.DocumentItem
	for _, it := range req.Items {
		if it.Quantity <= 0 {
			writeError(w, http.StatusBadRequest, "количество должно быть положительным")
			return
		}
		items = append(items, domain.DocumentItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			Price:     it.Price,
		})
	}

	if err := h.docRepo.CreateWithItems(r.Context(), doc, items); err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка создания документа: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, doc)
}

func (h *DocumentHandler) Post(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "неверный ID")
		return
	}

	if err := h.docRepo.PostDocument(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	doc, _ := h.docRepo.GetByID(r.Context(), id)
	writeJSON(w, http.StatusOK, doc)
}

func (h *DocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "неверный ID")
		return
	}

	if err := h.docRepo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "документ удалён"})
}

func (h *DocumentHandler) Movements(w http.ResponseWriter, r *http.Request) {
	productID := queryInt64Ptr(r, "product_id")
	movementType := r.URL.Query().Get("type")
	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "page_size", 20)

	var from, to *time.Time
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = &t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			end := t.Add(24*time.Hour - time.Nanosecond)
			to = &end
		}
	}

	movements, total, err := h.docRepo.GetMovements(r.Context(), productID, movementType, from, to, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения движений")
		return
	}
	if movements == nil {
		movements = []domain.Movement{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": movements, "total": total, "page": page})
}
