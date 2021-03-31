package sqlite

import (
	"context"
	"fmt"
	"strings"

	journal "github.com/bertinatto/journal3"
)

var _ journal.UserService = (*UserService)(nil)

type UserService struct {
	db *DB
}

func NewUserService(db *DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (u *UserService) CreateUser(ctx context.Context, user *journal.User) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	user.CreatedAt = tx.now
	user.UpdatedAt = tx.now

	result, err := tx.ExecContext(ctx, `
		INSERT INTO user (
			api_key,
			name,
			email,
			password,
			created_at,
			updated_at
		)
		VALUES (?,?,?,?,?,?)
	`,
		user.APIKey,
		user.Name,
		user.Email,
		user.Password,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("could create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("could not get last inserted id: %w", err)
	}
	user.ID = int(id)

	return tx.Commit()
}

func (u *UserService) UpdateUser(ctx context.Context, id int, updated *journal.UserUpdate) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	user, err := findUserByID(ctx, tx, id)
	if err != nil {
		return err
	}

	if v := updated.Name; v != nil {
		user.Name = *v
	}

	if v := updated.Email; v != nil {
		user.Email = *v
	}

	if v := updated.Password; v != nil {
		user.Password = *v
	}

	user.UpdatedAt = tx.now

	_, err = tx.ExecContext(ctx, `
		UPDATE user
        SET name = ?,
			email = ?,
			password = ?,
			updated_at = ?
		WHERE id = ?
	`,
		user.Name,
		user.Email,
		user.Password,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("could not run query: %v", err)
	}

	return tx.Commit()

}

func (u *UserService) DeleteUser(ctx context.Context, id int) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	user, err := findUserByID(ctx, tx, id)
	if err != nil {
		return err
	}

	// TODO: check if user is the same as the one in context

	_, err = tx.ExecContext(ctx, `DELETE FROM user WHERE id = ?`, user.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (u *UserService) FindUsers(ctx context.Context) ([]*journal.User, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	users, n, err := findUsers(ctx, tx, &journal.UserFilter{})
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, &journal.Error{Code: journal.ENOTFOUND, Message: "There are no users available"}
	}

	return users, nil
}

func (u *UserService) FindUserByID(ctx context.Context, id int) (*journal.User, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := findUserByID(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	return user, err
}

func (u *UserService) FindUserByEmail(ctx context.Context, email string) (*journal.User, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := findUserByEmail(ctx, tx, email)
	if err != nil {
		return nil, err
	}

	return user, err
}

func findUserByID(ctx context.Context, tx *Tx, id int) (*journal.User, error) {
	users, n, err := findUsers(ctx, tx, &journal.UserFilter{ID: &id})
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, &journal.Error{Code: journal.ENOTFOUND, Message: "User(s) not found"}
	}

	return users[0], nil
}

func findUserByEmail(ctx context.Context, tx *Tx, email string) (*journal.User, error) {
	users, n, err := findUsers(ctx, tx, &journal.UserFilter{Email: &email})
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, &journal.Error{Code: journal.ENOTFOUND, Message: "User(s) not found"}
	}

	return users[0], nil
}

func findUsers(ctx context.Context, tx *Tx, filter *journal.UserFilter) ([]*journal.User, int, error) {
	// where and args should always be mutate together
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.Email; v != nil {
		where, args = append(where, "email = ?"), append(args, *v)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
		    id,
		    api_key,
		    name,
		    email,
		    password,
		    created_at,
		    updated_at,
		    COUNT(*) OVER()
		FROM user
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
	users := make([]*journal.User, 0)
	for rows.Next() {
		var user journal.User
		if err := rows.Scan(
			&user.ID,
			&user.APIKey,
			&user.Name,
			&user.Email,
			&user.Password,
			&user.CreatedAt,
			&user.UpdatedAt,
			&n,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, n, nil
}
