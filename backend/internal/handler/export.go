package handler

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

type ExportHandler struct {
	db *pgxpool.Pool
}

func NewExportHandler(db *pgxpool.Pool) *ExportHandler {
	return &ExportHandler{db: db}
}

func (h *ExportHandler) ExportExcel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	f := excelize.NewFile()
	defer f.Close()

	// Sheet 1: Current Stock
	f.SetSheetName("Sheet1", "Остатки")
	headers := []string{"Артикул", "Товар", "Категория", "Ед.", "Остаток", "Мин.", "Цена", "Стоимость"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Остатки", cell, h)
	}

	rows, err := h.db.Query(ctx, `
		SELECT p.sku, p.name, COALESCE(c.name, ''), p.unit, p.quantity, p.min_stock, COALESCE(p.purchase_price, 0)
		FROM products p LEFT JOIN categories c ON c.id = p.category_id ORDER BY p.name
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка экспорта")
		return
	}
	defer rows.Close()

	row := 2
	for rows.Next() {
		var sku, name, cat, unit string
		var qty, minStock int
		var price float64
		rows.Scan(&sku, &name, &cat, &unit, &qty, &minStock, &price)

		f.SetCellValue("Остатки", fmt.Sprintf("A%d", row), sku)
		f.SetCellValue("Остатки", fmt.Sprintf("B%d", row), name)
		f.SetCellValue("Остатки", fmt.Sprintf("C%d", row), cat)
		f.SetCellValue("Остатки", fmt.Sprintf("D%d", row), unit)
		f.SetCellValue("Остатки", fmt.Sprintf("E%d", row), qty)
		f.SetCellValue("Остатки", fmt.Sprintf("F%d", row), minStock)
		f.SetCellValue("Остатки", fmt.Sprintf("G%d", row), price)
		f.SetCellValue("Остатки", fmt.Sprintf("H%d", row), float64(qty)*price)
		row++
	}

	// Sheet 2: Turnover (if dates provided)
	if from != "" && to != "" {
		f.NewSheet("Обороты")
		turnHeaders := []string{"Артикул", "Товар", "Приход", "Расход", "Корр. +", "Корр. -"}
		for i, h := range turnHeaders {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue("Обороты", cell, h)
		}

		tRows, err := h.db.Query(ctx, `
			SELECT p.sku, p.name,
				COALESCE(SUM(CASE WHEN m.movement_type='incoming' THEN m.quantity END), 0),
				COALESCE(SUM(CASE WHEN m.movement_type='outgoing' THEN m.quantity END), 0),
				COALESCE(SUM(CASE WHEN m.movement_type='correction_plus' THEN m.quantity END), 0),
				COALESCE(SUM(CASE WHEN m.movement_type='correction_minus' THEN m.quantity END), 0)
			FROM products p
			LEFT JOIN movements m ON m.product_id = p.id AND m.created_at >= $1::date AND m.created_at < ($2::date + interval '1 day')
			GROUP BY p.id, p.sku, p.name ORDER BY p.name
		`, from, to)
		if err == nil {
			defer tRows.Close()
			tRow := 2
			for tRows.Next() {
				var sku, name string
				var in_, out_, cp, cm int
				tRows.Scan(&sku, &name, &in_, &out_, &cp, &cm)
				f.SetCellValue("Обороты", fmt.Sprintf("A%d", tRow), sku)
				f.SetCellValue("Обороты", fmt.Sprintf("B%d", tRow), name)
				f.SetCellValue("Обороты", fmt.Sprintf("C%d", tRow), in_)
				f.SetCellValue("Обороты", fmt.Sprintf("D%d", tRow), out_)
				f.SetCellValue("Обороты", fmt.Sprintf("E%d", tRow), cp)
				f.SetCellValue("Обороты", fmt.Sprintf("F%d", tRow), cm)
				tRow++
			}
		}
	}

	// Style headers
	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"#D5E8F0"}, Pattern: 1}})
	for _, sheet := range f.GetSheetList() {
		for i := 1; i <= 8; i++ {
			cell, _ := excelize.CoordinatesToCellName(i, 1)
			f.SetCellStyle(sheet, cell, cell, style)
		}
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=simplestock-report.xlsx")
	f.Write(w)
}
