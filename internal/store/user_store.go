package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Akarshit7/competition-platform-go/internal/models"
)

type UserStore struct {
	pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{pool: pool}
}

func (s *UserStore) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx, `SELECT id, phone, name, created_at FROM users WHERE phone=$1`, phone).
		Scan(&u.ID, &u.Phone, &u.Name, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) FindByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx, `SELECT id, phone, name, created_at FROM users WHERE id=$1`, id).
		Scan(&u.ID, &u.Phone, &u.Name, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) UpsertByPhone(ctx context.Context, phone string) (*models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (phone) VALUES ($1) ON CONFLICT (phone) DO UPDATE SET updated_at=NOW() RETURNING id, phone, name, created_at`, phone).
		Scan(&u.ID, &u.Phone, &u.Name, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
