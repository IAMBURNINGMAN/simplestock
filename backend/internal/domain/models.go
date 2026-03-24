package domain

import (
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Product struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	SKU           string     `json:"sku"`
	CategoryID    *int64     `json:"category_id"`
	CategoryName  string     `json:"category_name,omitempty"`
	Unit          string     `json:"unit"`
	MinStock      int        `json:"min_stock"`
	Quantity      int        `json:"quantity"`
	PurchasePrice *float64   `json:"purchase_price"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Document struct {
	ID           int64          `json:"id"`
	DocType      string         `json:"doc_type"`
	DocNumber    string         `json:"doc_number"`
	Counterparty *string        `json:"counterparty"`
	ExpenseType  *string        `json:"expense_type"`
	Status       string         `json:"status"`
	UserID       int64          `json:"user_id"`
	DocDate      time.Time      `json:"doc_date"`
	CreatedAt    time.Time      `json:"created_at"`
	Items        []DocumentItem `json:"items,omitempty"`
}

type DocumentItem struct {
	ID          int64    `json:"id"`
	DocumentID  int64    `json:"document_id"`
	ProductID   int64    `json:"product_id"`
	ProductName string   `json:"product_name,omitempty"`
	ProductSKU  string   `json:"product_sku,omitempty"`
	Quantity    int      `json:"quantity"`
	Price       *float64 `json:"price"`
}

type Movement struct {
	ID           int64     `json:"id"`
	ProductID    int64     `json:"product_id"`
	ProductName  string    `json:"product_name,omitempty"`
	ProductSKU   string    `json:"product_sku,omitempty"`
	DocumentID   *int64    `json:"document_id"`
	InventoryID  *int64    `json:"inventory_id"`
	MovementType string    `json:"movement_type"`
	Quantity     int       `json:"quantity"`
	CreatedAt    time.Time `json:"created_at"`
}

type Inventory struct {
	ID          int64           `json:"id"`
	InvNumber   string          `json:"inv_number"`
	Status      string          `json:"status"`
	UserID      int64           `json:"user_id"`
	StartedAt   time.Time       `json:"started_at"`
	CompletedAt *time.Time      `json:"completed_at"`
	Items       []InventoryItem `json:"items,omitempty"`
}

type InventoryItem struct {
	ID               int64  `json:"id"`
	InventoryID      int64  `json:"inventory_id"`
	ProductID        int64  `json:"product_id"`
	ProductName      string `json:"product_name,omitempty"`
	ProductSKU       string `json:"product_sku,omitempty"`
	ExpectedQuantity int    `json:"expected_quantity"`
	ActualQuantity   int    `json:"actual_quantity"`
	Difference       int    `json:"difference"`
}
