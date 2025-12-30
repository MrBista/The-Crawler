package conf

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	DBConfig DBConfig
}

type DBConfig struct {
	Host     string
	Password string
	User     string
	DBName   string
	Port     string
	SSLMode  string
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// v.set

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("[ERROR_ENV] failed to read config because %v", err)
			return nil, err
		}
	}

	var config Config

	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
