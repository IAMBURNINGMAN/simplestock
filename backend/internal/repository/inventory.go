package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"simplestock/internal/domain"
)

type InventoryRepo struct {
	db *pgxpool.Pool
}

func NewInventoryRepo(db *pgxpool.Pool) *InventoryRepo {
	return &InventoryRepo{db: db}
}

func (r *InventoryRepo) List(ctx context.Context) ([]domain.Inventory, error) {
	rows, err := r.db.Query(ctx, `SELECT id, inv_number, status, user_id, started_at, completed_at FROM inventories ORDER BY started_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invs []domain.Inventory
	for rows.Next() {
		var inv domain.Inventory
		if err := rows.Scan(&inv.ID, &inv.InvNumber, &inv.Status, &inv.UserID, &inv.StartedAt, &inv.CompletedAt); err != nil {
			return nil, err
		}
		invs = append(invs, inv)
	}
	return invs, nil
}

func (r *InventoryRepo) GetByID(ctx context.Context, id int64) (*domain.Inventory, error) {
	var inv domain.Inventory
	err := r.db.QueryRow(ctx,
		`SELECT id, inv_number, status, user_id, started_at, completed_at FROM inventories WHERE id = $1`, id,
	).Scan(&inv.ID, &inv.InvNumber, &inv.Status, &inv.UserID, &inv.StartedAt, &inv.CompletedAt)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT ii.id, ii.inventory_id, ii.product_id, p.name, p.sku, ii.expected_quantity, ii.actual_quantity, ii.difference
		FROM inventory_items ii
		JOIN products p ON p.id = ii.product_id
		WHERE ii.inventory_id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.InventoryItem
		if err := rows.Scan(&item.ID, &item.InventoryID, &item.ProductID, &item.ProductName, &item.ProductSKU, &item.ExpectedQuantity, &item.ActualQuantity, &item.Difference); err != nil {
			return nil, err
		}
		inv.Items = append(inv.Items, item)
	}
	return &inv, nil
}

func (r *InventoryRepo) Create(ctx context.Context, inv *domain.Inventory) error {
	var count int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM inventories WHERE status = 'active'`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("уже есть активная инвентаризация")
	}

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) + 1 FROM inventories`).Scan(&total); err != nil {
		return err
	}
	inv.InvNumber = fmt.Sprintf("ИНВ-%05d", total)

	return r.db.QueryRow(ctx,
		`INSERT INTO inventories (inv_number, status, user_id) VALUES ($1, $2, $3) RETURNING id, started_at`,
		inv.InvNumber, "active", inv.UserID,
	).Scan(&inv.ID, &inv.StartedAt)
}

func (r *InventoryRepo) AddItem(ctx context.Context, invID int64, item *domain.InventoryItem) error {
	var status string
	if err := r.db.QueryRow(ctx, `SELECT status FROM inventories WHERE id = $1`, invID).Scan(&status); err != nil {
		return err
	}
	if status != "active" {
		return fmt.Errorf("инвентаризация уже завершена")
	}

	var currentQty int
	if err := r.db.QueryRow(ctx, `SELECT quantity FROM products WHERE id = $1`, item.ProductID).Scan(&currentQty); err != nil {
		return err
	}
	item.ExpectedQuantity = currentQty

	return r.db.QueryRow(ctx,
		`INSERT INTO inventory_items (inventory_id, product_id, expected_quantity, actual_quantity)
		 VALUES ($1, $2, $3, $4) RETURNING id, difference`,
		invID, item.ProductID, item.ExpectedQuantity, item.ActualQuantity,
	).Scan(&item.ID, &item.Difference)
}

func (r *InventoryRepo) Complete(ctx context.Context, invID int64) (*domain.Inventory, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM inventories WHERE id = $1 FOR UPDATE`, invID).Scan(&status); err != nil {
		return nil, err
	}
	if status != "active" {
		return nil, fmt.Errorf("инвентаризация уже завершена")
	}

	rows, err := tx.Query(ctx, `SELECT product_id, expected_quantity, actual_quantity, difference FROM inventory_items WHERE inventory_id = $1`, invID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type diffItem struct {
		productID  int64
		expected   int
		actual     int
		difference int
	}
	var diffs []diffItem
	for rows.Next() {
		var d diffItem
		if err := rows.Scan(&d.productID, &d.expected, &d.actual, &d.difference); err != nil {
			return nil, err
		}
		diffs = append(diffs, d)
	}
	rows.Close()

	for _, d := range diffs {
		if d.difference == 0 {
			continue
		}

		_, err = tx.Exec(ctx, `UPDATE products SET quantity = $1, updated_at = now() WHERE id = $2`, d.actual, d.productID)
		if err != nil {
			return nil, err
		}

		movementType := "correction_plus"
		qty := d.difference
		if d.difference < 0 {
			movementType = "correction_minus"
			qty = -d.difference
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO movements (product_id, inventory_id, movement_type, quantity) VALUES ($1, $2, $3, $4)`,
			d.productID, invID, movementType, qty)
		if err != nil {
			return nil, err
		}
	}

	now := time.Now()
	_, err = tx.Exec(ctx, `UPDATE inventories SET status = 'completed', completed_at = $1 WHERE id = $2`, now, invID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, invID)
}
