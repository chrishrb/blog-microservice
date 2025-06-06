package inmemory

import (
	"sync"

	"k8s.io/utils/clock"

	"github.com/chrishrb/blog-microservice/user-service/store"
	"github.com/google/uuid"
)

// Store is an in-memory implementation of the store.Engine interface. As everything
// is stored in memory it is not stateless and cannot be used if running >1 instances.
// It is primarily provided to support unit testing.
type Store struct {
	sync.Mutex
	clock  clock.PassiveClock
	users  map[uuid.UUID]*store.User
	tokens map[string]*store.Token
}

func NewStore(clock clock.PassiveClock) *Store {
	return &Store{
		clock:  clock,
		users:  make(map[uuid.UUID]*store.User),
		tokens: make(map[string]*store.Token),
	}
}
