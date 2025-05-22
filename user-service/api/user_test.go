package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chrishrb/blog-microservice/internal/testutil"
	"github.com/chrishrb/blog-microservice/internal/transport"
	"github.com/chrishrb/blog-microservice/user-service/api"
	"github.com/chrishrb/blog-microservice/user-service/service"
	"github.com/chrishrb/blog-microservice/user-service/store"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	server, r, engine, _, _, producer := setupServer(t)
	defer server.Close()

	d := api.UserCreate{
		Email:     openapi_types.Email("test@example.com"),
		FirstName: "John",
		LastName:  "Doe",
		Password:  "password123",
	}
	jsonData, err := json.Marshal(d)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/users",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Result().StatusCode)
	var res api.User
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	// Check the response
	assert.NotEmpty(t, res.Id)
	assert.Equal(t, d.Email, res.Email)
	assert.Equal(t, "John", res.FirstName)
	assert.Equal(t, "Doe", res.LastName)
	assert.Equal(t, api.UserRoleUser, res.Role)
	assert.Equal(t, api.UserStatusPending, res.Status)

	// Check the database
	dbUser, err := engine.LookupUser(req.Context(), res.Id)
	require.NoError(t, err)
	assert.Equal(t, res.Id, dbUser.ID)
	assert.Equal(t, "test@example.com", dbUser.Email)
	assert.Equal(t, "John", dbUser.FirstName)
	assert.Equal(t, "Doe", dbUser.LastName)
	assert.Equal(t, store.RoleUser, dbUser.Role)
	assert.Equal(t, store.StatusPending, dbUser.Status)

	// Verify event was produced
	require.NotNil(t, producer.ProducedMessages)
	assert.Len(t, producer.ProducedMessages, 1)
	assert.NotEmpty(t, producer.ProducedMessages[0].Message.ID)
	assert.Equal(t, transport.VerifyAccountTopic, producer.ProducedMessages[0].Topic)
	assert.NotEmpty(t, producer.ProducedMessages[0].Message.Data)

	// Verify message content
	var resetEvent transport.VerifyAccountEvent
	err = json.Unmarshal(producer.ProducedMessages[0].Message.Data, &resetEvent)
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", resetEvent.Recipient)
	assert.Equal(t, "email", resetEvent.Channel)
	assert.NotEmpty(t, resetEvent.Token)
	assert.Equal(t, "John", resetEvent.FirstName)
	assert.Equal(t, "Doe", resetEvent.LastName)
}

func TestListUsers(t *testing.T) {
	server, r, engine, _, _, _ := setupServer(t)
	defer server.Close()

	userID1 := uuid.New()
	err := engine.SetUser(t.Context(), &store.User{
		ID:           userID1,
		Email:        "john@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: "hashedpassword",
		Status:       store.StatusActive,
		Role:         store.RoleAdmin,
	})
	require.NoError(t, err)

	userID2 := uuid.New()
	err = engine.SetUser(t.Context(), &store.User{
		ID:           userID2,
		Email:        "jane@example.com",
		FirstName:    "Jane",
		LastName:     "Doe",
		PasswordHash: "hashedpassword",
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	// Setup mock behavior
	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	var resList []api.User
	err = json.NewDecoder(rr.Body).Decode(&resList)

	require.NoError(t, err)
	require.Equal(t, 2, len(resList))

	// Check that our created users exist in the response
	expected := map[uuid.UUID]struct {
		Email openapi_types.Email
		Role  api.UserRole
	}{
		userID1: {
			Email: openapi_types.Email("john@example.com"),
			Role:  api.UserRoleAdmin,
		},
		userID2: {
			Email: openapi_types.Email("jane@example.com"),
			Role:  api.UserRoleUser,
		},
	}

	require.Equal(t, len(expected), len(resList))

	for _, user := range resList {
		exp, ok := expected[user.Id]
		require.True(t, ok, "unexpected user ID: %s", user.Id)
		assert.Equal(t, exp.Email, user.Email)
		assert.Equal(t, exp.Role, user.Role)
	}
}

func TestDeleteUser(t *testing.T) {
	server, r, engine, _, _, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	err := engine.SetUser(t.Context(), &store.User{
		ID:           userID,
		Email:        "john@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: "hashedpassword",
		Status:       store.StatusActive,
		Role:         store.RoleAdmin,
	})
	require.NoError(t, err)

	// Delete the user
	req := httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("/users/%s", userID),
		nil,
	)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Result().StatusCode)
}

func TestLookupUser(t *testing.T) {
	server, r, engine, _, _, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	err := engine.SetUser(t.Context(), &store.User{
		ID:           userID,
		Email:        "john@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: "hashedpassword",
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	// Lookup the user
	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/users/%s", userID),
		nil,
	)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	var res api.User
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	// Check the response
	assert.Equal(t, userID, res.Id)
	assert.Equal(t, openapi_types.Email("john@example.com"), res.Email)
	assert.Equal(t, "John", res.FirstName)
	assert.Equal(t, "Doe", res.LastName)
	assert.Equal(t, api.UserRoleUser, res.Role)
	assert.Equal(t, api.UserStatusActive, res.Status)
}

func TestLookupUser_NotFound(t *testing.T) {
	server, r, _, _, _, _ := setupServer(t)
	defer server.Close()

	// Lookup a non-existent user
	userID := uuid.New()
	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/users/%s", userID),
		nil,
	)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Result().StatusCode)
}

