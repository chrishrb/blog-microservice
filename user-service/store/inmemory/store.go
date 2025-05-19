package inmemory

import (
	"sync"

	"k8s.io/utils/clock"

	"github.com/chrishrb/blog-microservice/user-service/store"
)

// Store is an in-memory implementation of the store.Engine interface. As everything
// is stored in memory it is not stateless and cannot be used if running >1 instances.
// It is primarily provided to support unit testing.
type Store struct {
	sync.Mutex
	clock clock.PassiveClock
	users map[string]*store.User
}

func NewStore(clock clock.PassiveClock) *Store {
	return &Store{
		clock: clock,
		users: make(map[string]*store.User),
	}
}
