package writeablecontext

import (
	"context"
	"net/http"
)

// Key is the type for the context key.
type Key string

// ContextKey is the key for the context.
const ContextKey Key = "writeablecontext"

// Store is a map of key/value pairs for the context.
type Store map[string]any

func NewStore() Store {
	return make(Store)
}

// Set sets the value for a key in the context.
func (s Store) Set(key string, value any) {
	if s == nil {
		s = NewStore()
	}

	s[key] = value
}

// Get returns the value for a key in the context.
func (s Store) Get(key string) (any, bool) {
	val, ok := s[key]

	return val, ok
}

// Middleware is a middleware that adds a writable context to the request.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newR := r.WithContext(context.WithValue(r.Context(), ContextKey, NewStore()))
		next.ServeHTTP(w, newR)
	})
}

// FromContext returns the writable context from the request.
func FromContext(ctx context.Context) Store {
	currStore, ok := ctx.Value(ContextKey).(Store)
	if !ok {
		return nil
	}
	return currStore
}
