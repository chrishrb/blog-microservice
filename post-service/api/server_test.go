package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chrishrb/blog-microservice/internal/writeablecontext"
	"github.com/chrishrb/blog-microservice/post-service/api"
	"github.com/chrishrb/blog-microservice/post-service/store"
	"github.com/chrishrb/blog-microservice/post-service/store/inmemory"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

func userIDContext(req *http.Request, userID uuid.UUID) *http.Request {
	store := writeablecontext.NewStore()
	store.Set("userID", userID.String())
	return req.WithContext(context.WithValue(req.Context(), writeablecontext.ContextKey, store))
}
