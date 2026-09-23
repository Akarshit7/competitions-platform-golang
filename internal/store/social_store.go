package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Akarshit7/competition-platform-go/internal/models"
)

type SocialStore struct {
	pool *pgxpool.Pool
}

func NewSocialStore(pool *pgxpool.Pool) *SocialStore {
	return &SocialStore{pool: pool}
}

// Submissions
func (s *SocialStore) CreateSubmission(ctx context.Context, userID string, videoID *string, title string, description *string) (*models.Submission, error) {
	var sub models.Submission
	err := s.pool.QueryRow(ctx,
		`INSERT INTO submissions (user_id, video_id, title, description) VALUES ($1,$2,$3,$4) RETURNING id, user_id, video_id, title, description, status, created_at`,
		userID, videoID, title, description).Scan(&sub.ID, &sub.UserID, &sub.VideoID, &sub.Title, &sub.Description, &sub.Status, &sub.CreatedAt)
	return &sub, err
}

func (s *SocialStore) ListSubmissions(ctx context.Context, limit, offset int) ([]models.Submission, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, user_id, video_id, title, description, status, created_at FROM submissions ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Submission
	for rows.Next() {
		var m models.Submission
		if err := rows.Scan(&m.ID, &m.UserID, &m.VideoID, &m.Title, &m.Description, &m.Status, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// Vote: upsert
func (s *SocialStore) Vote(ctx context.Context, userID, submissionID string, value int) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO votes (user_id, submission_id, value) VALUES ($1,$2,$3) ON CONFLICT (user_id, submission_id) DO UPDATE SET value=EXCLUDED.value`,
		userID, submissionID, value)
	return err
}

func (s *SocialStore) SubmissionExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM submissions WHERE id=$1)`, id).Scan(&exists)
	return exists, err
}

// Like toggle: if exists delete else insert; returns true if now liked
func (s *SocialStore) ToggleLike(ctx context.Context, userID, videoID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM likes WHERE user_id=$1 AND video_id=$2)`, userID, videoID).Scan(&exists)
	if err != nil {
		return false, err
	}
	if exists {
		_, err = s.pool.Exec(ctx, `DELETE FROM likes WHERE user_id=$1 AND video_id=$2`, userID, videoID)
		return false, err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO likes (user_id, video_id) VALUES ($1,$2)`, userID, videoID)
	return true, err
}

// Comments
func (s *SocialStore) CreateComment(ctx context.Context, userID, videoID, body string) (*models.Comment, error) {
	var c models.Comment
	err := s.pool.QueryRow(ctx, `INSERT INTO comments (user_id, video_id, body) VALUES ($1,$2,$3) RETURNING id, user_id, video_id, body, created_at`, userID, videoID, body).
		Scan(&c.ID, &c.UserID, &c.VideoID, &c.Body, &c.CreatedAt)
	return &c, err
}

func (s *SocialStore) ListComments(ctx context.Context, videoID string, limit, offset int) ([]models.Comment, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, user_id, video_id, body, created_at FROM comments WHERE video_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, videoID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Comment
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(&c.ID, &c.UserID, &c.VideoID, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *SocialStore) CreateReply(ctx context.Context, userID, commentID, body string) (*models.Reply, error) {
	var r models.Reply
	err := s.pool.QueryRow(ctx, `INSERT INTO comment_replies (user_id, comment_id, body) VALUES ($1,$2,$3) RETURNING id, user_id, comment_id, body, created_at`, userID, commentID, body).
		Scan(&r.ID, &r.UserID, &r.CommentID, &r.Body, &r.CreatedAt)
	return &r, err
}

func (s *SocialStore) ListReplies(ctx context.Context, commentID string) ([]models.Reply, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, user_id, comment_id, body, created_at FROM comment_replies WHERE comment_id=$1 ORDER BY created_at ASC`, commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Reply
	for rows.Next() {
		var r models.Reply
		if err := rows.Scan(&r.ID, &r.UserID, &r.CommentID, &r.Body, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *SocialStore) CommentExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM comments WHERE id=$1)`, id).Scan(&exists)
	return exists, err
}
