package config

import (
	"bufio"
	"io"
	"os"

	"github.com/go-playground/validator/v10"
	"sigs.k8s.io/yaml"
)

// BaseConfig provides the data structures that represent the configuration
// and provides the ability to load the configuration from a YAML file.
type BaseConfig struct {
	General       GeneralSettingsConfig       `mapstructure:"general" json:"general"`
	Transport     TransportSettingsConfig     `mapstructure:"transport" json:"transport" validate:"required"`
	Observability ObservabilitySettingsConfig `mapstructure:"observability" json:"observability" validate:"required"`
	Channels      ChannelsSettingsConfig      `mapstructure:"channels" json:"channels"`
}

// DefaultConfig provides the default configuration. The configuration
// read from the YAML file will overlay this configuration.
var DefaultConfig = BaseConfig{
	General: GeneralSettingsConfig{
		OrgName:        "Blog Microservice",
		WebsiteBaseURL: "https://example.com",
	},
	Transport: TransportSettingsConfig{
		Type: "kafka",
		Kafka: &KafkaSettingsConfig{
			Urls:           []string{"localhost:9092"},
			Group:          "post-service",
			ConnectTimeout: "10s",
		},
	},
	Observability: ObservabilitySettingsConfig{
		LogFormat: "text",
	},
	Channels: ChannelsSettingsConfig{
		Email: EmailSettingsConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "myuser",
			Password: "mypassword",
			FromAddr: "myuser@example.com",
		},
	},
}

// Load reads YAML configuration from a reader.
func (c *BaseConfig) Load(reader io.Reader) error {
	b, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(b, c); err != nil {
		return err
	}
	return nil
}

// LoadFromFile reads YAML configuration from a file.
func (c *BaseConfig) LoadFromFile(configFile string) error {
	//#nosec G304 - only files specified by the person running the application will be loaded
	f, err := os.Open(configFile)
	if err != nil {
		return err
	}
	err = c.Load(bufio.NewReader(f))
	return err
}

// Validate ensures that the configuration is structurally valid.
func (c *BaseConfig) Validate() error {
	validate := validator.New()

	return validate.Struct(c)
}
