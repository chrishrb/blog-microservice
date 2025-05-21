package inmemory_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/chrishrb/blog-microservice/user-service/store"
	"github.com/chrishrb/blog-microservice/user-service/store/inmemory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clock_testing "k8s.io/utils/clock/testing"
)

func TestSetUser(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	userID := uuid.New()
	user := &store.User{
		ID:           userID,
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		PasswordHash: "hashedPassword",
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	}

	err := engine.SetUser(t.Context(), user)
	require.NoError(t, err)

	savedUser, err := engine.LookupUser(t.Context(), userID)
	assert.NoError(t, err)
	assert.NotNil(t, savedUser)
	assert.Equal(t, userID, savedUser.ID)
	assert.Equal(t, "test@example.com", savedUser.Email)
	assert.Equal(t, "Test", savedUser.FirstName)
	assert.Equal(t, "User", savedUser.LastName)
	assert.Equal(t, "hashedPassword", savedUser.PasswordHash)
	assert.Equal(t, store.StatusActive, savedUser.Status)
	assert.Equal(t, store.RoleUser, savedUser.Role)
	assert.Equal(t, fakeClock.Now(), savedUser.CreatedAt)
	assert.Equal(t, fakeClock.Now(), savedUser.UpdatedAt)
}

func TestLookupUser(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	userID := uuid.New()
	user := &store.User{
		ID:           userID,
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		PasswordHash: "hashedPassword",
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	}

	err := engine.SetUser(t.Context(), user)
	require.NoError(t, err)

	savedUser, err := engine.LookupUser(t.Context(), userID)
	require.NoError(t, err)
	assert.NotNil(t, savedUser)
	assert.Equal(t, userID, savedUser.ID)

	nonExistentID := uuid.New()
	nonExistentUser, err := engine.LookupUser(t.Context(), nonExistentID)
	require.NoError(t, err)
	assert.Nil(t, nonExistentUser)
}

func TestLookupUserByEmail(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	userID := uuid.New()
	user := &store.User{
		ID:           userID,
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		PasswordHash: "hashedPassword",
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	}

	err := engine.SetUser(t.Context(), user)
	require.NoError(t, err)

	savedUser, err := engine.LookupUserByEmail(t.Context(), "test@example.com")
	require.NoError(t, err)
	assert.NotNil(t, savedUser)
	assert.Equal(t, userID, savedUser.ID)
	assert.Equal(t, "test@example.com", savedUser.Email)

	nonExistentUser, err := engine.LookupUserByEmail(t.Context(), "nonexistent@example.com")
	require.NoError(t, err)
	assert.Nil(t, nonExistentUser)
}

func TestListUsers(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	for i := range 5 {
		user := &store.User{
			ID:           uuid.New(),
			Email:        fmt.Sprintf("user%d@example.com", i),
			FirstName:    "Test",
			LastName:     fmt.Sprintf("User %d", i),
			PasswordHash: "hashedPassword",
			Status:       store.StatusActive,
			Role:         store.RoleUser,
		}
		err := engine.SetUser(t.Context(), user)
		require.NoError(t, err)
	}

	users, err := engine.ListUsers(t.Context(), 0, 3)
	require.NoError(t, err)
	assert.Len(t, users, 3)

	users, err = engine.ListUsers(t.Context(), 3, 4)
	require.NoError(t, err)
	assert.Len(t, users, 2)

	users, err = engine.ListUsers(t.Context(), 0, 10)
	require.NoError(t, err)
	assert.Len(t, users, 5)

	users, err = engine.ListUsers(t.Context(), 5, 10)
	require.NoError(t, err)
	assert.Len(t, users, 0)
}

func TestDeleteUser(t *testing.T) {
	fakeClock := clock_testing.NewFakeClock(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))
	engine := inmemory.NewStore(fakeClock)

	userID := uuid.New()
	user := &store.User{
		ID:           userID,
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		PasswordHash: "hashedPassword",
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	}

	err := engine.SetUser(t.Context(), user)
	require.NoError(t, err)

	savedUser, err := engine.LookupUser(t.Context(), userID)
	require.NoError(t, err)
	assert.NotNil(t, savedUser)

	err = engine.DeleteUser(t.Context(), userID)
	require.NoError(t, err)

	deletedUser, err := engine.LookupUser(t.Context(), userID)
	require.NoError(t, err)
	assert.Nil(t, deletedUser)

	nonExistentID := uuid.New()
	err = engine.DeleteUser(t.Context(), nonExistentID)
	assert.NoError(t, err)
}
