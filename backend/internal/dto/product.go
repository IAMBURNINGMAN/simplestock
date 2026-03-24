package dto

type CreateProductRequest struct {
	Name          string   `json:"name"`
	SKU           string   `json:"sku"`
	CategoryID    *int64   `json:"category_id"`
	Unit          string   `json:"unit"`
	MinStock      int      `json:"min_stock"`
	PurchasePrice *float64 `json:"purchase_price"`
}

type UpdateProductRequest struct {
	Name          string   `json:"name"`
	SKU           string   `json:"sku"`
	CategoryID    *int64   `json:"category_id"`
	Unit          string   `json:"unit"`
	MinStock      int      `json:"min_stock"`
	PurchasePrice *float64 `json:"purchase_price"`
}

type ProductListParams struct {
	Search     string
	CategoryID *int64
	Page       int
	PageSize   int
}
