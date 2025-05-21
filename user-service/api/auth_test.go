package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chrishrb/blog-microservice/internal/auth"
	"github.com/chrishrb/blog-microservice/user-service/api"
	"github.com/chrishrb/blog-microservice/user-service/service"
	"github.com/chrishrb/blog-microservice/user-service/store"
	"github.com/chrishrb/blog-microservice/user-service/store/inmemory"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/ecdsafile"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/clock"
	clockTest "k8s.io/utils/clock/testing"
)

func TestLoginUser(t *testing.T) {
	server, r, engine, jwsSigner := setupServerWithJWS(t)
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
	assert.Equal(t, int(jwsSigner.GetAccessTokenExpiresIn().Seconds()), authRes.ExpiresIn)
}

func TestLoginUser_WrongPassword(t *testing.T) {
	server, r, engine, _ := setupServer(t)
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
	server, r, _, _ := setupServer(t)
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
	server, r, engine, _ := setupServer(t)
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
	server, r, engine, _ := setupServerWithJWS(t)
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
	server, r, engine, jwsSigner := setupServerWithJWS(t)
	defer server.Close()

	userID := uuid.New()

	// Create a refresh token and store it
	refreshToken, err := jwsSigner.CreateRefreshJWS(userID)
	require.NoError(t, err)
	err = engine.SetToken(context.Background(), &store.Token{
		UserID:  userID,
		Token:   refreshToken,
		TTL:     jwsSigner.GetRefreshTokenExpiresIn(),
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
	server, r, engine, jwsSigner := setupServerWithJWS(t)
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
	refreshToken, err := jwsSigner.CreateRefreshJWS(userID)
	require.NoError(t, err)
	err = engine.SetToken(context.Background(), &store.Token{
		UserID:  userID,
		Token:   refreshToken,
		TTL:     jwsSigner.GetRefreshTokenExpiresIn(),
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
	assert.Equal(t, int(jwsSigner.GetAccessTokenExpiresIn().Seconds()), authRes.ExpiresIn)
}

func TestRefreshToken_RevokedToken(t *testing.T) {
	server, r, engine, jwsSigner := setupServerWithJWS(t)
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
	refreshToken, err := jwsSigner.CreateRefreshJWS(userID)
	require.NoError(t, err)

	// Store the token as revoked
	err = engine.SetToken(context.Background(), &store.Token{
		UserID:  userID,
		Token:   refreshToken,
		TTL:     jwsSigner.GetRefreshTokenExpiresIn(),
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
	server, r, _, _ := setupServerWithJWS(t)
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
	server, r, engine, jwsSigner := setupServerWithJWS(t)
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
	refreshToken, err := jwsSigner.CreateRefreshJWS(userID)
	require.NoError(t, err)

	// Store the token
	err = engine.SetToken(context.Background(), &store.Token{
		UserID:  userID,
		Token:   refreshToken,
		TTL:     jwsSigner.GetRefreshTokenExpiresIn(),
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

type MockJWSSigner struct{}

const PrivateKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIN2dALnjdcZaIZg4QuA6Dw+kxiSW502kJfmBN3priIhPoAoGCCqGSM49
AwEHoUQDQgAE4pPyvrB9ghqkT1Llk0A42lixkugFd/TBdOp6wf69O9Nndnp4+HcR
s9SlG/8hjB2Hz42v4p3haKWv3uS1C6ahCQ==
-----END EC PRIVATE KEY-----`

func (s *MockJWSSigner) CreateJWS(userID uuid.UUID, claims []string) (string, error) {
	privKey, err := ecdsafile.LoadEcdsaPrivateKey([]byte(PrivateKey))
	if err != nil {
		return "", fmt.Errorf("loading PEM private key: %w", err)
	}
	token, err := jwt.
		NewBuilder().
		Subject(userID.String()).
		Claim("permissions", claims).
		Build()
	if err != nil {
		return "", err
	}
	t, err := jwt.Sign(token, jwa.ES256, privKey)
	if err != nil {
		return "", err
	}
	return string(t), nil
}

func (s *MockJWSSigner) CreateRefreshJWS(userID uuid.UUID) (string, error) {
	privKey, err := ecdsafile.LoadEcdsaPrivateKey([]byte(PrivateKey))
	if err != nil {
		return "", fmt.Errorf("loading PEM private key: %w", err)
	}
	token, err := jwt.
		NewBuilder().
		Subject(userID.String()).
		Claim("permissions", []string{}).
		Build()
	if err != nil {
		return "", err
	}
	t, err := jwt.Sign(token, jwa.ES256, privKey)
	if err != nil {
		return "", err
	}
	return string(t), nil
}

func (s *MockJWSSigner) GetAccessTokenExpiresIn() time.Duration {
	d, _ := time.ParseDuration("10m")
	return d
}

func (s *MockJWSSigner) GetRefreshTokenExpiresIn() time.Duration {
	d, _ := time.ParseDuration("1h")
	return d
}

type MockJWSVerifier struct{}

func (v *MockJWSVerifier) ValidateJWS(jws string) (jwt.Token, error) {
	return jwt.Parse([]byte(jws))
}

func setupServerWithJWS(t *testing.T) (*httptest.Server, *chi.Mux, store.Engine, auth.JWSSigner) {
	engine := inmemory.NewStore(clock.RealClock{})

	jwsSigner := &MockJWSSigner{}
	jwsVerifier := &MockJWSVerifier{}

	now := time.Now().UTC()
	c := clockTest.NewFakePassiveClock(now)
	srv, err := api.NewServer(engine, c, jwsVerifier, jwsSigner)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Mount("/", api.Handler(srv))
	server := httptest.NewServer(r)

	return server, r, engine, jwsSigner
}
