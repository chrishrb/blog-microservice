package config_test

import (
	"testing"

	"github.com/chrishrb/blog-microservice/user-service/config"
	clone "github.com/huandu/go-clone/generic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigure(t *testing.T) {
	cfg := clone.Clone(&config.DefaultConfig)

	settings, err := config.Configure(t.Context(), cfg)
	require.NoError(t, err)

	wantApiSettings := config.ApiSettings{
		Addr:    "localhost:9410",
		Host:    "localhost",
		OrgName: "chrishrb",
	}

	assert.Equal(t, wantApiSettings, settings.Api)
	assert.NotNil(t, settings.Tracer)
	assert.NotNil(t, settings.TracerProvider)
	assert.NotNil(t, settings.Storage)
	assert.NotNil(t, settings.MsgProducer)
	assert.NotNil(t, settings.JWSVerifier)
	assert.NotNil(t, settings.JWSSigner)
}

func TestConfigureInMemoryStorage(t *testing.T) {
	cfg := clone.Clone(&config.DefaultConfig)
	cfg.Storage.Type = "in_memory"

	settings, err := config.Configure(t.Context(), cfg)
	require.NoError(t, err)
	require.NotNil(t, settings.Storage)
}
