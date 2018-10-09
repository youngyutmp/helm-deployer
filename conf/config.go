package conf

import (
	log "github.com/sirupsen/logrus"
)

// Config the application's configuration
type Config struct {
	API struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"api"`

	APP struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"app"`

	ChartRepository struct {
		BaseURL string `mapstructure:"baseUrl"`
	} `mapstructure:"chartRepository"`

	DB struct {
		Path string `mapstructure:"path"`
	} `mapstructure:"db"`

	K8S struct {
		ConfigPath string `configPath:"host"`
	} `mapstructure:"k8s"`

	LogConfig struct {
		Level     string `mapstructure:"level"`
		File      string `mapstructure:"file"`
		Formatter string `mapstructure:"formatter"`
		Logger    *log.Entry
	} `mapstructure:"log_config"`

	Tiller struct {
		Host string `mapstructure:"host"`
	} `mapstructure:"tiller"`
}

// ValidateConfig validates config
func (c *Config) ValidateConfig() error {
	if c.API.Port == 0 {
		c.API.Port = 8080
	}
	if c.API.Host == "" {
		c.API.Host = "localhost"
	}

	return nil
}
