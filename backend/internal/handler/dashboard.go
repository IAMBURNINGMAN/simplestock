package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardHandler struct {
	db *pgxpool.Pool
}

func NewDashboardHandler(db *pgxpool.Pool) *DashboardHandler {
	return &DashboardHandler{db: db}
}

func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var totalProducts int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM products`).Scan(&totalProducts)

	var lowStockCount int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE quantity <= min_stock AND min_stock > 0`).Scan(&lowStockCount)

	var todayMovements int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM movements WHERE created_at::date = CURRENT_DATE`).Scan(&todayMovements)

	var todayIncoming int
	h.db.QueryRow(ctx, `SELECT COALESCE(SUM(quantity), 0) FROM movements WHERE movement_type = 'incoming' AND created_at::date = CURRENT_DATE`).Scan(&todayIncoming)

	var todayOutgoing int
	h.db.QueryRow(ctx, `SELECT COALESCE(SUM(quantity), 0) FROM movements WHERE movement_type = 'outgoing' AND created_at::date = CURRENT_DATE`).Scan(&todayOutgoing)

	var totalStockValue float64
	h.db.QueryRow(ctx, `SELECT COALESCE(SUM(quantity * COALESCE(purchase_price, 0)), 0) FROM products`).Scan(&totalStockValue)

	writeJSON(w, http.StatusOK, map[string]any{
		"total_products":    totalProducts,
		"low_stock_count":   lowStockCount,
		"today_movements":   todayMovements,
		"today_incoming":    todayIncoming,
		"today_outgoing":    todayOutgoing,
		"total_stock_value": totalStockValue,
	})
}
