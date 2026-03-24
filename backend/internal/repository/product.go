package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"simplestock/internal/domain"
	"simplestock/internal/dto"
)

type ProductRepo struct {
	db *pgxpool.Pool
}

func NewProductRepo(db *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) List(ctx context.Context, params dto.ProductListParams) ([]domain.Product, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(p.name ILIKE $%d OR p.sku ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}
	if params.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("p.category_id = $%d", argIdx))
		args = append(args, *params.CategoryID)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products p %s", where)
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	offset := (params.Page - 1) * params.PageSize

	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.sku, p.category_id, COALESCE(c.name, ''), p.unit, p.min_stock, p.quantity, p.purchase_price, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		%s
		ORDER BY p.name
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, params.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.SKU, &p.CategoryID, &p.CategoryName, &p.Unit, &p.MinStock, &p.Quantity, &p.PurchasePrice, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, nil
}

func (r *ProductRepo) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	var p domain.Product
	err := r.db.QueryRow(ctx, `
		SELECT p.id, p.name, p.sku, p.category_id, COALESCE(c.name, ''), p.unit, p.min_stock, p.quantity, p.purchase_price, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1
	`, id).Scan(&p.ID, &p.Name, &p.SKU, &p.CategoryID, &p.CategoryName, &p.Unit, &p.MinStock, &p.Quantity, &p.PurchasePrice, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) Create(ctx context.Context, p *domain.Product) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO products (name, sku, category_id, unit, min_stock, purchase_price) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, quantity, created_at, updated_at`,
		p.Name, p.SKU, p.CategoryID, p.Unit, p.MinStock, p.PurchasePrice,
	).Scan(&p.ID, &p.Quantity, &p.CreatedAt, &p.UpdatedAt)
}

func (r *ProductRepo) Update(ctx context.Context, p *domain.Product) error {
	_, err := r.db.Exec(ctx,
		`UPDATE products SET name=$1, sku=$2, category_id=$3, unit=$4, min_stock=$5, purchase_price=$6, updated_at=now() WHERE id=$7`,
		p.Name, p.SKU, p.CategoryID, p.Unit, p.MinStock, p.PurchasePrice, p.ID,
	)
	return err
}

func (r *ProductRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	return err
}

func (r *ProductRepo) LowStock(ctx context.Context) ([]domain.Product, error) {
	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.name, p.sku, p.category_id, COALESCE(c.name, ''), p.unit, p.min_stock, p.quantity, p.purchase_price, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.quantity <= p.min_stock AND p.min_stock > 0
		ORDER BY (p.quantity::float / NULLIF(p.min_stock, 0)) ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.SKU, &p.CategoryID, &p.CategoryName, &p.Unit, &p.MinStock, &p.Quantity, &p.PurchasePrice, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}
