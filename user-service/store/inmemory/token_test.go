package inmemory_test

import (
	"testing"
	"time"

	"github.com/chrishrb/blog-microservice/user-service/store"
	"github.com/chrishrb/blog-microservice/user-service/store/inmemory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clock_testing "k8s.io/utils/clock/testing"
)

func TestSetToken(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	userID := uuid.New()
	ttl, err := time.ParseDuration("5m")
	require.NoError(t, err)

	err = engine.SetToken(t.Context(), &store.Token{
		UserID:  userID,
		Token:   "some-refresh-token",
		TTL:     ttl,
		Revoked: false,
	})
	require.NoError(t, err)

	token, err := engine.GetToken(t.Context(), "some-refresh-token")
	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, userID, token.UserID)
	assert.Equal(t, "some-refresh-token", token.Token)
	assert.Equal(t, ttl, token.TTL)
	assert.Equal(t, false, token.Revoked)
	assert.Equal(t, fakeClock.Now(), token.CreatedAt)
	assert.Equal(t, fakeClock.Now(), token.UpdatedAt)
}

func TestIsTokenRevoked(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	userID := uuid.New()
	ttl, err := time.ParseDuration("5m")
	require.NoError(t, err)

	err = engine.SetToken(t.Context(), &store.Token{
		Token:   "some-refresh-token",
		UserID:  userID,
		TTL:     ttl,
		Revoked: false,
	})
	require.NoError(t, err)

	isRevoked, err := engine.IsTokenRevoked(t.Context(), "some-refresh-token")
	require.NoError(t, err)
	assert.False(t, isRevoked)

	err = engine.SetToken(t.Context(), &store.Token{
		Token:   "another-refresh-token",
		UserID:  userID,
		TTL:     ttl,
		Revoked: true,
	})
	require.NoError(t, err)

	isRevoked, err = engine.IsTokenRevoked(t.Context(), "another-refresh-token")
	require.NoError(t, err)
	assert.True(t, isRevoked)

	isRevoked, err = engine.IsTokenRevoked(t.Context(), "non-existent-token")
	require.NoError(t, err)
	assert.False(t, isRevoked)
}

func TestSetTokenRevoked(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	userID := uuid.New()
	ttl, err := time.ParseDuration("5m")
	require.NoError(t, err)

	err = engine.SetToken(t.Context(), &store.Token{
		Token:   "some-refresh-token",
		UserID:  userID,
		TTL:     ttl,
		Revoked: false,
	})
	require.NoError(t, err)

	err = engine.SetToken(t.Context(), &store.Token{
		Token:   "another-refresh-token",
		UserID:  userID,
		TTL:     ttl,
		Revoked: false,
	})
	require.NoError(t, err)

	err = engine.SetTokenRevoked(t.Context(), userID)
	require.NoError(t, err)

	isRevoked, err := engine.IsTokenRevoked(t.Context(), "some-refresh-token")
	require.NoError(t, err)
	assert.True(t, isRevoked)

	isRevoked, err = engine.IsTokenRevoked(t.Context(), "another-refresh-token")
	require.NoError(t, err)
	assert.True(t, isRevoked)
}
