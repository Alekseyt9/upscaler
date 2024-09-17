package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/Alekseyt9/upscaler/internal/proc/services/processor"
	"github.com/IBM/sarama"
)

type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	brokers       []string
	topic         string
	group         string
}

// NewConsumer создает новый экземпляр ConsumerService
func NewConsumer(brokers []string, topic, group string, proc *processor.ProcessorService) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_6_0_0

	consumerGroup, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания группы потребителей: %w", err)
	}

	handler := ConsumerGroupHandler{proc: proc}
	cs := &Consumer{
		consumerGroup: consumerGroup,
		brokers:       brokers,
		topic:         topic,
		group:         group,
	}

	go func() {
		for {
			ctx := context.Background()
			if err := cs.consumerGroup.Consume(ctx, []string{cs.topic}, &handler); err != nil {
				log.Fatalf("Ошибка при обработке сообщений: %v", err)
			}
		}
	}()

	return cs, nil
}

// Close завершает работу consumer group
func (cs *Consumer) Close() {
	if err := cs.consumerGroup.Close(); err != nil {
		log.Printf("Ошибка при закрытии consumer group: %v", err)
	}
}

// ConsumerGroupHandler реализует интерфейс sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	proc *processor.ProcessorService
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
			log.Printf("Ошибка десериализации сообщения: %v", err)
			continue
		}

		idempotencyKey := string(message.Key)
		fmt.Printf("Получено сообщение: TaskId=%d, SrcFileURL=%s, DestFileURL=%s\n", brokerMessage.TaskId, brokerMessage.SrcFileURL, brokerMessage.DestFileURL)

		err = h.proc.Process(context.Background(), brokerMessage, idempotencyKey)
		if err != nil {
			log.Printf("Ошибка при обработке сообщения: %v", err)
			continue
		}

		session.MarkMessage(message, "")
	}

	return nil
}