func TestUpdateUser(t *testing.T) {
	server, r, engine, _, _, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	err := engine.SetUser(t.Context(), &store.User{
		ID:           userID,
		Email:        "john@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: "hashedpassword",
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	newEmail := openapi_types.Email("updated@example.com")

	// Update the user
	d := api.UserUpdate{
		Email:     &newEmail,
		FirstName: testutil.Ptr("Updated"),
		Role:      testutil.Ptr(api.UserUpdateRoleAdmin),
	}
	jsonData, err := json.Marshal(d)
	require.NoError(t, err)
	req := httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/users/%s", userID),
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	var res api.User
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	// Check the response
	assert.Equal(t, userID, res.Id)
	assert.Equal(t, newEmail, res.Email)
	assert.Equal(t, "Updated", res.FirstName)
	assert.Equal(t, "Doe", res.LastName)
	assert.Equal(t, api.UserRoleAdmin, res.Role)

	// Check the database
	dbUser, err := engine.LookupUser(req.Context(), res.Id)
	require.NoError(t, err)
	assert.Equal(t, res.Id, dbUser.ID)
	assert.Equal(t, "updated@example.com", dbUser.Email)
	assert.Equal(t, "Updated", dbUser.FirstName)
	assert.Equal(t, "Doe", dbUser.LastName)
	assert.Equal(t, store.RoleAdmin, dbUser.Role)
	assert.Equal(t, store.StatusActive, dbUser.Status)
}

func TestUpdateUser_NotFound(t *testing.T) {
	server, r, _, _, _, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()

	d := api.UserUpdate{
		FirstName: testutil.Ptr("Updated"),
	}
	jsonData, err := json.Marshal(d)
	require.NoError(t, err)

	// Update a non-existent user
	req := httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/users/%s", userID),
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Result().StatusCode)
}

func TestGetCurrentUser(t *testing.T) {
	server, r, engine, _, _, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	err := engine.SetUser(t.Context(), &store.User{
		ID:           userID,
		Email:        "john@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: "hashedpassword",
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	// Get current user
	req := httptest.NewRequest(
		http.MethodGet,
		"/users/me",
		nil,
	)
	req = userIDContext(req, userID)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	var res api.User
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	// Check the response
	assert.Equal(t, userID, res.Id)
	assert.Equal(t, openapi_types.Email("john@example.com"), res.Email)
	assert.Equal(t, "John", res.FirstName)
	assert.Equal(t, "Doe", res.LastName)
	assert.Equal(t, api.UserRoleUser, res.Role)
	assert.Equal(t, api.UserStatusActive, res.Status)
}

func TestUpdateCurrentUser(t *testing.T) {
	server, r, engine, _, _, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	passwordHash, err := service.HashPassword("currentPassword")
	require.NoError(t, err)

	// Create a user first
	err = engine.SetUser(t.Context(), &store.User{
		ID:           userID,
		Email:        "john@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: passwordHash,
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	// Update current user
	newEmail := openapi_types.Email("updated@example.com")
	newPassword := "newPassword"

	d := api.UserUpdateCurrent{
		Email:           &newEmail,
		Password:        &newPassword,
		CurrentPassword: "currentPassword",
	}
	jsonData, err := json.Marshal(d)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPut,
		"/users/me",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	req = userIDContext(req, userID)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	var res api.User
	err = json.NewDecoder(rr.Body).Decode(&res)
	require.NoError(t, err)

	// Check the response
	assert.Equal(t, userID, res.Id)
	assert.Equal(t, newEmail, res.Email)

	// Check the database
	dbUser, err := engine.LookupUser(req.Context(), res.Id)
	require.NoError(t, err)
	assert.Equal(t, res.Id, dbUser.ID)
	assert.Equal(t, "updated@example.com", dbUser.Email)
	assert.Equal(t, "John", dbUser.FirstName)
	assert.Equal(t, "Doe", dbUser.LastName)
	assert.True(t, service.VerifyPassword("newPassword", dbUser.PasswordHash))
	assert.Equal(t, store.RoleUser, dbUser.Role)
	assert.Equal(t, store.StatusActive, dbUser.Status)
}

func TestUpdateCurrentUser_IncorrectPassword(t *testing.T) {
	server, r, engine, _, _, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	passwordHash, err := service.HashPassword("correctPassword")
	require.NoError(t, err)

	// Create a user first
	err = engine.SetUser(t.Context(), &store.User{
		ID:           userID,
		Email:        "john@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: passwordHash,
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	// Update current user with incorrect password
	d := api.UserUpdateCurrent{
		CurrentPassword: "wrongPassword",
		FirstName:       testutil.Ptr("Updated"),
	}
	jsonData, err := json.Marshal(d)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPut,
		"/users/me",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	req = userIDContext(req, userID)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
}
