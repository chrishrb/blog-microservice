package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Token struct {
	UserID    uuid.UUID
	Token     string
	TTL       time.Duration
	Revoked   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type JWTBlacklistStore interface {
	SetToken(ctx context.Context, token *Token) error
	GetToken(ctx context.Context, token string) (*Token, error)
	SetTokenRevoked(ctx context.Context, userID uuid.UUID) error
	IsTokenRevoked(ctx context.Context, token string) (bool, error)
}
