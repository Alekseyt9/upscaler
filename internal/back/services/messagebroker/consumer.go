package messagebroker

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

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
func NewConsumer(brokers []string, topic, group string) (*ConsumerService, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_6_0_0

	consumerGroup, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return nil, fmt.Errorf("Ошибка создания группы потребителей: %v", err)
	}

	return &ConsumerService{
		consumerGroup: consumerGroup,
		brokers:       brokers,
		topic:         topic,
		group:         group,
	}, nil
}

// Start запускает консьюмера и начинает обработку сообщений
func (cs *ConsumerService) Start() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)

	handler := ConsumerGroupHandler{}

	go func() {
		for {
			if err := cs.consumerGroup.Consume(nil, []string{cs.topic}, &handler); err != nil {
				log.Fatalf("Ошибка при обработке сообщений: %v", err)
			}
		}
	}()

	log.Println("Консьюмер запущен. Чтение сообщений из Kafka...")
	<-sigchan
	log.Println("Получен сигнал остановки. Завершение работы...")
	cs.consumerGroup.Close()
}

// ConsumerGroupHandler реализует интерфейс sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct{}

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
		// Десериализация сообщения в структуру BrokerMessageResult
		var result cmodel.BrokerMessageResult
		err := json.Unmarshal(message.Value, &result)
		if err != nil {
			log.Printf("Ошибка десериализации сообщения: %v", err)
			continue
		}

		// Обработка десериализованного сообщения
		fmt.Printf("Получено сообщение: TaskId=%d, Result=%s, Error=%s\n", result.TaskId, result.Result, result.Error)

		// Маркируем сообщение как прочитанное
		session.MarkMessage(message, "")
	}
	return nil
}
