package config_test

import (
	"testing"

	"github.com/chrishrb/blog-microservice/notification-service/config"
	clone "github.com/huandu/go-clone/generic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigure(t *testing.T) {
	cfg := clone.Clone(&config.DefaultConfig)

	settings, err := config.Configure(t.Context(), cfg)
	require.NoError(t, err)

	assert.NotNil(t, settings.Tracer)
	assert.NotNil(t, settings.TracerProvider)
	assert.NotNil(t, settings.MsgProducer)
	assert.NotNil(t, settings.MsgConsumer)
	assert.NotNil(t, settings.PasswordResetHandler)
}
