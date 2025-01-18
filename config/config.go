package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config structure holds the configuration from the config file
type Config struct {
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		DBName   string `mapstructure:"dbname"`
		SSLMode  string `mapstructure:"sslMode"`
		Timeout  int    `mapstructure:"connectionTimeout"`
	} `mapstructure:"database"`
	Context struct {
		Timeout int `mapstructure:"timeout"`
	} `mapstructure:"context"`
	AccessToken struct {
		Secret string `mapstructure:"secret"`
	} `mapstructure:"access_token"`
	Server struct {
		LogLevel string `mapstructure:"logLevel"`
		AppEnv   string `mapstructure:"appEnv"`
		Address  string `mapstructure:"address"`
	} `mapstructure:"server"`
	Redis struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Password string `mapstructure:"password"`
		Timeout  string `mapstructure:"timeout"`
		SSL      bool   `mapstructure:"ssl"`
	} `mapstructure:"redis"`
}

// LoadConfig reads configuration from file and environment variables
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config") // config.json or config.yaml
	viper.SetConfigType("toml")   // Config file format (toml, json, yaml)
	viper.AddConfigPath("./env")  // Path to look for config file

	// If you want to read environment variables
	viper.AutomaticEnv()

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
		return nil, err
	}

	return &config, nil
}
