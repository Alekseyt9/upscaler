package config

import (
	"log"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("warning: .env file not found, using system environment variables")
	}

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
