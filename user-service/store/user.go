package store

import (
	"context"
	"time"
)

type User struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserStore interface {
	SetUser(ctx context.Context, user *User) error
	LookupUser(ctx context.Context, ID string) (*User, error)
	ListUsers(ctx context.Context, offset, limit int) ([]*User, error)
	DeleteUser(ctx context.Context, ID string) error
}
