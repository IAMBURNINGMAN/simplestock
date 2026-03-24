package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"simplestock/internal/domain"
	"simplestock/internal/dto"
)

type DocumentRepo struct {
	db *pgxpool.Pool
}

func NewDocumentRepo(db *pgxpool.Pool) *DocumentRepo {
	return &DocumentRepo{db: db}
}

func (r *DocumentRepo) List(ctx context.Context, params dto.DocumentListParams) ([]domain.Document, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if params.DocType != "" {
		conditions = append(conditions, fmt.Sprintf("doc_type = $%d", argIdx))
		args = append(args, params.DocType)
		argIdx++
	}
	if params.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, params.Status)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := r.db.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM documents %s", where), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	query := fmt.Sprintf(`SELECT id, doc_type, doc_number, counterparty, expense_type, status, user_id, doc_date, created_at
		FROM documents %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, params.PageSize, (params.Page-1)*params.PageSize)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var docs []domain.Document
	for rows.Next() {
		var d domain.Document
		if err := rows.Scan(&d.ID, &d.DocType, &d.DocNumber, &d.Counterparty, &d.ExpenseType, &d.Status, &d.UserID, &d.DocDate, &d.CreatedAt); err != nil {
			return nil, 0, err
		}
		docs = append(docs, d)
	}
	return docs, total, nil
}

func (r *DocumentRepo) GetByID(ctx context.Context, id int64) (*domain.Document, error) {
	var d domain.Document
	err := r.db.QueryRow(ctx,
		`SELECT id, doc_type, doc_number, counterparty, expense_type, status, user_id, doc_date, created_at FROM documents WHERE id = $1`, id,
	).Scan(&d.ID, &d.DocType, &d.DocNumber, &d.Counterparty, &d.ExpenseType, &d.Status, &d.UserID, &d.DocDate, &d.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT di.id, di.document_id, di.product_id, p.name, p.sku, di.quantity, di.price
		FROM document_items di
		JOIN products p ON p.id = di.product_id
		WHERE di.document_id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.DocumentItem
		if err := rows.Scan(&item.ID, &item.DocumentID, &item.ProductID, &item.ProductName, &item.ProductSKU, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		d.Items = append(d.Items, item)
	}
	return &d, nil
}

func (r *DocumentRepo) NextDocNumber(ctx context.Context, docType string) (string, error) {
	var prefix string
	switch docType {
	case "incoming":
		prefix = "ПН"
	case "outgoing":
		prefix = "РН"
	default:
		prefix = "ДОК"
	}

	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) + 1 FROM documents WHERE doc_type = $1`, docType).Scan(&count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%05d", prefix, count), nil
}

func (r *DocumentRepo) CreateWithItems(ctx context.Context, d *domain.Document, items []domain.DocumentItem) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO documents (doc_type, doc_number, counterparty, expense_type, status, user_id, doc_date) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`,
		d.DocType, d.DocNumber, d.Counterparty, d.ExpenseType, d.Status, d.UserID, d.DocDate,
	).Scan(&d.ID, &d.CreatedAt)
	if err != nil {
		return err
	}

	for i := range items {
		items[i].DocumentID = d.ID
		err = tx.QueryRow(ctx,
			`INSERT INTO document_items (document_id, product_id, quantity, price) VALUES ($1, $2, $3, $4) RETURNING id`,
			items[i].DocumentID, items[i].ProductID, items[i].Quantity, items[i].Price,
		).Scan(&items[i].ID)
		if err != nil {
			return err
		}
	}
	d.Items = items

	return tx.Commit(ctx)
}

