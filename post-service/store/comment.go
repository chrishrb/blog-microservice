package store

import (
	"context"
	"time"
)

type Comment struct {
	ID        string
	AuthorID  string
	PostID    string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CommentStore interface {
	SetComment(ctx context.Context, comment *Comment) error
	LookupComment(ctx context.Context, postId, ID string) (*Comment, error)
	ListCommentsByPostID(ctx context.Context, postID string, offset, limit int) ([]*Comment, error)
	DeleteComment(ctx context.Context, postID, ID string) error
}
