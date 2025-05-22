package config

type EmailSettingsConfig struct {
	Host     string `mapstructure:"host" json:"host" validate:"required"`
	Port     int    `mapstructure:"port" json:"port" validate:"required"`
	Username string `mapstructure:"username,omitempty" json:"username,omitempty"`
	Password string `mapstructure:"password,omitempty" json:"password,omitempty"`
	FromAddr string `mapstructure:"from_addr" json:"from_addr" validate:"required"`
}

type ChannelsSettingsConfig struct {
	Email EmailSettingsConfig `mapstructure:"email" json:"email" validate:"required"`
}
