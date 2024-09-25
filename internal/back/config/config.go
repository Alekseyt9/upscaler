// Package config provides functionality to load configuration settings
// from environment variables into a Config struct.
package config

import (
	"github.com/Alekseyt9/upscaler/internal/common/utils"
	"github.com/caarlos0/env"
)

// Config holds the configuration settings required for the application.
type Config struct {
	BackAddress        string `env:"BACK_ADDRESS"`         // Address of the backend service.
	PgDataBaseDSN      string `env:"DATABASE_DSN"`         // Data source name for PostgreSQL database connection.
	S3AccessKeyID      string `env:"S3_ACCESSKEYID"`       // Access key ID for the S3 storage.
	S3SecretAccessKey  string `env:"S3_SECRETACCESSKEY"`   // Secret access key for the S3 storage.
	S3BucketName       string `env:"S3_BUCKETNAME"`        // Name of the S3 bucket.
	JWTSecret          string `env:"WT_SECRET"`            // Secret key used for JWT token generation.
	KafkaAddr          string `env:"KAFKA_ADDRESS"`        // Address of the Kafka server.
	KafkaTopic         string `env:"KAFKA_TOPIC"`          // Name of the Kafka topic for sending messages.
	KafkaTopicResult   string `env:"KAFKA_TOPIC_RESULT"`   // Name of the Kafka topic for receiving results.
	KafkeCunsumerGroup string `env:"KAFKA_CONSUMER_GROUP"` // Name of the Kafka consumer group.
}

// LoadConfig loads the configuration settings from environment variables
//
// Returns:
//   - A pointer to a Config instance.
//   - An error if the environment variables cannot be loaded or parsed.
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
