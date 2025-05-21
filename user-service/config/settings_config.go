package config

type CorsConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins" json:"allowed_origins" validate:"required"`
	AllowedMethods []string `mapstructure:"allowed_methods" json:"allowed_methods" validate:"required"`
	AllowedHeaders []string `mapstructure:"allowed_headers" json:"allowed_headers" validate:"required"`
}

type ApiSettingsConfig struct {
	Addr    string      `mapstructure:"addr" json:"addr" validate:"required"`
	Host    string      `mapstructure:"host,omitempty" json:"host,omitempty"`
	OrgName string      `mapstructure:"org_name,omitempty" json:"org_name,omitempty"`
	Cors    *CorsConfig `mapstructure:"cors,omitempty" json:"cors,omitempty"`
}

type ObservabilitySettingsConfig struct {
	LogFormat         string `mapstructure:"log_format" json:"log_format" validate:"required"`
	OtelCollectorAddr string `mapstructure:"otel_collector_addr" json:"otel_collector_addr"`
	TlsKeylogFile     string `mapstructure:"tls_keylog_file" json:"tls_keylog_file"`
}
