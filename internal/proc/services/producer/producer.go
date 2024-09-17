package producer

import (
	"log"

	"github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/IBM/sarama"
)

type Producer struct {
	topic    string
	producer sarama.SyncProducer
}

// NewProducer создает новый продюсер, который будет отправлять сообщения в указанный топик.
func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		topic:    topic,
		producer: producer,
	}, nil
}

// Send отправляет одно сообщение в Kafka топик.
func (b *Producer) Send(msg model.BrokerMessageResult) error {
	message := &sarama.ProducerMessage{
		Topic: b.topic,
		Value: sarama.StringEncoder(msg.Result),
	}

	_, _, err := b.producer.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения в Kafka: %v", err)
		return err
	}

	return nil
}

// Close закрывает продюсер Kafka.
func (b *Producer) Close() error {
	return b.producer.Close()
}
