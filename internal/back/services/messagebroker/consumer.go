package messagebroker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Alekseyt9/upscaler/internal/back/services/userserv"
	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
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
	us  *userserv.UserService
	log *slog.Logger
}

// NewConsumer создает новый экземпляр ConsumerService
func NewConsumer(us *userserv.UserService, log *slog.Logger, opt cmodel.BrokerOptions) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_6_0_0

	consumerGroup, err := sarama.NewConsumerGroup(opt.KafkaBrokers, opt.ConsumerGroup, config)
	if err != nil {
		return nil, fmt.Errorf("oшибка создания группы потребителей: %w", err)
	}

	handler := ConsumerGroupHandler{
		us:  us,
		log: log,
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

	var msgs []cmodel.BrokerMessageResult

	for message := range claim.Messages() {
		var result cmodel.BrokerMessageResult
		err := json.Unmarshal(message.Value, &result)
		if err != nil {
			h.log.Info("consumer ConsumeClaim", "Ошибка десериализации сообщения", err)
			continue
		}
		h.log.Info("consumer ConsumeClaim Получено сообщение",
			"TaskId", result.TaskId, "Result", result.Result, "Error", result.Error)
		msgs = append(msgs, result)
	}

	h.log.Info("consumer ConsumeClaim", "messages", msgs)

	err := h.us.FinishTasks(context.Background(), msgs)
	if err != nil {
		return fmt.Errorf("consumer ConsumeClaim us.FinishTasks %w", err)
	}

	for message := range claim.Messages() {
		session.MarkMessage(message, "")
	}

	return nil
}
