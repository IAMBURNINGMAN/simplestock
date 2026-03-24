package dto

type AddInventoryItemRequest struct {
	ProductID      int64 `json:"product_id"`
	ActualQuantity int   `json:"actual_quantity"`
}
