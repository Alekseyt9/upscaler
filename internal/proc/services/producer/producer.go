package producer

import (
	"log/slog"

	"github.com/Alekseyt9/upscaler/internal/common/model"
	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/IBM/sarama"
)

type Producer struct {
	topic    string
	producer sarama.SyncProducer
	log      *slog.Logger
}

// NewProducer создает новый продюсер, который будет отправлять сообщения в указанный топик.
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

// Send отправляет одно сообщение в Kafka топик.
func (p *Producer) Send(msg model.BrokerMessageResult) error {
	message := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(msg.Result),
	}

	_, _, err := p.producer.SendMessage(message)
	if err != nil {
		p.log.Error("producer Send Ошибка при отправке сообщения в Kafka", "error", err)
		return err
	}

	return nil
}

// Close закрывает продюсер Kafka.
func (b *Producer) Close() error {
	return b.producer.Close()
}
