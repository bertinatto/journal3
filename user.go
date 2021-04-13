package journal

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	APIKey    string    `json:"api_key"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (u *User) Validate() error {
	if len(u.Password) < 6 {
		return fmt.Errorf("password must be at least 6 char long")
	}
	if !strings.Contains(u.Email, "@") {
		return fmt.Errorf("invalid email address")
	}
	return nil
}

type UserFilter struct {
	ID     *int    `json:"id"`
	Email  *string `json:"email"`
	Offset int     `json:"offset"`
	Limit  int     `json:"limit"`
}

type UserUpdate struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type UserService interface {
	CreateUser(ctx context.Context, user *User) (err error)
	UpdateUser(ctx context.Context, id int, updated *UserUpdate) (err error)
	DeleteUser(ctx context.Context, id int) (err error)
	FindUsers(ctx context.Context) (users []*User, err error)
	FindUserByID(ctx context.Context, id int) (user *User, err error)
	FindUserByEmail(ctx context.Context, email string) (user *User, err error)
}
