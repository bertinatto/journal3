package sqlite

import (
	"context"
	"fmt"
	"strings"

	journal "github.com/bertinatto/journal3"
	"k8s.io/klog/v2"
)

var _ journal.JournalService = &JournalService{}

type JournalService struct {
	db *DB
}

func NewJournalService(db *DB) *JournalService {
	return &JournalService{
		db: db,
	}
}

func (j *JournalService) CreatePost(ctx context.Context, post *journal.Post) error {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	post.CreatedAt = tx.now
	post.UpdatedAt = tx.now

	result, err := tx.ExecContext(ctx, `
		INSERT INTO posts (
			title,
			content,
			created_at,
			updated_at
		)
		VALUES (?,?,?,?)
	`,
		post.Title,
		post.Content,
		post.CreatedAt,
		post.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("could not run query: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("could not get last inserted id: %v", err)
	}
	post.ID = int(id)

	return tx.Commit()
}

func (j *JournalService) FindPostByID(ctx context.Context, id int) (*journal.Post, error) {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	p, err := findPostByID(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	klog.Infof("fjb: %+v", *p)

	return p, err

}

func findPostByID(ctx context.Context, tx *Tx, id int) (*journal.Post, error) {
	posts, _, err := findPosts(ctx, tx, &journal.PostFilter{ID: &id})
	if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return nil, fmt.Errorf("%s", "not found")
	}

	return posts[0], nil
}

func findPosts(ctx context.Context, tx *Tx, filter *journal.PostFilter) ([]*journal.Post, int, error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}

	// Execue query with limiting WHERE clause and LIMIT/OFFSET injected.
	rows, err := tx.QueryContext(ctx, `
		SELECT
		    id,
		    title,
            content,
		    created_at,
		    updated_at,
		    COUNT(*) OVER()
		FROM posts
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY id ASC
		`+FormatLimitOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Iterate over rows and deserialize into Dial objects.
	var n int
	posts := make([]*journal.Post, 0)
	for rows.Next() {
		var post journal.Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
			&n,
		); err != nil {
			return nil, 0, err
		}
		posts = append(posts, &post)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return posts, n, nil
}

// FormatLimitOffset returns a SQL string for a given limit & offset.
// Clauses are only added if limit and/or offset are greater than zero.
func FormatLimitOffset(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf(`LIMIT %d OFFSET %d`, limit, offset)
	} else if limit > 0 {
		return fmt.Sprintf(`LIMIT %d`, limit)
	} else if offset > 0 {
		return fmt.Sprintf(`OFFSET %d`, offset)
	}
	return ""
}
