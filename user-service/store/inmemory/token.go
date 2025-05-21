package inmemory

import (
	"context"

	"github.com/chrishrb/blog-microservice/user-service/store"
	"github.com/google/uuid"
)

func (s *Store) SetToken(ctx context.Context, token *store.Token) error {
	s.Lock()
	defer s.Unlock()

	// Set timestamps
	now := s.clock.Now()
	token.CreatedAt = now
	token.UpdatedAt = now

	s.tokens[token.Token] = token
	return nil
}

func (s *Store) GetToken(ctx context.Context, token string) (*store.Token, error) {
	s.Lock()
	defer s.Unlock()

	t, ok := s.tokens[token]
	if !ok {
		return nil, nil
	}

	return t, nil
}

func (s *Store) IsTokenRevoked(ctx context.Context, token string) (bool, error) {
	s.Lock()
	defer s.Unlock()

	t, ok := s.tokens[token]
	if ok && t.Revoked {
		return true, nil
	}

	return false, nil
}

func (s *Store) SetTokenRevoked(ctx context.Context, userID uuid.UUID) error {
	s.Lock()
	defer s.Unlock()


	for k, v := range s.tokens {
		if v.UserID == userID {
			s.tokens[k].UpdatedAt = s.clock.Now()
			s.tokens[k].Revoked = true
		}
	}

	return nil
}
