package config

import (
	"log"

	"github.com/spf13/viper"
)

var config Config

type Config struct {
	Env      string   `yaml:"env"`
	App      App      `yaml:"app"`
	DB       DB       `yaml:"db"`
	RabbitMQ RabbitMQ `yaml:"rabbitmq"`
}

func Load(configFile string) {
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not load config file %s: %v", configFile, err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Failed to unmarshal the config into struct: %s", err)
	}

	// Debug viper values directly
	log.Printf("Debug - Viper rabbitmq.username: %s", viper.GetString("rabbitmq.queue"))
	log.Printf("Debug - RabbitMQ config loaded: Host=%s, Port=%d, User=%s, Password=%s",
		config.RabbitMQ.Host, config.RabbitMQ.Port, config.RabbitMQ.UserName, config.RabbitMQ.Queue)
}

func GetConfig() *Config {
	return &config
}
