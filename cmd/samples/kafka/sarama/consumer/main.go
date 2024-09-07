package main

import (
	"context"
	"log"

	"github.com/IBM/sarama"
)

func main() {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	consumerGroup, err := sarama.NewConsumerGroup([]string{"localhost:9092"}, "test_group", config)
	if err != nil {
		log.Fatalf("Failed to create consumer group: %v", err)
	}
	defer consumerGroup.Close()

	consumer := Consumer{}
	ctx := context.Background()

	for {
		err := consumerGroup.Consume(ctx, []string{"test_topic"}, &consumer)
		if err != nil {
			log.Fatalf("Error consuming messages: %v", err)
		}
	}
}

type Consumer struct{}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s",
			string(msg.Value), msg.Timestamp, msg.Topic)
		session.MarkMessage(msg, "")
	}
	return nil
}
