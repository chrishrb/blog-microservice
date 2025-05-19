package store

import (
	"context"
	"time"
)

type Post struct {
	ID        string
	AuthorID  string
	Title     string
	Content   string
	Tags      []string
	Published bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PostStore interface {
	SetPost(ctx context.Context, post *Post) error
	LookupPost(ctx context.Context, ID string) (*Post, error)
	ListPosts(ctx context.Context) ([]*Post, error)
	DeletePost(ctx context.Context, ID string) error
}
