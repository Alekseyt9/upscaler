package messagebroker

import (
	"context"
	"log"
	"time"

	"github.com/Alekseyt9/upscaler/internal/back/model"
	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	"github.com/IBM/sarama"
	_ "github.com/lib/pq"
)

type Producer struct {
	store    store.Store
	producer sarama.SyncProducer
	interval time.Duration
	quit     chan struct{}
}

func NewProducer(store store.Store, kafkaBrokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Flush.Frequency = 500 * time.Millisecond

	producer, err := sarama.NewSyncProducer(kafkaBrokers, config)
	if err != nil {
		return nil, err
	}

	sender := &Producer{
		store:    store,
		producer: producer,
		interval: 3 * time.Second,
		quit:     make(chan struct{}),
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
				log.Println("error", err) // TODO slog
			}
		case <-s.quit:
			log.Println("Остановка отправки сообщений")
			return
		}
	}
}

func (s *Producer) sendMessagesBatch() error {
	err := s.store.SendTasksToBroker(context.Background(), func(items []model.OutboxItem) error {
		var messages []*sarama.ProducerMessage
		for _, item := range items {
			msg := &sarama.ProducerMessage{
				Topic: "file_processing",
				Key:   sarama.StringEncoder(item.IdKey),
				Value: sarama.StringEncoder(item.Payload),
			}
			messages = append(messages, msg)
		}

		err := s.producer.SendMessages(messages)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *Producer) Stop() {
	close(s.quit)
	s.producer.Close()
}
