package dto

type CreateDocumentRequest struct {
	DocType      string               `json:"doc_type"`
	Counterparty *string              `json:"counterparty"`
	ExpenseType  *string              `json:"expense_type"`
	DocDate      string               `json:"doc_date"`
	Items        []CreateDocItemRequest `json:"items"`
}

type CreateDocItemRequest struct {
	ProductID int64    `json:"product_id"`
	Quantity  int      `json:"quantity"`
	Price     *float64 `json:"price"`
}

type DocumentListParams struct {
	DocType string
	Status  string
	Page    int
	PageSize int
}
