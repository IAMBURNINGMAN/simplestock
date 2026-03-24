package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"simplestock/internal/domain"
)

type CategoryRepo struct {
	db *pgxpool.Pool
}

func NewCategoryRepo(db *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

func (r *CategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO categories (name) VALUES ($1) RETURNING id`,
		c.Name,
	).Scan(&c.ID)
}

func (r *CategoryRepo) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	var c domain.Category
	err := r.db.QueryRow(ctx, `SELECT id, name FROM categories WHERE id = $1`, id).Scan(&c.ID, &c.Name)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
