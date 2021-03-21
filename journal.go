package journal

import (
	"context"
	"time"
)

type Post struct {
	ID        int       `json:"id"`
	Permalink string    `json:"permalink"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PostFilter struct {
	ID        *int    `json:"id"`
	Permalink *string `json:"permalink"`
	Offset    int     `json:"offset"`
	Limit     int     `json:"limit"`
}

type JournalService interface {
	CreatePost(ctx context.Context, post *Post) (err error)
	FindPostByID(ctx context.Context, id int) (post *Post, err error)
	FindPosts(ctx context.Context) (posts []*Post, err error)
	FindPostByPermalink(ctx context.Context, permalink string) (post *Post, err error)
}

type Now struct {
	ID           int       `json:"id"`
	Content      string    `json:"content"`
	FromLocation string    `json:"from_location"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type NowFilter struct {
	ID     *int `json:"id"`
	Offset int  `json:"offset"`
	Limit  int  `json:"limit"`
}

type NowService interface {
	CreateNow(ctx context.Context, now *Now) (err error)
	FindLatestNow(ctx context.Context) (now *Now, err error)
}
