package model

type BrokerMessage struct {
	SrcFileURL    string
	DestFileURL   string
	DestFileKey   string
	FileID        int64
	FileExtension string
	UserID        int64
	QueueID       int64
}

type BrokerMessageResult struct {
	Result      string
	Error       string
	DestFileKey string
	UserID      int64
	FileID      int64
	QueueID     int64
}

type BrokerOptions struct {
	Topic         string
	KafkaBrokers  []string
	ConsumerGroup string
}
