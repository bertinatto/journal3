package journal

import (
	"context"
	"time"
)

type Post struct {
	ID int `json:"id"`

	Title   string `json:"title"`
	Content string `json:"content"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PostFilter struct {
	ID *int `json:"id"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type JournalService interface {
	CreatePost(ctx context.Context, post *Post) error
	FindPostByID(ctx context.Context, id int) (*Post, error)
}
