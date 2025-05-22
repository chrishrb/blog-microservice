package config_test

import (
	"testing"

	"github.com/chrishrb/blog-microservice/notification-service/config"
	clone "github.com/huandu/go-clone/generic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	cfg := clone.Clone(&config.DefaultConfig)
	err := cfg.LoadFromFile("testdata/config.yaml")
	require.NoError(t, err)

	want := &config.BaseConfig{
		General: config.GeneralSettingsConfig{
			OrgName:        "Blog Microservices",
			WebsiteBaseURL: "https://example.com",
		},
		Transport: config.TransportSettingsConfig{
			Type: "kafka",
			Kafka: &config.KafkaSettingsConfig{
				Urls:           []string{"localhost:9092"},
				Group:          "notification-service",
				ConnectTimeout: "10s",
			},
		},
		Observability: config.ObservabilitySettingsConfig{
			LogFormat:         "text",
			OtelCollectorAddr: "localhost:4317",
			TlsKeylogFile:     "/keylog/notification-service.log",
		},
		Channels: config.ChannelsSettingsConfig{
			Email: config.EmailSettingsConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "myuser",
				Password: "mypassword",
				FromAddr: "myuser@example.com",
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
