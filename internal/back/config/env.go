package config

import (
	"github.com/Alekseyt9/upscaler/internal/common/envutils"
	"github.com/caarlos0/env"
)

func LoadConfig() (*Config, error) {
	err := envutils.LoadEnv()
	if err != nil {
		return nil, err
	}

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
