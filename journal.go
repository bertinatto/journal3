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

type PostUpdate struct {
	Title     *string    `json:"title"`
	Content   *string    `json:"content"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

type JournalService interface {
	CreatePost(ctx context.Context, post *Post) (err error)
	UpdatePost(ctx context.Context, permalink string, updated *PostUpdate) (err error)
	FindPostByID(ctx context.Context, id int) (post *Post, err error)
	FindPosts(ctx context.Context) (posts []*Post, err error)
	FindPostByPermalink(ctx context.Context, permalink string) (post *Post, err error)
}

type contextKey int

const (
	userContextKey = contextKey(iota + 1)
)

func NewContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserIDFromContext(ctx context.Context) int {
	user := UserFromContext(ctx)
	if user == nil {
		return 0
	}
	return user.ID
}

func UserFromContext(ctx context.Context) *User {
	user, _ := ctx.Value(userContextKey).(*User)
	return user
}
