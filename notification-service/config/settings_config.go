package config

type GeneralSettingsConfig struct {
	OrgName        string `mapstructure:"org_name,omitempty" json:"org_name,omitempty"`
	WebsiteBaseURL string `mapstructure:"website_base_url,omitempty" json:"website_base_url,omitempty"`
}

type ObservabilitySettingsConfig struct {
	LogFormat         string `mapstructure:"log_format" json:"log_format" validate:"required"`
	OtelCollectorAddr string `mapstructure:"otel_collector_addr" json:"otel_collector_addr"`
	TlsKeylogFile     string `mapstructure:"tls_keylog_file" json:"tls_keylog_file"`
}
