package options

import (
	"github.com/spf13/pflag"

	"github.com/GBA-BI/tes-filer/internal/application"
	"github.com/GBA-BI/tes-filer/internal/infra/repo"
	"github.com/GBA-BI/tes-filer/pkg/log"
	"github.com/GBA-BI/tes-filer/pkg/viper"
)

type Options struct {
	Log        *log.Config
	AppFiler   *application.Config
	RepoConfig *repo.Config
}

func NewOptions() *Options {
	return &Options{
		Log:        log.NewConfig(),
		AppFiler:   application.NewConfig(),
		RepoConfig: repo.NewConfig(),
	}
}

func (o *Options) Validate() error {
	if err := o.Log.Validate(); err != nil {
		return err
	}
	if err := o.AppFiler.Validate(); err != nil {
		return err
	}
	if err := o.RepoConfig.Validate(); err != nil {
		return err
	}
	return nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.Log.AddFlags(fs)
	o.RepoConfig.AddFlags(fs)
}

func NewFromENV() *Options {
	opt := NewOptions()
	viper.SetConfigFromEnv(opt)
	return opt
}
