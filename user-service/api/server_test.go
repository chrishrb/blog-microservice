package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chrishrb/blog-microservice/internal/auth"
	"github.com/chrishrb/blog-microservice/internal/source"
	"github.com/chrishrb/blog-microservice/internal/transport"
	"github.com/chrishrb/blog-microservice/internal/writeablecontext"
	"github.com/chrishrb/blog-microservice/user-service/api"
	"github.com/chrishrb/blog-microservice/user-service/store"
	"github.com/chrishrb/blog-microservice/user-service/store/inmemory"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/clock"
	clockTest "k8s.io/utils/clock/testing"
)

const PrivateKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIN2dALnjdcZaIZg4QuA6Dw+kxiSW502kJfmBN3priIhPoAoGCCqGSM49
AwEHoUQDQgAE4pPyvrB9ghqkT1Llk0A42lixkugFd/TBdOp6wf69O9Nndnp4+HcR
s9SlG/8hjB2Hz42v4p3haKWv3uS1C6ahCQ==
-----END EC PRIVATE KEY-----`

const PublicKey = `
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE4pPyvrB9ghqkT1Llk0A42lixkugF
d/TBdOp6wf69O9Nndnp4+HcRs9SlG/8hjB2Hz42v4p3haKWv3uS1C6ahCQ==
-----END PUBLIC KEY-----`

type MockProducer struct {
	ProducedMessages []ProducedMessage
}

type ProducedMessage struct {
	Topic   string
	Message *transport.Message
}

func (p *MockProducer) Produce(ctx context.Context, topic string, message *transport.Message) error {
	p.ProducedMessages = append(p.ProducedMessages, ProducedMessage{
		Topic:   topic,
		Message: message,
	})
	return nil
}

func setupServer(t *testing.T) (*httptest.Server, *chi.Mux, store.Engine, clock.PassiveClock, auth.JWSSigner, *MockProducer) {
	engine := inmemory.NewStore(clock.RealClock{})

	issuer, audience := "example.com", "example.com"
	jwsSigner, err := auth.NewLocalJWSSigner(
		source.StringSourceProvider{Data: PrivateKey},
		issuer,
		audience,
		time.Duration(5*time.Minute),
		time.Duration(24*time.Hour),
	)
	require.NoError(t, err)

	jwsVerifier, err := auth.NewLocalJWSVerifier(
		source.StringSourceProvider{Data: PublicKey},
		issuer,
		audience,
	)
	require.NoError(t, err)

	mockProducer := &MockProducer{}

	now := time.Now().UTC()
	c := clockTest.NewFakePassiveClock(now)
	srv, err := api.NewServer(engine, c, jwsVerifier, jwsSigner, mockProducer)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Mount("/", api.Handler(srv))
	server := httptest.NewServer(r)

	return server, r, engine, c, jwsSigner, mockProducer
}

func userIDContext(req *http.Request, userID uuid.UUID) *http.Request {
	store := writeablecontext.NewStore()
	store.Set("userID", userID.String())
	return req.WithContext(context.WithValue(req.Context(), writeablecontext.ContextKey, store))
}
