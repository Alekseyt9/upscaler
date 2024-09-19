// Package model contains data structures used for messaging and configuration
// in the message broker system
package model

// BrokerMessage represents a message to be sent to a broker
type BrokerMessage struct {
	SrcFileURL    string
	DestFileURL   string
	DestFileKey   string
	FileID        int64
	FileExtension string
	UserID        int64
	QueueID       int64
}

// BrokerMessageResult represents the result of processing a broker message
type BrokerMessageResult struct {
	Result      string
	Error       string
	DestFileKey string
	UserID      int64
	FileID      int64
	QueueID     int64
}

// BrokerOptions defines the configuration options for a broker
type BrokerOptions struct {
	Topic         string
	KafkaBrokers  []string
	ConsumerGroup string
}
