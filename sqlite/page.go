package sqlite

import (
	"context"
	"log"
	"strings"

	journal "github.com/bertinatto/journal3"
)

var _ journal.PageService = (*PageService)(nil)

type PageService struct {
	db *DB
}

func NewPageService(db *DB) *PageService {
	return &PageService{
		db: db,
	}
}

func (p *PageService) CreatePage(ctx context.Context, page *journal.Page) error {
	log.Println("CreatePage start")
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	page.CreatedAt = tx.now
	page.UpdatedAt = tx.now

	result, err := tx.ExecContext(ctx, `
		INSERT INTO page (
			name,
			content,
			created_at,
			updated_at
		)
		VALUES (?,?,?,?)
	`,
		page.Name,
		page.Content,
		page.CreatedAt,
		page.UpdatedAt,
	)
	if err != nil {
		return err
	}
	log.Println("CreatePage inserted")

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	page.ID = int(id)

	return tx.Commit()
}

func (p *PageService) UpdatePage(ctx context.Context, name string, updated *journal.PageUpdate) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	page, err := findPageByName(ctx, tx, name)
	if err != nil {
		return err
	}

	if v := updated.Content; v != nil {
		page.Content = *v
	}

	page.UpdatedAt = tx.now

	_, err = tx.ExecContext(ctx, `
		UPDATE page
        SET content = ?,
			updated_at = ?
		WHERE id = ?
	`,
		page.Content,
		page.UpdatedAt,
		page.ID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (p *PageService) FindPageByName(ctx context.Context, name string) (*journal.Page, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	page, err := findPageByName(ctx, tx, name)
	if err != nil {
		return nil, err
	}

	return page, err
}

func findPageByName(ctx context.Context, tx *Tx, name string) (*journal.Page, error) {
	pages, n, err := findPages(ctx, tx, &journal.PageFilter{Name: &name})
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, &journal.Error{Code: journal.ENOTFOUND, Message: "Page not found"}
	}

	return pages[0], nil
}

func findPages(ctx context.Context, tx *Tx, filter *journal.PageFilter) ([]*journal.Page, int, error) {
	// where and args should always be mutate together
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.Name; v != nil {
		where, args = append(where, "name = ?"), append(args, *v)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
		    id,
		    name,
            content,
		    created_at,
		    updated_at,
		    COUNT(*) OVER()
		FROM page
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY id DESC
		`+formatLimitAndOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var n int
	pages := make([]*journal.Page, 0)
	for rows.Next() {
		var page journal.Page
		if err := rows.Scan(
			&page.ID,
			&page.Name,
			&page.Content,
			&page.CreatedAt,
			&page.UpdatedAt,
			&n,
		); err != nil {
			return nil, 0, err
		}
		pages = append(pages, &page)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return pages, n, nil
}
