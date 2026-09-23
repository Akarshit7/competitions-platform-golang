package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Akarshit7/competition-platform-go/internal/models"
)

type AdminStore struct {
	pool *pgxpool.Pool
}

func NewAdminStore(pool *pgxpool.Pool) *AdminStore {
	return &AdminStore{pool: pool}
}

func (s *AdminStore) FindByEmail(ctx context.Context, email string) (*models.Admin, error) {
	var a models.Admin
	err := s.pool.QueryRow(ctx, `SELECT id, email, password_hash, name, created_at FROM admins WHERE email=$1`, email).
		Scan(&a.ID, &a.Email, &a.PasswordHash, &a.Name, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *AdminStore) FindByID(ctx context.Context, id string) (*models.Admin, error) {
	var a models.Admin
	err := s.pool.QueryRow(ctx, `SELECT id, email, password_hash, name, created_at FROM admins WHERE id=$1`, id).
		Scan(&a.ID, &a.Email, &a.PasswordHash, &a.Name, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
