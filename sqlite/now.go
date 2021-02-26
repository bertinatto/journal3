package sqlite

import (
	"context"
	"fmt"
	"strings"

	journal "github.com/bertinatto/journal3"
)

var _ journal.NowService = &NowService{}

type NowService struct {
	db *DB
}

func NewNowService(db *DB) *NowService {
	return &NowService{
		db: db,
	}
}

func (j *NowService) CreateNow(ctx context.Context, now *journal.Now) error {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now.CreatedAt = tx.now
	now.UpdatedAt = tx.now

	result, err := tx.ExecContext(ctx, `
		INSERT INTO now (
			content,
            location,
			created_at,
			updated_at
		)
		VALUES (?,?,?,?)
	`,
		now.Content,
		now.FromLocation,
		now.CreatedAt,
		now.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("could not run query: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("could not get last inserted id: %v", err)
	}
	now.ID = int(id)

	return tx.Commit()
}

func (j *NowService) FindLatestNow(ctx context.Context) (*journal.Now, error) {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	p, err := findLatestNow(ctx, tx)
	if err != nil {
		return nil, err
	}

	return p, err

}

func findLatestNow(ctx context.Context, tx *Tx) (*journal.Now, error) {
	nows, n, err := findNows(ctx, tx, &journal.NowFilter{Limit: 1, Offset: 0})
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, fmt.Errorf("%s", "now content not found")
	}

	return nows[0], nil
}

func findNows(ctx context.Context, tx *Tx, filter *journal.NowFilter) ([]*journal.Now, int, error) {
	// where and args should always be mutate together
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
		    id,
            content,
            location,
		    created_at,
		    updated_at,
		    COUNT(*) OVER()
		FROM now
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
	nows := make([]*journal.Now, 0)
	for rows.Next() {
		var now journal.Now
		if err := rows.Scan(
			&now.ID,
			&now.Content,
			&now.FromLocation,
			&now.CreatedAt,
			&now.UpdatedAt,
			&n,
		); err != nil {
			return nil, 0, err
		}
		nows = append(nows, &now)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return nows, n, nil
}
