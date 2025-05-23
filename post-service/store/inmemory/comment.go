package inmemory

import (
	"context"

	"github.com/chrishrb/blog-microservice/post-service/store"
	"github.com/google/uuid"
)

func (s *Store) SetComment(ctx context.Context, comment *store.Comment) error {
	s.Lock()
	defer s.Unlock()

	// Verify the post exists
	if _, ok := s.posts[comment.PostID]; !ok {
		return nil
	}

	// Set timestamps
	now := s.clock.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now

	// Store the comment
	if _, ok := s.comments[comment.PostID]; !ok {
		s.comments[comment.PostID] = make(map[uuid.UUID]*store.Comment)
	}
	s.comments[comment.PostID][comment.ID] = comment
	return nil
}

func (s *Store) LookupComment(ctx context.Context, postID, ID uuid.UUID) (*store.Comment, error) {
	s.Lock()
	defer s.Unlock()

	comment, ok := s.comments[postID][ID]
	if !ok {
		return nil, nil
	}
	return comment, nil
}

func (s *Store) ListCommentsByPostID(ctx context.Context, postID uuid.UUID, offset, limit int) ([]*store.Comment, error) {
	s.Lock()
	defer s.Unlock()

	// Check if post exists
	if _, ok := s.posts[postID]; !ok {
		return nil, nil
	}

	var comments []*store.Comment
	for _, comment := range s.comments[postID] {
		comments = append(comments, comment)
	}

	if offset >= len(comments) {
		return []*store.Comment{}, nil
	}

	end := min(offset+limit, len(comments))
	return comments[offset:end], nil
}

func (s *Store) DeleteComment(ctx context.Context, postID, ID uuid.UUID) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.comments[postID][ID]; !ok {
		return nil
	}

	delete(s.comments[postID], ID)
	return nil
}
