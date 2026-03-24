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

func (h *DashboardHandler) StockReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rows, err := h.db.Query(ctx, `
		SELECT p.id, p.name, p.sku, COALESCE(c.name, ''), p.unit, p.min_stock, p.quantity, COALESCE(p.purchase_price, 0)
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		ORDER BY p.name
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка формирования отчёта")
		return
	}
	defer rows.Close()

	type stockRow struct {
		ID       int64   `json:"id"`
		Name     string  `json:"name"`
		SKU      string  `json:"sku"`
		Category string  `json:"category"`
		Unit     string  `json:"unit"`
		MinStock int     `json:"min_stock"`
		Quantity int     `json:"quantity"`
		Price    float64 `json:"price"`
		Value    float64 `json:"value"`
	}

	var result []stockRow
	for rows.Next() {
		var sr stockRow
		rows.Scan(&sr.ID, &sr.Name, &sr.SKU, &sr.Category, &sr.Unit, &sr.MinStock, &sr.Quantity, &sr.Price)
		sr.Value = float64(sr.Quantity) * sr.Price
		result = append(result, sr)
	}
	if result == nil {
		result = []stockRow{}
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *DashboardHandler) TurnoverReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		writeError(w, http.StatusBadRequest, "параметры from и to обязательны (формат YYYY-MM-DD)")
		return
	}

	rows, err := h.db.Query(ctx, `
		SELECT p.id, p.name, p.sku, p.unit,
			COALESCE(SUM(CASE WHEN m.movement_type = 'incoming' THEN m.quantity ELSE 0 END), 0) as total_in,
			COALESCE(SUM(CASE WHEN m.movement_type = 'outgoing' THEN m.quantity ELSE 0 END), 0) as total_out,
			COALESCE(SUM(CASE WHEN m.movement_type = 'correction_plus' THEN m.quantity ELSE 0 END), 0) as corrections_plus,
			COALESCE(SUM(CASE WHEN m.movement_type = 'correction_minus' THEN m.quantity ELSE 0 END), 0) as corrections_minus
		FROM products p
		LEFT JOIN movements m ON m.product_id = p.id AND m.created_at >= $1::date AND m.created_at < ($2::date + interval '1 day')
		GROUP BY p.id, p.name, p.sku, p.unit
		HAVING SUM(m.quantity) > 0 OR SUM(m.quantity) IS NOT NULL
		ORDER BY p.name
	`, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка формирования отчёта: "+err.Error())
		return
	}
	defer rows.Close()

	type turnoverRow struct {
		ID              int64  `json:"id"`
		Name            string `json:"name"`
		SKU             string `json:"sku"`
		Unit            string `json:"unit"`
		TotalIn         int    `json:"total_in"`
		TotalOut        int    `json:"total_out"`
		CorrectionsPlus int    `json:"corrections_plus"`
		CorrectionsMinus int   `json:"corrections_minus"`
	}

	var result []turnoverRow
	for rows.Next() {
		var tr turnoverRow
		rows.Scan(&tr.ID, &tr.Name, &tr.SKU, &tr.Unit, &tr.TotalIn, &tr.TotalOut, &tr.CorrectionsPlus, &tr.CorrectionsMinus)
		result = append(result, tr)
	}
	if result == nil {
		result = []turnoverRow{}
	}
	writeJSON(w, http.StatusOK, result)
}
