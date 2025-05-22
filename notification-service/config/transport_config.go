package config

type KafkaSettingsConfig struct {
	Urls           []string `mapstructure:"urls" json:"urls" validate:"required,dive,required"`
	Group          string   `mapstructure:"group" json:"group" validate:"required"`
	ConnectTimeout string   `mapstructure:"connect_timeout" json:"connect_timeout" validate:"required"`
}

type TransportSettingsConfig struct {
	Type  string               `mapstructure:"type" json:"type" validate:"required,oneof=kafka"`
	Kafka *KafkaSettingsConfig `mapstructure:"kafka,omitempty" json:"kafka,omitempty" validate:"required_if=Type kafka"`
}
