package journal

import (
	"context"
	"time"
)

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
