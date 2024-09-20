// Package messagebroker provides functionality for producing messages to Kafka
// from a store, allowing tasks to be sent in batches to a Kafka topic.
package messagebroker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Alekseyt9/upscaler/internal/back/model"
	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/IBM/sarama"
)

// Producer is a struct that represents a Kafka message producer.
type Producer struct {
	store    store.Store         // Store interface for retrieving tasks to send to Kafka.
	producer sarama.SyncProducer // Kafka SyncProducer for sending messages synchronously.
	topic    string              // Kafka topic to which the messages will be sent.
	interval time.Duration       // Interval for batching and sending messages.
	quit     chan struct{}       // Channel to signal stopping the message producer.
	log      *slog.Logger        // Logger for capturing logs and errors.
}

// NewProducer creates and returns a new instance of Producer.
// It initializes the Kafka producer, starts a goroutine for sending messages in batches,
// and returns the configured Producer instance.
//
// Parameters:
//   - store: The store from which tasks are retrieved for sending to Kafka.
//   - log: A logger for capturing logs and errors.
//   - opt: BrokerOptions containing Kafka broker addresses and the topic.
//
// Returns:
//   - A pointer to a Producer instance.
//   - An error if there is a problem creating the Kafka SyncProducer.
func NewProducer(store store.Store, log *slog.Logger, opt cmodel.BrokerOptions) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Flush.Frequency = 500 * time.Millisecond

	producer, err := sarama.NewSyncProducer(opt.KafkaBrokers, config)
	if err != nil {
		return nil, err
	}

	sender := &Producer{
		store:    store,
		producer: producer,
		topic:    opt.Topic,
		interval: 3 * time.Second,
		quit:     make(chan struct{}),
		log:      log,
	}

	go sender.startSending()
	return sender, nil
}

// startSending starts a loop that triggers message sending at regular intervals.
func (s *Producer) startSending() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := s.sendMessagesBatch()
			if err != nil {
				s.log.Error("sendMessagesBatch", "error", err)
			}
		case <-s.quit:
			s.log.Info("producer Stopping message sending")
			return
		}
	}
}

// sendMessagesBatch retrieves tasks from the store and sends them to Kafka in batches.
func (s *Producer) sendMessagesBatch() error {
	err := s.store.SendTasksToBroker(context.Background(), func(items []model.OutboxItem) error {
		var messages []*sarama.ProducerMessage
		for _, item := range items {
			msg := &sarama.ProducerMessage{
				Topic: s.topic,
				Key:   sarama.StringEncoder(item.IdKey),
				Value: sarama.StringEncoder(item.Payload),
			}
			messages = append(messages, msg)
		}

		err := s.producer.SendMessages(messages)
		if err != nil {
			return fmt.Errorf("producer.SendMessages %w", err)
		}
		s.log.Info("producer sendMessagesBatch", "messages", messages)

		return nil
	})

	if err != nil {
		return fmt.Errorf("sendMessagesBatch store.SendTasksToBroker %w", err)
	}

	return nil
}

// Close stops the producer and gracefully shuts down the Kafka producer.
func (s *Producer) Close() {
	close(s.quit)
	s.producer.Close()
}
