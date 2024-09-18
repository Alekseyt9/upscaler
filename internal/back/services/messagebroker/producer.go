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

type Producer struct {
	store    store.Store
	producer sarama.SyncProducer
	topic    string
	interval time.Duration
	quit     chan struct{}
	log      *slog.Logger
}

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

func (s *Producer) Close() {
	close(s.quit)
	s.producer.Close()
}
