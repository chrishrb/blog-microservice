package inmemory

import (
	"sync"

	"k8s.io/utils/clock"

	"github.com/chrishrb/blog-microservice/post-service/store"
	"github.com/google/uuid"
)

// Store is an in-memory implementation of the store.Engine interface. As everything
// is stored in memory it is not stateless and cannot be used if running >1 instances.
// It is primarily provided to support unit testing.
type Store struct {
	sync.Mutex
	clock    clock.PassiveClock
	posts    map[uuid.UUID]*store.Post
	comments map[uuid.UUID]map[uuid.UUID]*store.Comment
}

func NewStore(clock clock.PassiveClock) *Store {
	return &Store{
		clock:    clock,
		posts:    make(map[uuid.UUID]*store.Post),
		comments: make(map[uuid.UUID]map[uuid.UUID]*store.Comment),
	}
}
