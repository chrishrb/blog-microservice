package config_test

import (
	"testing"

	"github.com/chrishrb/blog-microservice/post-service/config"
	clone "github.com/huandu/go-clone/generic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	cfg := clone.Clone(&config.DefaultConfig)
	err := cfg.LoadFromFile("testdata/config.yaml")
	require.NoError(t, err)

	want := &config.BaseConfig{
		Api: config.ApiSettingsConfig{
			Addr:    ":9411",
			Host:    "example.com",
			OrgName: "Example",
		},
		Transport: config.TransportConfig{
			Type: "kafka",
			Kafka: &config.KafkaSettingsConfig{
				Urls:           []string{"localhost:9092"},
				Group:          "post-service",
				ConnectTimeout: "10s",
			},
		},
		Observability: config.ObservabilitySettingsConfig{
			LogFormat:         "text",
			OtelCollectorAddr: "localhost:4317",
			TlsKeylogFile:     "/keylog/post-service.log",
		},
		Storage: config.StorageConfig{
			Type: "in_memory",
		},
		Auth: config.AuthConfig{
			Issuer:   "auth.example.com",
			Audience: "blog-microservice",
			PublicKeySource: &config.LocalSourceConfig{
				Type: "file",
				File: "testdata/jwt.pub.pem",
			},
		},
	}

	assert.Equal(t, want, cfg)

	err = cfg.Validate()
	assert.NoError(t, err)
}

func TestValidateConfig(t *testing.T) {
	cfg := clone.Clone(&config.DefaultConfig)
	err := cfg.Validate()
	assert.NoError(t, err)
}
