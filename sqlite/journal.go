package sqlite

import (
	"context"
	"fmt"
	"strings"

	journal "github.com/bertinatto/journal3"
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
			permalink,
			title,
			content,
			created_at,
			updated_at
		)
		VALUES (?,?,?,?,?)
	`,
		post.Permalink,
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
	return p, err
}

func (j *JournalService) FindPostByPermalink(ctx context.Context, permalink string) (*journal.Post, error) {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	p, err := findPostByPermalink(ctx, tx, permalink)
	if err != nil {
		return nil, err
	}
	return p, err
}

func (j *JournalService) FindPosts(ctx context.Context) ([]*journal.Post, error) {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	posts, n, err := findPosts(ctx, tx, &journal.PostFilter{})
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, &journal.Error{Code: journal.ENOTFOUND, Message: "There are no posts available"}
	}

	return posts, nil
}

func findPostByPermalink(ctx context.Context, tx *Tx, permalink string) (*journal.Post, error) {
	posts, n, err := findPosts(ctx, tx, &journal.PostFilter{Permalink: &permalink})
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, &journal.Error{Code: journal.ENOTFOUND, Message: "There are no posts available with permalink"}
	}

	return posts[0], nil
}

func findPostByID(ctx context.Context, tx *Tx, id int) (*journal.Post, error) {
	posts, n, err := findPosts(ctx, tx, &journal.PostFilter{ID: &id})
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, fmt.Errorf("%s", "post not found")
	}

	return posts[0], nil
}

func findPosts(ctx context.Context, tx *Tx, filter *journal.PostFilter) ([]*journal.Post, int, error) {
	// where and args should always be mutate together
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.Permalink; v != nil {
		where, args = append(where, "permalink = ?"), append(args, *v)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
		    id,
		    permalink,
		    title,
            content,
		    created_at,
		    updated_at,
		    COUNT(*) OVER()
		FROM posts
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
	posts := make([]*journal.Post, 0)
	for rows.Next() {
		var post journal.Post
		if err := rows.Scan(
			&post.ID,
			&post.Permalink,
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

func formatLimitAndOffset(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf(`LIMIT %d OFFSET %d`, limit, offset)
	} else if limit > 0 {
		return fmt.Sprintf(`LIMIT %d`, limit)
	} else if offset > 0 {
		return fmt.Sprintf(`OFFSET %d`, offset)
	}
	return ""
}
