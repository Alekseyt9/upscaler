// Package messagebroker provides functionality for consuming messages from Kafka
// and processing them using the UserService. It handles message consumption in
// a consumer group and processes the results of file tasks.
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

// Consumer is a struct that represents a Kafka consumer that operates within a consumer group.
type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	brokers       []string
	topic         string
	group         string
	log           *slog.Logger
	ctx           context.Context
	cancel        context.CancelFunc
}

// ConsumerGroupHandler implements the sarama.ConsumerGroupHandler interface.
// It processes messages consumed from Kafka and interacts with the UserService to finalize tasks.
type ConsumerGroupHandler struct {
	us  userserv.UserService
	log *slog.Logger
}

// NewConsumer creates and returns a new instance of ConsumerService for consuming messages from Kafka.
//
// Parameters:
//   - us: The UserService used to process tasks based on consumed messages.
//   - log: A logger for capturing logs and errors.
//   - opt: BrokerOptions containing Kafka broker addresses, topic, and consumer group.
//
// Returns:
//   - A pointer to a Consumer instance.
//   - An error if there is an issue creating the Kafka consumer group.
func NewConsumer(us userserv.UserService, log *slog.Logger, opt cmodel.BrokerOptions) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_6_0_0

	consumerGroup, err := sarama.NewConsumerGroup(opt.KafkaBrokers, opt.ConsumerGroup, config)
	if err != nil {
		return nil, fmt.Errorf("error creating consumer group: %w", err)
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
					log.Error("consumerGroup.Consume: Error processing messages", "error", err)
				}
			}
		}
	}()

	return cs, nil
}

// Close shuts down the consumer group and cancels the context.
func (cs *Consumer) Close() {
	cs.cancel()

	if err := cs.consumerGroup.Close(); err != nil {
		cs.log.Error("Error closing consumer group", "error", err)
	}
}

// Setup is called before the consumer starts consuming messages.
// It is part of the sarama.ConsumerGroupHandler interface.
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is called after the consumer has finished consuming messages.
// It is part of the sarama.ConsumerGroupHandler interface.
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim is called for each claim in the consumer group.
// It processes each message, deserializes it, and calls the UserService to finalize the tasks.
//
// Parameters:
//   - session: The Kafka consumer group session.
//   - claim: The claim containing the Kafka messages to be consumed.
//
// Returns:
//   - An error if there is a failure during message processing.
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var result cmodel.BrokerMessageResult
		err := json.Unmarshal(message.Value, &result)
		if err != nil {
			h.log.Info("consumer ConsumeClaim", "Error deserializing message", err)
			continue
		}
		h.log.Info("consumer ConsumeClaim: Message received",
			"FileID", result.FileID, "Result", result.Result, "Error", result.Error)

		h.log.Info("consumer ConsumeClaim", "message", result)

		err = h.us.FinishTasks(context.Background(), []cmodel.BrokerMessageResult{result})
		if err != nil {
			h.log.Error("consumer ConsumeClaim us.FinishTasks", "error", err)
		}

		session.MarkMessage(message, "")
	}

	return nil
}
