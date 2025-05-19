package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive  = "active"
	StatusPending = "pending"
	StatusBanned  = "banned"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID           uuid.UUID
	Email        string
	FirstName    string
	LastName     string
	PasswordHash string
	Status       string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserStore interface {
	SetUser(ctx context.Context, user *User) error
	LookupUser(ctx context.Context, ID uuid.UUID) (*User, error)
	LookupUserByEmail(ctx context.Context, email string) (*User, error)
	ListUsers(ctx context.Context, offset, limit int) ([]*User, error)
	DeleteUser(ctx context.Context, ID uuid.UUID) error
}
