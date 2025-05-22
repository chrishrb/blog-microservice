package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chrishrb/blog-microservice/internal/transport"
	"github.com/chrishrb/blog-microservice/user-service/api"
	"github.com/chrishrb/blog-microservice/user-service/service"
	"github.com/chrishrb/blog-microservice/user-service/store"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginUser(t *testing.T) {
	server, r, engine, _, _, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	passwordHash, err := service.HashPassword("password123")
	require.NoError(t, err)

	// Create a user
	err = engine.SetUser(context.Background(), &store.User{
		ID:           userID,
		Email:        "test@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: passwordHash,
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	// Test login with valid credentials
	loginReq := api.LoginRequest{
		Email:    openapi_types.Email("test@example.com"),
		Password: "password123",
	}
	jsonData, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify success response
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	var authRes api.AuthResponse
	err = json.NewDecoder(rr.Body).Decode(&authRes)
	require.NoError(t, err)

	// Check that tokens are not empty
	assert.NotEmpty(t, authRes.AccessToken)
	assert.NotEmpty(t, authRes.RefreshToken)
	assert.Equal(t, int(time.Duration(5*time.Minute).Seconds()), authRes.ExpiresIn)
}

func TestLoginUser_WrongPassword(t *testing.T) {
	server, r, engine, _, _, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	passwordHash, err := service.HashPassword("password123")
	require.NoError(t, err)

	// Create a user
	err = engine.SetUser(context.Background(), &store.User{
		ID:           userID,
		Email:        "test@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: passwordHash,
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	// Test login with wrong password
	loginReq := api.LoginRequest{
		Email:    openapi_types.Email("test@example.com"),
		Password: "wrongpassword",
	}
	jsonData, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify unauthorized response
	assert.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)
}

func TestLoginUser_UserNotFound(t *testing.T) {
	server, r, _, _, _, _ := setupServer(t)
	defer server.Close()

	// Test login with non-existent user
	loginReq := api.LoginRequest{
		Email:    openapi_types.Email("nonexistent@example.com"),
		Password: "password123",
	}
	jsonData, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify unauthorized response
	assert.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)
}

func TestLoginUser_InactiveUser(t *testing.T) {
	server, r, engine, _, _, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	passwordHash, err := service.HashPassword("password123")
	require.NoError(t, err)

	// Create an inactive user
	err = engine.SetUser(context.Background(), &store.User{
		ID:           userID,
		Email:        "inactive@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: passwordHash,
		Status:       store.StatusBanned,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	// Test login with inactive user
	loginReq := api.LoginRequest{
		Email:    openapi_types.Email("inactive@example.com"),
		Password: "password123",
	}
	jsonData, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify unauthorized response
	assert.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)
}

func TestLoginUser_Admin(t *testing.T) {
	server, r, engine, _, _, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()
	passwordHash, err := service.HashPassword("adminpass")
	require.NoError(t, err)

	// Create an admin user
	err = engine.SetUser(context.Background(), &store.User{
		ID:           userID,
		Email:        "admin@example.com",
		FirstName:    "Admin",
		LastName:     "User",
		PasswordHash: passwordHash,
		Status:       store.StatusActive,
		Role:         store.RoleAdmin,
	})
	require.NoError(t, err)

	// Test login with admin credentials
	loginReq := api.LoginRequest{
		Email:    openapi_types.Email("admin@example.com"),
		Password: "adminpass",
	}
	jsonData, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify success response
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)

	// We'd need to verify that the JWT contains the admin claims, but that would
	// require decoding the JWT which might be outside the scope of this test
}

func TestLogoutUser(t *testing.T) {
	server, r, engine, _, jwsSigner, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()

	// Create a refresh token and store it
	refreshToken, refreshTokenExpiresIn, err := jwsSigner.CreateRefreshToken(userID)
	require.NoError(t, err)
	err = engine.SetToken(context.Background(), &store.Token{
		UserID:  userID,
		Token:   refreshToken,
		TTL:     refreshTokenExpiresIn,
		Revoked: false,
	})
	require.NoError(t, err)

	// Make logout request
	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/logout",
		nil,
	)
	req = userIDContext(req, userID)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify success response
	assert.Equal(t, http.StatusNoContent, rr.Result().StatusCode)

	// Check that refresh token was revoked
	tokens, err := engine.ListTokens(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
	assert.True(t, tokens[0].Revoked)
}

func TestRefreshToken(t *testing.T) {
	server, r, engine, _, jwsSigner, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()

	// Create a user
	err := engine.SetUser(context.Background(), &store.User{
		ID:           userID,
		Email:        "test@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: "hashedpassword",
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	// Create a refresh token and store it
	refreshToken, refreshTokenExpiresIn, err := jwsSigner.CreateRefreshToken(userID)
	require.NoError(t, err)
	err = engine.SetToken(context.Background(), &store.Token{
		UserID:  userID,
		Token:   refreshToken,
		TTL:     refreshTokenExpiresIn,
		Revoked: false,
	})
	require.NoError(t, err)

	// Make refresh token request
	refreshReq := api.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}
	jsonData, err := json.Marshal(refreshReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/refresh",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify success response
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	var authRes api.AuthResponse
	err = json.NewDecoder(rr.Body).Decode(&authRes)
	require.NoError(t, err)

	// Check that new tokens are not empty
	assert.NotEmpty(t, authRes.AccessToken)
	assert.NotEmpty(t, authRes.RefreshToken)
	assert.Equal(t, int(time.Duration(5*time.Minute).Seconds()), authRes.ExpiresIn)
}

func TestRefreshToken_RevokedToken(t *testing.T) {
	server, r, engine, _, jwsSigner, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()

	// Create a user
	err := engine.SetUser(context.Background(), &store.User{
		ID:           userID,
		Email:        "test@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: "hashedpassword",
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	// Create a refresh token
	refreshToken, refreshTokenExpiresIn, err := jwsSigner.CreateRefreshToken(userID)
	require.NoError(t, err)

	// Store the token as revoked
	err = engine.SetToken(context.Background(), &store.Token{
		UserID:  userID,
		Token:   refreshToken,
		TTL:     refreshTokenExpiresIn,
		Revoked: true,
	})
	require.NoError(t, err)

	// Make refresh token request with revoked token
	refreshReq := api.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}
	jsonData, err := json.Marshal(refreshReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/refresh",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify unauthorized response
	assert.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	server, r, _, _, _, _ := setupServer(t)
	defer server.Close()

	// Make refresh token request with invalid token
	refreshReq := api.RefreshTokenRequest{
		RefreshToken: "invalid-token",
	}
	jsonData, err := json.Marshal(refreshReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/refresh",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify unauthorized response
	assert.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)
}

func TestRefreshToken_InactiveUser(t *testing.T) {
	server, r, engine, _, jwsSigner, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()

	// Create an inactive user
	err := engine.SetUser(context.Background(), &store.User{
		ID:           userID,
		Email:        "inactive@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: "hashedpassword",
		Status:       store.StatusBanned,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	// Create a refresh token
	refreshToken, refreshTokenExpiresIn, err := jwsSigner.CreateRefreshToken(userID)
	require.NoError(t, err)

	// Store the token
	err = engine.SetToken(context.Background(), &store.Token{
		UserID:  userID,
		Token:   refreshToken,
		TTL:     refreshTokenExpiresIn,
		Revoked: false,
	})
	require.NoError(t, err)

	// Make refresh token request
	refreshReq := api.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}
	jsonData, err := json.Marshal(refreshReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/refresh",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify unauthorized response for inactive user
	assert.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)
}

func TestRequestPasswordReset(t *testing.T) {
	server, r, engine, _, _, producer := setupServer(t)
	defer server.Close()

	userID := uuid.New()

	// Create a user
	err := engine.SetUser(context.Background(), &store.User{
		ID:           userID,
		Email:        "test@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: "hashedpassword",
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	// Test password reset request
	resetReq := api.PasswordResetRequest{
		Email: openapi_types.Email("test@example.com"),
	}
	jsonData, err := json.Marshal(resetReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/password-reset",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify success response
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)

	// Verify message was produced
	require.NotNil(t, producer.ProducedMessages)
	assert.Len(t, producer.ProducedMessages, 1)
	assert.NotEmpty(t, producer.ProducedMessages[0].Message.ID)
	assert.Equal(t, transport.PasswordResetTopic, producer.ProducedMessages[0].Topic)
	assert.NotEmpty(t, producer.ProducedMessages[0].Message.Data)

	// Verify message content
	var resetEvent transport.PasswordResetEvent
	err = json.Unmarshal(producer.ProducedMessages[0].Message.Data, &resetEvent)
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", resetEvent.Recipient)
	assert.Equal(t, "email", resetEvent.Channel)
	assert.NotEmpty(t, resetEvent.Token)
	assert.Equal(t, "John", resetEvent.FirstName)
	assert.Equal(t, "Doe", resetEvent.LastName)
}

func TestRequestPasswordReset_UserNotFound(t *testing.T) {
	server, r, _, _, _, _ := setupServer(t)
	defer server.Close()

	// Test with non-existent user
	resetReq := api.PasswordResetRequest{
		Email: openapi_types.Email("nonexistent@example.com"),
	}
	jsonData, err := json.Marshal(resetReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/password-reset",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify bad request response
	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestResetPassword(t *testing.T) {
	server, r, engine, _, jwsSigner, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()

	// Create a user
	err := engine.SetUser(context.Background(), &store.User{
		ID:           userID,
		Email:        "test@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: "oldhash",
		Status:       store.StatusActive,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	// Create a password reset token
	resetToken, _, err := jwsSigner.CreatePasswordResetToken(userID)
	require.NoError(t, err)

	// Test password reset
	resetReq := api.PasswordResetConfirmation{
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}
	jsonData, err := json.Marshal(resetReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/password-reset/"+resetToken,
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify success response
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)

	// Verify password was updated
	user, err := engine.LookupUser(context.Background(), userID)
	require.NoError(t, err)
	assert.NotEqual(t, "oldhash", user.PasswordHash)
	assert.True(t, service.VerifyPassword("newpassword123", user.PasswordHash))
}

func TestResetPassword_PasswordMismatch(t *testing.T) {
	server, r, _, _, jwsSigner, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()

	// Create a password reset token
	resetToken, _, err := jwsSigner.CreatePasswordResetToken(userID)
	require.NoError(t, err)

	// Test password reset with mismatched passwords
	resetReq := api.PasswordResetConfirmation{
		NewPassword:     "newpassword123",
		ConfirmPassword: "differentpassword",
	}
	jsonData, err := json.Marshal(resetReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/password-reset/"+resetToken,
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify bad request response
	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestResetPassword_InvalidToken(t *testing.T) {
	server, r, _, _, _, _ := setupServer(t)
	defer server.Close()

	// Test password reset with invalid token
	resetReq := api.PasswordResetConfirmation{
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}
	jsonData, err := json.Marshal(resetReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/password-reset/invalid-token",
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify bad request response
	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestResetPassword_InactiveUser(t *testing.T) {
	server, r, engine, _, jwsSigner, _ := setupServer(t)
	defer server.Close()

	userID := uuid.New()

	// Create an inactive user
	err := engine.SetUser(context.Background(), &store.User{
		ID:           userID,
		Email:        "inactive@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: "oldhash",
		Status:       store.StatusBanned,
		Role:         store.RoleUser,
	})
	require.NoError(t, err)

	// Create a password reset token
	resetToken, _, err := jwsSigner.CreatePasswordResetToken(userID)
	require.NoError(t, err)

	// Test password reset for inactive user
	resetReq := api.PasswordResetConfirmation{
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}
	jsonData, err := json.Marshal(resetReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/password-reset/"+resetToken,
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set("content-type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Verify bad request response
	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
}
