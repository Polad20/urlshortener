package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type URLConfig struct {
	Domain  string `mapstructure:"DOMAIN"`
	Length  string `mapstructure:"LENGTH"`
	Charset string `mapstructure:"CHARSET"`
}

type RepositoryConfig struct {
	Type  string `mapstructure:"TYPE"`
	PgURL string `mapstructure:"PG_URL"`
}

type AuthConfig struct {
	Key string `mapstructure:"KEY"`
}

type Logger struct {
	Level   string `mapstructure:"LEVEL"`
	Format  string `mapstructure:"FORMAT"`
	LogPath string `mapstructure:"PATH"`
}

type Config struct {
	URL        URLConfig        `mapstructure:"url"`
	Repository RepositoryConfig `mapstructure:"repository"`
	Auth       AuthConfig       `mapstructure:"auth"`
	Log        Logger           `mapstructure:"log"`
}

var AppConfig Config

func LoadConfig() error {
	err := godotenv.Load()
	if err != nil {
		log.Err(err).Msg("Can`t read .env file")
	}
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AutomaticEnv()
	err = v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Configuration file not found")
		} else {
			return fmt.Errorf("Some unexpected error reading config: %v. Using .env variables", err)
		}
	}
	err = v.Unmarshal(&AppConfig)
	if err != nil {
		return fmt.Errorf("Error unmarshalling config data: %v", err)
	}
	return nil
}
