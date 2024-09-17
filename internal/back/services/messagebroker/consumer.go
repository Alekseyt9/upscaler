package messagebroker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Alekseyt9/upscaler/internal/back/services/userserv"
	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/IBM/sarama"
)

type ConsumerService struct {
	consumerGroup sarama.ConsumerGroup
	brokers       []string
	topic         string
	group         string
}

// NewConsumer создает новый экземпляр ConsumerService
func NewConsumer(brokers []string, topic, group string, us *userserv.UserService) (*ConsumerService, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_6_0_0

	consumerGroup, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return nil, fmt.Errorf("oшибка создания группы потребителей: %w", err)
	}

	handler := ConsumerGroupHandler{us: us}
	cs := &ConsumerService{
		consumerGroup: consumerGroup,
		brokers:       brokers,
		topic:         topic,
		group:         group,
	}

	go func() {
		for {
			if err := cs.consumerGroup.Consume(nil, []string{cs.topic}, &handler); err != nil {
				log.Fatalf("Ошибка при обработке сообщений: %v", err)
			}
		}
	}()

	return cs, nil
}

func (cs *ConsumerService) Close() {
	cs.consumerGroup.Close()
}

// ConsumerGroupHandler реализует интерфейс sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	us *userserv.UserService
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

	var msgs []cmodel.BrokerMessageResult

	for message := range claim.Messages() {
		var result cmodel.BrokerMessageResult
		err := json.Unmarshal(message.Value, &result)
		if err != nil {
			log.Printf("Ошибка десериализации сообщения: %v", err)
			continue
		}
		fmt.Printf("Получено сообщение: TaskId=%d, Result=%s, Error=%s\n", result.TaskId, result.Result, result.Error)
		msgs = append(msgs, result)
	}

	err := h.us.FinishTasks(context.Background(), msgs)
	if err != nil {
		return fmt.Errorf("us.FinishTasks %w", err)
	}

	for message := range claim.Messages() {
		session.MarkMessage(message, "")
	}

	return nil
}
