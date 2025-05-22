package config

type EmailSettingsConfig struct {
	Host     string `mapstructure:"host" json:"host" validate:"required"`
	Port     int    `mapstructure:"port" json:"port" validate:"required"`
	Username string `mapstructure:"username" json:"username" validate:"required"`
	Password string `mapstructure:"password" json:"password" validate:"required"`
	FromAddr string `mapstructure:"from_addr" json:"from_addr" validate:"required"`
}

type ChannelsSettingsConfig struct {
	Email EmailSettingsConfig `mapstructure:"email" json:"email" validate:"required"`
}
