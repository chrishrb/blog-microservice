package inmemory

import (
	"context"

	"github.com/chrishrb/blog-microservice/post-service/store"
	"github.com/google/uuid"
)

func (s *Store) SetPost(ctx context.Context, post *store.Post) error {
	s.Lock()
	defer s.Unlock()

	// Set timestamps
	now := s.clock.Now()
	post.CreatedAt = now
	post.UpdatedAt = now

	// Store the post
	s.posts[post.ID] = post
	return nil
}

func (s *Store) LookupPost(ctx context.Context, ID uuid.UUID) (*store.Post, error) {
	s.Lock()
	defer s.Unlock()

	post, ok := s.posts[ID]
	if !ok {
		return nil, nil
	}
	return post, nil
}

func (s *Store) ListPosts(ctx context.Context, offset, limit int) ([]*store.Post, error) {
	s.Lock()
	defer s.Unlock()

	var posts []*store.Post
	for _, post := range s.posts {
		posts = append(posts, post)
	}

	end := min(offset+limit, len(posts))
	return posts[offset:end], nil
}

func (s *Store) DeletePost(ctx context.Context, ID uuid.UUID) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.posts[ID]; !ok {
		return nil
	}

	delete(s.posts, ID)
	return nil
}
