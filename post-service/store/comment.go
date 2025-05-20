package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID
	AuthorID  uuid.UUID
	PostID    uuid.UUID
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CommentStore interface {
	SetComment(ctx context.Context, comment *Comment) error
	LookupComment(ctx context.Context, postId, ID uuid.UUID) (*Comment, error)
	ListCommentsByPostID(ctx context.Context, postID uuid.UUID, offset, limit int) ([]*Comment, error)
	DeleteComment(ctx context.Context, postID, ID uuid.UUID) error
}
