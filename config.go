package client

import (
	"github.com/spf13/viper"
	"time"
)

// Config contains required configuration for the client connection
type Config struct {
	// Port is the port used by the gateway
	Port int `mapstructure:"port" validate:"required"`

	// Hostname is the hostname used by the client
	Hostname string `mapstructure:"hostname" validate:"hostname"`

	// PathBase is the base of the api path it might contain the version
	PathBase string `mapstructure:"path_base"`

	// APIVersion defines the api version
	APIVersion int `mapstructure:"api_version"`

	// HTTPS defines if the connection is over the https protocol
	HTTPS bool `mapstructure:"https"`

	// Timeout defines the client timeout
	Timeout time.Duration `mapstructure:"timeout"`
}

// ReadConfig reads the config from the provided path and name
func ReadConfig(name, path string) (cfg *Config, err error) {
	v := viper.New()

	v.AddConfigPath(path)
	v.SetConfigName(name)
	defaultConfig(v)

	if err = v.ReadInConfig(); err != nil {
		return
	}

	cfg = &Config{}
	if err = v.Unmarshal(cfg); err != nil {
		return
	}

	return
}

func defaultConfig(v *viper.Viper) {
	v.Set("timeout", time.Second*20)
	v.Set("port", 80)
	v.Set("api_version", 1)
}
