package viper

import (
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestSetConfigFromEnv(t *testing.T) {

	type test struct {
		name     string
		envVars  map[string]string
		config   interface{}
		expected interface{}
	}

	tests := []test{
		{
			name: "set string field from env var",
			envVars: map[string]string{
				"APP_NAME": "myapp",
			},
			config: &struct {
				AppName string `env:"APP_NAME"`
			}{},
			expected: &struct {
				AppName string `env:"APP_NAME"`
			}{
				AppName: "myapp",
			},
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			for envVarName, envVarValue := range tc.envVars {
				os.Setenv(envVarName, envVarValue)
				defer os.Unsetenv(envVarName)
			}
			SetConfigFromEnv(tc.config)
			convey.So(tc.config, convey.ShouldResemble, tc.expected)
		})
	}

}

func TestSetConfigFromFileINI(t *testing.T) {
	type test struct {
		name       string
		configFile string
		section    string
		expected   struct {
			Key string `mapstructure:"key"`
		}
	}

	tests := []test{
		{
			name:       "set string field from ini file",
			configFile: "./test.ini",
			section:    "",
			expected: struct {
				Key string `mapstructure:"key"`
			}{
				Key: "value",
			},
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			var conf struct {
				Key string `mapstructure:"key"`
			}
			err := SetConfigFromFileINI(tc.configFile, tc.section, &conf)
			convey.So(err, convey.ShouldBeNil)
			convey.So(conf, convey.ShouldResemble, tc.expected)
		})
	}
}
