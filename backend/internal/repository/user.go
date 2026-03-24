package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"simplestock/internal/domain"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx,
		`SELECT id, username, password, full_name, role, created_at FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.Password, &u.FullName, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx,
		`SELECT id, username, password, full_name, role, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Username, &u.Password, &u.FullName, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO users (username, password, full_name, role) VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
		u.Username, u.Password, u.FullName, u.Role,
	).Scan(&u.ID, &u.CreatedAt)
}