func (r *DocumentRepo) PostDocument(ctx context.Context, id int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var doc domain.Document
	err = tx.QueryRow(ctx,
		`SELECT id, doc_type, status FROM documents WHERE id = $1 FOR UPDATE`, id,
	).Scan(&doc.ID, &doc.DocType, &doc.Status)
	if err != nil {
		return err
	}
	if doc.Status != "draft" {
		return fmt.Errorf("документ уже проведён")
	}

	rows, err := tx.Query(ctx,
		`SELECT di.product_id, di.quantity, p.quantity as stock FROM document_items di JOIN products p ON p.id = di.product_id WHERE di.document_id = $1`, id)
	if err != nil {
		return err
	}
	defer rows.Close()

	type itemInfo struct {
		productID int64
		qty       int
		stock     int
	}
	var itemInfos []itemInfo
	for rows.Next() {
		var ii itemInfo
		if err := rows.Scan(&ii.productID, &ii.qty, &ii.stock); err != nil {
			return err
		}
		itemInfos = append(itemInfos, ii)
	}
	rows.Close()

	for _, ii := range itemInfos {
		if doc.DocType == "outgoing" && ii.stock < ii.qty {
			return fmt.Errorf("недостаточно товара (ID %d): на складе %d, запрошено %d", ii.productID, ii.stock, ii.qty)
		}

		var delta int
		var movementType string
		if doc.DocType == "incoming" {
			delta = ii.qty
			movementType = "incoming"
		} else {
			delta = -ii.qty
			movementType = "outgoing"
		}

		_, err = tx.Exec(ctx, `UPDATE products SET quantity = quantity + $1, updated_at = now() WHERE id = $2`, delta, ii.productID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO movements (product_id, document_id, movement_type, quantity) VALUES ($1, $2, $3, $4)`,
			ii.productID, id, movementType, ii.qty)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, `UPDATE documents SET status = 'posted' WHERE id = $1`, id)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *DocumentRepo) Delete(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, `DELETE FROM documents WHERE id = $1 AND status = 'draft'`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("документ не найден или уже проведён")
	}
	return nil
}

func (r *DocumentRepo) GetMovements(ctx context.Context, productID *int64, movementType string, from, to *time.Time, page, pageSize int) ([]domain.Movement, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if productID != nil {
		conditions = append(conditions, fmt.Sprintf("m.product_id = $%d", argIdx))
		args = append(args, *productID)
		argIdx++
	}
	if movementType != "" {
		conditions = append(conditions, fmt.Sprintf("m.movement_type = $%d", argIdx))
		args = append(args, movementType)
		argIdx++
	}
	if from != nil {
		conditions = append(conditions, fmt.Sprintf("m.created_at >= $%d", argIdx))
		args = append(args, *from)
		argIdx++
	}
	if to != nil {
		conditions = append(conditions, fmt.Sprintf("m.created_at <= $%d", argIdx))
		args = append(args, *to)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := r.db.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM movements m %s", where), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}

	query := fmt.Sprintf(`
		SELECT m.id, m.product_id, p.name, p.sku, m.document_id, m.inventory_id, m.movement_type, m.quantity, m.created_at
		FROM movements m
		JOIN products p ON p.id = m.product_id
		%s ORDER BY m.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, pageSize, (page-1)*pageSize)

	mvRows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer mvRows.Close()

	var movements []domain.Movement
	for mvRows.Next() {
		var mv domain.Movement
		if err := mvRows.Scan(&mv.ID, &mv.ProductID, &mv.ProductName, &mv.ProductSKU, &mv.DocumentID, &mv.InventoryID, &mv.MovementType, &mv.Quantity, &mv.CreatedAt); err != nil {
			return nil, 0, err
		}
		movements = append(movements, mv)
	}
	return movements, total, nil
}

// Used by inventory completion
func (r *DocumentRepo) CreateMovementTx(ctx context.Context, tx pgx.Tx, mv *domain.Movement) error {
	return tx.QueryRow(ctx,
		`INSERT INTO movements (product_id, inventory_id, movement_type, quantity) VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
		mv.ProductID, mv.InventoryID, mv.MovementType, mv.Quantity,
	).Scan(&mv.ID, &mv.CreatedAt)
}
