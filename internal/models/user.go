package models

import "time"

type User struct {
	ID        string    `json:"id"`
	Phone     string    `json:"phone"`
	Name      *string   `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Video struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  *string   `json:"description"`
	URL          *string   `json:"url"`
	ThumbnailURL *string   `json:"thumbnail_url"`
	UserID       *string   `json:"user_id"`
	ViewCount    int       `json:"view_count"`
	CreatedAt    time.Time `json:"created_at"`
	LikeCount    int       `json:"like_count,omitempty"`
	CommentCount int       `json:"comment_count,omitempty"`
}

type Submission struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	VideoID     *string   `json:"video_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	VoteScore   int       `json:"vote_score,omitempty"`
}

type Comment struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	VideoID   string    `json:"video_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type Reply struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CommentID string    `json:"comment_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}
