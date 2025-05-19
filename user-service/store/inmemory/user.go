package inmemory

import (
	"context"

	"github.com/chrishrb/blog-microservice/user-service/store"
	"github.com/google/uuid"
)

func (s *Store) SetUser(ctx context.Context, user *store.User) error {
	s.Lock()
	defer s.Unlock()

	// Set timestamps
	now := s.clock.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Store the user
	s.users[user.ID] = user
	return nil
}

func (s *Store) LookupUser(ctx context.Context, ID uuid.UUID) (*store.User, error) {
	s.Lock()
	defer s.Unlock()

	user, ok := s.users[ID]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (s *Store) LookupUserByEmail(ctx context.Context, email string) (*store.User, error) {
	s.Lock()
	defer s.Unlock()

	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, nil
}

func (s *Store) ListUsers(ctx context.Context, offset, limit int) ([]*store.User, error) {
	s.Lock()
	defer s.Unlock()

	var users []*store.User
	for _, user := range s.users {
		users = append(users, user)

		if len(users) >= limit {
			break
		}
	}

	end := min(offset+limit, len(users))
	return users[offset:end], nil
}

func (s *Store) DeleteUser(ctx context.Context, ID uuid.UUID) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.users[ID]; !ok {
		return nil
	}

	delete(s.users, ID)
	return nil
}
