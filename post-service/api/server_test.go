package api_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chrishrb/blog-microservice/post-service/api"
	"github.com/chrishrb/blog-microservice/post-service/store"
	"github.com/chrishrb/blog-microservice/post-service/store/inmemory"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/clock"
	clockTest "k8s.io/utils/clock/testing"
)

func setupServer(t *testing.T) (*httptest.Server, *chi.Mux, store.Engine, clock.PassiveClock) {
	engine := inmemory.NewStore(clock.RealClock{})

	now := time.Now().UTC()
	c := clockTest.NewFakePassiveClock(now)
	srv, err := api.NewServer(engine, c)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Mount("/", api.Handler(srv))
	server := httptest.NewServer(r)

	return server, r, engine, c
}
