package config

import (
	"github.com/Alekseyt9/upscaler/internal/common/envutils"
	"github.com/caarlos0/env"
)

type Config struct {
	RedisAddr          string `env:"REDIS_ADDRESS"`
	KafkaAddr          string `env:"KAFKA_ADDRESS"`
	KafkaTopic         string `env:"KAFKA_TOPIC"`
	KafkaTopicResult   string `env:"KAFKA_TOPIC_RESULT"`
	KafkeCunsumerGroup string `env:"KAFKA_CONSUMER_GROUP"`
	S3AccessKeyID      string `env:"S3_ACCESSKEYID"`
	S3SecretAccessKey  string `env:"S3_SECRETACCESSKEY"`
	S3BucketName       string `env:"S3_BUCKETNAME"`
}

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
