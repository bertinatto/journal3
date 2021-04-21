package journal

import (
	"context"
	"time"
)

type Page struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PageFilter struct {
	Name   *string `json:"name"`
	Offset int     `json:"offset"`
	Limit  int     `json:"limit"`
}

type PageUpdate struct {
	Content   *string    `json:"content"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

type PageService interface {
	CreatePage(ctx context.Context, page *Page) (err error)
	UpdatePage(ctx context.Context, name string, updated *PageUpdate) (err error)
	FindPageByName(ctx context.Context, name string) (page *Page, err error)
}
