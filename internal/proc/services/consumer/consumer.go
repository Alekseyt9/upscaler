package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/Alekseyt9/upscaler/internal/proc/services/processor"
	"github.com/IBM/sarama"
)

type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	brokers       []string
	topic         string
	group         string
	log           *slog.Logger
	ctx           context.Context
	cancel        context.CancelFunc
}

// ConsumerGroupHandler реализует интерфейс sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	proc *processor.ProcessorService
	log  *slog.Logger
}

// NewConsumer создает новый экземпляр ConsumerService
func NewConsumer(proc *processor.ProcessorService, log *slog.Logger, opt cmodel.BrokerOptions) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_6_0_0

	consumerGroup, err := sarama.NewConsumerGroup(opt.KafkaBrokers, opt.ConsumerGroup, config)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания группы потребителей: %w", err)
	}

	handler := ConsumerGroupHandler{
		proc: proc,
		log:  log,
	}

	ctx, cancel := context.WithCancel(context.Background())

	cs := &Consumer{
		consumerGroup: consumerGroup,
		brokers:       opt.KafkaBrokers,
		topic:         opt.Topic,
		group:         opt.ConsumerGroup,
		log:           log,
		ctx:           ctx,
		cancel:        cancel,
	}

	go func() {
		for {
			select {
			case <-cs.ctx.Done():
				return
			default:
				if err := cs.consumerGroup.Consume(cs.ctx, []string{cs.topic}, &handler); err != nil {
					log.Error("consumerGroup.Consume Ошибка при обработке сообщений", "error", err)
				}
			}
		}
	}()

	return cs, nil
}

// Close завершает работу consumer group
func (cs *Consumer) Close() {
	cs.cancel()
	if err := cs.consumerGroup.Close(); err != nil {
		cs.log.Error("Ошибка при закрытии consumer group", "error", err)
	}
}

// Setup вызывается перед началом потребления
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup вызывается после завершения потребления
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim вызывается для каждого утверждения (claim) в группе
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var brokerMessage cmodel.BrokerMessage
		err := json.Unmarshal(message.Value, &brokerMessage)
		if err != nil {
			h.log.Error("consumer ConsumeClaim Ошибка десериализации сообщения", "error", err)
			continue
		}

		idempotencyKey := string(message.Key)
		h.log.Info("consumer ConsumeClaim Получено сообщение", "FileID", brokerMessage.FileID,
			"SrcFileURL", brokerMessage.SrcFileURL, "DestFileURL", brokerMessage.DestFileURL)

		err = h.proc.Process(context.Background(), brokerMessage, idempotencyKey)
		if err != nil {
			h.log.Error("consumer ConsumeClaim proc.Process Ошибка при обработке сообщения", "error", err)
			continue
		}

		session.MarkMessage(message, "")
	}

	return nil
}
