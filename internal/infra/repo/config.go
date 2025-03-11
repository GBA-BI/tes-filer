package repo

import "github.com/spf13/pflag"

type Config struct {
	S3ConfigPath         string `env:"S3SDK_CONFIG_FILE"`
	ExpirationConfigPath string `env:"AWS_CREDENTIALS_EXPIRED_TIME_FILE"`
	S3SecretPath         string `env:"AWS_SHARED_CREDENTIALS_FILE"`

	OffloadType string `env:"OFFLOAD_TYPE"`

	IsMountTOS string `env:"IS_MOUNT_TOS"`
}

func NewConfig() *Config {
	return &Config{
		OffloadType: "pvc",
	}
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) AddFlags(fs *pflag.FlagSet) {
	return
}
