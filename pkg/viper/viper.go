package viper

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/spf13/viper"
)

func SetConfigFromEnv(config interface{}) {
	configValue := reflect.ValueOf(config)
	if configValue.Kind() == reflect.Ptr {
		configValue = configValue.Elem()
	}
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		field := configType.Field(i)
		envVarName := field.Tag.Get("env")

		if envVarName != "" {
			viper.BindEnv(envVarName)
			envVarValue := viper.Get(envVarName)
			if envVarValue != nil {
				configValue.Field(i).Set(reflect.ValueOf(envVarValue).Convert(field.Type))
			}
		} else {
			fieldValue := configValue.Field(i)
			if fieldValue.Kind() == reflect.Ptr {
				fieldValue = fieldValue.Elem()
			}
			if fieldValue.Kind() == reflect.Struct {
				SetConfigFromEnv(fieldValue.Addr().Interface())
			}
		}
	}
}

var mutex = &sync.Mutex{}

func SetConfigFromFileINI(configFile string, section string, conf interface{}) error {
	mutex.Lock()
	defer mutex.Unlock()

	viper.SetConfigFile(configFile)
	viper.SetConfigType("ini")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	sec := "default"
	if section != "" {
		sec = section
	}
	if err := viper.UnmarshalKey(sec, &conf); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}
