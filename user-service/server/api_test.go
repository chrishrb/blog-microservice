package server_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chrishrb/blog-microservice/user-service/config"
	"github.com/chrishrb/blog-microservice/user-service/server"
	"github.com/chrishrb/blog-microservice/user-service/store/inmemory"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/clock"
)

func TestHealthHandler(t *testing.T) {
	handler := server.NewApiHandler(config.ApiSettings{}, inmemory.NewStore(clock.RealClock{}), nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()
	defer func() {
		err := res.Body.Close()
		if err != nil {
			t.Errorf("closing body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status code: want %d, got %d", http.StatusOK, res.StatusCode)
	}
}

func TestMetricsHandler(t *testing.T) {
	handler := server.NewApiHandler(config.ApiSettings{}, inmemory.NewStore(clock.RealClock{}), nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()
	defer func() {
		err := res.Body.Close()
		if err != nil {
			t.Errorf("closing body: %v", err)
		}
	}()

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	if res.StatusCode != http.StatusOK {
		t.Errorf("status code: want %d, got %d", http.StatusOK, res.StatusCode)
	}

	assert.Contains(t, string(b), "go_goroutines", "metrics should contain go_goroutines")
}

func TestSwaggerHandler(t *testing.T) {
	handler := server.NewApiHandler(config.ApiSettings{}, inmemory.NewStore(clock.RealClock{}), nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/user-service/openapi.json", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()
	defer func() {
		err := res.Body.Close()
		if err != nil {
			t.Errorf("closing body: %v", err)
		}
	}()

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	if res.StatusCode != http.StatusOK {
		t.Errorf("status code: want %d, got %d", http.StatusOK, res.StatusCode)
	}

	var jsonData map[string]any
	err = json.Unmarshal(b, &jsonData)
	require.NoError(t, err)
	require.Equal(t, jsonData["info"].(map[string]any)["title"], "User Service API")
}

type mockJWSVerifier struct{}

func (m *mockJWSVerifier) ValidateToken(jws string) (jwt.Token, error) {
	return nil, errors.New("unauthorized")
}

func (m *mockJWSVerifier) ValidatePasswordResetToken(jws string) (jwt.Token, error) {
	return nil, errors.New("unauthorized")
}

func TestAuthMiddleware(t *testing.T) {
	jwsVerifier := &mockJWSVerifier{}
	handler := server.NewApiHandler(config.ApiSettings{}, inmemory.NewStore(clock.RealClock{}), jwsVerifier, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/user-service/v1/users", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()
	defer func() {
		err := res.Body.Close()
		if err != nil {
			t.Errorf("closing body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("status code: want %d, got %d", http.StatusOK, res.StatusCode)
	}
}
