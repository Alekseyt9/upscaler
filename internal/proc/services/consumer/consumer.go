// Package consumer provides a Kafka consumer implementation that consumes messages
// from a Kafka topic using a consumer group. It processes messages using the provided
// ProcessorService and supports graceful shutdown.
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

// Consumer is a Kafka consumer that uses a consumer group to read messages from a Kafka topic.
// It processes the messages using a ProcessorService and logs its operations.
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
// It defines how messages in the consumer group are processed.
type ConsumerGroupHandler struct {
	proc *processor.ProcessorService
	log  *slog.Logger
}

// NewConsumer creates a new instance of Consumer.
//
// Parameters:
//   - proc: The ProcessorService for processing incoming messages.
//   - log: A logger for capturing and outputting log messages.
//   - opt: BrokerOptions containing configuration such as the Kafka brokers, topic, and consumer group.
//
// Returns:
//   - A pointer to a Consumer instance.
//   - An error if the consumer group cannot be created.
func NewConsumer(proc *processor.ProcessorService, log *slog.Logger, opt cmodel.BrokerOptions) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_6_0_0

	consumerGroup, err := sarama.NewConsumerGroup(opt.KafkaBrokers, opt.ConsumerGroup, config)
	if err != nil {
		return nil, fmt.Errorf("error creating consumer group: %w", err)
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
					log.Error("consumerGroup.Consume: error processing messages", "error", err)
				}
			}
		}
	}()

	return cs, nil
}

// Close shuts down the consumer group, ensuring that it stops consuming messages gracefully.
func (cs *Consumer) Close() {
	cs.cancel()
	if err := cs.consumerGroup.Close(); err != nil {
		cs.log.Error("Error closing consumer group", "error", err)
	}
}

// Setup is called before the start of the message consumption. It's part of the sarama.ConsumerGroupHandler interface.
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is called after the consumption ends. It's part of the sarama.ConsumerGroupHandler interface.
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim is called for each claim in the consumer group and handles the processing of messages.
//
// Parameters:
//   - session: The session for the consumer group.
//   - claim: The claim containing the messages to consume.
//
// Returns:
//   - An error if there is a problem processing the messages.
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var brokerMessage cmodel.BrokerMessage
		err := json.Unmarshal(message.Value, &brokerMessage)
		if err != nil {
			h.log.Error("consumer ConsumeClaim: error deserializing message", "error", err)
			continue
		}

		idempotencyKey := string(message.Key)
		h.log.Info("consumer ConsumeClaim: message received", "FileID", brokerMessage.FileID,
			"SrcFileURL", brokerMessage.SrcFileURL, "DestFileURL", brokerMessage.DestFileURL)

		err = h.proc.Process(context.Background(), brokerMessage, idempotencyKey)
		if err != nil {
			h.log.Error("consumer ConsumeClaim proc.Process: error processing message", "error", err)
			continue
		}

		session.MarkMessage(message, "")
	}

	return nil
}
