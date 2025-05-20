package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chrishrb/blog-microservice/user-service/config"
	"github.com/chrishrb/blog-microservice/user-service/server"
	"github.com/chrishrb/blog-microservice/user-service/store/inmemory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/clock"
)

func TestHealthHandler(t *testing.T) {
	handler := server.NewApiHandler(config.ApiSettings{}, inmemory.NewStore(clock.RealClock{}), nil, nil)

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
	handler := server.NewApiHandler(config.ApiSettings{}, inmemory.NewStore(clock.RealClock{}), nil, nil)

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
	handler := server.NewApiHandler(config.ApiSettings{}, inmemory.NewStore(clock.RealClock{}), nil, nil)

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
