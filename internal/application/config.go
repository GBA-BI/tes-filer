package application

import (
	"strings"

	apperror "github.com/GBA-BI/tes-filer/pkg/error"
	utilsstrings "github.com/GBA-BI/tes-filer/pkg/utils/strings"
)

type Config struct {
	Path string `env:"POD_INFO_ANNOTATIONS_FILE"`
	Mode string `env:"FILER_MODE"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Validate() error {
	if flag := utilsstrings.Contains([]string{"inputs", "outputs", "all"}, strings.ToLower(c.Mode)); !flag {
		return apperror.NewInvalidArgumentError("Config.Mode", c.Mode)
	}
	if c.Path == "" {
		return apperror.NewInvalidArgumentError("Config.Path", c.Path)
	}
	return nil
}
