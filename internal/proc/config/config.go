// Package config provides functionality to load configuration settings
// from environment variables into a Config struct.
package config

import (
	"github.com/Alekseyt9/upscaler/internal/common/utils"
	"github.com/caarlos0/env"
)

// Config holds the configuration settings required for the application.
type Config struct {
	RedisAddr          string `env:"REDIS_ADDRESS"`        // Address of the Redis server.
	KafkaAddr          string `env:"KAFKA_ADDRESS"`        // Address of the Kafka server.
	KafkaTopic         string `env:"KAFKA_TOPIC"`          // Name of the Kafka topic for sending messages.
	KafkaTopicResult   string `env:"KAFKA_TOPIC_RESULT"`   // Name of the Kafka topic for receiving results.
	KafkeCunsumerGroup string `env:"KAFKA_CONSUMER_GROUP"` // Name of the Kafka consumer group.
	S3AccessKeyID      string `env:"S3_ACCESSKEYID"`       // Access key ID for the S3 storage.
	S3SecretAccessKey  string `env:"S3_SECRETACCESSKEY"`   // Secret access key for the S3 storage.
	S3BucketName       string `env:"S3_BUCKETNAME"`        // Name of the S3 bucket.
}

// LoadConfig loads the configuration settings from environment variables.
func LoadConfig() (*Config, error) {
	err := utils.LoadEnv()
	if err != nil {
		return nil, err
	}

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
