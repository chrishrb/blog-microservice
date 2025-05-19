package config

type ApiSettingsConfig struct {
	Addr    string `mapstructure:"addr" json:"addr" validate:"required"`
	Host    string `mapstructure:"host,omitempty" json:"host,omitempty"`
	OrgName string `mapstructure:"org_name,omitempty" json:"org_name,omitempty"`
}

type ObservabilitySettingsConfig struct {
	LogFormat         string `mapstructure:"log_format" json:"log_format" validate:"required"`
	OtelCollectorAddr string `mapstructure:"otel_collector_addr" json:"otel_collector_addr"`
	TlsKeylogFile     string `mapstructure:"tls_keylog_file" json:"tls_keylog_file"`
}
