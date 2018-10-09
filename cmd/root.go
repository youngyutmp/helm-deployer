package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/entwico/helm-deployer/conf"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

const (
	serviceName      = "helm-deployer"
	serviceEnvPrefix = "HELM_DEPLOYER"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   serviceName,
	Short: "Redeploys helm charts",
	Long: `helm-deployer
listens to webhooks from GitLab/Nexus and redeploys helm-charts`,

	Run: func(cmd *cobra.Command, args []string) {
		executeWithConfig(serve)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func executeWithConfig(fn func(config *conf.Config)) {
	var config conf.Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := configureLogging(&config); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fn(&config)
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name "config" (without extension).
		viper.AddConfigPath("./")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix(serviceEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "using config file:", viper.ConfigFileUsed())
	}
}

func configureLogging(config *conf.Config) error {
	logConfig := config.LogConfig
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	textFormatter := &log.TextFormatter{
		FullTimestamp:          true,
		ForceColors:            true,
		DisableLevelTruncation: true,
	}
	log.SetFormatter(textFormatter)
	if strings.ToLower(logConfig.Formatter) == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	}
	logger := log.WithFields(log.Fields{"hostname": hostname, "service_name": serviceName})

	// use a file if you want
	if logConfig.File != "" {
		f, errOpen := os.OpenFile(logConfig.File, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
		if errOpen != nil {
			return errOpen
		}
		log.WithField("output_file", logConfig.File).Info("set output file")
		textFormatter.ForceColors = false
		log.SetOutput(f)
	}

	if logConfig.Level != "" {
		level, err := log.ParseLevel(strings.ToUpper(logConfig.Level))
		if err != nil {
			return err
		}
		log.SetLevel(level)
		log.WithField("log_level", log.GetLevel().String()).Info("set log level")
	}

	config.LogConfig.Logger = logger
	return nil
}
