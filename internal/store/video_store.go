package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Akarshit7/competition-platform-go/internal/models"
)

type VideoStore struct {
	pool *pgxpool.Pool
}

func NewVideoStore(pool *pgxpool.Pool) *VideoStore {
	return &VideoStore{pool: pool}
}

func (s *VideoStore) ListLatest(ctx context.Context, limit, offset int) ([]models.Video, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, title, description, url, thumbnail_url, user_id, view_count, created_at FROM videos ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var videos []models.Video
	for rows.Next() {
		var v models.Video
		if err := rows.Scan(&v.ID, &v.Title, &v.Description, &v.URL, &v.ThumbnailURL, &v.UserID, &v.ViewCount, &v.CreatedAt); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, rows.Err()
}

func (s *VideoStore) ListNew(ctx context.Context, limit, offset int) ([]models.Video, error) {
	// New = last 48h else fallback latest; for simplicity same as latest but could filter
	rows, err := s.pool.Query(ctx, `SELECT id, title, description, url, thumbnail_url, user_id, view_count, created_at FROM videos WHERE created_at > NOW() - INTERVAL '48 hours' ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var videos []models.Video
	for rows.Next() {
		var v models.Video
		if err := rows.Scan(&v.ID, &v.Title, &v.Description, &v.URL, &v.ThumbnailURL, &v.UserID, &v.ViewCount, &v.CreatedAt); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	if len(videos) == 0 {
		return s.ListLatest(ctx, limit, offset)
	}
	return videos, rows.Err()
}

func (s *VideoStore) ListTrending(ctx context.Context, limit, offset int) ([]models.Video, error) {
	// Trending = view_count desc + like count recent 7 days
	rows, err := s.pool.Query(ctx, `
		SELECT v.id, v.title, v.description, v.url, v.thumbnail_url, v.user_id, v.view_count, v.created_at,
		       COALESCE(l.cnt,0) as like_cnt
		FROM videos v
		LEFT JOIN (SELECT video_id, COUNT(*) cnt FROM likes WHERE created_at > NOW() - INTERVAL '7 days' GROUP BY video_id) l ON l.video_id = v.id
		WHERE v.created_at > NOW() - INTERVAL '30 days'
		ORDER BY (v.view_count + COALESCE(l.cnt,0)*10) DESC, v.created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var videos []models.Video
	for rows.Next() {
		var v models.Video
		var likeCnt int
		if err := rows.Scan(&v.ID, &v.Title, &v.Description, &v.URL, &v.ThumbnailURL, &v.UserID, &v.ViewCount, &v.CreatedAt, &likeCnt); err != nil {
			return nil, err
		}
		v.LikeCount = likeCnt
		videos = append(videos, v)
	}
	return videos, rows.Err()
}

func (s *VideoStore) FindByID(ctx context.Context, id string) (*models.Video, error) {
	var v models.Video
	err := s.pool.QueryRow(ctx, `SELECT id, title, description, url, thumbnail_url, user_id, view_count, created_at FROM videos WHERE id=$1`, id).
		Scan(&v.ID, &v.Title, &v.Description, &v.URL, &v.ThumbnailURL, &v.UserID, &v.ViewCount, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (s *VideoStore) Exists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM videos WHERE id=$1)`, id).Scan(&exists)
	return exists, err
}
