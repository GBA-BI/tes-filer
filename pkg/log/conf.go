package log

import (
	"errors"

	"github.com/spf13/pflag"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level       string   `env:"LOG_LEVEL"`
	Encoding    string   `env:"LOG_ENCODING"`
	OutputPaths []string `env:"LOG_OUTPUT_PATHS"`
	Compress    bool     `env:"LOG_COMPRESS"`
}

func NewConfig() *Config {
	return &Config{
		Level:       "info",
		Encoding:    "console",
		OutputPaths: []string{"stdout"},
	}
}

func (c *Config) Validate() error {
	// Check if Level is a valid log level
	if _, err := zapcore.ParseLevel(c.Level); err != nil {
		return err
	}

	// Check if Encoding is a valid encoding
	if c.Encoding != "json" && c.Encoding != "console" {
		return errors.New("invalid encoding")
	}

	// Check if OutputPaths and ErrorOutputPaths are not empty
	if len(c.OutputPaths) == 0 {
		return errors.New("output paths cannot be empty")
	}

	return nil
}

func (c *Config) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.Level, "log-level", "info", "log level")
	fs.StringVar(&c.Encoding, "log-encoding", "console", "log encoding")
	fs.StringSliceVar(&c.OutputPaths, "log-file", []string{"stdout"}, "log output paths")
}
