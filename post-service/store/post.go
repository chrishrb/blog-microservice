package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID        uuid.UUID
	AuthorID  uuid.UUID
	Title     string
	Content   string
	Tags      []string
	Published bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PostStore interface {
	SetPost(ctx context.Context, post *Post) error
	LookupPost(ctx context.Context, ID uuid.UUID) (*Post, error)
	ListPosts(ctx context.Context, offset, limit int) ([]*Post, error)
	DeletePost(ctx context.Context, ID uuid.UUID) error
}
