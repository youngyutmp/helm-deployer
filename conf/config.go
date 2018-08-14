package conf

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Config the application's configuration
type Config struct {
	API struct {
		Host string `mapstructure:"host" json:"host"`
		Port int    `mapstructure:"port" json:"port"`
	} `mapstructure:"api" json:"api"`

	APP struct {
		Username string `mapstructure:"username" json:"username"`
		Password string `mapstructure:"password" json:"password"`
	} `mapstructure:"app" json:"app"`

	ChartRepository struct {
		BaseURL string `mapstructure:"baseUrl" json:"baseUrl"`
	} `mapstructure:"chartRepository" json:"chartRepository"`

	DB struct {
		Path string `mapstructure:"path" json:"path"`
	} `mapstructure:"db" json:"db"`

	K8S struct {
		ConfigPath string `configPath:"host" json:"configPath"`
	} `mapstructure:"k8s" json:"k8s"`

	LogConfig struct {
		Level string `mapstructure:"level"`
		File  string `mapstructure:"file"`
	} `mapstructure:"log_config" json:"log_config"`

	Tiller struct {
		Host string `mapstructure:"host" json:"host"`
	} `mapstructure:"tiller" json:"tiller"`
}

// Load will construct the config from the file
func Load(configFile string) (*Config, error) {
	viper.SetConfigType("json")

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("./") // ./config.[json | toml]
	}

	viper.SetEnvPrefix("HELM_DEPLOYER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrap(err, "reading configuration from files")
	}

	config := new(Config)
	if err := viper.Unmarshal(config); err != nil {
		return nil, errors.Wrap(err, "unmarshaling configuration")
	}

	if err := configureLogging(config); err != nil {
		return nil, errors.Wrap(err, "configure logging")
	}

	return validateConfig(config)
}

func configureLogging(config *Config) error {
	// always use the full timestamp
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: false,
	})

	// use a file if you want
	if config.LogConfig.File != "" {
		f, errOpen := os.OpenFile(config.LogConfig.File, os.O_RDWR|os.O_APPEND, 0660)
		if errOpen != nil {
			return errOpen
		}
		logrus.SetOutput(bufio.NewWriter(f))
		logrus.Infof("Set output file to %s", config.LogConfig.File)
	}

	if config.LogConfig.Level != "" {
		level, err := logrus.ParseLevel(strings.ToUpper(config.LogConfig.Level))
		if err != nil {
			return err
		}
		logrus.SetLevel(level)
		logrus.Debug("Set log level to: " + logrus.GetLevel().String())
	}

	return nil
}

func validateConfig(config *Config) (*Config, error) {
	if config.API.Port == 0 && os.Getenv("PORT") != "" {
		port, err := strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			return nil, errors.Wrap(err, "formatting PORT into int")
		}

		config.API.Port = port
	}

	if config.API.Port == 0 && config.API.Host == "" {
		config.API.Port = 8080
	}

	return config, nil
}
