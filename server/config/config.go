package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Port int
		Env  string
	}
	Mongo struct {
		URI    string
		DBName string
	}
	CORS struct {
		AllowOrigins []string
		AllowMethods []string
		AllowHeaders []string
	}
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/wisp/")
	viper.AddConfigPath("$HOME/.wisp")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
