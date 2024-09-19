// Package producer provides functionality for sending messages to a Kafka topic.
// It uses the Sarama library to produce messages in a synchronous manner.
package producer

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Alekseyt9/upscaler/internal/common/model"
	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/IBM/sarama"
)

// Producer is a struct that wraps a Kafka producer and provides methods
// for sending messages to a Kafka topic.
type Producer struct {
	topic    string
	producer sarama.SyncProducer
	log      *slog.Logger
}

// NewProducer creates a new Producer that will send messages to the specified Kafka topic.
//
// Parameters:
//   - log: A logger instance for logging errors and other information.
//   - opt: BrokerOptions containing configuration such as the Kafka brokers and topic.
//
// Returns:
//   - A pointer to a Producer instance.
//   - An error if the producer could not be created.
func NewProducer(log *slog.Logger, opt cmodel.BrokerOptions) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	producer, err := sarama.NewSyncProducer(opt.KafkaBrokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		topic:    opt.Topic,
		producer: producer,
		log:      log,
	}, nil
}

// Send sends a single message to the Kafka topic.
//
// Parameters:
//   - msg: The message to be sent, encapsulated in a BrokerMessageResult struct.
//
// Returns:
//   - An error if the message could not be serialized or sent.
func (p *Producer) Send(msg model.BrokerMessageResult) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("producer Send: error serializing message: %w", err)
	}

	message := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(msgBytes),
	}

	_, _, err = p.producer.SendMessage(message)
	if err != nil {
		p.log.Error("producer Send: error sending message to Kafka", "error", err)
		return err
	}

	return nil
}

// Close closes the Kafka producer.
//
// Returns:
//   - An error if the producer could not be closed.
func (b *Producer) Close() error {
	return b.producer.Close()
}
