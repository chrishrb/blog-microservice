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
	Api                       ApiSettingsConfig               `mapstructure:"api" json:"api" validate:"required"`
	Transport                 TransportConfig                 `mapstructure:"transport" json:"transport" validate:"required"`
	Observability             ObservabilitySettingsConfig     `mapstructure:"observability" json:"observability" validate:"required"`
	Storage                   StorageConfig                   `mapstructure:"storage" json:"storage" validate:"required"`
}

// DefaultConfig provides the default configuration. The configuration
// read from the YAML file will overlay this configuration.
var DefaultConfig = BaseConfig{
	Api: ApiSettingsConfig{
		Addr:    "localhost:9410",
		Host:    "localhost",
		OrgName: "chrishrb",
	},
	Transport: TransportConfig{
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
	Storage: StorageConfig{
		Type: "in_memory",
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
