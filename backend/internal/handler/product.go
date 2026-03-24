package handler

import (
	"net/http"

	"simplestock/internal/domain"
	"simplestock/internal/dto"
	"simplestock/internal/repository"
)

type ProductHandler struct {
	productRepo *repository.ProductRepo
}

func NewProductHandler(productRepo *repository.ProductRepo) *ProductHandler {
	return &ProductHandler{productRepo: productRepo}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	params := dto.ProductListParams{
		Search:     r.URL.Query().Get("search"),
		CategoryID: queryInt64Ptr(r, "category_id"),
		Page:       queryInt(r, "page", 1),
		PageSize:   queryInt(r, "page_size", 20),
	}

	products, total, err := h.productRepo.List(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения товаров")
		return
	}
	if products == nil {
		products = []domain.Product{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  products,
		"total": total,
		"page":  params.Page,
	})
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "неверный ID")
		return
	}

	product, err := h.productRepo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "товар не найден")
		return
	}

	writeJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if req.Name == "" || req.SKU == "" {
		writeError(w, http.StatusBadRequest, "название и артикул обязательны")
		return
	}
	if req.Unit == "" {
		req.Unit = "шт"
	}

	p := &domain.Product{
		Name:          req.Name,
		SKU:           req.SKU,
		CategoryID:    req.CategoryID,
		Unit:          req.Unit,
		MinStock:      req.MinStock,
		PurchasePrice: req.PurchasePrice,
	}

	if err := h.productRepo.Create(r.Context(), p); err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка создания товара: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "неверный ID")
		return
	}

	var req dto.UpdateProductRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if req.Name == "" || req.SKU == "" {
		writeError(w, http.StatusBadRequest, "название и артикул обязательны")
		return
	}

	p := &domain.Product{
		ID:            id,
		Name:          req.Name,
		SKU:           req.SKU,
		CategoryID:    req.CategoryID,
		Unit:          req.Unit,
		MinStock:      req.MinStock,
		PurchasePrice: req.PurchasePrice,
	}

	if err := h.productRepo.Update(r.Context(), p); err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка обновления товара")
		return
	}

	updated, _ := h.productRepo.GetByID(r.Context(), id)
	writeJSON(w, http.StatusOK, updated)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "неверный ID")
		return
	}

	if err := h.productRepo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка удаления товара")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "товар удалён"})
}

func (h *ProductHandler) LowStock(w http.ResponseWriter, r *http.Request) {
	products, err := h.productRepo.LowStock(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения данных")
		return
	}
	if products == nil {
		products = []domain.Product{}
	}
	writeJSON(w, http.StatusOK, products)
}
