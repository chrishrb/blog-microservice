package config

type LocalSourceConfig struct {
	Type string `mapstructure:"type" json:"type" validate:"required,oneof=file"`
	File string `mapstructure:"file,omitempty" json:"file,omitempty" validate:"required_if=Type file"`
}

type AuthConfig struct {
	Issuer           string             `mapstructure:"issuer" json:"issuer" validate:"required"`
	Audience         string             `mapstructure:"audience" json:"audience" validate:"required"`
	ExpiresIn        string             `mapstructure:"expires_in" json:"expires_in" validate:"required"`
	PublicKeySource  *LocalSourceConfig `mapstructure:"public_key" json:"public_key" validate:"required"`
	PrivateKeySource *LocalSourceConfig `mapstructure:"private_key" json:"private_key" validate:"required"`
}
